package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func TestIntegrationArtifactsLifecycleAndTransitions(t *testing.T) {
	h := requireIntegration(t)

	owner := h.createVerifiedSession(t, "artifacts_owner")
	project := h.createProject(t, owner.AccessToken, "Artifacts Lifecycle Integration")
	base := "/api/v1/projects/" + project.Slug

	route := "/api/v1/projects/{projectId}/stories"
	missBefore := h.metricCounterValue(t, "superapi_cache_operations_total", route, "miss")
	hitBefore := h.metricCounterValue(t, "superapi_cache_operations_total", route, "hit")

	storiesListPath := base + "/stories"
	storiesFirst := h.requestJSON(t, http.MethodGet, storiesListPath, owner.AccessToken, nil)
	if storiesFirst.Status != http.StatusOK || !storiesFirst.Envelope.Success {
		t.Fatalf("stories first list status=%d body=%s", storiesFirst.Status, storiesFirst.Body)
	}
	storiesSecond := h.requestJSON(t, http.MethodGet, storiesListPath, owner.AccessToken, nil)
	if storiesSecond.Status != http.StatusOK || !storiesSecond.Envelope.Success {
		t.Fatalf("stories second list status=%d body=%s", storiesSecond.Status, storiesSecond.Body)
	}

	missAfter := h.metricCounterValue(t, "superapi_cache_operations_total", route, "miss")
	hitAfter := h.metricCounterValue(t, "superapi_cache_operations_total", route, "hit")
	if missAfter < missBefore+1 {
		t.Fatalf("expected stories cache miss increment: before=%.0f after=%.0f", missBefore, missAfter)
	}
	if hitAfter < hitBefore+1 {
		t.Fatalf("expected stories cache hit increment: before=%.0f after=%.0f", hitBefore, hitAfter)
	}

	artifactTag := fmt.Sprintf("artifacts.project|path.projectId=%s", url.QueryEscape(project.Slug))
	versionBefore := h.cacheTagVersion(t, artifactTag)

	storyCreate := h.requestJSON(t, http.MethodPost, storiesListPath, owner.AccessToken, map[string]any{"title": "Integration Story"})
	if storyCreate.Status != http.StatusCreated || !storyCreate.Envelope.Success {
		t.Fatalf("create story status=%d body=%s", storyCreate.Status, storyCreate.Body)
	}
	storyData := mustDataMap(t, storyCreate)
	storySlug := mustString(t, storyData["id"], "story.id")
	storyUUID, storyStatus, storyRevision := queryArtifactStateBySlug(t, h, "stories", project.UUID, storySlug)
	if storyStatus != "Draft" || storyRevision < 1 {
		t.Fatalf("story initial state status=%q revision=%d", storyStatus, storyRevision)
	}
	_ = mustFindArtifactDocument(t, h, "story_documents", storyUUID)

	storyLockResp := h.requestJSON(t, http.MethodPatch, base+"/stories/"+storySlug, owner.AccessToken, map[string]any{
		"story": map[string]any{"status": "Locked"},
	})
	if storyLockResp.Status != http.StatusOK || !storyLockResp.Envelope.Success {
		t.Fatalf("lock story status=%d body=%s", storyLockResp.Status, storyLockResp.Body)
	}
	_, storyStatus, _ = queryArtifactStateBySlug(t, h, "stories", project.UUID, storySlug)
	if storyStatus != "Locked" {
		t.Fatalf("story status=%q want=Locked", storyStatus)
	}

	storyImmutable := h.requestJSON(t, http.MethodPatch, base+"/stories/"+storySlug, owner.AccessToken, map[string]any{
		"story": map[string]any{"title": "Should Fail"},
	})
	if storyImmutable.Status != http.StatusBadRequest || storyImmutable.Envelope.Success {
		t.Fatalf("locked story mutation status=%d body=%s", storyImmutable.Status, storyImmutable.Body)
	}

	storyArchive := h.requestJSON(t, http.MethodPatch, base+"/stories/"+storySlug, owner.AccessToken, map[string]any{
		"story": map[string]any{"status": "Archived"},
	})
	if storyArchive.Status != http.StatusOK || !storyArchive.Envelope.Success {
		t.Fatalf("archive story status=%d body=%s", storyArchive.Status, storyArchive.Body)
	}

	journeyCreate := h.requestJSON(t, http.MethodPost, base+"/journeys", owner.AccessToken, map[string]any{"title": "Integration Journey"})
	if journeyCreate.Status != http.StatusCreated || !journeyCreate.Envelope.Success {
		t.Fatalf("create journey status=%d body=%s", journeyCreate.Status, journeyCreate.Body)
	}
	journeyData := mustDataMap(t, journeyCreate)
	journeySlug := mustString(t, journeyData["id"], "journey.id")
	journeyUUID, journeyStatus, _ := queryArtifactStateBySlug(t, h, "journeys", project.UUID, journeySlug)
	if journeyStatus != "Draft" {
		t.Fatalf("journey status=%q want=Draft", journeyStatus)
	}
	_ = mustFindArtifactDocument(t, h, "journey_documents", journeyUUID)

	journeyArchive := h.requestJSON(t, http.MethodPatch, base+"/journeys/"+journeySlug, owner.AccessToken, map[string]any{
		"journey": map[string]any{"status": "Archived"},
	})
	if journeyArchive.Status != http.StatusOK || !journeyArchive.Envelope.Success {
		t.Fatalf("archive journey status=%d body=%s", journeyArchive.Status, journeyArchive.Body)
	}
	journeyImmutable := h.requestJSON(t, http.MethodPatch, base+"/journeys/"+journeySlug, owner.AccessToken, map[string]any{
		"journey": map[string]any{"title": "Forbidden"},
	})
	if journeyImmutable.Status != http.StatusBadRequest || journeyImmutable.Envelope.Success {
		t.Fatalf("archived journey mutation status=%d body=%s", journeyImmutable.Status, journeyImmutable.Body)
	}

	problemCreate := h.requestJSON(t, http.MethodPost, base+"/problems", owner.AccessToken, map[string]any{"statement": "Integration Problem"})
	if problemCreate.Status != http.StatusCreated || !problemCreate.Envelope.Success {
		t.Fatalf("create problem status=%d body=%s", problemCreate.Status, problemCreate.Body)
	}
	problemData := mustDataMap(t, problemCreate)
	problemSlug := mustString(t, problemData["id"], "problem.id")
	problemUUID, problemStatus, _ := queryArtifactStateBySlug(t, h, "problems", project.UUID, problemSlug)
	if problemStatus != "Draft" {
		t.Fatalf("problem status=%q want=Draft", problemStatus)
	}
	_ = mustFindArtifactDocument(t, h, "problem_documents", problemUUID)

	problemLock := h.requestJSON(t, http.MethodPatch, base+"/problems/"+problemSlug, owner.AccessToken, map[string]any{
		"state": map[string]any{"status": "Locked"},
	})
	if problemLock.Status != http.StatusOK || !problemLock.Envelope.Success {
		t.Fatalf("lock problem status=%d body=%s", problemLock.Status, problemLock.Body)
	}
	locked, problemStatus := queryProblemLockState(t, h, project.UUID, problemSlug)
	if !locked || problemStatus != "Locked" {
		t.Fatalf("problem lock state locked=%v status=%q", locked, problemStatus)
	}

	ideaCreate := h.requestJSON(t, http.MethodPost, base+"/ideas", owner.AccessToken, map[string]any{"title": "Integration Idea"})
	if ideaCreate.Status != http.StatusCreated || !ideaCreate.Envelope.Success {
		t.Fatalf("create idea status=%d body=%s", ideaCreate.Status, ideaCreate.Body)
	}
	ideaData := mustDataMap(t, ideaCreate)
	ideaSlug := mustString(t, ideaData["id"], "idea.id")
	ideaUUID, ideaStatus, _ := queryArtifactStateBySlug(t, h, "ideas", project.UUID, ideaSlug)
	if ideaStatus != "Considered" {
		t.Fatalf("idea status=%q want=Considered", ideaStatus)
	}
	_ = mustFindArtifactDocument(t, h, "idea_documents", ideaUUID)

	ideaLink := h.requestJSON(t, http.MethodPatch, base+"/ideas/"+ideaSlug, owner.AccessToken, map[string]any{
		"state": map[string]any{"selectedProblemId": problemSlug},
	})
	if ideaLink.Status != http.StatusOK || !ideaLink.Envelope.Success {
		t.Fatalf("link idea to problem status=%d body=%s", ideaLink.Status, ideaLink.Body)
	}

	ideaSelect := h.requestJSON(t, http.MethodPatch, base+"/ideas/"+ideaSlug, owner.AccessToken, map[string]any{
		"state": map[string]any{"status": "Selected"},
	})
	if ideaSelect.Status != http.StatusOK || !ideaSelect.Envelope.Success {
		t.Fatalf("select idea status=%d body=%s", ideaSelect.Status, ideaSelect.Body)
	}
	ideaReject := h.requestJSON(t, http.MethodPatch, base+"/ideas/"+ideaSlug, owner.AccessToken, map[string]any{
		"state": map[string]any{"status": "Rejected"},
	})
	if ideaReject.Status != http.StatusBadRequest || ideaReject.Envelope.Success {
		t.Fatalf("reject selected idea status=%d body=%s", ideaReject.Status, ideaReject.Body)
	}

	ideaRejectedCreate := h.requestJSON(t, http.MethodPost, base+"/ideas", owner.AccessToken, map[string]any{"title": "Rejected Integration Idea"})
	if ideaRejectedCreate.Status != http.StatusCreated || !ideaRejectedCreate.Envelope.Success {
		t.Fatalf("create rejected-branch idea status=%d body=%s", ideaRejectedCreate.Status, ideaRejectedCreate.Body)
	}
	ideaRejectedSlug := mustString(t, mustDataMap(t, ideaRejectedCreate)["id"], "idea.rejected.id")
	ideaRejectAllowed := h.requestJSON(t, http.MethodPatch, base+"/ideas/"+ideaRejectedSlug, owner.AccessToken, map[string]any{
		"state": map[string]any{"status": "Rejected"},
	})
	if ideaRejectAllowed.Status != http.StatusOK || !ideaRejectAllowed.Envelope.Success {
		t.Fatalf("reject considered idea status=%d body=%s", ideaRejectAllowed.Status, ideaRejectAllowed.Body)
	}

	ideaImmutable := h.requestJSON(t, http.MethodPatch, base+"/ideas/"+ideaRejectedSlug, owner.AccessToken, map[string]any{
		"state": map[string]any{"title": "Should Fail"},
	})
	if ideaImmutable.Status != http.StatusBadRequest || ideaImmutable.Envelope.Success {
		t.Fatalf("immutable idea update status=%d body=%s", ideaImmutable.Status, ideaImmutable.Body)
	}

	problemArchive := h.requestJSON(t, http.MethodPatch, base+"/problems/"+problemSlug, owner.AccessToken, map[string]any{
		"state": map[string]any{"status": "Archived"},
	})
	if problemArchive.Status != http.StatusOK || !problemArchive.Envelope.Success {
		t.Fatalf("archive problem status=%d body=%s", problemArchive.Status, problemArchive.Body)
	}
	problemRelock := h.requestJSON(t, http.MethodPatch, base+"/problems/"+problemSlug, owner.AccessToken, map[string]any{
		"state": map[string]any{"status": "Locked"},
	})
	if problemRelock.Status != http.StatusOK || !problemRelock.Envelope.Success {
		t.Fatalf("relock archived problem status=%d body=%s", problemRelock.Status, problemRelock.Body)
	}
	_, relockedProblemStatus := queryProblemLockState(t, h, project.UUID, problemSlug)
	if relockedProblemStatus != "Locked" {
		t.Fatalf("relocked problem status=%q want=Locked", relockedProblemStatus)
	}

	taskCreate := h.requestJSON(t, http.MethodPost, base+"/tasks", owner.AccessToken, map[string]any{"title": "Integration Task"})
	if taskCreate.Status != http.StatusCreated || !taskCreate.Envelope.Success {
		t.Fatalf("create task status=%d body=%s", taskCreate.Status, taskCreate.Body)
	}
	taskData := mustDataMap(t, taskCreate)
	taskSlug := mustString(t, taskData["id"], "task.id")
	taskUUID, taskStatus, _ := queryArtifactStateBySlug(t, h, "tasks", project.UUID, taskSlug)
	if taskStatus != "Planned" {
		t.Fatalf("task status=%q want=Planned", taskStatus)
	}
	_ = mustFindArtifactDocument(t, h, "task_documents", taskUUID)

	taskProgress := h.requestJSON(t, http.MethodPatch, base+"/tasks/"+taskSlug, owner.AccessToken, map[string]any{
		"state": map[string]any{"status": "In Progress"},
	})
	if taskProgress.Status != http.StatusOK || !taskProgress.Envelope.Success {
		t.Fatalf("task in-progress status=%d body=%s", taskProgress.Status, taskProgress.Body)
	}
	taskComplete := h.requestJSON(t, http.MethodPatch, base+"/tasks/"+taskSlug, owner.AccessToken, map[string]any{
		"state": map[string]any{"status": "Completed"},
	})
	if taskComplete.Status != http.StatusOK || !taskComplete.Envelope.Success {
		t.Fatalf("task completed status=%d body=%s", taskComplete.Status, taskComplete.Body)
	}
	taskBack := h.requestJSON(t, http.MethodPatch, base+"/tasks/"+taskSlug, owner.AccessToken, map[string]any{
		"state": map[string]any{"status": "Planned"},
	})
	if taskBack.Status != http.StatusBadRequest || taskBack.Envelope.Success {
		t.Fatalf("task invalid transition status=%d body=%s", taskBack.Status, taskBack.Body)
	}

	feedbackCreate := h.requestJSON(t, http.MethodPost, base+"/feedback", owner.AccessToken, map[string]any{"title": "Integration Feedback"})
	if feedbackCreate.Status != http.StatusCreated || !feedbackCreate.Envelope.Success {
		t.Fatalf("create feedback status=%d body=%s", feedbackCreate.Status, feedbackCreate.Body)
	}
	feedbackData := mustDataMap(t, feedbackCreate)
	feedbackSlug := mustString(t, feedbackData["id"], "feedback.id")
	feedbackUUID, _ := queryFeedbackStateBySlug(t, h, project.UUID, feedbackSlug)
	_ = mustFindArtifactDocument(t, h, "feedback_documents", feedbackUUID)

	feedbackUpdate := h.requestJSON(t, http.MethodPatch, base+"/feedback/"+feedbackSlug, owner.AccessToken, map[string]any{
		"state": map[string]any{"outcome": "Validated"},
	})
	if feedbackUpdate.Status != http.StatusOK || !feedbackUpdate.Envelope.Success {
		t.Fatalf("update feedback outcome status=%d body=%s", feedbackUpdate.Status, feedbackUpdate.Body)
	}

	feedbackGet := h.requestJSON(t, http.MethodGet, base+"/feedback/"+feedbackSlug, owner.AccessToken, nil)
	if feedbackGet.Status != http.StatusOK || !feedbackGet.Envelope.Success {
		t.Fatalf("get feedback status=%d body=%s", feedbackGet.Status, feedbackGet.Body)
	}

	outcome, revision := queryFeedbackOutcome(t, h, project.UUID, feedbackSlug)
	if outcome != "Validated" {
		t.Fatalf("feedback outcome=%q want=Validated", outcome)
	}
	if revision < 2 {
		t.Fatalf("feedback revision=%d want>=2", revision)
	}

	versionAfter := h.cacheTagVersion(t, artifactTag)
	if versionAfter <= versionBefore {
		t.Fatalf("expected artifacts cache tag version bump for %s: before=%d after=%d", artifactTag, versionBefore, versionAfter)
	}
}

