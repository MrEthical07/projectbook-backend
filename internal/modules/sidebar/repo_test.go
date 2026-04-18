package sidebar

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/MrEthical07/superapi/internal/core/storage"
	pagesmod "github.com/MrEthical07/superapi/internal/modules/pages"
)

type fakeRelationalStore struct {
	projectUUID      string
	projectSlug      string
	artifactUUID     string
	loggedArtifactID string
}

func (f *fakeRelationalStore) Kind() storage.Kind { return storage.KindRelational }

func (f *fakeRelationalStore) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func (f *fakeRelationalStore) Execute(ctx context.Context, op storage.RelationalOperation) error {
	return op.ExecuteRelational(ctx, f)
}

func (f *fakeRelationalStore) Exec(ctx context.Context, query string, args ...any) error {
	if strings.Contains(query, "INSERT INTO activity_log") {
		if len(args) > 3 {
			id, _ := args[3].(string)
			f.loggedArtifactID = id
		}
	}
	return nil
}

func (f *fakeRelationalStore) QueryRow(ctx context.Context, query string, scan func(storage.RowScanner) error, args ...any) error {
	if strings.Contains(query, "FROM projects WHERE id::text = $1") {
		return scan(fakeRow{values: []any{f.projectUUID, f.projectSlug}})
	}
	if strings.Contains(query, "SELECT id::text FROM problems") {
		return scan(fakeRow{values: []any{f.artifactUUID}})
	}
	return fmt.Errorf("unexpected query: %s", query)
}

func (f *fakeRelationalStore) Query(ctx context.Context, query string, scan func(storage.RowScanner) error, args ...any) error {
	return nil
}

type fakeDocumentStore struct{}

func (f *fakeDocumentStore) Kind() storage.Kind { return storage.KindDocument }

func (f *fakeDocumentStore) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func (f *fakeDocumentStore) Execute(ctx context.Context, op storage.DocumentOperation) error {
	return nil
}

type fakeRow struct {
	values []any
}

func (r fakeRow) Scan(dest ...any) error {
	if len(dest) != len(r.values) {
		return fmt.Errorf("scan mismatch: got %d destinations, want %d", len(dest), len(r.values))
	}
	for i := range dest {
		target := reflect.ValueOf(dest[i])
		if target.Kind() != reflect.Ptr || target.IsNil() {
			return fmt.Errorf("destination %d is not a writable pointer", i)
		}
		value := reflect.ValueOf(r.values[i])
		elem := target.Elem()
		if value.Type().AssignableTo(elem.Type()) {
			elem.Set(value)
			continue
		}
		if value.Type().ConvertibleTo(elem.Type()) {
			elem.Set(value.Convert(elem.Type()))
			continue
		}
		return fmt.Errorf("cannot assign %s to %s", value.Type(), elem.Type())
	}
	return nil
}

type fakeSidebarArtifactsRepo struct {
	createProblemResult map[string]any
	updateProblemResult map[string]any
}

func (f *fakeSidebarArtifactsRepo) CreateStory(ctx context.Context, projectID, actorUserID, title string, content map[string]any) (map[string]any, error) {
	return map[string]any{"id": "story-1", "title": title}, nil
}

func (f *fakeSidebarArtifactsRepo) CreateJourney(ctx context.Context, projectID, actorUserID, title string, content map[string]any) (map[string]any, error) {
	return map[string]any{"id": "journey-1", "title": title}, nil
}

func (f *fakeSidebarArtifactsRepo) CreateProblem(ctx context.Context, projectID, actorUserID, statement string, content map[string]any) (map[string]any, error) {
	if f.createProblemResult != nil {
		return f.createProblemResult, nil
	}
	return map[string]any{"id": "problem-1", "title": statement}, nil
}

func (f *fakeSidebarArtifactsRepo) CreateIdea(ctx context.Context, projectID, actorUserID, title string, content map[string]any) (map[string]any, error) {
	return map[string]any{"id": "idea-1", "title": title}, nil
}

func (f *fakeSidebarArtifactsRepo) CreateTask(ctx context.Context, projectID, actorUserID, title string, content map[string]any) (map[string]any, error) {
	return map[string]any{"id": "task-1", "title": title}, nil
}

func (f *fakeSidebarArtifactsRepo) CreateFeedback(ctx context.Context, projectID, actorUserID, title string, content map[string]any) (map[string]any, error) {
	return map[string]any{"id": "feedback-1", "title": title}, nil
}

func (f *fakeSidebarArtifactsRepo) UpdateStory(ctx context.Context, projectID, storyID, actorUserID string, patch map[string]any) (map[string]any, error) {
	return map[string]any{"id": storyID, "title": patch["title"]}, nil
}

