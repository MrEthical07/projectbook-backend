package calendar

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	apperr "github.com/MrEthical07/superapi/internal/core/errors"
	"github.com/MrEthical07/superapi/internal/core/storage"
	"github.com/jackc/pgx/v5"
)

type Repo interface {
	ListCalendarData(ctx context.Context, projectID string, query listQuery) (map[string]any, error)
	CreateCalendarEvent(ctx context.Context, projectID, actorUserID string, req createCalendarEventRequest) (map[string]any, error)
	GetCalendarEvent(ctx context.Context, projectID, eventID string) (map[string]any, error)
	UpdateCalendarEvent(ctx context.Context, projectID, eventID, actorUserID string, patch map[string]any) (map[string]any, error)
	DeleteCalendarEvent(ctx context.Context, projectID, eventID, actorUserID string) (map[string]any, error)
}

type repo struct {
	store storage.RelationalStore
}

type projectIdentity struct {
	UUID string
	Slug string
}

type calendarRecord struct {
	ID              string
	Title           string
	EventType       string
	Start           string
	End             string
	AllDay          bool
	StartTime       string
	EndTime         string
	Owner           string
	Phase           string
	ArtifactType    string
	Description     string
	Location        string
	EventKind       string
	LinkedArtifacts []any
	Tags            []any
	SourceTitle     string
	CreatedAt       string
	UpdatedAt       string
}

func NewRepo(store storage.RelationalStore) Repo {
	return &repo{store: store}
}

func (r *repo) ListCalendarData(ctx context.Context, projectID string, query listQuery) (map[string]any, error) {
	identity, err := r.resolveProjectIdentity(ctx, projectID)
	if err != nil {
		return nil, err
	}

	rows := make([]map[string]any, 0, query.Limit)
	err = r.store.Execute(ctx, storage.RelationalQueryMany(
		`SELECT
			e.id::text,
			e.title,
			e.event_type::text,
			COALESCE(to_char(e.starts_at, 'YYYY-MM-DD'), ''),
			COALESCE(to_char(e.ends_at, 'YYYY-MM-DD'), ''),
			e.all_day,
			COALESCE(e.start_time, ''),
			COALESCE(e.end_time, ''),
			COALESCE(u.name, ''),
			e.phase::text,
			COALESCE(e.artifact_type::text, 'Manual'),
			COALESCE(e.description, ''),
			COALESCE(e.location, ''),
			COALESCE(e.event_kind, ''),
			COALESCE(e.linked_artifacts, '[]'::jsonb),
			COALESCE(e.tags, '[]'::jsonb),
			COALESCE(e.source_title, ''),
			COALESCE(to_char(e.created_at, 'YYYY-MM-DD'), ''),
			COALESCE(to_char(e.updated_at, 'YYYY-MM-DD'), '')
		 FROM calendar_events e
		 LEFT JOIN users u ON u.id = e.owner_user_id
		 WHERE e.project_id = $1::uuid
		 ORDER BY e.starts_at ASC, e.created_at ASC
		 LIMIT $2`,
		func(row storage.RowScanner) error {
			rec, scanErr := scanCalendarRecord(row)
			if scanErr != nil {
				return scanErr
			}
			rows = append(rows, asCalendarListItem(rec))
			return nil
		},
		identity.UUID,
		query.Limit,
	))
	if err != nil {
		return nil, wrapRepoError("list calendar", err)
	}

	return map[string]any{
		"events": rows,
		"reference": map[string]any{
			"phaseChoices":          defaultPhaseChoices,
			"manualKinds":           defaultManualKinds,
			"linkedArtifactOptions": []any{},
		},
	}, nil
}

