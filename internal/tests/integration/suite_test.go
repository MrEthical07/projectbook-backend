package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	coredb "github.com/MrEthical07/superapi/internal/core/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	defaultPostgresAdminURL = "postgres://superapi:superapi@127.0.0.1:5432/postgres?sslmode=disable"
	defaultRedisAddr        = "127.0.0.1:6379"
	defaultMongoURL         = "mongodb://127.0.0.1:27017"
	defaultSignupPassword   = "Passw0rd!2026"
)

var sharedHarness *integrationHarness

type integrationHarness struct {
	repoRoot string
	baseURL  string

	metricsAuthToken string

	httpClient *http.Client

	apiCancel context.CancelFunc
	apiCmd    *exec.Cmd
	apiStdout bytes.Buffer
	apiStderr bytes.Buffer

	pgAdminURL string
	pgURL      string
	pgDBName   string
	pgPool     *pgxpool.Pool

	redisAddr     string
	redisPassword string
	redisDB       int
	redisClient   *redis.Client

	mongoURL    string
	mongoDBName string
	mongoClient *mongo.Client
	mongoDB     *mongo.Database
}

type apiEnvelope struct {
	Success   bool            `json:"success"`
	Data      json.RawMessage `json:"data"`
	Error     *apiErrorBody   `json:"error,omitempty"`
	RequestID string          `json:"request_id,omitempty"`
}

type apiErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type apiResponse struct {
	Status   int
	Header   http.Header
	Envelope apiEnvelope
	Body     string
}

type authSession struct {
	UserID       string
	Name         string
	Email        string
	Password     string
	AccessToken  string
	RefreshToken string
}

type projectFixture struct {
	Slug string
	UUID string
}

func TestMain(m *testing.M) {
	if !integrationEnabled() {
		os.Exit(m.Run())
	}

	harness, err := setupIntegrationHarness(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "integration setup failed: %v\n", err)
		os.Exit(1)
	}
	sharedHarness = harness

	code := m.Run()
	if closeErr := harness.Close(context.Background()); closeErr != nil {
		fmt.Fprintf(os.Stderr, "integration cleanup failed: %v\n", closeErr)
	}

	os.Exit(code)
}

func integrationEnabled() bool {
	raw := strings.ToLower(strings.TrimSpace(os.Getenv("INTEGRATION_TESTS")))
	return raw == "1" || raw == "true" || raw == "yes"
}

func requireIntegration(t *testing.T) *integrationHarness {
	t.Helper()
	if !integrationEnabled() {
		t.Skip("integration tests disabled; set INTEGRATION_TESTS=1")
	}
	if sharedHarness == nil {
		t.Fatal("integration harness is not initialized")
	}
	return sharedHarness
}

func setupIntegrationHarness(ctx context.Context) (*integrationHarness, error) {
	repoRoot, err := findRepoRoot()
	if err != nil {
		return nil, err
	}

	h := &integrationHarness{
		repoRoot: repoRoot,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	if err := h.setupPostgres(ctx); err != nil {
		_ = h.Close(context.Background())
		return nil, err
	}

	if err := h.setupRedis(ctx); err != nil {
		_ = h.Close(context.Background())
		return nil, err
	}

	if err := h.setupMongo(ctx); err != nil {
		_ = h.Close(context.Background())
		return nil, err
	}

	if err := h.startAPI(ctx); err != nil {
		_ = h.Close(context.Background())
		return nil, err
	}

	return h, nil
}

func (h *integrationHarness) setupPostgres(ctx context.Context) error {
	adminURL := strings.TrimSpace(os.Getenv("IT_PG_ADMIN_URL"))
	if adminURL == "" {
		base := strings.TrimSpace(os.Getenv("POSTGRES_URL"))
		if base == "" {
			base = defaultPostgresAdminURL
		}
		resolved, err := withPostgresDatabase(base, "postgres")
		if err != nil {
			return fmt.Errorf("resolve postgres admin url: %w", err)
		}
		adminURL = resolved
	}

	dbName := makeDatabaseName("it_projectbook")
	testURL, err := withPostgresDatabase(adminURL, dbName)
	if err != nil {
		return fmt.Errorf("build postgres test url: %w", err)
	}

	adminPool, err := pgxpool.New(ctx, adminURL)
	if err != nil {
		return fmt.Errorf("connect postgres admin db: %w", err)
	}
	defer adminPool.Close()

	if _, err := adminPool.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s", quoteIdentifier(dbName))); err != nil {
		return fmt.Errorf("create postgres test db %q: %w", dbName, err)
	}

	sourceURL, err := coredb.MigrationSourceURL(filepath.Join(h.repoRoot, "db", "migrations"))
	if err != nil {
		return fmt.Errorf("resolve migration source: %w", err)
	}

	runner, err := coredb.NewMigrationRunner(testURL, sourceURL)
	if err != nil {
		return fmt.Errorf("create migration runner: %w", err)
	}
	defer func() {
		_ = runner.Close()
	}()

	if _, err := runner.Up(); err != nil {
		return fmt.Errorf("run postgres migrations: %w", err)
	}

	pool, err := pgxpool.New(ctx, testURL)
	if err != nil {
		return fmt.Errorf("connect postgres test db: %w", err)
	}

	h.pgAdminURL = adminURL
	h.pgURL = testURL
	h.pgDBName = dbName
	h.pgPool = pool

	return nil
}

