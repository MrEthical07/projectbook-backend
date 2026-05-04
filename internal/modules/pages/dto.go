package pages

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	apperr "github.com/MrEthical07/superapi/internal/core/errors"
	"github.com/MrEthical07/superapi/internal/core/patchx"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const maxListLimit = 50

var slugSanitizer = regexp.MustCompile(`[^a-z0-9]+`)

var pageStatuses = map[string]struct{}{
	"Draft":    {},
	"Archived": {},
}

var pagePatchRules = map[string]patchx.FieldRule{
	"status":          {},
	"title":           {AllowNull: true},
	"owner":           {AllowNull: true},
	"description":     {AllowNull: true},
	"tags":            {AllowNull: true},
	"linkedArtifacts": {AllowNull: true},
	"docHeading":      {AllowNull: true},
	"docBody":         {AllowNull: true},
	"views":           {AllowNull: true},
	"activeViewId":    {AllowNull: true},
	"tableData":       {AllowNull: true},
	"tableColumns":    {AllowNull: true},
	"tableRows":       {AllowNull: true},
	"databaseItems":   {AllowNull: true},
}

type listQuery struct {
	Status string
	Offset int
	Limit  int
}

type PageListItem struct {
	ID                   string `json:"id"`
	Title                string `json:"title"`
	Owner                string `json:"owner"`
	LastEdited           string `json:"lastEdited"`
	LinkedArtifactsCount int    `json:"linkedArtifactsCount"`
	Status               string `json:"status"`
	IsOrphan             bool   `json:"isOrphan"`
}

type ListPagesResponse struct {
	Items      []PageListItem `json:"items"`
	NextCursor *string        `json:"next_cursor,omitempty"`
}

type PageSummary struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Status     string `json:"status"`
	Owner      string `json:"owner"`
	LastEdited string `json:"lastEdited"`
}

type PageView struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type PageDetail struct {
	Description     string            `json:"description"`
	Tags            []string          `json:"tags"`
	LinkedArtifacts []string          `json:"linkedArtifacts"`
	DocHeading      string            `json:"docHeading"`
	DocBody         string            `json:"docBody"`
	Views           []PageView        `json:"views"`
	ActiveViewID    string            `json:"activeViewId"`
	TableData       []json.RawMessage `json:"tableData"`
	TableColumns    []json.RawMessage `json:"tableColumns"`
	TableRows       []json.RawMessage `json:"tableRows"`
	DatabaseItems   []json.RawMessage `json:"databaseItems"`
}

type PageReference struct {
	TagOptions            []string `json:"tagOptions"`
	LinkedArtifactOptions []string `json:"linkedArtifactOptions"`
}

type GetPageResponse struct {
	Page      PageSummary   `json:"page"`
	Detail    PageDetail    `json:"detail"`
	Reference PageReference `json:"reference"`
}

type RenamePageResponse struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	LastEdited string `json:"lastEdited"`
}

type DeletePageResponse struct {
	ID string `json:"id"`
}

type createPageRequest struct {
	Title string `json:"title"`
}

func (r createPageRequest) Validate() error {
	if strings.TrimSpace(r.Title) == "" {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "title is required")
	}
	return nil
}

type updatePageRequest struct {
	State map[string]any `json:"state"`
}

func (r updatePageRequest) Validate() error {
	if r.State == nil {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "state payload is required")
	}
	if err := validatePatchPayload(r.State, pagePatchRules, "page state"); err != nil {
		return err
	}
	if status := toString(r.State["status"]); status != "" {
		canonicalStatus, ok := normalizeAllowedStatus(pageStatuses, status)
		if !ok {
			return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "invalid page status")
		}
		r.State["status"] = canonicalStatus
	}
	return nil
}

type renamePageRequest struct {
	Title string `json:"title"`
}

func (r renamePageRequest) Validate() error {
	if strings.TrimSpace(r.Title) == "" {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "title is required")
	}
	return nil
}

func parseOptionalIntQuery(raw string, fallback int, name string) (int, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(trimmed)
	if err != nil {
		return 0, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, name+" must be an integer")
	}
	if parsed < 0 {
		return 0, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, name+" must be non-negative")
	}
	return parsed, nil
}

func normalizeLimit(limit int) int {
	if limit <= 0 {
		return 20
	}
	if limit > maxListLimit {
		return maxListLimit
	}
	return limit
}

func normalizeAllowedStatus(allowed map[string]struct{}, raw string) (string, bool) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", false
	}
	if _, ok := allowed[trimmed]; ok {
		return trimmed, true
	}
	for candidate := range allowed {
		if strings.EqualFold(candidate, trimmed) {
			return candidate, true
		}
	}
	return "", false
}

func validatePatchPayload(payload map[string]any, rules map[string]patchx.FieldRule, payloadName string) error {
	if err := patchx.ValidatePatch(payload, rules); err != nil {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, fmt.Sprintf("%s contains %s", payloadName, err.Error()))
	}
	return nil
}

func toString(v any) string {
	s, _ := v.(string)
	return strings.TrimSpace(s)
}

func toSlice(v any) []any {
	switch val := v.(type) {
	case []any:
		return val
	case primitive.A:
		return []any(val)
	case *[]any:
		if val != nil {
			return *val
		}
	case *primitive.A:
		if val != nil {
			return []any(*val)
		}
	}
	return []any{}
}

