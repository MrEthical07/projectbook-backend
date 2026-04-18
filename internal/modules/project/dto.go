package project

import (
	"net/http"
	"strings"

	apperr "github.com/MrEthical07/superapi/internal/core/errors"
)

var allowedProjectStatus = map[string]struct{}{
	"Active":   {},
	"Archived": {},
}

var allowedDeliveryChannel = map[string]struct{}{
	"In-app": {},
	"Email":  {},
}

type updateProjectSettingsRequest struct {
	Settings projectSettingsPatch `json:"settings"`
}

func (r updateProjectSettingsRequest) Validate() error {
	return r.Settings.Validate()
}

type projectSettingsPatch struct {
	ProjectName                 string  `json:"projectName"`
	ProjectDescription          *string `json:"projectDescription"`
	ProjectStatus               string  `json:"projectStatus"`
	WhiteboardsEnabled          *bool   `json:"whiteboardsEnabled"`
	AdvancedDatabasesEnabled    *bool   `json:"advancedDatabasesEnabled"`
	CalendarManualEventsEnabled *bool   `json:"calendarManualEventsEnabled"`
	ResourceVersioningEnabled   *bool   `json:"resourceVersioningEnabled"`
	FeedbackAggregationEnabled  *bool   `json:"feedbackAggregationEnabled"`
	NotifyArtifactCreated       *bool   `json:"notifyArtifactCreated"`
	NotifyArtifactLocked        *bool   `json:"notifyArtifactLocked"`
	NotifyFeedbackAdded         *bool   `json:"notifyFeedbackAdded"`
	NotifyResourceUpdated       *bool   `json:"notifyResourceUpdated"`
	DeliveryChannel             *string `json:"deliveryChannel"`
}

func (r projectSettingsPatch) Validate() error {
	if strings.TrimSpace(r.ProjectName) == "" {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "projectName is required")
	}
	if _, ok := allowedProjectStatus[strings.TrimSpace(r.ProjectStatus)]; !ok {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "projectStatus is invalid")
	}
	if r.DeliveryChannel != nil {
		if _, ok := allowedDeliveryChannel[strings.TrimSpace(*r.DeliveryChannel)]; !ok {
			return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "deliveryChannel is invalid")
		}
	}
	return nil
}

type projectDashboardResponse struct {
	Project     dashboardProject      `json:"project"`
	Me          dashboardUser         `json:"me"`
	Summary     dashboardSummary      `json:"summary"`
	MyTasks     []dashboardTask       `json:"myTasks"`
	MyFeedback  []dashboardFeedback   `json:"myFeedback"`
	Events      []dashboardEvent      `json:"events"`
	Activity    []dashboardActivity   `json:"activity"`
	RecentEdits []dashboardRecentEdit `json:"recentEdits"`
}

type projectDashboardSummaryResponse struct {
	Project dashboardProject `json:"project"`
	Summary dashboardSummary `json:"summary"`
}

type projectDashboardMyWorkResponse struct {
	Me          dashboardUser         `json:"me"`
	MyTasks     []dashboardTask       `json:"myTasks"`
	MyFeedback  []dashboardFeedback   `json:"myFeedback"`
	RecentEdits []dashboardRecentEdit `json:"recentEdits"`
}

type projectDashboardEventsResponse struct {
	Events []dashboardEvent `json:"events"`
}

type projectDashboardActivityResponse struct {
	Activity []dashboardActivity `json:"activity"`
}

type dashboardSummary struct {
	Stories                   int `json:"stories"`
	Journeys                  int `json:"journeys"`
	Problems                  int `json:"problems"`
	Ideas                     int `json:"ideas"`
	Tasks                     int `json:"tasks"`
	Feedback                  int `json:"feedback"`
	OrphanStories             int `json:"orphanStories"`
	OrphanJourneys            int `json:"orphanJourneys"`
	LockedProblems            int `json:"lockedProblems"`
	ProblemsWithoutIdeas      int `json:"problemsWithoutIdeas"`
	SelectedIdeas             int `json:"selectedIdeas"`
	SelectedIdeasWithoutTasks int `json:"selectedIdeasWithoutTasks"`
	OpenTasks                 int `json:"openTasks"`
	OverdueTasks              int `json:"overdueTasks"`
	CompletedTasks            int `json:"completedTasks"`
	BlockedOrAbandonedTasks   int `json:"blockedOrAbandonedTasks"`
	CompletedTasksNoFeedback  int `json:"completedTasksNoFeedback"`
	FeedbackNeedsIteration    int `json:"feedbackNeedsIteration"`
}