func TestIntegrationCalendarEventLifecycle(t *testing.T) {
	h := requireIntegration(t)

	owner := h.createVerifiedSession(t, "calendar_owner")
	project := h.createProject(t, owner.AccessToken, "Calendar Lifecycle Integration")
	base := "/api/v1/projects/" + project.Slug + "/calendar"

	createResp := h.requestJSON(t, http.MethodPost, base, owner.AccessToken, map[string]any{
		"title":           "Integration Calendar Event",
		"start":           "2026-01-10",
		"end":             "2026-01-10",
		"allDay":          false,
		"startTime":       "10:30",
		"endTime":         "11:30",
		"owner":           "Integration Owner",
		"phase":           "Define",
		"description":     "Calendar integration create",
		"location":        "Room A",
		"eventKind":       "Workshop",
		"linkedArtifacts": []any{},
		"tags":            []any{"integration"},
	})
	if createResp.Status != http.StatusCreated || !createResp.Envelope.Success {
		t.Fatalf("create calendar event status=%d body=%s", createResp.Status, createResp.Body)
	}
	created := mustDataMap(t, createResp)
	eventID := mustString(t, created["id"], "calendar.event.id")

	listResp := h.requestJSON(t, http.MethodGet, base, owner.AccessToken, nil)
	if listResp.Status != http.StatusOK || !listResp.Envelope.Success {
		t.Fatalf("list calendar events status=%d body=%s", listResp.Status, listResp.Body)
	}
	listData := mustDataMap(t, listResp)
	events := mustSliceField(t, listData, "items")
	if len(events) == 0 {
		t.Fatal("expected at least one calendar event in list")
	}

	getResp := h.requestJSON(t, http.MethodGet, base+"/"+eventID, owner.AccessToken, nil)
	if getResp.Status != http.StatusOK || !getResp.Envelope.Success {
		t.Fatalf("get calendar event status=%d body=%s", getResp.Status, getResp.Body)
	}

	updateResp := h.requestJSON(t, http.MethodPatch, base+"/"+eventID, owner.AccessToken, map[string]any{
		"state": map[string]any{
			"title":       "Integration Calendar Event Updated",
			"phase":       "Ideate",
			"allDay":      true,
			"start":       "2026-01-12",
			"end":         "2026-01-12",
			"eventKind":   "Review",
			"description": "Calendar integration updated",
		},
	})
	if updateResp.Status != http.StatusOK || !updateResp.Envelope.Success {
		t.Fatalf("update calendar event status=%d body=%s", updateResp.Status, updateResp.Body)
	}

	getUpdated := h.requestJSON(t, http.MethodGet, base+"/"+eventID, owner.AccessToken, nil)
	if getUpdated.Status != http.StatusOK || !getUpdated.Envelope.Success {
		t.Fatalf("get updated calendar event status=%d body=%s", getUpdated.Status, getUpdated.Body)
	}
	updatedData := mustDataMap(t, getUpdated)
	eventPayload := mustMap(t, updatedData["event"], "calendar.event")
	if got := mustString(t, eventPayload["title"], "calendar.event.title"); got != "Integration Calendar Event Updated" {
		t.Fatalf("updated calendar event title=%q", got)
	}

	deleteResp := h.requestJSON(t, http.MethodDelete, base+"/"+eventID, owner.AccessToken, nil)
	if deleteResp.Status != http.StatusOK || !deleteResp.Envelope.Success {
		t.Fatalf("delete calendar event status=%d body=%s", deleteResp.Status, deleteResp.Body)
	}

	getMissing := h.requestJSON(t, http.MethodGet, base+"/"+eventID, owner.AccessToken, nil)
	if getMissing.Status != http.StatusNotFound || getMissing.Envelope.Success {
		t.Fatalf("expected deleted calendar event to be missing status=%d body=%s", getMissing.Status, getMissing.Body)
	}

	deleteAgain := h.requestJSON(t, http.MethodDelete, base+"/"+eventID, owner.AccessToken, nil)
	if deleteAgain.Status != http.StatusNotFound || deleteAgain.Envelope.Success {
		t.Fatalf("expected second calendar delete to be not-found status=%d body=%s", deleteAgain.Status, deleteAgain.Body)
	}

	invalidCreate := h.requestJSON(t, http.MethodPost, base, owner.AccessToken, map[string]any{
		"title":     "Bad Calendar Event",
		"start":     "2026-01-10",
		"end":       "2026-01-10",
		"allDay":    true,
		"owner":     "Integration Owner",
		"phase":     "Invalid",
		"eventKind": "Workshop",
	})
	if invalidCreate.Status != http.StatusBadRequest || invalidCreate.Envelope.Success {
		t.Fatalf("invalid calendar create status=%d body=%s", invalidCreate.Status, invalidCreate.Body)
	}
}

