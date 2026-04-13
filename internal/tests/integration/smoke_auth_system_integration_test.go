package integration

import (
	"net/http"
	"testing"
)

func TestIntegrationSmokeHealthAndReady(t *testing.T) {
	h := requireIntegration(t)

	health := h.requestJSON(t, http.MethodGet, "/healthz", "", nil)
	if health.Status != http.StatusOK || !health.Envelope.Success {
		t.Fatalf("healthz failed status=%d body=%s", health.Status, health.Body)
	}
	healthData := mustDataMap(t, health)
	if status := mustString(t, healthData["status"], "health.status"); status != "ok" {
		t.Fatalf("health status=%q want=ok", status)
	}

	ready := h.requestJSON(t, http.MethodGet, "/readyz", "", nil)
	if ready.Status != http.StatusOK || !ready.Envelope.Success {
		t.Fatalf("readyz failed status=%d body=%s", ready.Status, ready.Body)
	}
	readyData := mustDataMap(t, ready)
	if status := mustString(t, readyData["status"], "ready.status"); status != "ready" {
		t.Fatalf("ready status=%q want=ready", status)
	}

	deps := mustMap(t, readyData["dependencies"], "ready.dependencies")
	assertDependencyOK(t, deps, "postgres")
	assertDependencyOK(t, deps, "redis")
	assertDependencyOK(t, deps, "mongo")
}

func TestIntegrationSystemParseDuration(t *testing.T) {
	h := requireIntegration(t)

	okResp := h.requestJSON(t, http.MethodPost, "/system/parse-duration", "", map[string]any{
		"duration": "1500ms",
	})
	if okResp.Status != http.StatusOK || !okResp.Envelope.Success {
		t.Fatalf("parse-duration success path failed status=%d body=%s", okResp.Status, okResp.Body)
	}

	data := mustDataMap(t, okResp)
	if normalized := mustString(t, data["duration"], "parse-duration.duration"); normalized != "1.5s" {
		t.Fatalf("parsed duration=%q want=1.5s", normalized)
	}
	if ms := intFromAny(t, data["milliseconds"], "parse-duration.milliseconds"); ms != 1500 {
		t.Fatalf("milliseconds=%d want=1500", ms)
	}

	invalidResp := h.requestJSON(t, http.MethodPost, "/system/parse-duration", "", map[string]any{
		"duration": "not-a-duration",
	})
	if invalidResp.Status != http.StatusBadRequest || invalidResp.Envelope.Success {
		t.Fatalf("parse-duration invalid path status=%d body=%s", invalidResp.Status, invalidResp.Body)
	}
	if invalidResp.Envelope.Error == nil || invalidResp.Envelope.Error.Code != "bad_request" {
		t.Fatalf("expected bad_request error body=%s", invalidResp.Body)
	}

	negativeResp := h.requestJSON(t, http.MethodPost, "/system/parse-duration", "", map[string]any{
		"duration": "-1s",
	})
	if negativeResp.Status != http.StatusOK || !negativeResp.Envelope.Success {
		t.Fatalf("parse-duration negative path failed status=%d body=%s", negativeResp.Status, negativeResp.Body)
	}
	negativeData := mustDataMap(t, negativeResp)
	if normalized := mustString(t, negativeData["duration"], "parse-duration.negative.duration"); normalized != "-1s" {
		t.Fatalf("negative parsed duration=%q want=-1s", normalized)
	}
	if ms := intFromAny(t, negativeData["milliseconds"], "parse-duration.negative.milliseconds"); ms != -1000 {
		t.Fatalf("negative milliseconds=%d want=-1000", ms)
	}

	overflowResp := h.requestJSON(t, http.MethodPost, "/system/parse-duration", "", map[string]any{
		"duration": "1000000000000000000000000h",
	})
	if overflowResp.Status != http.StatusBadRequest || overflowResp.Envelope.Success {
		t.Fatalf("parse-duration overflow path status=%d body=%s", overflowResp.Status, overflowResp.Body)
	}
	if overflowResp.Envelope.Error == nil || overflowResp.Envelope.Error.Code != "bad_request" {
		t.Fatalf("expected bad_request for overflow duration body=%s", overflowResp.Body)
	}
}

