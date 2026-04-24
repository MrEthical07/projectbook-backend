package artifacts

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	apperr "github.com/MrEthical07/superapi/internal/core/errors"
	"github.com/MrEthical07/superapi/internal/core/patchx"
)

const maxListLimit = 50

var slugSanitizer = regexp.MustCompile(`[^a-z0-9]+`)

var storyStatuses = map[string]struct{}{
	"Draft":    {},
	"Locked":   {},
	"Archived": {},
}

var journeyStatuses = map[string]struct{}{
	"Draft":    {},
	"Locked":   {},
	"Archived": {},
}

var problemStatuses = map[string]struct{}{
	"Draft":    {},
	"Locked":   {},
	"Archived": {},
}

var ideaStatuses = map[string]struct{}{
	"Considered": {},
	"Selected":   {},
	"Rejected":   {},
	"Archived":   {},
}

var taskStatuses = map[string]struct{}{
	"Planned":     {},
	"In Progress": {},
	"Completed":   {},
	"Abandoned":   {},
}

var feedbackStatuses = map[string]struct{}{
	"Active":   {},
	"Archived": {},
}

var feedbackOutcomes = map[string]struct{}{
	"Validated":       {},
	"Invalidated":     {},
	"Needs Iteration": {},
}

var artifactTypeMap = map[string]string{
	"story":             "story",
	"stories":           "story",
	"user story":        "story",
	"user stories":      "story",
	"journey":           "journey",
	"journeys":          "journey",
	"user journey":      "journey",
	"user journeys":     "journey",
	"problem":           "problem",
	"problems":          "problem",
	"problem statement": "problem",
	"idea":              "idea",
	"ideas":             "idea",
	"task":              "task",
	"tasks":             "task",
	"feedback":          "feedback",
}

var personaPatchRules = map[string]patchx.FieldRule{
	"name": {AllowNull: true},
	"bio":  {AllowNull: true},
	"role": {AllowNull: true},
	"age":  {AllowNull: true},
	"job":  {AllowNull: true},
	"edu":  {AllowNull: true},
}

var empathyPatchRules = map[string]patchx.FieldRule{
	"says":   {AllowNull: true},
	"thinks": {AllowNull: true},
	"does":   {AllowNull: true},
	"feels":  {AllowNull: true},
}

var storyPatchRules = map[string]patchx.FieldRule{
	"title":         {AllowNull: true},
	"description":   {AllowNull: true},
	"status":        {},
	"persona":       {AllowNull: true, Nested: personaPatchRules},
	"context":       {AllowNull: true},
	"empathyMap":    {AllowNull: true, Nested: empathyPatchRules},
	"painPoints":    {AllowNull: true},
	"hypothesis":    {AllowNull: true},
	"addOnSections": {AllowNull: true},
	"notes":         {AllowNull: true},
}

var journeyPatchRules = map[string]patchx.FieldRule{
	"title":       {AllowNull: true},
	"description": {AllowNull: true},
	"status":      {},
	"persona":     {AllowNull: true, Nested: personaPatchRules},
	"context":     {AllowNull: true},
	"stages":      {AllowNull: true},
	"notes":       {AllowNull: true},
}

var problemPatchRules = map[string]patchx.FieldRule{
	"title":              {AllowNull: true},
	"description":        {AllowNull: true},
	"status":             {},
	"finalStatement":     {AllowNull: true},
	"orphanAcknowledged": {AllowNull: true},
	"selectedPainPoints": {AllowNull: true},
	"customPainPoints":   {AllowNull: true},
	"linkedSources":      {AllowNull: true},
	"activeModules":      {AllowNull: true},
	"moduleContent":      {AllowNull: true, AllowAnyNested: true},
	"notesText":          {AllowNull: true},
	"notes":              {AllowNull: true},
}