func (r *repo) CreateCalendarEvent(ctx context.Context, projectID, actorUserID string, req createCalendarEventRequest) (map[string]any, error) {
	identity, err := r.resolveProjectIdentity(ctx, projectID)
	if err != nil {
		return nil, err
	}
	ownerName, err := r.resolveUserName(ctx, actorUserID)
	if err != nil {
		return nil, err
	}
	allDay := true
	if req.AllDay != nil {
		allDay = *req.AllDay
	}
	startsAt, endsAt, err := buildEventTimestamps(req.Start, req.End, allDay, req.StartTime, req.EndTime)
	if err != nil {
		return nil, err
	}
	linkedJSON, err := encodeArrayJSON(req.LinkedArtifacts)
	if err != nil {
		return nil, err
	}
	tagsJSON, err := encodeArrayJSON(req.Tags)
	if err != nil {
		return nil, err
	}

	var id, title, eventType, start, end, phase, createdAt string
	var outAllDay bool
	var startTime, endTime string
	err = r.store.Execute(ctx, storage.RelationalQueryOne(
		`INSERT INTO calendar_events (
			project_id,
			title,
			description,
			event_type,
			phase,
			artifact_type,
			starts_at,
			ends_at,
			owner_user_id,
			all_day,
			start_time,
			end_time,
			location,
			event_kind,
			linked_artifacts,
			tags,
			source_title
		) VALUES (
			$1::uuid,
			$2,
			NULLIF($3, ''),
			'Manual'::calendar_event_type,
			$4::calendar_phase,
			'Manual'::calendar_artifact_type,
			$5::timestamptz,
			$6::timestamptz,
			$7::uuid,
			$8,
			NULLIF($9, ''),
			NULLIF($10, ''),
			NULLIF($11, ''),
			NULLIF($12, ''),
			$13::jsonb,
			$14::jsonb,
			NULL
		)
		RETURNING id::text, title, event_type::text, COALESCE(to_char(starts_at, 'YYYY-MM-DD'), ''), COALESCE(to_char(ends_at, 'YYYY-MM-DD'), ''), all_day, COALESCE(start_time, ''), COALESCE(end_time, ''), phase::text, COALESCE(to_char(created_at, 'YYYY-MM-DD'), '')`,
		func(row storage.RowScanner) error {
			return row.Scan(&id, &title, &eventType, &start, &end, &outAllDay, &startTime, &endTime, &phase, &createdAt)
		},
		identity.UUID,
		strings.TrimSpace(req.Title),
		strings.TrimSpace(req.Description),
		strings.TrimSpace(req.Phase),
		startsAt,
		endsAt,
		strings.TrimSpace(actorUserID),
		allDay,
		strings.TrimSpace(req.StartTime),
		strings.TrimSpace(req.EndTime),
		strings.TrimSpace(req.Location),
		strings.TrimSpace(req.EventKind),
		linkedJSON,
		tagsJSON,
	))
	if err != nil {
		return nil, wrapRepoError("create calendar event", err)
	}

	if err := r.logActivity(ctx, identity.UUID, actorUserID, id, "created Calendar Event", map[string]any{
		"artifact": title,
		"href":     fmt.Sprintf("/project/%s/calendar/%s", identity.Slug, id),
	}); err != nil {
		return nil, err
	}

	result := map[string]any{
		"id":           id,
		"title":        title,
		"type":         eventType,
		"start":        start,
		"end":          end,
		"allDay":       outAllDay,
		"owner":        ownerName,
		"phase":        phase,
		"artifactType": "Manual",
		"createdAt":    createdAt,
	}
	if !outAllDay {
		result["startTime"] = startTime
		result["endTime"] = endTime
	}
	return result, nil
}