func TestIntegrationTeamHomeAndProjectLifecycle(t *testing.T) {
	h := requireIntegration(t)

	owner := h.createVerifiedSession(t, "team_owner")
	invitee := h.createVerifiedSession(t, "team_invitee")
	declinee := h.createVerifiedSession(t, "team_declinee")
	cancelled := h.createVerifiedSession(t, "team_cancelled")
	project := h.createProject(t, owner.AccessToken, "Team Home Project Integration")

	teamBase := "/api/v1/projects/" + project.Slug + "/team"

	membersResp := h.requestJSON(t, http.MethodGet, teamBase+"/members", owner.AccessToken, nil)
	if membersResp.Status != http.StatusOK || !membersResp.Envelope.Success {
		t.Fatalf("list team members status=%d body=%s", membersResp.Status, membersResp.Body)
	}

	createInviteResp := h.requestJSON(t, http.MethodPost, teamBase+"/invites", owner.AccessToken, map[string]any{
		"email": invitee.Email,
		"role":  "Viewer",
	})
	if createInviteResp.Status != http.StatusCreated || !createInviteResp.Envelope.Success {
		t.Fatalf("create invite status=%d body=%s", createInviteResp.Status, createInviteResp.Body)
	}

	duplicateInvite := h.requestJSON(t, http.MethodPost, teamBase+"/invites", owner.AccessToken, map[string]any{
		"email": invitee.Email,
		"role":  "Viewer",
	})
	if duplicateInvite.Status != http.StatusConflict || duplicateInvite.Envelope.Success {
		t.Fatalf("duplicate invite status=%d body=%s", duplicateInvite.Status, duplicateInvite.Body)
	}

	inviteeInvitesResp := h.requestJSON(t, http.MethodGet, "/api/v1/home/invites", invitee.AccessToken, nil)
	if inviteeInvitesResp.Status != http.StatusOK || !inviteeInvitesResp.Envelope.Success {
		t.Fatalf("invitee home invites status=%d body=%s", inviteeInvitesResp.Status, inviteeInvitesResp.Body)
	}
	inviteID := findInviteIDForProject(t, mustDataSlice(t, inviteeInvitesResp), project.Slug)

	acceptResp := h.requestJSON(t, http.MethodPost, "/api/v1/home/invites/"+inviteID+"/accept", invitee.AccessToken, nil)
	if acceptResp.Status != http.StatusOK || !acceptResp.Envelope.Success {
		t.Fatalf("accept invite status=%d body=%s", acceptResp.Status, acceptResp.Body)
	}

	memberID, role, mask := queryProjectMemberState(t, h, project.UUID, invitee.UserID)
	if role != "Viewer" {
		t.Fatalf("invitee role=%q want=Viewer", role)
	}
	if mask == 0 {
		t.Fatal("expected non-zero viewer permission mask")
	}

	editorMask := queryRoleMask(t, h, project.UUID, "Editor")
	memberUpdateResp := h.requestJSON(t, http.MethodPatch, teamBase+"/members/"+memberID+"/permissions", owner.AccessToken, map[string]any{
		"role":           "Editor",
		"isCustom":       true,
		"permissionMask": editorMask,
	})
	if memberUpdateResp.Status != http.StatusOK || !memberUpdateResp.Envelope.Success {
		t.Fatalf("update member permissions status=%d body=%s", memberUpdateResp.Status, memberUpdateResp.Body)
	}

	_, updatedRole, updatedMask := queryProjectMemberState(t, h, project.UUID, invitee.UserID)
	if updatedRole != "Editor" {
		t.Fatalf("updated member role=%q want=Editor", updatedRole)
	}
	if updatedMask != int64(editorMask) {
		t.Fatalf("updated member mask=%d want=%d", updatedMask, editorMask)
	}

	viewerMask := queryRoleMask(t, h, project.UUID, "Viewer")
	roleUpdateResp := h.requestJSON(t, http.MethodPatch, teamBase+"/roles/viewer/permissions", owner.AccessToken, map[string]any{
		"role":           "Viewer",
		"permissionMask": viewerMask,
	})
	if roleUpdateResp.Status != http.StatusOK || !roleUpdateResp.Envelope.Success {
		t.Fatalf("update role permissions status=%d body=%s", roleUpdateResp.Status, roleUpdateResp.Body)
	}

	declineInviteCreate := h.requestJSON(t, http.MethodPost, teamBase+"/invites", owner.AccessToken, map[string]any{
		"email": declinee.Email,
		"role":  "Member",
	})
	if declineInviteCreate.Status != http.StatusCreated || !declineInviteCreate.Envelope.Success {
		t.Fatalf("create decline invite status=%d body=%s", declineInviteCreate.Status, declineInviteCreate.Body)
	}

	declineeInvites := h.requestJSON(t, http.MethodGet, "/api/v1/home/invites", declinee.AccessToken, nil)
	if declineeInvites.Status != http.StatusOK || !declineeInvites.Envelope.Success {
		t.Fatalf("declinee home invites status=%d body=%s", declineeInvites.Status, declineeInvites.Body)
	}
	declineInviteID := findInviteIDForProject(t, mustDataSlice(t, declineeInvites), project.Slug)

	declineResp := h.requestJSON(t, http.MethodPost, "/api/v1/home/invites/"+declineInviteID+"/decline", declinee.AccessToken, nil)
	if declineResp.Status != http.StatusOK || !declineResp.Envelope.Success {
		t.Fatalf("decline invite status=%d body=%s", declineResp.Status, declineResp.Body)
	}
	if status := queryInviteStatusByEmail(t, h, project.UUID, declinee.Email); strings.ToLower(status) != "declined" {
		t.Fatalf("declined invite status=%q want=declined", status)
	}

	cancelInviteCreate := h.requestJSON(t, http.MethodPost, teamBase+"/invites", owner.AccessToken, map[string]any{
		"email": cancelled.Email,
		"role":  "Member",
	})
	if cancelInviteCreate.Status != http.StatusCreated || !cancelInviteCreate.Envelope.Success {
		t.Fatalf("create cancel invite status=%d body=%s", cancelInviteCreate.Status, cancelInviteCreate.Body)
	}

	cancelPath := teamBase + "/invites/" + url.PathEscape(cancelled.Email)
	cancelResp := h.requestJSON(t, http.MethodDelete, cancelPath, owner.AccessToken, nil)
	if cancelResp.Status != http.StatusOK || !cancelResp.Envelope.Success {
		t.Fatalf("cancel invite status=%d body=%s", cancelResp.Status, cancelResp.Body)
	}
	if status := queryInviteStatusByEmail(t, h, project.UUID, cancelled.Email); strings.ToLower(status) != "cancelled" {
		t.Fatalf("cancelled invite status=%q want=cancelled", status)
	}

	batchResp := h.requestJSON(t, http.MethodPost, teamBase+"/invites/batch", owner.AccessToken, map[string]any{
		"invites": []map[string]any{
			{"email": "invalid-email", "role": "Editor"},
			{"email": fmt.Sprintf("batch_%d@example.com", time.Now().UnixNano()), "role": "Editor"},
		},
	})
	if batchResp.Status != http.StatusMultiStatus || !batchResp.Envelope.Success {
		t.Fatalf("batch invite partial status=%d body=%s", batchResp.Status, batchResp.Body)
	}

	inviteeProjectsResp := h.requestJSON(t, http.MethodGet, "/api/v1/home/projects", invitee.AccessToken, nil)
	if inviteeProjectsResp.Status != http.StatusOK || !inviteeProjectsResp.Envelope.Success {
		t.Fatalf("invitee home projects status=%d body=%s", inviteeProjectsResp.Status, inviteeProjectsResp.Body)
	}
	assertHomeProjectPresent(t, mustDataSlice(t, inviteeProjectsResp), project.Slug)

	accountUpdate := h.requestJSON(t, http.MethodPatch, "/api/v1/home/account", owner.AccessToken, map[string]any{
		"settings": map[string]any{
			"displayName":        "Integration Owner Updated",
			"bio":                "Owner bio",
			"theme":              "Dark",
			"density":            "Compact",
			"landing":            "Project Selector",
			"timeFormat":         "12-hour",
			"inAppNotifications": true,
			"emailNotifications": false,
		},
	})
	if accountUpdate.Status != http.StatusOK || !accountUpdate.Envelope.Success {
		t.Fatalf("update account status=%d body=%s", accountUpdate.Status, accountUpdate.Body)
	}

	accountGet := h.requestJSON(t, http.MethodGet, "/api/v1/home/account", owner.AccessToken, nil)
	if accountGet.Status != http.StatusOK || !accountGet.Envelope.Success {
		t.Fatalf("get account after update status=%d body=%s", accountGet.Status, accountGet.Body)
	}
	accountData := mustDataMap(t, accountGet)
	if got := mustString(t, accountData["displayName"], "account.displayName"); got != "Integration Owner Updated" {
		t.Fatalf("account displayName=%q want=Integration Owner Updated", got)
	}

	storySeed := h.requestJSON(t, http.MethodPost, "/api/v1/projects/"+project.Slug+"/stories", owner.AccessToken, map[string]any{"title": "Delete Cascade Story"})
	if storySeed.Status != http.StatusCreated || !storySeed.Envelope.Success {
		t.Fatalf("seed story for project delete status=%d body=%s", storySeed.Status, storySeed.Body)
	}
	resourceSeed := h.requestJSON(t, http.MethodPost, "/api/v1/projects/"+project.Slug+"/resources", owner.AccessToken, map[string]any{"name": "Delete Cascade Resource", "docType": "Specification"})
	if resourceSeed.Status != http.StatusCreated || !resourceSeed.Envelope.Success {
		t.Fatalf("seed resource for project delete status=%d body=%s", resourceSeed.Status, resourceSeed.Body)
	}
	pageSeed := h.requestJSON(t, http.MethodPost, "/api/v1/projects/"+project.Slug+"/pages", owner.AccessToken, map[string]any{"title": "Delete Cascade Page"})
	if pageSeed.Status != http.StatusCreated || !pageSeed.Envelope.Success {
		t.Fatalf("seed page for project delete status=%d body=%s", pageSeed.Status, pageSeed.Body)
	}

	archiveResp := h.requestJSON(t, http.MethodPost, "/api/v1/projects/"+project.Slug+"/archive", owner.AccessToken, nil)
	if archiveResp.Status != http.StatusOK || !archiveResp.Envelope.Success {
		t.Fatalf("archive project status=%d body=%s", archiveResp.Status, archiveResp.Body)
	}
	rearchiveResp := h.requestJSON(t, http.MethodPost, "/api/v1/projects/"+project.Slug+"/archive", owner.AccessToken, nil)
	if rearchiveResp.Status != http.StatusConflict || rearchiveResp.Envelope.Success {
		t.Fatalf("rearchive project status=%d body=%s", rearchiveResp.Status, rearchiveResp.Body)
	}

	deleteResp := h.requestJSON(t, http.MethodDelete, "/api/v1/projects/"+project.Slug, owner.AccessToken, nil)
	if deleteResp.Status != http.StatusOK || !deleteResp.Envelope.Success {
		t.Fatalf("delete project status=%d body=%s", deleteResp.Status, deleteResp.Body)
	}

	assertProjectDeletedCascade(t, h, project.UUID)
}