var ideaPatchRules = map[string]patchx.FieldRule{
	"title":              {AllowNull: true},
	"description":        {AllowNull: true},
	"status":             {},
	"summary":            {AllowNull: true},
	"summaryTitle":       {AllowNull: true},
	"summaryDescription": {AllowNull: true},
	"notes":              {AllowNull: true},
	"notesText":          {AllowNull: true},
	"selectedProblemId":  {AllowNull: true},
	"activeModules":      {AllowNull: true},
	"moduleContent":      {AllowNull: true, AllowAnyNested: true},
}

var taskPatchRules = map[string]patchx.FieldRule{
	"title":          {AllowNull: true},
	"status":         {},
	"assignedToId":   {AllowNull: true},
	"assignedToIds":  {AllowNull: true},
	"selectedIdeaId": {AllowNull: true},
	"deadline":       {AllowNull: true},
	"hypothesis":     {AllowNull: true},
	"planItems":      {AllowNull: true},
	"executionLinks": {AllowNull: true},
	"notes":          {AllowNull: true},
	"notesText":      {AllowNull: true},
	"activeModules":  {AllowNull: true},
	"abandonReason":  {AllowNull: true},
	"hasFeedback":    {AllowNull: true},
}

var feedbackPatchRules = map[string]patchx.FieldRule{
	"title":           {AllowNull: true},
	"description":     {AllowNull: true},
	"status":          {},
	"outcome":         {},
	"isArchived":      {AllowNull: true},
	"linkedArtifacts": {AllowNull: true},
	"activeModules":   {AllowNull: true},
	"moduleContent":   {AllowNull: true, AllowAnyNested: true},
	"notes":           {AllowNull: true},
	"notesText":       {AllowNull: true},
	"observation":     {AllowNull: true},
	"interpretation":  {AllowNull: true},
	"evidenceText":    {AllowNull: true},
	"evidenceLocked":  {AllowNull: true},
	"nextStepsText":   {AllowNull: true},
}

type listQuery struct {
	Status  string
	Outcome string
	Offset  int
	Limit   int
}

type ArtifactMetadata struct {
	Owner        string `json:"owner"`
	CreatedBy    string `json:"createdBy"`
	CreatedAt    string `json:"createdAt"`
	LastEditedBy string `json:"lastEditedBy"`
	LastEditedAt string `json:"lastEditedAt"`
	LastUpdated  string `json:"lastUpdated"`
}

type StoryListItem struct {
	ID                     string `json:"id"`
	Title                  string `json:"title"`
	PersonaName            string `json:"personaName"`
	PainPointsCount        int    `json:"painPointsCount"`
	ProblemHypothesesCount int    `json:"problemHypothesesCount"`
	Owner                  string `json:"owner"`
	LastUpdated            string `json:"lastUpdated"`
	Status                 string `json:"status"`
	IsOrphan               bool   `json:"isOrphan"`
}

type JourneyListItem struct {
	ID              string   `json:"id"`
	Title           string   `json:"title"`
	LinkedPersonas  []string `json:"linkedPersonas"`
	StagesCount     int      `json:"stagesCount"`
	PainPointsCount int      `json:"painPointsCount"`
	Owner           string   `json:"owner"`
	LastUpdated     string   `json:"lastUpdated"`
	Status          string   `json:"status"`
	IsOrphan        bool     `json:"isOrphan"`
}

type ProblemListItem struct {
	ID              string   `json:"id"`
	Statement       string   `json:"statement"`
	LinkedSources   []string `json:"linkedSources"`
	PainPointsCount int      `json:"painPointsCount"`
	IdeasCount      int      `json:"ideasCount"`
	Status          string   `json:"status"`
	Owner           string   `json:"owner"`
	LastUpdated     string   `json:"lastUpdated"`
	IsOrphan        bool     `json:"isOrphan"`
}

type IdeaListItem struct {
	ID                     string `json:"id"`
	Title                  string `json:"title"`
	LinkedProblemStatement string `json:"linkedProblemStatement"`
	Persona                string `json:"persona"`
	Status                 string `json:"status"`
	TasksCount             int    `json:"tasksCount"`
	Owner                  string `json:"owner"`
	LastUpdated            string `json:"lastUpdated"`
	LinkedProblemLocked    bool   `json:"linkedProblemLocked"`
	IsOrphan               bool   `json:"isOrphan"`
}