func (r *repo) GetCalendarEvent(ctx context.Context, projectID, eventID string) (map[string]any, error) {
	identity, err := r.resolveProjectIdentity(ctx, projectID)
	if err != nil {
		return nil, err
	}
	rec, err := r.loadCalendarEvent(ctx, identity.UUID, eventID)
	if err != nil {
		return nil, err
	}

	event := map[string]any{
		"id":              rec.ID,
		"title":           rec.Title,
		"type":            rec.EventType,
		"date":            rec.Start,
		"allDay":          rec.AllDay,
		"owner":           rec.Owner,
		"eventKind":       rec.EventKind,
		"description":     rec.Description,
		"location":        rec.Location,
		"linkedArtifacts": rec.LinkedArtifacts,
		"tags":            rec.Tags,
		"createdAt":       rec.CreatedAt,
		"lastEdited":      rec.UpdatedAt,
	}
	if !rec.AllDay {
		event["startTime"] = rec.StartTime
		event["endTime"] = rec.EndTime
	}

	return map[string]any{
		"event": event,
		"reference": map[string]any{
			"phaseChoices":          defaultPhaseChoices,
			"manualKinds":           defaultManualKinds,
			"linkedArtifactOptions": []any{},
			"permissions":           map[string]any{},
		},
	}, nil
}

func (r *repo) UpdateCalendarEvent(ctx context.Context, projectID, eventID, actorUserID string, patch map[string]any) (map[string]any, error) {
	identity, err := r.resolveProjectIdentity(ctx, projectID)
	if err != nil {
		return nil, err
	}
	current, err := r.loadCalendarEvent(ctx, identity.UUID, eventID)
	if err != nil {
		return nil, err
	}
	if strings.EqualFold(current.EventType, "Derived") {
		return nil, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "derived calendar events cannot be edited")
	}
	updated, err := mergeCalendarPatch(current, patch)
	if err != nil {
		return nil, err
	}
	startsAt, endsAt, err := buildEventTimestamps(updated.Start, updated.End, updated.AllDay, updated.StartTime, updated.EndTime)
	if err != nil {
		return nil, err
	}
	linkedJSON, err := encodeArrayJSON(updated.LinkedArtifacts)
	if err != nil {
		return nil, err
	}
	tagsJSON, err := encodeArrayJSON(updated.Tags)
	if err != nil {
		return nil, err
	}

	var id, title, lastEdited string
	err = r.store.Execute(ctx, storage.RelationalQueryOne(
		`UPDATE calendar_events
		 SET title = $3,
		     description = NULLIF($4, ''),
		     phase = $5::calendar_phase,
		     starts_at = $6::timestamptz,
		     ends_at = $7::timestamptz,
		     all_day = $8,
		     start_time = NULLIF($9, ''),
		     end_time = NULLIF($10, ''),
		     location = NULLIF($11, ''),
		     event_kind = NULLIF($12, ''),
		     linked_artifacts = $13::jsonb,
		     tags = $14::jsonb,
		     updated_at = NOW()
		 WHERE project_id = $1::uuid AND id::text = $2
		 RETURNING id::text, title, COALESCE(to_char(updated_at, 'YYYY-MM-DD'), '')`,
		func(row storage.RowScanner) error {
			return row.Scan(&id, &title, &lastEdited)
		},
		identity.UUID,
		strings.TrimSpace(eventID),
		updated.Title,
		updated.Description,
		updated.Phase,
		startsAt,
		endsAt,
		updated.AllDay,
		updated.StartTime,
		updated.EndTime,
		updated.Location,
		updated.EventKind,
		linkedJSON,
		tagsJSON,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.New(apperr.CodeNotFound, http.StatusNotFound, "calendar event not found")
		}
		return nil, wrapRepoError("update calendar event", err)
	}

	if err := r.logActivity(ctx, identity.UUID, actorUserID, id, "updated Calendar Event", map[string]any{
		"artifact": title,
		"href":     fmt.Sprintf("/project/%s/calendar/%s", identity.Slug, id),
	}); err != nil {
		return nil, err
	}

	return map[string]any{"id": id, "title": title, "lastEdited": lastEdited}, nil
}