func mustDataSlice(t *testing.T, resp apiResponse) []any {
	t.Helper()
	if len(resp.Envelope.Data) == 0 {
		t.Fatalf("response data is empty body=%s", resp.Body)
	}
	var out []any
	if err := json.Unmarshal(resp.Envelope.Data, &out); err != nil {
		t.Fatalf("decode response data slice: %v body=%s", err, resp.Body)
	}
	return out
}

func mustSliceField(t *testing.T, payload map[string]any, field string) []any {
	t.Helper()
	value, ok := payload[field]
	if !ok {
		t.Fatalf("missing field %q", field)
	}
	items, ok := value.([]any)
	if !ok {
		t.Fatalf("field %q is not a slice", field)
	}
	return items
}

func queryArtifactStateBySlug(t *testing.T, h *integrationHarness, table, projectUUID, slug string) (id, status string, revision int) {
	t.Helper()

	allowedTables := map[string]bool{
		"stories":  true,
		"journeys": true,
		"problems": true,
		"ideas":    true,
		"tasks":    true,
		"feedback": true,
	}
	if !allowedTables[table] {
		t.Fatalf("unsupported artifact table %q", table)
	}

	query := fmt.Sprintf(`SELECT id::text, status::text, document_revision FROM %s WHERE project_id = $1::uuid AND (id::text = $2 OR slug = $2) LIMIT 1`, table)
	err := h.pgPool.QueryRow(context.Background(), query, projectUUID, slug).Scan(&id, &status, &revision)
	if err != nil {
		t.Fatalf("query artifact state table=%s identifier=%s: %v", table, slug, err)
	}

	return id, status, revision
}

