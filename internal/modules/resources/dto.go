package resources

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	apperr "github.com/MrEthical07/superapi/internal/core/errors"
	"github.com/MrEthical07/superapi/internal/core/patchx"
)

const maxListLimit = 100

var slugSanitizer = regexp.MustCompile(`[^a-z0-9]+`)

var resourceStatuses = map[string]struct{}{
	"Active":   {},
	"Archived": {},
}

var resourcePatchRules = map[string]patchx.FieldRule{
	"name":            {AllowNull: true},
	"title":           {AllowNull: true},
	"docType":         {AllowNull: true},
	"description":     {AllowNull: true},
	"notes":           {AllowNull: true},
	"notesText":       {AllowNull: true},
	"linkedArtifacts": {AllowNull: true},
	"versions":        {AllowNull: true},
	"fileType":        {AllowNull: true},
	"status":          {},
}

type listQuery struct {
	Status  string
	DocType string
	Sort    string
	Order   string
	Offset  int
	Limit   int
}

type ResourceLinkedArtifact struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Type   string `json:"type"`
	Phase  string `json:"phase"`
	Href   string `json:"href"`
	Status string `json:"status,omitempty"`
}

type ResourceVersion struct {
	Version     string `json:"version"`
	UploadedBy  string `json:"uploadedBy"`
	UploadDate  string `json:"uploadDate"`
	Description string `json:"description"`
}

type ResourceListItem struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	FileType    string `json:"fileType"`
	DocType     string `json:"docType"`
	Owner       string `json:"owner"`
	Version     string `json:"version"`
	LastUpdated string `json:"lastUpdated"`
	LinkedCount int    `json:"linkedCount"`
	Status      string `json:"status"`
}

type ResourceListReference struct {
	DocTypes       []string `json:"docTypes"`
	FileTypes      []string `json:"fileTypes"`
	Owners         []string `json:"owners"`
	SortOptions    []string `json:"sortOptions"`
	StoryOptions   []string `json:"storyOptions"`
	ProblemOptions []string `json:"problemOptions"`
	IdeaOptions    []string `json:"ideaOptions"`
	TaskOptions    []string `json:"taskOptions"`
}

type ListResourcesResponse struct {
	Items      []ResourceListItem    `json:"items"`
	NextCursor *string               `json:"next_cursor,omitempty"`
	Reference  ResourceListReference `json:"reference"`
}

type ResourceSummary struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	FileType string `json:"fileType"`
	DocType  string `json:"docType"`
	Status   string `json:"status"`
	Owner    string `json:"owner"`
}

type ResourceDetail struct {
	Description     string                   `json:"description"`
	FileSize        string                   `json:"fileSize"`
	LinkedArtifacts []ResourceLinkedArtifact `json:"linkedArtifacts"`
	Versions        []ResourceVersion        `json:"versions"`
	NotesText       string                   `json:"notesText"`
}

type ResourceDetailReference struct {
	StoryOptions   []ResourceLinkedArtifact `json:"storyOptions"`
	ProblemOptions []ResourceLinkedArtifact `json:"problemOptions"`
	IdeaOptions    []ResourceLinkedArtifact `json:"ideaOptions"`
	TaskOptions    []ResourceLinkedArtifact `json:"taskOptions"`
}

type ResourceMeta struct {
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type GetResourceResponse struct {
	Resource  ResourceSummary         `json:"resource"`
	Detail    ResourceDetail          `json:"detail"`
	Reference ResourceDetailReference `json:"reference"`
	Meta      ResourceMeta            `json:"meta"`
}

type UpdateResourceStatusResponse struct {
	ID          string `json:"id"`
	Status      string `json:"status"`
	LastUpdated string `json:"lastUpdated"`
}

type createResourceRequest struct {
	Name    string `json:"name"`
	DocType string `json:"docType"`
}

func (r createResourceRequest) Validate() error {
	if strings.TrimSpace(r.Name) == "" {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "name is required")
	}
	if strings.TrimSpace(r.DocType) == "" {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "docType is required")
	}
	return nil
}

type updateResourceRequest struct {
	State map[string]any `json:"state"`
}

func (r updateResourceRequest) Validate() error {
	if r.State == nil {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "state payload is required")
	}
	if err := validatePatchPayload(r.State, resourcePatchRules, "resource state"); err != nil {
		return err
	}
	if status := toString(r.State["status"]); status != "" {
		canonicalStatus, ok := normalizeAllowedStatus(resourceStatuses, status)
		if !ok {
			return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "invalid resource status")
		}
		r.State["status"] = canonicalStatus
	}
	return nil
}

type updateResourceStatusRequest struct {
	Status string `json:"status"`
}

func (r updateResourceStatusRequest) Validate() error {
	canonicalStatus, ok := normalizeAllowedStatus(resourceStatuses, r.Status)
	if !ok {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "invalid resource status")
	}
	r.Status = canonicalStatus
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

func normalizeSort(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "name":
		return "name"
	case "uploaddate":
		return "uploadDate"
	default:
		return "lastUpdated"
	}
}

func normalizeOrder(raw string) string {
	if strings.EqualFold(strings.TrimSpace(raw), "asc") {
		return "asc"
	}
	return "desc"
}

func isAllowedStatus(allowed map[string]struct{}, raw string) bool {
	_, ok := normalizeAllowedStatus(allowed, raw)
	return ok
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
	s, _ := v.([]any)
	return s
}

func toMap(v any) map[string]any {
	m, _ := v.(map[string]any)
	return m
}

func asStringSlice(values []any) []string {
	items := make([]string, 0, len(values))
	for _, value := range values {
		text := toString(value)
		if text != "" {
			items = append(items, text)
		}
	}
	return items
}

