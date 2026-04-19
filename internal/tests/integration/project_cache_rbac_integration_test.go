package integration

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestIntegrationProjectSettingsRBACAndCache(t *testing.T) {
	h := requireIntegration(t)

	owner := h.createVerifiedSession(t, "project_owner")
	limitedMember := h.createVerifiedSession(t, "project_limited")
	outsider := h.createVerifiedSession(t, "project_outsider")

	project := h.createProject(t, owner.AccessToken, "RBAC Cache Integration")
	h.upsertCustomMember(t, project.UUID, limitedMember.UserID, 0)

	settingsPath := "/api/v1/projects/" + project.Slug + "/settings"

	unauthorized := h.requestJSON(t, http.MethodGet, settingsPath, "", nil)
	if unauthorized.Status != http.StatusUnauthorized || unauthorized.Envelope.Success {
		t.Fatalf("unauthenticated settings read status=%d body=%s", unauthorized.Status, unauthorized.Body)
	}

	nonMember := h.requestJSON(t, http.MethodGet, settingsPath, outsider.AccessToken, nil)
	if nonMember.Status != http.StatusForbidden || nonMember.Envelope.Success {
		t.Fatalf("non-member settings read status=%d body=%s", nonMember.Status, nonMember.Body)
	}

	noPermission := h.requestJSON(t, http.MethodGet, settingsPath, limitedMember.AccessToken, nil)
	if noPermission.Status != http.StatusForbidden || noPermission.Envelope.Success {
		t.Fatalf("limited member settings read status=%d body=%s", noPermission.Status, noPermission.Body)
	}

	route := "/api/v1/projects/{projectId}/settings"
	missBefore := h.metricCounterValue(t, "superapi_cache_operations_total", route, "miss")
	hitBefore := h.metricCounterValue(t, "superapi_cache_operations_total", route, "hit")
	setBefore := h.metricCounterValue(t, "superapi_cache_operations_total", route, "set")

	ownerFirst := h.requestJSON(t, http.MethodGet, settingsPath, owner.AccessToken, nil)
	if ownerFirst.Status != http.StatusOK || !ownerFirst.Envelope.Success {
		t.Fatalf("owner first settings read status=%d body=%s", ownerFirst.Status, ownerFirst.Body)
	}
	cacheControl := ownerFirst.Header.Get("Cache-Control")
	if !strings.Contains(cacheControl, "private") || !strings.Contains(cacheControl, "max-age=300") {
		t.Fatalf("unexpected Cache-Control header: %q", cacheControl)
	}

	ownerSecond := h.requestJSON(t, http.MethodGet, settingsPath, owner.AccessToken, nil)
	if ownerSecond.Status != http.StatusOK || !ownerSecond.Envelope.Success {
		t.Fatalf("owner second settings read status=%d body=%s", ownerSecond.Status, ownerSecond.Body)
	}

	missAfter := h.metricCounterValue(t, "superapi_cache_operations_total", route, "miss")
	hitAfter := h.metricCounterValue(t, "superapi_cache_operations_total", route, "hit")
	setAfter := h.metricCounterValue(t, "superapi_cache_operations_total", route, "set")

	if missAfter < missBefore+1 {
		t.Fatalf("expected cache miss increment: before=%.0f after=%.0f", missBefore, missAfter)
	}
	if setAfter < setBefore+1 {
		t.Fatalf("expected cache set increment: before=%.0f after=%.0f", setBefore, setAfter)
	}
	if hitAfter < hitBefore+1 {
		t.Fatalf("expected cache hit increment: before=%.0f after=%.0f", hitBefore, hitAfter)
	}

	tag := fmt.Sprintf("project.settings|project=%s", url.QueryEscape(project.UUID))
	versionBefore := h.cacheTagVersion(t, tag)

	updatedName := "RBAC Cache Integration Updated"
	updatedDescription := fmt.Sprintf("updated-at-%d", time.Now().UnixNano())
	updateResp := h.requestJSON(t, http.MethodPatch, settingsPath, owner.AccessToken, map[string]any{
		"settings": map[string]any{
			"projectName":        updatedName,
			"projectDescription": updatedDescription,
			"projectStatus":      "Active",
			"deliveryChannel":    "In-app",
		},
	})
	if updateResp.Status != http.StatusOK || !updateResp.Envelope.Success {
		t.Fatalf("update settings failed status=%d body=%s", updateResp.Status, updateResp.Body)
	}

	versionAfter := h.cacheTagVersion(t, tag)
	if versionAfter <= versionBefore {
		t.Fatalf("expected cache tag version bump for %s: before=%d after=%d", tag, versionBefore, versionAfter)
	}

	postUpdateRead := h.requestJSON(t, http.MethodGet, settingsPath, owner.AccessToken, nil)
	if postUpdateRead.Status != http.StatusOK || !postUpdateRead.Envelope.Success {
		t.Fatalf("post-update settings read status=%d body=%s", postUpdateRead.Status, postUpdateRead.Body)
	}

	settings := mustDataMap(t, postUpdateRead)
	if got := mustString(t, settings["projectName"], "settings.projectName"); got != updatedName {
		t.Fatalf("projectName=%q want=%q", got, updatedName)
	}
}