func (h *integrationHarness) setupRedis(ctx context.Context) error {
	addr := strings.TrimSpace(os.Getenv("IT_REDIS_ADDR"))
	if addr == "" {
		addr = strings.TrimSpace(os.Getenv("REDIS_ADDR"))
	}
	if addr == "" {
		addr = defaultRedisAddr
	}

	password := strings.TrimSpace(os.Getenv("IT_REDIS_PASSWORD"))
	if password == "" {
		password = strings.TrimSpace(os.Getenv("REDIS_PASSWORD"))
	}

	redisDB := envInt("IT_REDIS_DB", int(time.Now().UnixNano()%8)+8)
	if redisDB < 0 {
		redisDB = 0
	}

	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           redisDB,
		DialTimeout:  2 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return fmt.Errorf("ping redis: %w", err)
	}

	if err := client.FlushDB(ctx).Err(); err != nil {
		_ = client.Close()
		return fmt.Errorf("flush redis test db: %w", err)
	}

	h.redisAddr = addr
	h.redisPassword = password
	h.redisDB = redisDB
	h.redisClient = client

	return nil
}

func (h *integrationHarness) setupMongo(ctx context.Context) error {
	mongoURL := strings.TrimSpace(os.Getenv("IT_MONGO_URL"))
	if mongoURL == "" {
		mongoURL = strings.TrimSpace(os.Getenv("MONGO_URL"))
	}
	if mongoURL == "" {
		mongoURL = defaultMongoURL
	}

	dbName := makeDatabaseName("projectbook_it")

	connectCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(connectCtx, options.Client().ApplyURI(mongoURL))
	if err != nil {
		return fmt.Errorf("connect mongo: %w", err)
	}

	if err := client.Ping(connectCtx, nil); err != nil {
		_ = client.Disconnect(context.Background())
		return fmt.Errorf("ping mongo: %w", err)
	}

	db := client.Database(dbName)
	if err := db.Drop(connectCtx); err != nil {
		_ = client.Disconnect(context.Background())
		return fmt.Errorf("drop stale mongo test db: %w", err)
	}

	h.mongoURL = mongoURL
	h.mongoDBName = dbName
	h.mongoClient = client
	h.mongoDB = db

	return nil
}

func (h *integrationHarness) startAPI(ctx context.Context) error {
	port, err := freeTCPPort()
	if err != nil {
		return err
	}

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	h.baseURL = "http://" + addr
	h.metricsAuthToken = "it-metrics-token"

	cmdCtx, cancel := context.WithCancel(ctx)
	cmd := exec.CommandContext(cmdCtx, "go", "run", "./cmd/api")
	cmd.Dir = h.repoRoot
	cmd.Stdout = &h.apiStdout
	cmd.Stderr = &h.apiStderr
	cmd.Env = mergeEnv(os.Environ(), map[string]string{
		"APP_ENV":                 "dev",
		"HTTP_ADDR":               addr,
		"LOG_LEVEL":               "warn",
		"LOG_FORMAT":              "json",
		"POSTGRES_ENABLED":        "true",
		"POSTGRES_URL":            h.pgURL,
		"REDIS_ENABLED":           "true",
		"REDIS_ADDR":              h.redisAddr,
		"REDIS_PASSWORD":          h.redisPassword,
		"REDIS_DB":                strconv.Itoa(h.redisDB),
		"MONGO_ENABLED":           "true",
		"MONGO_URL":               h.mongoURL,
		"MONGO_DB":                h.mongoDBName,
		"MONGO_BOOTSTRAP_ENABLED": "true",
		"AUTH_ENABLED":            "true",
		"AUTH_MODE":               "hybrid",
		"RATELIMIT_ENABLED":       "true",
		"CACHE_ENABLED":           "true",
		"PERMISSIONS_ENABLED":     "true",
		"METRICS_ENABLED":         "true",
		"METRICS_AUTH_TOKEN":      h.metricsAuthToken,
	})

	if err := cmd.Start(); err != nil {
		cancel()
		return fmt.Errorf("start api process: %w", err)
	}

	h.apiCancel = cancel
	h.apiCmd = cmd

	if err := h.waitForReady(90 * time.Second); err != nil {
		return err
	}

	return nil
}