func queryProblemLockState(t *testing.T, h *integrationHarness, projectUUID, slug string) (locked bool, status string) {
	t.Helper()
	err := h.pgPool.QueryRow(
		context.Background(),
		`SELECT is_locked, status::text FROM problems WHERE project_id = $1::uuid AND (id::text = $2 OR slug = $2) LIMIT 1`,
		projectUUID,
		slug,
	).Scan(&locked, &status)
	if err != nil {
		t.Fatalf("query problem lock state identifier=%s: %v", slug, err)
	}
	return locked, status
}

func queryFeedbackOutcome(t *testing.T, h *integrationHarness, projectUUID, slug string) (outcome string, revision int) {
	t.Helper()
	err := h.pgPool.QueryRow(
		context.Background(),
		`SELECT COALESCE(outcome::text, ''), document_revision FROM feedback WHERE project_id = $1::uuid AND (id::text = $2 OR slug = $2) LIMIT 1`,
		projectUUID,
		slug,
	).Scan(&outcome, &revision)
	if err != nil {
		t.Fatalf("query feedback outcome identifier=%s: %v", slug, err)
	}
	return outcome, revision
}

func queryFeedbackStateBySlug(t *testing.T, h *integrationHarness, projectUUID, slug string) (id string, revision int) {
	t.Helper()
	err := h.pgPool.QueryRow(
		context.Background(),
		`SELECT id::text, document_revision FROM feedback WHERE project_id = $1::uuid AND (id::text = $2 OR slug = $2) LIMIT 1`,
		projectUUID,
		slug,
	).Scan(&id, &revision)
	if err != nil {
		t.Fatalf("query feedback state identifier=%s: %v", slug, err)
	}
	return id, revision
}