func (r *repo) DeleteCalendarEvent(ctx context.Context, projectID, eventID, actorUserID string) (map[string]any, error) {
	identity, err := r.resolveProjectIdentity(ctx, projectID)
	if err != nil {
		return nil, err
	}
	current, err := r.loadCalendarEvent(ctx, identity.UUID, eventID)
	if err != nil {
		return nil, err
	}
	if strings.EqualFold(current.EventType, "Derived") {
		return nil, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "derived calendar events cannot be deleted")
	}

	var deletedID string
	err = r.store.Execute(ctx, storage.RelationalQueryOne(
		`DELETE FROM calendar_events WHERE project_id = $1::uuid AND id::text = $2 RETURNING id::text`,
		func(row storage.RowScanner) error { return row.Scan(&deletedID) },
		identity.UUID,
		strings.TrimSpace(eventID),
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.New(apperr.CodeNotFound, http.StatusNotFound, "calendar event not found")
		}
		return nil, wrapRepoError("delete calendar event", err)
	}

	if err := r.logActivity(ctx, identity.UUID, actorUserID, deletedID, "deleted Calendar Event", map[string]any{
		"artifact": current.Title,
		"href":     fmt.Sprintf("/project/%s/calendar/%s", identity.Slug, deletedID),
	}); err != nil {
		return nil, err
	}

	return map[string]any{"eventId": deletedID}, nil
}

func (r *repo) loadCalendarEvent(ctx context.Context, projectUUID, eventID string) (calendarRecord, error) {
	rec := calendarRecord{}
	err := r.store.Execute(ctx, storage.RelationalQueryOne(
		`SELECT
			e.id::text,
			e.title,
			e.event_type::text,
			COALESCE(to_char(e.starts_at, 'YYYY-MM-DD'), ''),
			COALESCE(to_char(e.ends_at, 'YYYY-MM-DD'), ''),
			e.all_day,
			COALESCE(e.start_time, ''),
			COALESCE(e.end_time, ''),
			COALESCE(u.name, ''),
			e.phase::text,
			COALESCE(e.artifact_type::text, 'Manual'),
			COALESCE(e.description, ''),
			COALESCE(e.location, ''),
			COALESCE(e.event_kind, ''),
			COALESCE(e.linked_artifacts, '[]'::jsonb),
			COALESCE(e.tags, '[]'::jsonb),
			COALESCE(e.source_title, ''),
			COALESCE(to_char(e.created_at, 'YYYY-MM-DD'), ''),
			COALESCE(to_char(e.updated_at, 'YYYY-MM-DD'), '')
		 FROM calendar_events e
		 LEFT JOIN users u ON u.id = e.owner_user_id
		 WHERE e.project_id = $1::uuid AND e.id::text = $2
		 LIMIT 1`,
		func(row storage.RowScanner) error {
			var err error
			rec, err = scanCalendarRecord(row)
			return err
		},
		projectUUID,
		strings.TrimSpace(eventID),
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return calendarRecord{}, apperr.New(apperr.CodeNotFound, http.StatusNotFound, "calendar event not found")
		}
		return calendarRecord{}, wrapRepoError("load calendar event", err)
	}
	return rec, nil
}

func scanCalendarRecord(row storage.RowScanner) (calendarRecord, error) {
	var rec calendarRecord
	var linkedRaw []byte
	var tagsRaw []byte
	if err := row.Scan(
		&rec.ID,
		&rec.Title,
		&rec.EventType,
		&rec.Start,
		&rec.End,
		&rec.AllDay,
		&rec.StartTime,
		&rec.EndTime,
		&rec.Owner,
		&rec.Phase,
		&rec.ArtifactType,
		&rec.Description,
		&rec.Location,
		&rec.EventKind,
		&linkedRaw,
		&tagsRaw,
		&rec.SourceTitle,
		&rec.CreatedAt,
		&rec.UpdatedAt,
	); err != nil {
		return calendarRecord{}, err
	}
	rec.LinkedArtifacts = decodeArrayJSON(linkedRaw)
	rec.Tags = decodeArrayJSON(tagsRaw)
	return rec, nil
}