type TaskListItem struct {
	ID                     string `json:"id"`
	Title                  string `json:"title"`
	LinkedIdea             string `json:"linkedIdea"`
	LinkedProblemStatement string `json:"linkedProblemStatement"`
	Persona                string `json:"persona"`
	Owner                  string `json:"owner"`
	Deadline               string `json:"deadline"`
	LastUpdated            string `json:"lastUpdated"`
	Status                 string `json:"status"`
	IdeaRejected           bool   `json:"ideaRejected"`
	HasFeedback            bool   `json:"hasFeedback"`
	IsOrphan               bool   `json:"isOrphan"`
}

type FeedbackListItem struct {
	ID               string   `json:"id"`
	Title            string   `json:"title"`
	LinkedArtifacts  []string `json:"linkedArtifacts"`
	Outcome          string   `json:"outcome"`
	LinkedTaskOrIdea string   `json:"linkedTaskOrIdea"`
	Owner            string   `json:"owner"`
	CreatedDate      string   `json:"createdDate"`
	HasTaskLink      bool     `json:"hasTaskLink"`
	IsOrphan         bool     `json:"isOrphan"`
}

type StoryListResponse struct {
	Items      []StoryListItem `json:"items"`
	NextCursor *string         `json:"next_cursor,omitempty"`
}

type JourneyListResponse struct {
	Items      []JourneyListItem `json:"items"`
	NextCursor *string           `json:"next_cursor,omitempty"`
}

type ProblemListResponse struct {
	Items      []ProblemListItem `json:"items"`
	NextCursor *string           `json:"next_cursor,omitempty"`
}

type IdeaListResponse struct {
	Items      []IdeaListItem `json:"items"`
	NextCursor *string        `json:"next_cursor,omitempty"`
}

type TaskListResponse struct {
	Items      []TaskListItem `json:"items"`
	NextCursor *string        `json:"next_cursor,omitempty"`
}

type FeedbackListResponse struct {
	Items      []FeedbackListItem `json:"items"`
	NextCursor *string            `json:"next_cursor,omitempty"`
}

type StoryPage struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Status      string `json:"status"`
	Owner       string `json:"owner"`
	LastUpdated string `json:"lastUpdated"`
}

type JourneyPage struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Status      string `json:"status"`
	Owner       string `json:"owner"`
	LastUpdated string `json:"lastUpdated"`
}

type ProblemPage struct {
	ID          string `json:"id"`
	Statement   string `json:"statement"`
	Title       string `json:"title"`
	Status      string `json:"status"`
	Owner       string `json:"owner"`
	LastUpdated string `json:"lastUpdated"`
}

type IdeaPage struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Status      string `json:"status"`
	Owner       string `json:"owner"`
	LastUpdated string `json:"lastUpdated"`
}

type TaskPage struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Status      string `json:"status"`
	Owner       string `json:"owner"`
	Deadline    string `json:"deadline"`
	LastUpdated string `json:"lastUpdated"`
}

type FeedbackPage struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Outcome     string `json:"outcome"`
	Status      string `json:"status"`
	Owner       string `json:"owner"`
	CreatedDate string `json:"createdDate"`
	LastUpdated string `json:"lastUpdated"`
}

type StoryPageResponse struct {
	Story         StoryPage        `json:"story"`
	Metadata      ArtifactMetadata `json:"metadata"`
	Detail        json.RawMessage  `json:"detail"`
	AddOnCatalog  json.RawMessage  `json:"addOnCatalog"`
	AddOnSections json.RawMessage  `json:"addOnSections"`
	Reference     json.RawMessage  `json:"reference"`
}