func mustFindArtifactDocument(t *testing.T, h *integrationHarness, collection, artifactUUID string) bson.M {
	t.Helper()
	var doc bson.M
	err := h.mongoDB.Collection(collection).FindOne(context.Background(), bson.M{"artifact_id": artifactUUID}).Decode(&doc)
	if err != nil {
		t.Fatalf("find artifact document collection=%s artifact_id=%s: %v", collection, artifactUUID, err)
	}
	return doc
}

func findInviteIDForProject(t *testing.T, invites []any, projectSlug string) string {
	t.Helper()
	for _, item := range invites {
		invite := mustMap(t, item, "home.invite")
		projectID, _ := invite["projectId"].(string)
		if strings.TrimSpace(projectID) == strings.TrimSpace(projectSlug) {
			return mustString(t, invite["id"], "home.invite.id")
		}
	}
	t.Fatalf("invite for project %s not found", projectSlug)
	return ""
}

func queryProjectMemberState(t *testing.T, h *integrationHarness, projectUUID, userID string) (memberID, role string, permissionMask int64) {
	t.Helper()
	err := h.pgPool.QueryRow(
		context.Background(),
		`SELECT id::text, role::text, permission_mask FROM project_members WHERE project_id = $1::uuid AND user_id = $2::uuid`,
		projectUUID,
		userID,
	).Scan(&memberID, &role, &permissionMask)
	if err != nil {
		t.Fatalf("query project member state user_id=%s: %v", userID, err)
	}
	return memberID, role, permissionMask
}