func (h *integrationHarness) waitForReady(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	var lastErr error

	for time.Now().Before(deadline) {
		resp, err := h.httpClient.Get(h.baseURL + "/readyz")
		if err != nil {
			lastErr = err
			time.Sleep(500 * time.Millisecond)
			continue
		}

		bodyBytes, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			return nil
		}

		lastErr = fmt.Errorf("readyz status=%d body=%s", resp.StatusCode, string(bodyBytes))
		time.Sleep(500 * time.Millisecond)
	}

	if h.apiCmd != nil {
		_ = h.apiCmd.Process.Kill()
	}

	return fmt.Errorf(
		"api did not become ready: %w\nstdout:\n%s\nstderr:\n%s",
		lastErr,
		h.apiStdout.String(),
		h.apiStderr.String(),
	)
}

func (h *integrationHarness) Close(ctx context.Context) error {
	var closeErr error

	if h.apiCancel != nil {
		h.apiCancel()
	}
	if h.apiCmd != nil && h.apiCmd.Process != nil {
		done := make(chan error, 1)
		go func() {
			done <- h.apiCmd.Wait()
		}()

		select {
		case err := <-done:
			if err != nil && !errors.Is(err, context.Canceled) {
				closeErr = errors.Join(closeErr, fmt.Errorf("wait api process: %w", err))
			}
		case <-time.After(10 * time.Second):
			_ = h.apiCmd.Process.Kill()
			select {
			case <-done:
			case <-time.After(5 * time.Second):
				// best-effort shutdown only
			}
		}
	}

	if h.mongoDB != nil {
		dropCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		if err := h.mongoDB.Drop(dropCtx); err != nil {
			closeErr = errors.Join(closeErr, fmt.Errorf("drop mongo db: %w", err))
		}
		cancel()
	}
	if h.mongoClient != nil {
		disconnectCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		if err := h.mongoClient.Disconnect(disconnectCtx); err != nil {
			closeErr = errors.Join(closeErr, fmt.Errorf("disconnect mongo: %w", err))
		}
		cancel()
	}

	if h.redisClient != nil {
		flushCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		if err := h.redisClient.FlushDB(flushCtx).Err(); err != nil {
			closeErr = errors.Join(closeErr, fmt.Errorf("flush redis db: %w", err))
		}
		cancel()
		if err := h.redisClient.Close(); err != nil {
			closeErr = errors.Join(closeErr, fmt.Errorf("close redis: %w", err))
		}
	}

	if h.pgPool != nil {
		h.pgPool.Close()
	}

	if strings.TrimSpace(h.pgAdminURL) != "" && strings.TrimSpace(h.pgDBName) != "" {
		adminPool, err := pgxpool.New(ctx, h.pgAdminURL)
		if err != nil {
			closeErr = errors.Join(closeErr, fmt.Errorf("connect postgres admin for cleanup: %w", err))
		} else {
			_, _ = adminPool.Exec(ctx, "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = $1 AND pid <> pg_backend_pid()", h.pgDBName)
			if _, err := adminPool.Exec(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s", quoteIdentifier(h.pgDBName))); err != nil {
				closeErr = errors.Join(closeErr, fmt.Errorf("drop postgres db %q: %w", h.pgDBName, err))
			}
			adminPool.Close()
		}
	}

	return closeErr
}