func asCalendarListItem(rec calendarRecord) map[string]any {
	item := map[string]any{
		"id":           rec.ID,
		"title":        rec.Title,
		"type":         rec.EventType,
		"start":        rec.Start,
		"end":          rec.End,
		"allDay":       rec.AllDay,
		"owner":        rec.Owner,
		"phase":        rec.Phase,
		"artifactType": rec.ArtifactType,
		"createdAt":    rec.CreatedAt,
	}
	if rec.SourceTitle != "" {
		item["sourceTitle"] = rec.SourceTitle
	}
	if rec.EventType == "Manual" {
		item["description"] = rec.Description
		item["location"] = rec.Location
		item["eventKind"] = rec.EventKind
		item["linkedArtifacts"] = rec.LinkedArtifacts
		item["tags"] = rec.Tags
		if !rec.AllDay {
			item["startTime"] = rec.StartTime
			item["endTime"] = rec.EndTime
		}
	}
	return item
}

func mergeCalendarPatch(current calendarRecord, patch map[string]any) (calendarRecord, error) {
	out := current
	if title := toString(patch["title"]); title != "" {
		out.Title = title
	}
	if description, ok := patch["description"].(string); ok {
		out.Description = strings.TrimSpace(description)
	}
	if location, ok := patch["location"].(string); ok {
		out.Location = strings.TrimSpace(location)
	}
	if eventKind := toString(patch["eventKind"]); eventKind != "" {
		out.EventKind = eventKind
	}
	if phase := toString(patch["phase"]); phase != "" {
		if !isAllowedPhase(phase) {
			return calendarRecord{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "invalid phase")
		}
		out.Phase = phase
	}
	if start := toString(patch["start"]); start != "" {
		out.Start = start
	}
	if end := toString(patch["end"]); end != "" {
		out.End = end
	}
	if allDay, ok := toBool(patch["allDay"]); ok {
		out.AllDay = allDay
	}
	if startTime, ok := patch["startTime"].(string); ok {
		out.StartTime = strings.TrimSpace(startTime)
	}
	if endTime, ok := patch["endTime"].(string); ok {
		out.EndTime = strings.TrimSpace(endTime)
	}
	if links := toSlice(patch["linkedArtifacts"]); links != nil {
		out.LinkedArtifacts = links
	}
	if tags := toSlice(patch["tags"]); tags != nil {
		out.Tags = tags
	}

	if strings.TrimSpace(out.Title) == "" {
		return calendarRecord{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "title is required")
	}
	if strings.TrimSpace(out.EventKind) == "" {
		return calendarRecord{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "eventKind is required")
	}
	if out.AllDay {
		out.StartTime = ""
		out.EndTime = ""
	} else {
		if !isValidHHMM(out.StartTime) || !isValidHHMM(out.EndTime) {
			return calendarRecord{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "startTime and endTime are required in HH:mm format when allDay is false")
		}
	}
	if _, err := parseISODate(out.Start); err != nil {
		return calendarRecord{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "start must be a valid ISO date")
	}
	if _, err := parseISODate(out.End); err != nil {
		return calendarRecord{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "end must be a valid ISO date")
	}
	if !isAllowedPhase(out.Phase) {
		return calendarRecord{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "invalid phase")
	}
	return out, nil
}

func buildEventTimestamps(startDateRaw, endDateRaw string, allDay bool, startTimeRaw, endTimeRaw string) (time.Time, time.Time, error) {
	startDate, err := parseISODate(startDateRaw)
	if err != nil {
		return time.Time{}, time.Time{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "start must be a valid ISO date")
	}
	endDate, err := parseISODate(endDateRaw)
	if err != nil {
		return time.Time{}, time.Time{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "end must be a valid ISO date")
	}
	if endDate.Before(startDate) {
		return time.Time{}, time.Time{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "end must not be before start")
	}
	if allDay {
		start := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.UTC)
		end := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 0, 0, 0, 0, time.UTC)
		return start, end, nil
	}
	startTime, err := time.Parse("15:04", strings.TrimSpace(startTimeRaw))
	if err != nil {
		return time.Time{}, time.Time{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "startTime must use HH:mm format")
	}
	endTime, err := time.Parse("15:04", strings.TrimSpace(endTimeRaw))
	if err != nil {
		return time.Time{}, time.Time{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "endTime must use HH:mm format")
	}
	start := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), startTime.Hour(), startTime.Minute(), 0, 0, time.UTC)
	end := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), endTime.Hour(), endTime.Minute(), 0, 0, time.UTC)
	if end.Before(start) {
		return time.Time{}, time.Time{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "end must not be before start")
	}
	return start, end, nil
}