func TestIntegrationAuthLifecycle(t *testing.T) {
	h := requireIntegration(t)

	session := h.createVerifiedSession(t, "auth_lifecycle")
	if session.UserID == "" || session.AccessToken == "" || session.RefreshToken == "" {
		t.Fatalf("expected non-empty auth session fields: %+v", session)
	}

	refreshResp := h.requestJSON(t, http.MethodPost, "/api/v1/auth/refresh", "", map[string]any{
		"refresh_token": session.RefreshToken,
	})
	if refreshResp.Status != http.StatusOK || !refreshResp.Envelope.Success {
		t.Fatalf("refresh failed status=%d body=%s", refreshResp.Status, refreshResp.Body)
	}
	refreshData := mustDataMap(t, refreshResp)
	newAccess := mustString(t, refreshData["access_token"], "refresh.access_token")
	if newAccess == "" {
		t.Fatal("expected non-empty refreshed access token")
	}

	invalidRefreshResp := h.requestJSON(t, http.MethodPost, "/api/v1/auth/refresh", "", map[string]any{
		"refresh_token": "definitely-invalid-token",
	})
	if invalidRefreshResp.Status != http.StatusUnauthorized || invalidRefreshResp.Envelope.Success {
		t.Fatalf("invalid refresh status=%d body=%s", invalidRefreshResp.Status, invalidRefreshResp.Body)
	}
	if invalidRefreshResp.Envelope.Error == nil || invalidRefreshResp.Envelope.Error.Code != "unauthorized" {
		t.Fatalf("expected unauthorized for invalid refresh body=%s", invalidRefreshResp.Body)
	}

	whoamiBefore := h.requestJSON(t, http.MethodGet, "/api/v1/system/whoami", session.AccessToken, nil)
	if whoamiBefore.Status != http.StatusOK || !whoamiBefore.Envelope.Success {
		t.Fatalf("whoami before logout failed status=%d body=%s", whoamiBefore.Status, whoamiBefore.Body)
	}

	logoutResp := h.requestJSON(t, http.MethodPost, "/api/v1/auth/logout", session.AccessToken, nil)
	if logoutResp.Status != http.StatusOK || !logoutResp.Envelope.Success {
		t.Fatalf("logout failed status=%d body=%s", logoutResp.Status, logoutResp.Body)
	}

	logoutNoAuth := h.requestJSON(t, http.MethodPost, "/api/v1/auth/logout", "", nil)
	if logoutNoAuth.Status != http.StatusUnauthorized || logoutNoAuth.Envelope.Success {
		t.Fatalf("logout without auth status=%d body=%s", logoutNoAuth.Status, logoutNoAuth.Body)
	}

	badVerifyResp := h.requestJSON(t, http.MethodPost, "/api/v1/auth/verify-email", "", map[string]any{"token": "bad-token"})
	if badVerifyResp.Status != http.StatusBadRequest || badVerifyResp.Envelope.Success {
		t.Fatalf("invalid verify-email status=%d body=%s", badVerifyResp.Status, badVerifyResp.Body)
	}

	invalidLoginResp := h.requestJSON(t, http.MethodPost, "/api/v1/auth/login", "", map[string]any{
		"email":    "invalid-email",
		"password": "bad",
	})
	if invalidLoginResp.Status != http.StatusBadRequest || invalidLoginResp.Envelope.Success {
		t.Fatalf("invalid login status=%d body=%s", invalidLoginResp.Status, invalidLoginResp.Body)
	}
	if invalidLoginResp.Envelope.Error == nil || invalidLoginResp.Envelope.Error.Code != "bad_request" {
		t.Fatalf("expected bad_request for invalid login body=%s", invalidLoginResp.Body)
	}
}

func assertDependencyOK(t *testing.T, deps map[string]any, name string) {
	t.Helper()
	raw, ok := deps[name]
	if !ok {
		t.Fatalf("dependency %q missing in readiness report", name)
	}
	dep := mustMap(t, raw, "dependency "+name)
	status := mustString(t, dep["status"], name+".status")
	if status != "ok" {
		t.Fatalf("dependency %s status=%q want=ok", name, status)
	}
}

func intFromAny(t *testing.T, value any, field string) int {
	t.Helper()
	floatVal, ok := value.(float64)
	if !ok {
		t.Fatalf("%s is not numeric", field)
	}
	return int(floatVal)
}