func (h *integrationHarness) requestJSON(t *testing.T, method, path, bearerToken string, payload any) apiResponse {
	t.Helper()

	var bodyReader io.Reader
	if payload != nil {
		bytesPayload, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("marshal request payload: %v", err)
		}
		bodyReader = bytes.NewReader(bytesPayload)
	}

	req, err := http.NewRequest(method, h.baseURL+path, bodyReader)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if strings.TrimSpace(bearerToken) != "" {
		req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(bearerToken))
	}

	resp, err := h.httpClient.Do(req)
	if err != nil {
		t.Fatalf("http request failed (%s %s): %v", method, path, err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}

	result := apiResponse{Status: resp.StatusCode, Header: resp.Header.Clone(), Body: string(bodyBytes)}
	if len(bodyBytes) == 0 {
		return result
	}

	if err := json.Unmarshal(bodyBytes, &result.Envelope); err != nil {
		t.Fatalf("decode api envelope for %s %s: %v body=%s", method, path, err, string(bodyBytes))
	}

	return result
}

func (h *integrationHarness) requestText(t *testing.T, method, path, bearerToken string) (int, string, http.Header) {
	t.Helper()

	req, err := http.NewRequest(method, h.baseURL+path, nil)
	if err != nil {
		t.Fatalf("create text request: %v", err)
	}
	if strings.TrimSpace(bearerToken) != "" {
		req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(bearerToken))
	}

	resp, err := h.httpClient.Do(req)
	if err != nil {
		t.Fatalf("text request failed (%s %s): %v", method, path, err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read text response body: %v", err)
	}

	return resp.StatusCode, string(bodyBytes), resp.Header.Clone()
}

func (h *integrationHarness) createVerifiedSession(t *testing.T, prefix string) authSession {
	t.Helper()

	email := fmt.Sprintf("%s_%d@example.com", sanitizeName(prefix), time.Now().UnixNano())
	name := fmt.Sprintf("%s User", strings.TrimSpace(prefix))

	signupResp := h.requestJSON(t, http.MethodPost, "/api/v1/auth/signup", "", map[string]any{
		"name":            name,
		"email":           email,
		"password":        defaultSignupPassword,
		"confirmPassword": defaultSignupPassword,
	})
	if signupResp.Status != http.StatusCreated || !signupResp.Envelope.Success {
		t.Fatalf("signup failed status=%d body=%s", signupResp.Status, signupResp.Body)
	}

	signupData := mustDataMap(t, signupResp)
	userData := mustMap(t, signupData["user"], "signup.user")
	userID := mustString(t, userData["id"], "signup.user.id")
	h.markUserEmailVerified(t, userID)

	loginResp := h.requestJSON(t, http.MethodPost, "/api/v1/auth/login", "", map[string]any{
		"email":    email,
		"password": defaultSignupPassword,
	})
	if loginResp.Status != http.StatusOK || !loginResp.Envelope.Success {
		t.Fatalf("login failed status=%d body=%s", loginResp.Status, loginResp.Body)
	}

	loginData := mustDataMap(t, loginResp)

	return authSession{
		UserID:       userID,
		Name:         name,
		Email:        email,
		Password:     defaultSignupPassword,
		AccessToken:  mustString(t, loginData["access_token"], "login.access_token"),
		RefreshToken: mustString(t, loginData["refresh_token"], "login.refresh_token"),
	}
}

func (h *integrationHarness) markUserEmailVerified(t *testing.T, userID string) {
	t.Helper()

	result, err := h.pgPool.Exec(
		context.Background(),
		`UPDATE users SET is_email_verified = TRUE, updated_at = NOW() WHERE id = $1::uuid`,
		strings.TrimSpace(userID),
	)
	if err != nil {
		t.Fatalf("mark user email verified: %v", err)
	}
	if result.RowsAffected() != 1 {
		t.Fatalf("mark user email verified updated %d rows, want 1", result.RowsAffected())
	}
}

func (h *integrationHarness) mustLatestVerificationToken(t *testing.T, email string) string {
	t.Helper()

	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		var link string
		err := h.pgPool.QueryRow(context.Background(),
			`SELECT link FROM auth_email_log WHERE recipient_email = $1 AND kind = 'verify' ORDER BY sent_at DESC LIMIT 1`,
			email,
		).Scan(&link)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				time.Sleep(250 * time.Millisecond)
				continue
			}
			t.Fatalf("query verification link: %v", err)
		}

		token := extractTokenFromLink(link)
		if strings.TrimSpace(token) != "" {
			return token
		}

		time.Sleep(200 * time.Millisecond)
	}

	t.Fatalf("verification token not found for %s", email)
	return ""
}

