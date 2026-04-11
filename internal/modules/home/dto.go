package home

import (
	"net/http"
	"strings"
	"unicode"

	apperr "github.com/MrEthical07/superapi/internal/core/errors"
)

var allowedProjectIcons = map[string]struct{}{
	"folderKanban":  {},
	"rocket":        {},
	"lightbulb":     {},
	"flaskConical":  {},
	"compass":       {},
	"target":        {},
	"briefcase":     {},
	"layoutGrid":    {},
	"notebookPen":   {},
	"sparkles":      {},
	"code":          {},
	"palette":       {},
	"zap":           {},
	"shieldCheck":   {},
	"chartLine":     {},
	"database":      {},
	"globe":         {},
	"megaphone":     {},
	"users":         {},
	"graduationCap": {},
	"handshake":     {},
	"wrench":        {},
	"cpu":           {},
	"bookOpen":      {},
	"flag":          {},
}

var allowedTheme = map[string]struct{}{
	"Light":  {},
	"Dark":   {},
	"System": {},
}

var allowedDensity = map[string]struct{}{
	"Comfortable": {},
	"Compact":     {},
}

var allowedLanding = map[string]struct{}{
	"Last Project":     {},
	"Project Selector": {},
}

var allowedTimeFormat = map[string]struct{}{
	"12-hour": {},
	"24-hour": {},
}

var allowedActivityTypes = map[string]struct{}{
	"Artifacts": {},
	"Tasks":     {},
	"Feedback":  {},
	"Comments":  {},
}

type createProjectRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Icon        string  `json:"icon"`
}

func (r createProjectRequest) Validate() error {
	name := strings.TrimSpace(r.Name)
	if name == "" {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "name is required")
	}
	if !containsAlphaNum(name) {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "name must contain letters or numbers")
	}

	icon := strings.TrimSpace(r.Icon)
	if icon == "" {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "icon is required")
	}
	if _, ok := allowedProjectIcons[icon]; !ok {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "icon is invalid")
	}

	return nil
}

type updateAccountRequest struct {
	Settings accountSettingsPatch `json:"settings"`
}

func (r updateAccountRequest) Validate() error {
	return r.Settings.Validate()
}

type accountSettingsPatch struct {
	DisplayName        string  `json:"displayName"`
	Bio                *string `json:"bio"`
	Theme              *string `json:"theme"`
	Density            *string `json:"density"`
	Landing            *string `json:"landing"`
	TimeFormat         *string `json:"timeFormat"`
	InAppNotifications *bool   `json:"inAppNotifications"`
	EmailNotifications *bool   `json:"emailNotifications"`
}

func (r accountSettingsPatch) Validate() error {
	if strings.TrimSpace(r.DisplayName) == "" {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "displayName is required")
	}

	if r.Theme != nil {
		if _, ok := allowedTheme[strings.TrimSpace(*r.Theme)]; !ok {
			return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "theme is invalid")
		}
	}

	if r.Density != nil {
		if _, ok := allowedDensity[strings.TrimSpace(*r.Density)]; !ok {
			return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "density is invalid")
		}
	}

	if r.Landing != nil {
		if _, ok := allowedLanding[strings.TrimSpace(*r.Landing)]; !ok {
			return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "landing is invalid")
		}
	}

	if r.TimeFormat != nil {
		if _, ok := allowedTimeFormat[strings.TrimSpace(*r.TimeFormat)]; !ok {
			return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "timeFormat is invalid")
		}
	}

	return nil
}

type homeUser struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type homeProject struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Organization  string `json:"organization"`
	Icon          string `json:"icon"`
	Description   string `json:"description"`
	Role          string `json:"role"`
	OpenTasks     int    `json:"openTasks"`
	LastVisitedAt string `json:"lastVisitedAt,omitempty"`
	LastUpdatedAt string `json:"lastUpdatedAt"`
	Status        string `json:"status"`
}

