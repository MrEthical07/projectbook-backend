package integration

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func TestIntegrationResourcesPersistenceAndInvalidation(t *testing.T) {
	h := requireIntegration(t)

	owner := h.createVerifiedSession(t, "resource_owner")
	outsider := h.createVerifiedSession(t, "resource_outsider")
	project := h.createProject(t, owner.AccessToken, "Resources Storage Integration")

	listPath := "/api/v1/projects/" + project.Slug + "/resources"

	unauthorized := h.requestJSON(t, http.MethodGet, listPath, "", nil)
	if unauthorized.Status != http.StatusUnauthorized || unauthorized.Envelope.Success {
		t.Fatalf("unauthenticated resources list status=%d body=%s", unauthorized.Status, unauthorized.Body)
	}

	nonMember := h.requestJSON(t, http.MethodGet, listPath, outsider.AccessToken, nil)
	if nonMember.Status != http.StatusForbidden || nonMember.Envelope.Success {
		t.Fatalf("non-member resources list status=%d body=%s", nonMember.Status, nonMember.Body)
	}

	route := "/api/v1/projects/{projectId}/resources"
	missBefore := h.metricCounterValue(t, "superapi_cache_operations_total", route, "miss")
	hitBefore := h.metricCounterValue(t, "superapi_cache_operations_total", route, "hit")

	ownerFirstList := h.requestJSON(t, http.MethodGet, listPath, owner.AccessToken, nil)
	if ownerFirstList.Status != http.StatusOK || !ownerFirstList.Envelope.Success {
		t.Fatalf("owner first resources list status=%d body=%s", ownerFirstList.Status, ownerFirstList.Body)
	}

	ownerSecondList := h.requestJSON(t, http.MethodGet, listPath, owner.AccessToken, nil)
	if ownerSecondList.Status != http.StatusOK || !ownerSecondList.Envelope.Success {
		t.Fatalf("owner second resources list status=%d body=%s", ownerSecondList.Status, ownerSecondList.Body)
	}

	missAfter := h.metricCounterValue(t, "superapi_cache_operations_total", route, "miss")
	hitAfter := h.metricCounterValue(t, "superapi_cache_operations_total", route, "hit")
	if missAfter < missBefore+1 {
		t.Fatalf("expected resources cache miss increment: before=%.0f after=%.0f", missBefore, missAfter)
	}
	if hitAfter < hitBefore+1 {
		t.Fatalf("expected resources cache hit increment: before=%.0f after=%.0f", hitBefore, hitAfter)
	}

	createResp := h.requestJSON(t, http.MethodPost, listPath, owner.AccessToken, map[string]any{
		"name":    "Integration Resource",
		"docType": "Specification",
	})
	if createResp.Status != http.StatusCreated || !createResp.Envelope.Success {
		t.Fatalf("create resource status=%d body=%s", createResp.Status, createResp.Body)
	}

	createdData := mustDataMap(t, createResp)
	resourceSlug := mustString(t, createdData["id"], "resource.id")
	resourceUUID := h.findResourceUUID(t, project.UUID, resourceSlug)

	var title, docType, status string
	err := h.pgPool.QueryRow(
		context.Background(),
		`SELECT title, doc_type, status::text FROM resources WHERE id = $1::uuid AND project_id = $2::uuid`,
		resourceUUID,
		project.UUID,
	).Scan(&title, &docType, &status)
	if err != nil {
		t.Fatalf("query created resource row: %v", err)
	}
	if title != "Integration Resource" || docType != "Specification" || status != "Active" {
		t.Fatalf("unexpected resource row values title=%q docType=%q status=%q", title, docType, status)
	}

	doc := h.mustFindResourceDocument(t, resourceUUID)
	if gotProjectID, _ := doc["project_id"].(string); gotProjectID != project.UUID {
		t.Fatalf("resource document project_id=%q want=%q", gotProjectID, project.UUID)
	}
	content := extractBsonMap(t, doc, "content")
	if gotDocType, _ := content["docType"].(string); gotDocType != "Specification" {
		t.Fatalf("resource document content.docType=%q want=Specification", gotDocType)
	}

	tag := fmt.Sprintf("resources.project|path.projectId=%s", url.QueryEscape(project.Slug))
	versionBefore := h.cacheTagVersion(t, tag)

	updateStatusPath := "/api/v1/projects/" + project.Slug + "/resources/" + resourceSlug + "/status"
	updateResp := h.requestJSON(t, http.MethodPut, updateStatusPath, owner.AccessToken, map[string]any{
		"status": "Archived",
	})
	if updateResp.Status != http.StatusOK || !updateResp.Envelope.Success {
		t.Fatalf("update resource status failed status=%d body=%s", updateResp.Status, updateResp.Body)
	}

	versionAfter := h.cacheTagVersion(t, tag)
	if versionAfter <= versionBefore {
		t.Fatalf("expected resources cache tag version bump for %s: before=%d after=%d", tag, versionBefore, versionAfter)
	}

	err = h.pgPool.QueryRow(context.Background(), `SELECT status::text FROM resources WHERE id = $1::uuid`, resourceUUID).Scan(&status)
	if err != nil {
		t.Fatalf("query archived resource row: %v", err)
	}
	if status != "Archived" {
		t.Fatalf("resource row status=%q want=Archived", status)
	}

	updatedDoc := h.mustFindResourceDocument(t, resourceUUID)
	updatedContent := extractBsonMap(t, updatedDoc, "content")
	if gotStatus, _ := updatedContent["status"].(string); gotStatus != "Archived" {
		t.Fatalf("resource document content.status=%q want=Archived", gotStatus)
	}

	revision := intFromBsonAny(updatedDoc["revision"])
	if revision < 2 {
		t.Fatalf("resource document revision=%d want>=2", revision)
	}
}

func extractBsonMap(t *testing.T, payload bson.M, field string) bson.M {
	t.Helper()
	value, ok := payload[field]
	if !ok {
		t.Fatalf("missing field %q in bson payload", field)
	}
	mapValue, ok := value.(bson.M)
	if ok {
		return mapValue
	}
	generic, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("field %q is not a map", field)
	}
	return bson.M(generic)
}

func intFromBsonAny(value any) int {
	switch typed := value.(type) {
	case int:
		return typed
	case int32:
		return int(typed)
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	default:
		return 0
	}
}