func (h *integrationHarness) createProject(t *testing.T, ownerToken, name string) projectFixture {
	t.Helper()

	resp := h.requestJSON(t, http.MethodPost, "/api/v1/home/projects", ownerToken, map[string]any{
		"name":        name,
		"icon":        "rocket",
		"description": "integration project",
	})
	if resp.Status != http.StatusCreated || !resp.Envelope.Success {
		t.Fatalf("create project failed status=%d body=%s", resp.Status, resp.Body)
	}

	data := mustDataMap(t, resp)
	slug := mustString(t, data["projectId"], "projectId")

	var projectUUID string
	err := h.pgPool.QueryRow(context.Background(), `SELECT id::text FROM projects WHERE slug = $1`, slug).Scan(&projectUUID)
	if err != nil {
		t.Fatalf("lookup project uuid for slug %s: %v", slug, err)
	}

	return projectFixture{Slug: slug, UUID: projectUUID}
}

func (h *integrationHarness) upsertCustomMember(t *testing.T, projectUUID, userID string, permissionMask uint64) {
	t.Helper()

	_, err := h.pgPool.Exec(context.Background(),
		`INSERT INTO project_members (project_id, user_id, role, permission_mask, is_custom, status, joined_at)
		 VALUES ($1::uuid, $2::uuid, 'Member'::project_role, $3, TRUE, 'Active', CURRENT_DATE)
		 ON CONFLICT (project_id, user_id)
		 DO UPDATE SET permission_mask = EXCLUDED.permission_mask, is_custom = TRUE, status = 'Active', updated_at = NOW()`,
		projectUUID,
		userID,
		int64(permissionMask),
	)
	if err != nil {
		t.Fatalf("upsert custom member failed: %v", err)
	}
}

func (h *integrationHarness) cacheTagVersion(t *testing.T, tag string) int64 {
	t.Helper()
	key := fmt.Sprintf("cver:dev:%s", strings.TrimSpace(tag))
	value, err := h.redisClient.Get(context.Background(), key).Int64()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0
		}
		t.Fatalf("read cache tag version %s: %v", key, err)
	}
	return value
}

func (h *integrationHarness) findResourceUUID(t *testing.T, projectUUID, slug string) string {
	t.Helper()

	var resourceID string
	err := h.pgPool.QueryRow(context.Background(),
		`SELECT id::text FROM resources WHERE project_id = $1::uuid AND slug = $2`,
		projectUUID,
		slug,
	).Scan(&resourceID)
	if err != nil {
		t.Fatalf("find resource id by slug=%s: %v", slug, err)
	}

	return resourceID
}

func (h *integrationHarness) metricCounterValue(t *testing.T, metricName, route, outcome string) float64 {
	t.Helper()
	status, body, _ := h.requestText(t, http.MethodGet, "/metrics", h.metricsAuthToken)
	if status != http.StatusOK {
		t.Fatalf("metrics endpoint status=%d body=%s", status, body)
	}

	lines := strings.Split(body, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if !strings.HasPrefix(line, metricName+"{") {
			continue
		}

		labelStart := strings.Index(line, "{")
		labelEnd := strings.LastIndex(line, "}")
		if labelStart < 0 || labelEnd <= labelStart {
			continue
		}

		labels := parsePromLabels(line[labelStart+1 : labelEnd])
		if labels["route"] != route || labels["outcome"] != outcome {
			continue
		}

		parts := strings.Fields(line[labelEnd+1:])
		if len(parts) == 0 {
			continue
		}

		value, err := strconv.ParseFloat(parts[0], 64)
		if err != nil {
			t.Fatalf("parse metric value from line %q: %v", line, err)
		}

		return value
	}

	return 0
}

func parsePromLabels(raw string) map[string]string {
	out := map[string]string{}
	for _, part := range strings.Split(raw, ",") {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.TrimSpace(kv[0])
		value := strings.Trim(strings.TrimSpace(kv[1]), `"`)
		out[key] = value
	}
	return out
}

func mustDataMap(t *testing.T, resp apiResponse) map[string]any {
	t.Helper()
	if len(resp.Envelope.Data) == 0 {
		t.Fatalf("response data is empty body=%s", resp.Body)
	}
	var out map[string]any
	if err := json.Unmarshal(resp.Envelope.Data, &out); err != nil {
		t.Fatalf("decode response data map: %v body=%s", err, resp.Body)
	}
	return out
}