func encodeArrayJSON(items []any) (string, error) {
	if items == nil {
		items = []any{}
	}
	bytes, err := json.Marshal(items)
	if err != nil {
		return "", apperr.WithCause(apperr.New(apperr.CodeInternal, http.StatusInternalServerError, "failed to encode array payload"), err)
	}
	return string(bytes), nil
}

func decodeArrayJSON(raw []byte) []any {
	if len(raw) == 0 {
		return []any{}
	}
	items := make([]any, 0)
	if err := json.Unmarshal(raw, &items); err != nil {
		return []any{}
	}
	return items
}

func (r *repo) resolveProjectIdentity(ctx context.Context, projectID string) (projectIdentity, error) {
	if err := r.requireStore(); err != nil {
		return projectIdentity{}, err
	}
	var identity projectIdentity
	err := r.store.Execute(ctx, storage.RelationalQueryOne(
		`SELECT id::text, slug FROM projects WHERE slug = $1 OR id::text = $1 LIMIT 1`,
		func(row storage.RowScanner) error {
			return row.Scan(&identity.UUID, &identity.Slug)
		},
		strings.TrimSpace(projectID),
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return projectIdentity{}, apperr.New(apperr.CodeNotFound, http.StatusNotFound, "project not found")
		}
		return projectIdentity{}, wrapRepoError("resolve project", err)
	}
	return identity, nil
}

func (r *repo) resolveUserName(ctx context.Context, userID string) (string, error) {
	var name string
	err := r.store.Execute(ctx, storage.RelationalQueryOne(
		`SELECT name FROM users WHERE id = $1::uuid LIMIT 1`,
		func(row storage.RowScanner) error { return row.Scan(&name) },
		strings.TrimSpace(userID),
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", apperr.New(apperr.CodeNotFound, http.StatusNotFound, "user not found")
		}
		return "", wrapRepoError("resolve user", err)
	}
	return name, nil
}

func (r *repo) logActivity(ctx context.Context, projectUUID, actorUserID, eventID, action string, payload map[string]any) error {
	bytes, err := json.Marshal(payload)
	if err != nil {
		return apperr.WithCause(apperr.New(apperr.CodeInternal, http.StatusInternalServerError, "failed to encode activity payload"), err)
	}
	if err := r.store.Execute(ctx, storage.RelationalExec(
		`INSERT INTO activity_log (project_id, actor_user_id, artifact_type, artifact_id, action, payload)
		 VALUES ($1::uuid, $2::uuid, 'calendar'::artifact_type, $3::uuid, $4, $5::jsonb)`,
		projectUUID,
		strings.TrimSpace(actorUserID),
		eventID,
		action,
		string(bytes),
	)); err != nil {
		return wrapRepoError("log activity", err)
	}
	return nil
}

func (r *repo) requireStore() error {
	if r == nil || r.store == nil {
		return apperr.New(apperr.CodeDependencyFailure, http.StatusServiceUnavailable, "calendar repository unavailable")
	}
	return nil
}

func wrapRepoError(action string, err error) error {
	if err == nil {
		return nil
	}
	if ae, ok := apperr.AsAppError(err); ok {
		return ae
	}
	return apperr.WithCause(apperr.New(apperr.CodeInternal, http.StatusInternalServerError, "failed to process calendar data"), fmt.Errorf("%s: %w", action, err))
}