func queryRoleMask(t *testing.T, h *integrationHarness, projectUUID, role string) uint64 {
	t.Helper()
	var mask int64
	err := h.pgPool.QueryRow(
		context.Background(),
		`SELECT permission_mask FROM role_permissions WHERE project_id = $1::uuid AND role = $2::project_role`,
		projectUUID,
		role,
	).Scan(&mask)
	if err != nil {
		t.Fatalf("query role mask role=%s: %v", role, err)
	}
	if mask < 0 {
		t.Fatalf("role mask is negative for role=%s", role)
	}
	return uint64(mask)
}

func queryInviteStatusByEmail(t *testing.T, h *integrationHarness, projectUUID, email string) string {
	t.Helper()
	var status string
	err := h.pgPool.QueryRow(
		context.Background(),
		`SELECT status::text FROM project_invites WHERE project_id = $1::uuid AND lower(email) = lower($2) ORDER BY sent_at DESC LIMIT 1`,
		projectUUID,
		email,
	).Scan(&status)
	if err != nil {
		t.Fatalf("query invite status email=%s: %v", email, err)
	}
	return status
}

func assertHomeProjectPresent(t *testing.T, projects []any, projectSlug string) {
	t.Helper()
	for _, item := range projects {
		project := mustMap(t, item, "home.project")
		id, _ := project["id"].(string)
		if strings.TrimSpace(id) == strings.TrimSpace(projectSlug) {
			return
		}
	}
	t.Fatalf("project %s not found in home projects response", projectSlug)
}