type homeInvite struct {
	ID                 string `json:"id"`
	ProjectName        string `json:"projectName"`
	ProjectDescription string `json:"projectDescription"`
	ProjectStatus      string `json:"projectStatus"`
	ProjectID          string `json:"projectId"`
	OrganizationName   string `json:"organizationName"`
	InviterName        string `json:"inviterName"`
	InviterRole        string `json:"inviterRole"`
	InviterEmail       string `json:"inviterEmail"`
	AssignedRole       string `json:"assignedRole"`
	SentAt             string `json:"sentAt"`
	ExpiresAt          string `json:"expiresAt"`
	ExpiresSoon        bool   `json:"expiresSoon"`
	Expired            bool   `json:"expired"`
}

type homeNotification struct {
	ID         string `json:"id"`
	Text       string `json:"text"`
	Timestamp  string `json:"timestamp"`
	URL        string `json:"url"`
	Read       bool   `json:"read"`
	SourceType string `json:"sourceType"`
	Dismissed  bool   `json:"dismissed"`
}

type homeActivityItem struct {
	ID           string `json:"id"`
	UserName     string `json:"userName"`
	UserInitials string `json:"userInitials"`
	Action       string `json:"action"`
	ArtifactName string `json:"artifactName,omitempty"`
	ArtifactURL  string `json:"artifactUrl,omitempty"`
	ProjectID    string `json:"projectId"`
	ProjectName  string `json:"projectName"`
	Type         string `json:"type"`
	Timestamp    string `json:"timestamp"`
	OccurredAt   string `json:"occurredAt"`
}

type dashboardActivityItem struct {
	ID           string `json:"id"`
	UserName     string `json:"userName"`
	UserInitials string `json:"userInitials"`
	Action       string `json:"action"`
	ProjectName  string `json:"projectName"`
	Timestamp    string `json:"timestamp"`
	OccurredAt   string `json:"occurredAt"`
	Involved     bool   `json:"involved"`
}

type homeDashboardResponse struct {
	User          homeUser                `json:"user"`
	Projects      []homeProject           `json:"projects"`
	Invites       []homeInvite            `json:"invites"`
	Notifications []homeNotification      `json:"notifications"`
	Activity      []dashboardActivityItem `json:"activity"`
}

type projectCreationResponse struct {
	ProjectID string      `json:"projectId"`
	Project   homeProject `json:"project"`
}

type projectReferenceResponse struct {
	ExistingProjects []string `json:"existingProjects"`
	ExistingUsers    []string `json:"existingUsers"`
}

type inviteAcceptResponse struct {
	InviteID  string `json:"inviteId"`
	ProjectID string `json:"projectId"`
}

type inviteDeclineResponse struct {
	InviteID string `json:"inviteId"`
}

type homeAccountSettingsResponse struct {
	DisplayName        string `json:"displayName"`
	Email              string `json:"email"`
	Bio                string `json:"bio"`
	Theme              string `json:"theme"`
	Density            string `json:"density"`
	Landing            string `json:"landing"`
	TimeFormat         string `json:"timeFormat"`
	InAppNotifications bool   `json:"inAppNotifications"`
	EmailNotifications bool   `json:"emailNotifications"`
}

type updateAccountResponse struct {
	UpdatedAt string `json:"updatedAt"`
}

type docsResponse struct {
	Sections []string `json:"sections"`
}

type activityFilter struct {
	Limit     int
	Type      string
	ProjectID string
}

func (f activityFilter) Validate() error {
	if f.Limit <= 0 {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "limit must be greater than zero")
	}
	if f.Limit > 200 {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "limit must be less than or equal to 200")
	}
	if f.Type != "" {
		if _, ok := allowedActivityTypes[f.Type]; !ok {
			return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "type is invalid")
		}
	}
	return nil
}

func containsAlphaNum(value string) bool {
	for _, ch := range value {
		if unicode.IsLetter(ch) || unicode.IsDigit(ch) {
			return true
		}
	}
	return false
}