func mustMap(t *testing.T, value any, field string) map[string]any {
	t.Helper()
	out, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("%s is not an object", field)
	}
	return out
}

func mustString(t *testing.T, value any, field string) string {
	t.Helper()
	text, ok := value.(string)
	if !ok || strings.TrimSpace(text) == "" {
		t.Fatalf("%s is not a non-empty string", field)
	}
	return strings.TrimSpace(text)
}

func findRepoRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}

	current := wd
	for i := 0; i < 10; i++ {
		if _, err := os.Stat(filepath.Join(current, "go.mod")); err == nil {
			return current, nil
		}
		next := filepath.Dir(current)
		if next == current {
			break
		}
		current = next
	}

	return "", fmt.Errorf("failed to locate repository root from %s", wd)
}

func mergeEnv(base []string, overrides map[string]string) []string {
	out := make([]string, 0, len(base)+len(overrides))
	seen := make(map[string]struct{}, len(overrides))

	for _, entry := range base {
		parts := strings.SplitN(entry, "=", 2)
		key := parts[0]
		if _, ok := overrides[key]; ok {
			continue
		}
		out = append(out, entry)
		seen[key] = struct{}{}
	}

	for key, value := range overrides {
		if _, ok := seen[key]; ok {
			continue
		}
		out = append(out, key+"="+value)
	}

	return out
}

func freeTCPPort() (int, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, fmt.Errorf("allocate tcp port: %w", err)
	}
	defer ln.Close()

	addr, ok := ln.Addr().(*net.TCPAddr)
	if !ok {
		return 0, fmt.Errorf("unexpected listener address type %T", ln.Addr())
	}
	return addr.Port, nil
}

func makeDatabaseName(prefix string) string {
	suffix := strings.ToLower(strconv.FormatInt(time.Now().UnixNano(), 36))
	name := fmt.Sprintf("%s_%s", sanitizeName(prefix), suffix)
	if len(name) > 60 {
		name = name[:60]
	}
	if len(name) == 0 {
		return "itdb"
	}
	if name[0] >= '0' && name[0] <= '9' {
		name = "it_" + name
	}
	return name
}

var identifierRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

func quoteIdentifier(identifier string) string {
	if !identifierRegex.MatchString(identifier) {
		panic(fmt.Sprintf("invalid SQL identifier: %q", identifier))
	}
	return `"` + identifier + `"`
}

func withPostgresDatabase(rawURL, dbName string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(parsed.Scheme) == "" {
		return "", fmt.Errorf("postgres url missing scheme")
	}
	parsed.Path = "/" + strings.TrimSpace(dbName)
	return parsed.String(), nil
}

func envInt(name string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}

func sanitizeName(raw string) string {
	trimmed := strings.TrimSpace(strings.ToLower(raw))
	if trimmed == "" {
		return "it"
	}
	b := strings.Builder{}
	for _, ch := range trimmed {
		switch {
		case ch >= 'a' && ch <= 'z':
			b.WriteRune(ch)
		case ch >= '0' && ch <= '9':
			b.WriteRune(ch)
		default:
			b.WriteRune('_')
		}
	}
	out := strings.Trim(b.String(), "_")
	if out == "" {
		return "it"
	}
	return out
}

func extractTokenFromLink(link string) string {
	trimmed := strings.TrimSpace(link)
	if trimmed == "" {
		return ""
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return ""
	}

	priorityKeys := []string{"token", "code", "verification_token", "verificationToken"}
	query := parsed.Query()
	for _, key := range priorityKeys {
		if value := strings.TrimSpace(query.Get(key)); value != "" {
			return value
		}
	}

	for key, values := range query {
		_ = key
		for _, value := range values {
			if strings.TrimSpace(value) != "" {
				return strings.TrimSpace(value)
			}
		}
	}

	pathParts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(pathParts) > 0 {
		candidate := strings.TrimSpace(pathParts[len(pathParts)-1])
		if candidate != "" {
			return candidate
		}
	}

	return ""
}

func (h *integrationHarness) mustFindResourceDocument(t *testing.T, resourceUUID string) bson.M {
	t.Helper()

	var doc bson.M
	err := h.mongoDB.Collection("resource_documents").FindOne(context.Background(), bson.M{"artifact_id": resourceUUID}).Decode(&doc)
	if err != nil {
		t.Fatalf("find resource document artifact_id=%s: %v", resourceUUID, err)
	}

	return doc
}