type JourneyPageResponse struct {
	Journey        JourneyPage      `json:"journey"`
	Metadata       ArtifactMetadata `json:"metadata"`
	Detail         json.RawMessage  `json:"detail"`
	EmotionOptions []string         `json:"emotionOptions"`
	Reference      json.RawMessage  `json:"reference"`
}

type ProblemPageResponse struct {
	Problem   ProblemPage      `json:"problem"`
	Metadata  ArtifactMetadata `json:"metadata"`
	Detail    json.RawMessage  `json:"detail"`
	Reference json.RawMessage  `json:"reference"`
}

type IdeaPageResponse struct {
	Idea      IdeaPage        `json:"idea"`
	Detail    json.RawMessage `json:"detail"`
	Reference json.RawMessage `json:"reference"`
}

type TaskPageResponse struct {
	Task      TaskPage        `json:"task"`
	Detail    json.RawMessage `json:"detail"`
	Reference json.RawMessage `json:"reference"`
}

type FeedbackPageResponse struct {
	Feedback  FeedbackPage     `json:"feedback"`
	Metadata  ArtifactMetadata `json:"metadata"`
	Detail    json.RawMessage  `json:"detail"`
	Reference json.RawMessage  `json:"reference"`
}

type ArtifactStatusResponse struct {
	ID          string `json:"id"`
	Status      string `json:"status"`
	LastUpdated string `json:"lastUpdated"`
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
	items := make([]string, 0)
	for _, item := range toSlice(v) {
		text := toString(item)
		if text != "" {
			items = append(items, text)
		}
	}
	return items
}

func toRawJSON(v any, fallback string) json.RawMessage {
	if v == nil {
		return json.RawMessage(fallback)
	}
	bytes, err := json.Marshal(v)
	if err != nil || len(bytes) == 0 {
		return json.RawMessage(fallback)
	}
	return json.RawMessage(bytes)
}

func decodeArtifactMetadata(raw map[string]any) ArtifactMetadata {
	return ArtifactMetadata{
		Owner:        toString(raw["owner"]),
		CreatedBy:    toString(raw["createdBy"]),
		CreatedAt:    toString(raw["createdAt"]),
		LastEditedBy: toString(raw["lastEditedBy"]),
		LastEditedAt: toString(raw["lastEditedAt"]),
		LastUpdated:  toString(raw["lastUpdated"]),
	}
}

func decodeStoryListItem(raw map[string]any) StoryListItem {
	return StoryListItem{
		ID:                     toString(raw["id"]),
		Title:                  toString(raw["title"]),
		PersonaName:            toString(raw["personaName"]),
		PainPointsCount:        toInt(raw["painPointsCount"]),
		ProblemHypothesesCount: toInt(raw["problemHypothesesCount"]),
		Owner:                  toString(raw["owner"]),
		LastUpdated:            toString(raw["lastUpdated"]),
		Status:                 toString(raw["status"]),
		IsOrphan:               toBool(raw["isOrphan"]),
	}
}

func decodeJourneyListItem(raw map[string]any) JourneyListItem {
	return JourneyListItem{
		ID:              toString(raw["id"]),
		Title:           toString(raw["title"]),
		LinkedPersonas:  toStringSlice(raw["linkedPersonas"]),
		StagesCount:     toInt(raw["stagesCount"]),
		PainPointsCount: toInt(raw["painPointsCount"]),
		Owner:           toString(raw["owner"]),
		LastUpdated:     toString(raw["lastUpdated"]),
		Status:          toString(raw["status"]),
		IsOrphan:        toBool(raw["isOrphan"]),
	}
}

func decodeProblemListItem(raw map[string]any) ProblemListItem {
	return ProblemListItem{
		ID:              toString(raw["id"]),
		Statement:       toString(raw["statement"]),
		LinkedSources:   toStringSlice(raw["linkedSources"]),
		PainPointsCount: toInt(raw["painPointsCount"]),
		IdeasCount:      toInt(raw["ideasCount"]),
		Status:          toString(raw["status"]),
		Owner:           toString(raw["owner"]),
		LastUpdated:     toString(raw["lastUpdated"]),
		IsOrphan:        toBool(raw["isOrphan"]),
	}
}