func toMap(v any) map[string]any {
	switch val := v.(type) {
	case map[string]any:
		return val
	case bson.M:
		return map[string]any(val)
	case *bson.M:
		if val != nil {
			return map[string]any(*val)
		}
	}
	return nil
}
func toInt(v any) int {
	switch typed := v.(type) {
	case int:
		return typed
	case int32:
		return int(typed)
	case int64:
		return int(typed)
	case float32:
		return int(typed)
	case float64:
		return int(typed)
	default:
		return 0
	}
}

func toBool(v any) bool {
	b, _ := v.(bool)
	return b
}

func toStringSlice(v any) []string {
	values := toSlice(v)
	items := make([]string, 0, len(values))
	for _, value := range values {
		text := toString(value)
		if text != "" {
			items = append(items, text)
		}
	}
	return items
}

func toStringSliceLoose(v any) []string {
	values := toSlice(v)
	items := make([]string, 0, len(values))
	for _, value := range values {
		if text := toString(value); text != "" {
			items = append(items, text)
			continue
		}
		row := toMap(value)
		if row == nil {
			continue
		}
		text := toString(row["title"])
		if text == "" {
			text = toString(row["label"])
		}
		if text != "" {
			items = append(items, text)
		}
	}
	return items
}

func toJSONRawMessages(v any) []json.RawMessage {
	values := toSlice(v)
	items := make([]json.RawMessage, 0, len(values))
	for _, value := range values {
		bytes, err := json.Marshal(value)
		if err != nil {
			continue
		}
		items = append(items, json.RawMessage(bytes))
	}
	return items
}

func decodePageView(value any) (PageView, bool) {
	row := toMap(value)
	if row == nil {
		return PageView{}, false
	}
	view := PageView{
		ID:   toString(row["id"]),
		Name: toString(row["name"]),
		Type: toString(row["type"]),
	}
	if view.ID == "" || view.Name == "" || view.Type == "" {
		return PageView{}, false
	}
	return view, true
}

func decodePageViews(value any) []PageView {
	rawViews := toSlice(value)
	views := make([]PageView, 0, len(rawViews))
	for _, raw := range rawViews {
		view, ok := decodePageView(raw)
		if !ok {
			continue
		}
		views = append(views, view)
	}
	return views
}

func decodePageListItem(value map[string]any) PageListItem {
	return PageListItem{
		ID:                   toString(value["id"]),
		Title:                toString(value["title"]),
		Owner:                toString(value["owner"]),
		LastEdited:           toString(value["lastEdited"]),
		LinkedArtifactsCount: toInt(value["linkedArtifactsCount"]),
		Status:               toString(value["status"]),
		IsOrphan:             toBool(value["isOrphan"]),
	}
}

func decodeGetPageResponse(payload map[string]any) GetPageResponse {
	pageMap := toMap(payload["page"])
	if pageMap == nil {
		pageMap = map[string]any{}
	}
	detailMap := toMap(payload["detail"])
	if detailMap == nil {
		detailMap = map[string]any{}
	}
	referenceMap := toMap(payload["reference"])
	if referenceMap == nil {
		referenceMap = map[string]any{}
	}

	return GetPageResponse{
		Page: PageSummary{
			ID:         toString(pageMap["id"]),
			Title:      toString(pageMap["title"]),
			Status:     toString(pageMap["status"]),
			Owner:      toString(pageMap["owner"]),
			LastEdited: toString(pageMap["lastEdited"]),
		},
		Detail: PageDetail{
			Description:     toString(detailMap["description"]),
			Tags:            toStringSlice(detailMap["tags"]),
			LinkedArtifacts: toStringSlice(detailMap["linkedArtifacts"]),
			DocHeading:      toString(detailMap["docHeading"]),
			DocBody:         toString(detailMap["docBody"]),
			Views:           decodePageViews(detailMap["views"]),
			ActiveViewID:    toString(detailMap["activeViewId"]),
			TableData:       toJSONRawMessages(detailMap["tableData"]),
			TableColumns:    toJSONRawMessages(detailMap["tableColumns"]),
			TableRows:       toJSONRawMessages(detailMap["tableRows"]),
			DatabaseItems:   toJSONRawMessages(detailMap["databaseItems"]),
		},
		Reference: PageReference{
			TagOptions:            toStringSliceLoose(referenceMap["tagOptions"]),
			LinkedArtifactOptions: toStringSliceLoose(referenceMap["linkedArtifactOptions"]),
		},
	}
}

func decodeRenamePageResponse(payload map[string]any) RenamePageResponse {
	return RenamePageResponse{
		ID:         toString(payload["id"]),
		Title:      toString(payload["title"]),
		LastEdited: toString(payload["lastEdited"]),
	}
}

func decodeDeletePageResponse(payload map[string]any) DeletePageResponse {
	return DeletePageResponse{ID: toString(payload["id"])}
}

func slugify(raw string) string {
	trimmed := strings.ToLower(strings.TrimSpace(raw))
	if trimmed == "" {
		return "page"
	}
	slug := slugSanitizer.ReplaceAllString(trimmed, "-")
	slug = strings.Trim(slug, "-")
	if slug == "" {
		return "page"
	}
	if len(slug) > 96 {
		slug = strings.Trim(slug[:96], "-")
		if slug == "" {
			return "page"
		}
	}
	return slug
}