func assertProjectDeletedCascade(t *testing.T, h *integrationHarness, projectUUID string) {
	t.Helper()

	var exists bool
	err := h.pgPool.QueryRow(
		context.Background(),
		`SELECT EXISTS(SELECT 1 FROM projects WHERE id = $1::uuid)`,
		projectUUID,
	).Scan(&exists)
	if err != nil {
		t.Fatalf("query project existence after delete: %v", err)
	}
	if exists {
		t.Fatalf("expected project %s to be deleted", projectUUID)
	}

	for _, table := range []string{"project_members", "project_settings", "stories", "resources", "pages", "calendar_events", "project_invites"} {
		count := queryProjectRowCount(t, h, table, projectUUID)
		if count != 0 {
			t.Fatalf("expected zero rows in %s after project delete, found %d", table, count)
		}
	}
}

func queryProjectRowCount(t *testing.T, h *integrationHarness, table, projectUUID string) int {
	t.Helper()

	allowedTables := map[string]bool{
		"project_members":  true,
		"project_settings": true,
		"stories":          true,
		"resources":        true,
		"pages":            true,
		"calendar_events":  true,
		"project_invites":  true,
	}
	if !allowedTables[table] {
		t.Fatalf("unsupported table for row count query: %s", table)
	}

	query := fmt.Sprintf(`SELECT COUNT(1) FROM %s WHERE project_id = $1::uuid`, table)
	var count int
	err := h.pgPool.QueryRow(context.Background(), query, projectUUID).Scan(&count)
	if err != nil {
		t.Fatalf("query row count table=%s: %v", table, err)
	}
	return count
}