type dashboardTask struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Status   string `json:"status"`
	Deadline string `json:"deadline"`
}

type dashboardFeedback struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Outcome string `json:"outcome"`
}

type dashboardProject struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type dashboardUser struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Initials string `json:"initials"`
}

type dashboardEvent struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Type     string `json:"type"`
	StartAt  string `json:"startAt"`
	Creator  string `json:"creator"`
	Initials string `json:"initials"`
}

type dashboardActivity struct {
	ID       string `json:"id"`
	User     string `json:"user"`
	Initials string `json:"initials"`
	Action   string `json:"action"`
	Artifact string `json:"artifact,omitempty"`
	Href     string `json:"href,omitempty"`
	At       string `json:"at"`
}

type dashboardRecentEdit struct {
	ID    string `json:"id"`
	Type  string `json:"type"`
	Title string `json:"title"`
	Href  string `json:"href"`
	At    string `json:"at"`
}

type projectAccessResponse struct {
	User        accessUser       `json:"user"`
	Role        string           `json:"role"`
	Permissions permissionMatrix `json:"permissions"`
}

type accessUser struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type permissionMatrix struct {
	Project  permissionSet `json:"project"`
	Story    permissionSet `json:"story"`
	Problem  permissionSet `json:"problem"`
	Idea     permissionSet `json:"idea"`
	Task     permissionSet `json:"task"`
	Feedback permissionSet `json:"feedback"`
	Resource permissionSet `json:"resource"`
	Page     permissionSet `json:"page"`
	Calendar permissionSet `json:"calendar"`
	Member   permissionSet `json:"member"`
}

type permissionSet struct {
	View         bool `json:"view"`
	Create       bool `json:"create"`
	Edit         bool `json:"edit"`
	Delete       bool `json:"delete"`
	Archive      bool `json:"archive"`
	StatusChange bool `json:"statusChange"`
}

type projectSidebarResponse struct {
	User      accessUser       `json:"user"`
	Projects  []sidebarProject `json:"projects"`
	Artifacts sidebarArtifacts `json:"artifacts"`
}

type sidebarProject struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Icon   string `json:"icon"`
	Status string `json:"status"`
}

type sidebarArtifacts struct {
	Stories  []sidebarArtifact `json:"stories"`
	Journeys []sidebarArtifact `json:"journeys"`
	Problems []sidebarArtifact `json:"problems"`
	Ideas    []sidebarArtifact `json:"ideas"`
	Tasks    []sidebarArtifact `json:"tasks"`
	Feedback []sidebarArtifact `json:"feedback"`
	Pages    []sidebarArtifact `json:"pages"`
}

type sidebarArtifact struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type projectSettingsResponse struct {
	ProjectName                 string `json:"projectName"`
	ProjectDescription          string `json:"projectDescription"`
	ProjectStatus               string `json:"projectStatus"`
	WhiteboardsEnabled          bool   `json:"whiteboardsEnabled"`
	AdvancedDatabasesEnabled    bool   `json:"advancedDatabasesEnabled"`
	CalendarManualEventsEnabled bool   `json:"calendarManualEventsEnabled"`
	ResourceVersioningEnabled   bool   `json:"resourceVersioningEnabled"`
	FeedbackAggregationEnabled  bool   `json:"feedbackAggregationEnabled"`
	NotifyArtifactCreated       bool   `json:"notifyArtifactCreated"`
	NotifyArtifactLocked        bool   `json:"notifyArtifactLocked"`
	NotifyFeedbackAdded         bool   `json:"notifyFeedbackAdded"`
	NotifyResourceUpdated       bool   `json:"notifyResourceUpdated"`
	DeliveryChannel             string `json:"deliveryChannel"`
}

type projectUpdateSettingsResponse struct {
	ProjectID string `json:"projectId"`
}

type projectArchiveResponse struct {
	ProjectID string `json:"projectId"`
	Status    string `json:"status"`
}

type projectDeleteResponse struct {
	ProjectID string `json:"projectId"`
	Status    string `json:"status"`
}
