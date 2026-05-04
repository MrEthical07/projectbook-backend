package calendar

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	apperr "github.com/MrEthical07/superapi/internal/core/errors"
)

const (
	defaultListLimit = 20
	maxListLimit     = 50
)

var hhmmRegex = regexp.MustCompile(`^(?:[01]\d|2[0-3]):[0-5]\d$`)

var allowedPhases = map[string]struct{}{
	"None":      {},
	"Empathize": {},
	"Define":    {},
	"Ideate":    {},
	"Prototype": {},
	"Test":      {},
}

var defaultManualKinds = []string{"Workshop", "Review", "Testing Session", "Meeting", "Other"}
var defaultPhaseChoices = []string{"None", "Empathize", "Define", "Ideate", "Prototype", "Test"}

type listQuery struct {
	Offset int
	Limit  int
}

type CalendarReference struct {
	PhaseChoices          []string         `json:"phaseChoices"`
	ManualKinds           []string         `json:"manualKinds"`
	LinkedArtifactOptions []LinkedArtifact `json:"linkedArtifactOptions"`
}

type CalendarListEvent struct {
	ID              string           `json:"id"`
	Title           string           `json:"title"`
	Type            string           `json:"type"`
	Start           string           `json:"start"`
	End             string           `json:"end"`
	AllDay          bool             `json:"allDay"`
	StartTime       string           `json:"startTime,omitempty"`
	EndTime         string           `json:"endTime,omitempty"`
	Owner           string           `json:"owner"`
	Phase           string           `json:"phase"`
	ArtifactType    string           `json:"artifactType"`
	SourceTitle     string           `json:"sourceTitle,omitempty"`
	Description     string           `json:"description,omitempty"`
	Location        string           `json:"location,omitempty"`
	EventKind       string           `json:"eventKind,omitempty"`
	LinkedArtifacts []LinkedArtifact `json:"linkedArtifacts,omitempty"`
	Tags            []string         `json:"tags,omitempty"`
	CreatedAt       string           `json:"createdAt"`
}

type ListCalendarDataResponse struct {
	Items      []CalendarListEvent `json:"items"`
	NextCursor *string             `json:"next_cursor,omitempty"`
	Reference  CalendarReference   `json:"reference"`
}

type CalendarEventDetail struct {
	ID              string           `json:"id"`
	Title           string           `json:"title"`
	Type            string           `json:"type"`
	Date            string           `json:"date"`
	AllDay          bool             `json:"allDay"`
	StartTime       string           `json:"startTime,omitempty"`
	EndTime         string           `json:"endTime,omitempty"`
	Owner           string           `json:"owner"`
	EventKind       string           `json:"eventKind,omitempty"`
	Description     string           `json:"description,omitempty"`
	Location        string           `json:"location,omitempty"`
	LinkedArtifacts []LinkedArtifact `json:"linkedArtifacts"`
	Tags            []string         `json:"tags"`
	CreatedAt       string           `json:"createdAt"`
	LastEdited      string           `json:"lastEdited"`
}

type GetCalendarEventResponse struct {
	Event     CalendarEventDetail `json:"event"`
	Reference CalendarReference   `json:"reference"`
}

type UpdateCalendarEventResponse struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	LastEdited string `json:"lastEdited"`
}

type DeleteCalendarEventResponse struct {
	EventID string `json:"eventId"`
}

type createCalendarEventRequest struct {
	Title           string           `json:"title"`
	Start           string           `json:"start"`
	End             string           `json:"end"`
	AllDay          *bool            `json:"allDay"`
	StartTime       string           `json:"startTime"`
	EndTime         string           `json:"endTime"`
	Owner           string           `json:"owner"`
	Phase           string           `json:"phase"`
	Description     string           `json:"description"`
	Location        string           `json:"location"`
	EventKind       string           `json:"eventKind"`
	LinkedArtifacts []LinkedArtifact `json:"linkedArtifacts"`
	Tags            []string         `json:"tags"`
}

func (r createCalendarEventRequest) Validate() error {
	if strings.TrimSpace(r.Title) == "" {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "title is required")
	}
	startDate, err := parseISODate(r.Start)
	if err != nil {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "start must be a valid ISO date")
	}
	endDate, err := parseISODate(r.End)
	if err != nil {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "end must be a valid ISO date")
	}
	if endDate.Before(startDate) {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "end must not be before start")
	}
	if r.AllDay == nil {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "allDay is required")
	}
	if !*r.AllDay {
		if !isValidHHMM(r.StartTime) || !isValidHHMM(r.EndTime) {
			return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "startTime and endTime are required in HH:mm format when allDay is false")
		}
	}
	if strings.TrimSpace(r.Owner) == "" {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "owner is required")
	}
	if !isAllowedPhase(r.Phase) {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "invalid phase")
	}
	if strings.TrimSpace(r.EventKind) == "" {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "eventKind is required")
	}
	return nil
}

type updateCalendarEventRequest struct {
	State map[string]any `json:"state"`
}

func (r updateCalendarEventRequest) Validate() error {
	if r.State == nil {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "state payload is required")
	}
	if phase := toString(r.State["phase"]); phase != "" && !isAllowedPhase(phase) {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "invalid phase")
	}
	if start := toString(r.State["start"]); start != "" {
		if _, err := parseISODate(start); err != nil {
			return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "start must be a valid ISO date")
		}
	}
	if end := toString(r.State["end"]); end != "" {
		if _, err := parseISODate(end); err != nil {
			return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "end must be a valid ISO date")
		}
	}
	if st := toString(r.State["startTime"]); st != "" && !isValidHHMM(st) {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "startTime must use HH:mm format")
	}
	if et := toString(r.State["endTime"]); et != "" && !isValidHHMM(et) {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "endTime must use HH:mm format")
	}
	return nil
}

func parseLimit(raw string) (int, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return defaultListLimit, nil
	}
	parsed, err := strconv.Atoi(trimmed)
	if err != nil {
		return 0, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "limit must be an integer")
	}
	if parsed < 0 {
		return 0, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "limit must be non-negative")
	}
	if parsed == 0 {
		return defaultListLimit, nil
	}
	if parsed > maxListLimit {
		return maxListLimit, nil
	}
	return parsed, nil
}

func parseISODate(raw string) (time.Time, error) {
	return time.Parse("2006-01-02", strings.TrimSpace(raw))
}

func isValidHHMM(raw string) bool {
	return hhmmRegex.MatchString(strings.TrimSpace(raw))
}

func isAllowedPhase(raw string) bool {
	_, ok := allowedPhases[strings.TrimSpace(raw)]
	return ok
}

func toString(v any) string {
	s, _ := v.(string)
	return strings.TrimSpace(s)
}

func toBool(v any) (bool, bool) {
	b, ok := v.(bool)
	return b, ok
}

func toStringSlice(v any) []string {
	s, ok := v.([]string)
	if ok {
		out := make([]string, 0, len(s))
		for _, item := range s {
			trimmed := strings.TrimSpace(item)
			if trimmed != "" {
				out = append(out, trimmed)
			}
		}
		return out
	}
	raw, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(raw))
	for _, item := range raw {
		if text, ok := item.(string); ok {
			trimmed := strings.TrimSpace(text)
			if trimmed != "" {
				out = append(out, trimmed)
			}
		}
	}
	return out
}