func (f *fakeSidebarArtifactsRepo) UpdateJourney(ctx context.Context, projectID, journeyID, actorUserID string, patch map[string]any) (map[string]any, error) {
	return map[string]any{"id": journeyID, "title": patch["title"]}, nil
}

func (f *fakeSidebarArtifactsRepo) UpdateProblem(ctx context.Context, projectID, problemID, actorUserID string, patch map[string]any) (map[string]any, error) {
	if f.updateProblemResult != nil {
		return f.updateProblemResult, nil
	}
	return map[string]any{"id": problemID, "title": patch["title"]}, nil
}

func (f *fakeSidebarArtifactsRepo) UpdateIdea(ctx context.Context, projectID, ideaID, actorUserID string, patch map[string]any) (map[string]any, error) {
	return map[string]any{"id": ideaID, "title": patch["title"]}, nil
}

func (f *fakeSidebarArtifactsRepo) UpdateTask(ctx context.Context, projectID, taskID, actorUserID string, patch map[string]any) (map[string]any, error) {
	return map[string]any{"id": taskID, "title": patch["title"]}, nil
}

func (f *fakeSidebarArtifactsRepo) UpdateFeedback(ctx context.Context, projectID, feedbackID, actorUserID string, patch map[string]any) (map[string]any, error) {
	return map[string]any{"id": feedbackID, "title": patch["title"]}, nil
}

type fakeSidebarPagesService struct{}

func (f *fakeSidebarPagesService) CreatePageForSidebar(ctx context.Context, projectID, actorUserID, title string) (pagesmod.PageListItem, error) {
	return pagesmod.PageListItem{ID: "page-1", Title: title}, nil
}

func (f *fakeSidebarPagesService) RenamePageForSidebar(ctx context.Context, projectID, pageID, actorUserID, title string) (pagesmod.RenamePageResponse, error) {
	return pagesmod.RenamePageResponse{ID: pageID, Title: title}, nil
}

func (f *fakeSidebarPagesService) DeletePageForSidebar(ctx context.Context, projectID, pageID, actorUserID string) (pagesmod.DeletePageResponse, error) {
	return pagesmod.DeletePageResponse{ID: pageID}, nil
}

func TestCreateSidebarArtifactLogsUUIDForActivity(t *testing.T) {
	store := &fakeRelationalStore{
		projectUUID:  "11111111-1111-1111-1111-111111111111",
		projectSlug:  "my-first-project-yey",
		artifactUUID: "22222222-2222-2222-2222-222222222222",
	}
	r := &repo{
		store:         store,
		docs:          &fakeDocumentStore{},
		artifactsRepo: &fakeSidebarArtifactsRepo{createProblemResult: map[string]any{"id": "my-first-ps", "title": "My First PS"}},
		pagesSvc:      &fakeSidebarPagesService{},
	}

	created, err := r.CreateSidebarArtifact(context.Background(), store.projectUUID, "33333333-3333-3333-3333-333333333333", "problem-statement", "My First PS")
	if err != nil {
		t.Fatalf("CreateSidebarArtifact returned error: %v", err)
	}
	if got := strings.TrimSpace(fmt.Sprint(created["id"])); got != "my-first-ps" {
		t.Fatalf("expected response id to stay slug, got %q", got)
	}
	if store.loggedArtifactID != store.artifactUUID {
		t.Fatalf("expected activity artifact_id %q, got %q", store.artifactUUID, store.loggedArtifactID)
	}
}

func TestRenameSidebarArtifactLogsUUIDForActivity(t *testing.T) {
	store := &fakeRelationalStore{
		projectUUID:  "11111111-1111-1111-1111-111111111111",
		projectSlug:  "my-first-project-yey",
		artifactUUID: "22222222-2222-2222-2222-222222222222",
	}
	r := &repo{
		store:         store,
		docs:          &fakeDocumentStore{},
		artifactsRepo: &fakeSidebarArtifactsRepo{updateProblemResult: map[string]any{"id": "my-first-ps", "title": "Renamed PS"}},
		pagesSvc:      &fakeSidebarPagesService{},
	}

	updated, err := r.RenameSidebarArtifact(context.Background(), store.projectUUID, "my-first-ps", "33333333-3333-3333-3333-333333333333", "problem-statement", "Renamed PS")
	if err != nil {
		t.Fatalf("RenameSidebarArtifact returned error: %v", err)
	}
	if got := strings.TrimSpace(fmt.Sprint(updated["id"])); got != "my-first-ps" {
		t.Fatalf("expected response id to remain slug, got %q", got)
	}
	if store.loggedArtifactID != store.artifactUUID {
		t.Fatalf("expected activity artifact_id %q, got %q", store.artifactUUID, store.loggedArtifactID)
	}
}
