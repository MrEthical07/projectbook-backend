package integration

import (
	"bufio"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

type routeMatrixEntry struct {
	Method string
	Path   string
	Pack   string
}

func TestIntegrationRouteMatrixProtectedAnonymous401(t *testing.T) {
	h := requireIntegration(t)
	entries := h.loadRouteMatrix(t)

	for _, entry := range entries {
		if !strings.HasPrefix(entry.Pack, "PROTECTED-") {
			continue
		}

		path := materializeRoutePath(entry.Path, "project-anon", map[string]string{})
		t.Run(entry.Method+" "+entry.Path, func(t *testing.T) {
			payload := writePayload(entry.Method)
			resp := h.requestJSON(t, entry.Method, path, "", payload)
			if resp.Status != http.StatusUnauthorized || resp.Envelope.Success {
				t.Fatalf("expected 401 for anonymous protected route status=%d body=%s", resp.Status, resp.Body)
			}
		})
	}
}

func TestIntegrationRouteMatrixProjectScopedNonMember403(t *testing.T) {
	h := requireIntegration(t)
	entries := h.loadRouteMatrix(t)

	owner := h.createVerifiedSession(t, "matrix_owner")
	outsider := h.createVerifiedSession(t, "matrix_outsider")
	project := h.createProject(t, owner.AccessToken, "Matrix Protected Routes")

	for _, entry := range entries {
		if !strings.HasPrefix(entry.Pack, "PROTECTED-") {
			continue
		}
		if !strings.Contains(entry.Path, "/projects/{projectId}") {
			continue
		}

		path := materializeRoutePath(entry.Path, project.Slug, map[string]string{})
		t.Run(entry.Method+" "+entry.Path, func(t *testing.T) {
			payload := writePayload(entry.Method)
			resp := h.requestJSON(t, entry.Method, path, outsider.AccessToken, payload)
			if resp.Status != http.StatusForbidden || resp.Envelope.Success {
				t.Fatalf("expected 403 for non-member on project-scoped route status=%d body=%s", resp.Status, resp.Body)
			}
		})
	}
}

func TestIntegrationRouteMatrixProjectScopedNoPermission403(t *testing.T) {
	h := requireIntegration(t)
	entries := h.loadRouteMatrix(t)

	owner := h.createVerifiedSession(t, "matrix_owner_write")
	limited := h.createVerifiedSession(t, "matrix_limited_write")
	project := h.createProject(t, owner.AccessToken, "Matrix No Permission Routes")
	h.upsertCustomMember(t, project.UUID, limited.UserID, 0)

	for _, entry := range entries {
		if entry.Pack != "PROTECTED-WRITE" {
			continue
		}
		if !strings.Contains(entry.Path, "/projects/{projectId}") {
			continue
		}

		path := materializeRoutePath(entry.Path, project.Slug, map[string]string{})
		t.Run(entry.Method+" "+entry.Path, func(t *testing.T) {
			payload := writePayload(entry.Method)
			resp := h.requestJSON(t, entry.Method, path, limited.AccessToken, payload)
			if resp.Status != http.StatusForbidden || resp.Envelope.Success {
				t.Fatalf("expected 403 for member without permissions on project write route status=%d body=%s", resp.Status, resp.Body)
			}
		})
	}
}

func TestIntegrationRouteMatrixProtectedReadAuthorizedNotAuthDenied(t *testing.T) {
	h := requireIntegration(t)
	entries := h.loadRouteMatrix(t)

	owner := h.createVerifiedSession(t, "matrix_read_owner")
	project := h.createProject(t, owner.AccessToken, "Matrix Read Authorized")

	for _, entry := range entries {
		if entry.Pack != "PROTECTED-READ" {
			continue
		}

		path := materializeRoutePath(entry.Path, project.Slug, map[string]string{})
		t.Run(entry.Method+" "+entry.Path, func(t *testing.T) {
			resp := h.requestJSON(t, entry.Method, path, owner.AccessToken, nil)
			if resp.Status == http.StatusUnauthorized || resp.Status == http.StatusForbidden {
				t.Fatalf("expected non-auth-denied status for authorized protected read route status=%d body=%s", resp.Status, resp.Body)
			}
		})
	}
}

func TestIntegrationRouteMatrixProtectedWriteAuthorizedNotAuthDenied(t *testing.T) {
	h := requireIntegration(t)
	entries := h.loadRouteMatrix(t)

	owner := h.createVerifiedSession(t, "matrix_write_owner")
	project := h.createProject(t, owner.AccessToken, "Matrix Write Authorized")

	for _, entry := range entries {
		if entry.Pack != "PROTECTED-WRITE" {
			continue
		}
		if entry.Method == http.MethodDelete && entry.Path == "/api/v1/projects/{projectId}" {
			continue
		}
		if entry.Method == http.MethodPost && entry.Path == "/api/v1/projects/{projectId}/archive" {
			continue
		}

		path := materializeRoutePath(entry.Path, project.Slug, map[string]string{})
		t.Run(entry.Method+" "+entry.Path, func(t *testing.T) {
			payload := writePayload(entry.Method)
			resp := h.requestJSON(t, entry.Method, path, owner.AccessToken, payload)
			if resp.Status == http.StatusUnauthorized || resp.Status == http.StatusForbidden {
				t.Fatalf("expected non-auth-denied status for authorized protected write route status=%d body=%s", resp.Status, resp.Body)
			}
		})
	}
}

func TestIntegrationAuthPublicInvalidPayloadMatrix(t *testing.T) {
	h := requireIntegration(t)

	cases := []struct {
		method string
		path   string
		body   map[string]any
	}{
		{method: http.MethodPost, path: "/api/v1/auth/signup", body: map[string]any{}},
		{method: http.MethodPost, path: "/api/v1/auth/login", body: map[string]any{}},
		{method: http.MethodPost, path: "/api/v1/auth/refresh", body: map[string]any{}},
		{method: http.MethodPost, path: "/api/v1/auth/verify-email", body: map[string]any{}},
		{method: http.MethodPost, path: "/api/v1/auth/resend-verification", body: map[string]any{}},
		{method: http.MethodPost, path: "/api/v1/auth/forgot-password", body: map[string]any{}},
		{method: http.MethodPost, path: "/api/v1/auth/reset-password", body: map[string]any{}},
	}

	for _, tc := range cases {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			resp := h.requestJSON(t, tc.method, tc.path, "", tc.body)
			if resp.Status != http.StatusBadRequest || resp.Envelope.Success {
				t.Fatalf("expected 400 for invalid auth-public payload status=%d body=%s", resp.Status, resp.Body)
			}
		})
	}
}