func toInt(value any) int {
	switch typed := value.(type) {
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

func decodeResourceListItem(value map[string]any) ResourceListItem {
	return ResourceListItem{
		ID:          toString(value["id"]),
		Name:        toString(value["name"]),
		FileType:    toString(value["fileType"]),
		DocType:     toString(value["docType"]),
		Owner:       toString(value["owner"]),
		Version:     toString(value["version"]),
		LastUpdated: toString(value["lastUpdated"]),
		LinkedCount: toInt(value["linkedCount"]),
		Status:      toString(value["status"]),
	}
}

func decodeLinkedArtifacts(value any) []ResourceLinkedArtifact {
	rawItems := toSlice(value)
	items := make([]ResourceLinkedArtifact, 0, len(rawItems))
	for _, raw := range rawItems {
		row := toMap(raw)
		if row == nil {
			continue
		}
		item := ResourceLinkedArtifact{
			ID:     toString(row["id"]),
			Title:  toString(row["title"]),
			Type:   toString(row["type"]),
			Phase:  toString(row["phase"]),
			Href:   toString(row["href"]),
			Status: toString(row["status"]),
		}
		if item.ID == "" || item.Title == "" || item.Href == "" {
			continue
		}
		items = append(items, item)
	}
	return items
}

func decodeResourceVersions(value any) []ResourceVersion {
	rawItems := toSlice(value)
	items := make([]ResourceVersion, 0, len(rawItems))
	for _, raw := range rawItems {
		row := toMap(raw)
		if row == nil {
			continue
		}
		item := ResourceVersion{
			Version:     toString(row["version"]),
			UploadedBy:  toString(row["uploadedBy"]),
			UploadDate:  toString(row["uploadDate"]),
			Description: toString(row["description"]),
		}
		if item.Version == "" || item.UploadedBy == "" || item.UploadDate == "" {
			continue
		}
		items = append(items, item)
	}
	return items
}

func decodeListResourcesResponse(payload map[string]any) ListResourcesResponse {
	rawItems := toSlice(payload["items"])
	items := make([]ResourceListItem, 0, len(rawItems))
	for _, raw := range rawItems {
		row := toMap(raw)
		if row == nil {
			continue
		}
		items = append(items, decodeResourceListItem(row))
	}

	var nextCursor *string
	if cursor := toString(payload["next_cursor"]); cursor != "" {
		nextCursor = &cursor
	}

	referenceMap := toMap(payload["reference"])
	if referenceMap == nil {
		referenceMap = map[string]any{}
	}

	return ListResourcesResponse{
		Items:      items,
		NextCursor: nextCursor,
		Reference: ResourceListReference{
			DocTypes:       asStringSlice(toSlice(referenceMap["docTypes"])),
			FileTypes:      asStringSlice(toSlice(referenceMap["fileTypes"])),
			Owners:         asStringSlice(toSlice(referenceMap["owners"])),
			SortOptions:    asStringSlice(toSlice(referenceMap["sortOptions"])),
			StoryOptions:   asStringSlice(toSlice(referenceMap["storyOptions"])),
			ProblemOptions: asStringSlice(toSlice(referenceMap["problemOptions"])),
			IdeaOptions:    asStringSlice(toSlice(referenceMap["ideaOptions"])),
			TaskOptions:    asStringSlice(toSlice(referenceMap["taskOptions"])),
		},
	}
}

func decodeGetResourceResponse(payload map[string]any) GetResourceResponse {
	resourceMap := toMap(payload["resource"])
	if resourceMap == nil {
		resourceMap = map[string]any{}
	}
	detailMap := toMap(payload["detail"])
	if detailMap == nil {
		detailMap = map[string]any{}
	}
	referenceMap := toMap(payload["reference"])
	if referenceMap == nil {
		referenceMap = map[string]any{}
	}
	metaMap := toMap(payload["meta"])
	if metaMap == nil {
		metaMap = map[string]any{}
	}

	notesText := toString(detailMap["notesText"])
	if notesText == "" {
		notesText = toString(detailMap["notes"])
	}

	return GetResourceResponse{
		Resource: ResourceSummary{
			ID:       toString(resourceMap["id"]),
			Name:     toString(resourceMap["name"]),
			FileType: toString(resourceMap["fileType"]),
			DocType:  toString(resourceMap["docType"]),
			Status:   toString(resourceMap["status"]),
			Owner:    toString(resourceMap["owner"]),
		},
		Detail: ResourceDetail{
			Description:     toString(detailMap["description"]),
			FileSize:        toString(detailMap["fileSize"]),
			LinkedArtifacts: decodeLinkedArtifacts(detailMap["linkedArtifacts"]),
			Versions:        decodeResourceVersions(detailMap["versions"]),
			NotesText:       notesText,
		},
		Reference: ResourceDetailReference{
			StoryOptions:   decodeLinkedArtifacts(referenceMap["storyOptions"]),
			ProblemOptions: decodeLinkedArtifacts(referenceMap["problemOptions"]),
			IdeaOptions:    decodeLinkedArtifacts(referenceMap["ideaOptions"]),
			TaskOptions:    decodeLinkedArtifacts(referenceMap["taskOptions"]),
		},
		Meta: ResourceMeta{
			CreatedAt: toString(metaMap["createdAt"]),
			UpdatedAt: toString(metaMap["updatedAt"]),
		},
	}
}

func slugify(raw string) string {
	trimmed := strings.ToLower(strings.TrimSpace(raw))
	if trimmed == "" {
		return "resource"
	}
	slug := slugSanitizer.ReplaceAllString(trimmed, "-")
	slug = strings.Trim(slug, "-")
	if slug == "" {
		return "resource"
	}
	if len(slug) > 96 {
		slug = strings.Trim(slug[:96], "-")
		if slug == "" {
			return "resource"
		}
	}
	return slug
}