func decodeIdeaListItem(raw map[string]any) IdeaListItem {
	return IdeaListItem{
		ID:                     toString(raw["id"]),
		Title:                  toString(raw["title"]),
		LinkedProblemStatement: toString(raw["linkedProblemStatement"]),
		Persona:                toString(raw["persona"]),
		Status:                 toString(raw["status"]),
		TasksCount:             toInt(raw["tasksCount"]),
		Owner:                  toString(raw["owner"]),
		LastUpdated:            toString(raw["lastUpdated"]),
		LinkedProblemLocked:    toBool(raw["linkedProblemLocked"]),
		IsOrphan:               toBool(raw["isOrphan"]),
	}
}

func decodeTaskListItem(raw map[string]any) TaskListItem {
	return TaskListItem{
		ID:                     toString(raw["id"]),
		Title:                  toString(raw["title"]),
		LinkedIdea:             toString(raw["linkedIdea"]),
		LinkedProblemStatement: toString(raw["linkedProblemStatement"]),
		Persona:                toString(raw["persona"]),
		Owner:                  toString(raw["owner"]),
		Deadline:               toString(raw["deadline"]),
		LastUpdated:            toString(raw["lastUpdated"]),
		Status:                 toString(raw["status"]),
		IdeaRejected:           toBool(raw["ideaRejected"]),
		HasFeedback:            toBool(raw["hasFeedback"]),
		IsOrphan:               toBool(raw["isOrphan"]),
	}
}

func decodeFeedbackListItem(raw map[string]any) FeedbackListItem {
	return FeedbackListItem{
		ID:               toString(raw["id"]),
		Title:            toString(raw["title"]),
		LinkedArtifacts:  toStringSlice(raw["linkedArtifacts"]),
		Outcome:          toString(raw["outcome"]),
		LinkedTaskOrIdea: toString(raw["linkedTaskOrIdea"]),
		Owner:            toString(raw["owner"]),
		CreatedDate:      toString(raw["createdDate"]),
		HasTaskLink:      toBool(raw["hasTaskLink"]),
		IsOrphan:         toBool(raw["isOrphan"]),
	}
}

type createStoryRequest struct {
	Title string `json:"title"`
}

func (r createStoryRequest) Validate() error {
	if strings.TrimSpace(r.Title) == "" {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "title is required")
	}
	return nil
}

type createJourneyRequest struct {
	Title string `json:"title"`
}

func (r createJourneyRequest) Validate() error {
	if strings.TrimSpace(r.Title) == "" {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "title is required")
	}
	return nil
}

type createProblemRequest struct {
	Statement string `json:"statement"`
}

func (r createProblemRequest) Validate() error {
	if strings.TrimSpace(r.Statement) == "" {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "statement is required")
	}
	return nil
}

type createIdeaRequest struct {
	Title string `json:"title"`
}

func (r createIdeaRequest) Validate() error {
	if strings.TrimSpace(r.Title) == "" {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "title is required")
	}
	return nil
}

type createTaskRequest struct {
	Title string `json:"title"`
}

func (r createTaskRequest) Validate() error {
	if strings.TrimSpace(r.Title) == "" {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "title is required")
	}
	return nil
}

type createFeedbackRequest struct {
	Title string `json:"title"`
}

func (r createFeedbackRequest) Validate() error {
	if strings.TrimSpace(r.Title) == "" {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "title is required")
	}
	return nil
}

type updateStoryRequest struct {
	Story map[string]any `json:"story"`
}

func (r updateStoryRequest) Validate() error {
	if r.Story == nil {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "story payload is required")
	}
	if err := validatePatchPayload(r.Story, storyPatchRules, "story"); err != nil {
		return err
	}
	if status := toString(r.Story["status"]); status != "" {
		canonicalStatus, ok := normalizeAllowedStatus(storyStatuses, status)
		if !ok {
			return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "invalid story status")
		}
		r.Story["status"] = canonicalStatus
	}
	return nil
}