func (h *integrationHarness) loadRouteMatrix(t *testing.T) []routeMatrixEntry {
	t.Helper()

	matrixPath := filepath.Join(h.repoRoot, "docs", "test", "route-coverage-matrix.md")
	file, err := os.Open(matrixPath)
	if err != nil {
		t.Fatalf("open route matrix file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	entries := make([]routeMatrixEntry, 0, 128)
	seen := map[string]struct{}{}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "|") {
			continue
		}
		if strings.Contains(line, "| Module | Method | Path | Scenario Pack |") {
			continue
		}
		if strings.HasPrefix(line, "|---") {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) < 6 {
			continue
		}

		method := strings.TrimSpace(parts[2])
		path := strings.TrimSpace(parts[3])
		pack := strings.TrimSpace(parts[4])
		if method == "" || path == "" || pack == "" {
			continue
		}

		key := method + " " + path
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}

		entries = append(entries, routeMatrixEntry{Method: method, Path: path, Pack: pack})
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("scan route matrix file: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("route matrix parser produced zero entries")
	}

	return entries
}

var placeholderRe = regexp.MustCompile(`\{([^}]+)\}`)

func materializeRoutePath(templatePath, projectID string, overrides map[string]string) string {
	defaults := map[string]string{
		"projectId":  projectID,
		"slug":       "sample-slug",
		"storyId":    "00000000-0000-0000-0000-000000000011",
		"journeyId":  "00000000-0000-0000-0000-000000000012",
		"problemId":  "00000000-0000-0000-0000-000000000013",
		"ideaId":     "00000000-0000-0000-0000-000000000014",
		"taskId":     "00000000-0000-0000-0000-000000000015",
		"feedbackId": "00000000-0000-0000-0000-000000000016",
		"resourceId": "00000000-0000-0000-0000-000000000017",
		"pageId":     "00000000-0000-0000-0000-000000000018",
		"eventId":    "00000000-0000-0000-0000-000000000019",
		"artifactId": "00000000-0000-0000-0000-000000000020",
		"inviteId":   "00000000-0000-0000-0000-000000000021",
		"memberId":   "00000000-0000-0000-0000-000000000022",
		"role":       "Member",
		"email":      "nobody@example.com",
	}

	for key, value := range overrides {
		defaults[key] = value
	}

	resolved := placeholderRe.ReplaceAllStringFunc(templatePath, func(token string) string {
		key := strings.TrimSuffix(strings.TrimPrefix(token, "{"), "}")
		if value, ok := defaults[key]; ok {
			return value
		}
		return "placeholder"
	})

	return resolved
}

func writePayload(method string) any {
	switch strings.ToUpper(strings.TrimSpace(method)) {
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		return map[string]any{}
	default:
		return nil
	}
}
