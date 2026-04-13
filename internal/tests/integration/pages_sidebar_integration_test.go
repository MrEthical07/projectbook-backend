package integration

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestIntegrationPagesLifecycleAndInvalidation(t *testing.T) {
	h := requireIntegration(t)

	owner := h.createVerifiedSession(t, "pages_owner")
	project := h.createProject(t, owner.AccessToken, "Pages Lifecycle Integration")

	listPath := "/api/v1/projects/" + project.Slug + "/pages"
	route := "/api/v1/projects/{projectId}/pages"

	missBefore := h.metricCounterValue(t, "superapi_cache_operations_total", route, "miss")
	hitBefore := h.metricCounterValue(t, "superapi_cache_operations_total", route, "hit")

	firstList := h.requestJSON(t, http.MethodGet, listPath, owner.AccessToken, nil)
	if firstList.Status != http.StatusOK || !firstList.Envelope.Success {
		t.Fatalf("first pages list status=%d body=%s", firstList.Status, firstList.Body)
	}

	secondList := h.requestJSON(t, http.MethodGet, listPath, owner.AccessToken, nil)
	if secondList.Status != http.StatusOK || !secondList.Envelope.Success {
		t.Fatalf("second pages list status=%d body=%s", secondList.Status, secondList.Body)
	}

	missAfter := h.metricCounterValue(t, "superapi_cache_operations_total", route, "miss")
	hitAfter := h.metricCounterValue(t, "superapi_cache_operations_total", route, "hit")
	if missAfter < missBefore+1 {
		t.Fatalf("expected pages cache miss increment: before=%.0f after=%.0f", missBefore, missAfter)
	}
	if hitAfter < hitBefore+1 {
		t.Fatalf("expected pages cache hit increment: before=%.0f after=%.0f", hitBefore, hitAfter)
	}

	tag := fmt.Sprintf("pages.project|path.projectId=%s", url.QueryEscape(project.Slug))
	versionBefore := h.cacheTagVersion(t, tag)

	pageTitle := "Integration Page"
	createResp := h.requestJSON(t, http.MethodPost, listPath, owner.AccessToken, map[string]any{"title": pageTitle})
	if createResp.Status != http.StatusCreated || !createResp.Envelope.Success {
		t.Fatalf("create page status=%d body=%s", createResp.Status, createResp.Body)
	}

	created := mustDataMap(t, createResp)
	pageSlug := mustString(t, created["id"], "page.id")
	pageUUID, status, isOrphan, revision := queryPageStateBySlug(t, h, project.UUID, pageSlug)
	if status != "Draft" {
		t.Fatalf("new page status=%q want=Draft", status)
	}
	if !isOrphan {
		t.Fatalf("new page isOrphan=%v want=true", isOrphan)
	}
	if revision < 1 {
		t.Fatalf("new page revision=%d want>=1", revision)
	}

	doc := mustFindPageDocument(t, h, pageUUID)
	if gotProjectID, _ := doc["project_id"].(string); gotProjectID != project.UUID {
		t.Fatalf("page document project_id=%q want=%q", gotProjectID, project.UUID)
	}
	initialContent := extractBsonMap(t, doc, "content")
	if heading, _ := initialContent["docHeading"].(string); heading != pageTitle {
		t.Fatalf("page document content.docHeading=%q want=%q", heading, pageTitle)
	}

	renamePath := "/api/v1/projects/" + project.Slug + "/pages/" + pageSlug + "/rename"
	renamedTitle := "Integration Page Renamed"
	renameResp := h.requestJSON(t, http.MethodPut, renamePath, owner.AccessToken, map[string]any{"title": renamedTitle})
	if renameResp.Status != http.StatusOK || !renameResp.Envelope.Success {
		t.Fatalf("rename page status=%d body=%s", renameResp.Status, renameResp.Body)
	}

	updatePath := "/api/v1/projects/" + project.Slug + "/pages/" + pageSlug
	archiveResp := h.requestJSON(t, http.MethodPut, updatePath, owner.AccessToken, map[string]any{
		"state": map[string]any{"status": "Archived"},
	})
	if archiveResp.Status != http.StatusOK || !archiveResp.Envelope.Success {
		t.Fatalf("archive page status=%d body=%s", archiveResp.Status, archiveResp.Body)
	}

	_, archivedStatus, _, archivedRevision := queryPageStateBySlug(t, h, project.UUID, pageSlug)
	if archivedStatus != "Archived" {
		t.Fatalf("archived page status=%q want=Archived", archivedStatus)
	}
	if archivedRevision <= revision {
		t.Fatalf("archived page revision=%d want>%d", archivedRevision, revision)
	}

	updatedDoc := mustFindPageDocument(t, h, pageUUID)
	updatedContent := extractBsonMap(t, updatedDoc, "content")
	if gotStatus, _ := updatedContent["status"].(string); gotStatus != "Archived" {
		t.Fatalf("page document content.status=%q want=Archived", gotStatus)
	}

	renameArchivedResp := h.requestJSON(t, http.MethodPut, renamePath, owner.AccessToken, map[string]any{"title": "Should Fail"})
	if renameArchivedResp.Status != http.StatusBadRequest || renameArchivedResp.Envelope.Success {
		t.Fatalf("rename archived page status=%d body=%s", renameArchivedResp.Status, renameArchivedResp.Body)
	}

	versionAfter := h.cacheTagVersion(t, tag)
	if versionAfter <= versionBefore {
		t.Fatalf("expected pages cache tag version bump for %s: before=%d after=%d", tag, versionBefore, versionAfter)
	}
}