type updateJourneyRequest struct {
	Journey map[string]any `json:"journey"`
}

func (r updateJourneyRequest) Validate() error {
	if r.Journey == nil {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "journey payload is required")
	}
	if err := validatePatchPayload(r.Journey, journeyPatchRules, "journey"); err != nil {
		return err
	}
	if status := toString(r.Journey["status"]); status != "" {
		canonicalStatus, ok := normalizeAllowedStatus(journeyStatuses, status)
		if !ok {
			return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "invalid journey status")
		}
		r.Journey["status"] = canonicalStatus
	}
	return nil
}

type updateProblemRequest struct {
	State map[string]any `json:"state"`
}

func (r updateProblemRequest) Validate() error {
	if r.State == nil {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "state payload is required")
	}
	if err := validatePatchPayload(r.State, problemPatchRules, "problem state"); err != nil {
		return err
	}
	if status := toString(r.State["status"]); status != "" {
		canonicalStatus, ok := normalizeAllowedStatus(problemStatuses, status)
		if !ok {
			return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "invalid problem status")
		}
		r.State["status"] = canonicalStatus
	}
	return nil
}

type updateIdeaRequest struct {
	State map[string]any `json:"state"`
}

func (r updateIdeaRequest) Validate() error {
	if r.State == nil {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "state payload is required")
	}
	if err := validatePatchPayload(r.State, ideaPatchRules, "idea state"); err != nil {
		return err
	}
	if status := toString(r.State["status"]); status != "" {
		canonicalStatus, ok := normalizeAllowedStatus(ideaStatuses, status)
		if !ok {
			return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "invalid idea status")
		}
		r.State["status"] = canonicalStatus
	}
	return nil
}

type updateTaskRequest struct {
	State map[string]any `json:"state"`
}

func (r updateTaskRequest) Validate() error {
	if r.State == nil {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "state payload is required")
	}
	if err := validatePatchPayload(r.State, taskPatchRules, "task state"); err != nil {
		return err
	}
	if status := toString(r.State["status"]); status != "" {
		canonicalStatus, ok := normalizeAllowedStatus(taskStatuses, status)
		if !ok {
			return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "invalid task status")
		}
		r.State["status"] = canonicalStatus
	}
	return nil
}

type updateFeedbackRequest struct {
	State map[string]any `json:"state"`
}

func (r updateFeedbackRequest) Validate() error {
	if r.State == nil {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "state payload is required")
	}
	if err := validatePatchPayload(r.State, feedbackPatchRules, "feedback state"); err != nil {
		return err
	}
	if status := toString(r.State["status"]); status != "" {
		canonicalStatus, ok := normalizeAllowedStatus(feedbackStatuses, status)
		if !ok {
			return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "invalid feedback status")
		}
		r.State["status"] = canonicalStatus
	}
	if outcome := toString(r.State["outcome"]); outcome != "" {
		canonicalOutcome, ok := normalizeAllowedStatus(feedbackOutcomes, outcome)
		if !ok {
			return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "invalid feedback outcome")
		}
		r.State["outcome"] = canonicalOutcome
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

func normalizeArtifactType(raw string) (string, bool) {
	key := strings.ToLower(strings.TrimSpace(raw))
	v, ok := artifactTypeMap[key]
	return v, ok
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

func countStringItems(values []any) int {
	count := 0
	for _, item := range values {
		if strings.TrimSpace(toString(item)) != "" {
			count++
		}
	}
	return count
}

func slugify(raw string) string {
	trimmed := strings.ToLower(strings.TrimSpace(raw))
	if trimmed == "" {
		return "artifact"
	}
	slug := slugSanitizer.ReplaceAllString(trimmed, "-")
	slug = strings.Trim(slug, "-")
	if slug == "" {
		return "artifact"
	}
	if len(slug) > 96 {
		slug = strings.Trim(slug[:96], "-")
		if slug == "" {
			return "artifact"
		}
	}
	return slug
}