func TestIntegrationSidebarPagesDeleteRemovesDocument(t *testing.T) {
	h := requireIntegration(t)

	owner := h.createVerifiedSession(t, "sidebar_pages_owner")
	project := h.createProject(t, owner.AccessToken, "Sidebar Pages Integration")

	tag := fmt.Sprintf("pages.project|path.projectId=%s", url.QueryEscape(project.Slug))
	versionBefore := h.cacheTagVersion(t, tag)

	createPath := "/api/v1/projects/" + project.Slug + "/sidebar/artifacts"
	createResp := h.requestJSON(t, http.MethodPost, createPath, owner.AccessToken, map[string]any{
		"prefix": "pages",
		"title":  "Sidebar Created Page",
	})
	if createResp.Status != http.StatusCreated || !createResp.Envelope.Success {
		t.Fatalf("create sidebar page artifact status=%d body=%s", createResp.Status, createResp.Body)
	}

	created := mustDataMap(t, createResp)
	pageSlug := mustString(t, created["id"], "sidebar.page.id")
	pageUUID, _, _, _ := queryPageStateBySlug(t, h, project.UUID, pageSlug)
	_ = mustFindPageDocument(t, h, pageUUID)

	deletePath := "/api/v1/projects/" + project.Slug + "/sidebar/artifacts/" + pageSlug
	deleteResp := h.requestJSON(t, http.MethodDelete, deletePath, owner.AccessToken, map[string]any{"prefix": "pages"})
	if deleteResp.Status != http.StatusOK || !deleteResp.Envelope.Success {
		t.Fatalf("delete sidebar page artifact status=%d body=%s", deleteResp.Status, deleteResp.Body)
	}

	deleted := mustDataMap(t, deleteResp)
	if got := mustString(t, deleted["id"], "sidebar.delete.id"); got != pageSlug {
		t.Fatalf("deleted id=%q want=%q", got, pageSlug)
	}

	var exists bool
	err := h.pgPool.QueryRow(
		context.Background(),
		`SELECT EXISTS(SELECT 1 FROM pages WHERE project_id = $1::uuid AND slug = $2)`,
		project.UUID,
		pageSlug,
	).Scan(&exists)
	if err != nil {
		t.Fatalf("query deleted page row existence: %v", err)
	}
	if exists {
		t.Fatalf("expected page %s to be deleted", pageSlug)
	}

	findErr := h.mongoDB.Collection("page_documents").FindOne(context.Background(), bson.M{"artifact_id": pageUUID}).Decode(&bson.M{})
	if !errors.Is(findErr, mongo.ErrNoDocuments) {
		t.Fatalf("expected page document deletion for artifact_id=%s, got err=%v", pageUUID, findErr)
	}

	versionAfter := h.cacheTagVersion(t, tag)
	if versionAfter <= versionBefore {
		t.Fatalf("expected pages cache tag version bump for %s: before=%d after=%d", tag, versionBefore, versionAfter)
	}

	deleteAgain := h.requestJSON(t, http.MethodDelete, deletePath, owner.AccessToken, map[string]any{"prefix": "pages"})
	if deleteAgain.Status != http.StatusNotFound || deleteAgain.Envelope.Success {
		t.Fatalf("expected not found when deleting already-removed page status=%d body=%s", deleteAgain.Status, deleteAgain.Body)
	}
}

func queryPageStateBySlug(t *testing.T, h *integrationHarness, projectUUID, pageSlug string) (id, status string, isOrphan bool, revision int) {
	t.Helper()

	err := h.pgPool.QueryRow(
		context.Background(),
		`SELECT id::text, status::text, is_orphan, document_revision FROM pages WHERE project_id = $1::uuid AND slug = $2`,
		projectUUID,
		pageSlug,
	).Scan(&id, &status, &isOrphan, &revision)
	if err != nil {
		t.Fatalf("query page state for slug=%s: %v", pageSlug, err)
	}

	return id, status, isOrphan, revision
}

func mustFindPageDocument(t *testing.T, h *integrationHarness, pageUUID string) bson.M {
	t.Helper()

	var doc bson.M
	err := h.mongoDB.Collection("page_documents").FindOne(context.Background(), bson.M{"artifact_id": pageUUID}).Decode(&doc)
	if err != nil {
		t.Fatalf("find page document artifact_id=%s: %v", pageUUID, err)
	}

	return doc
}
