package rbac

// Permission bit constants for the canonical ProjectBook 64-bit RBAC model.
//
// Allocation is append-only and follows docs/ProjectBookDocs/rbac.md.
const (
	PermProjectView uint64 = 1 << iota
	PermProjectCreate
	PermProjectEdit
	PermProjectDelete
	PermProjectArchive
	PermProjectStatusChange

	PermStoryView
	PermStoryCreate
	PermStoryEdit
	PermStoryDelete
	PermStoryArchive
	PermStoryStatusChange

	PermProblemView
	PermProblemCreate
	PermProblemEdit
	PermProblemDelete
	PermProblemArchive
	PermProblemStatusChange

	PermIdeaView
	PermIdeaCreate
	PermIdeaEdit
	PermIdeaDelete
	PermIdeaArchive
	PermIdeaStatusChange

	PermTaskView
	PermTaskCreate
	PermTaskEdit
	PermTaskDelete
	PermTaskArchive
	PermTaskStatusChange

	PermFeedbackView
	PermFeedbackCreate
	PermFeedbackEdit
	PermFeedbackDelete
	PermFeedbackArchive
	PermFeedbackStatusChange

	PermResourceView
	PermResourceCreate
	PermResourceEdit
	PermResourceDelete
	PermResourceArchive
	PermResourceStatusChange

	PermPageView
	PermPageCreate
	PermPageEdit
	PermPageDelete
	PermPageArchive
	PermPageStatusChange

	PermCalendarView
	PermCalendarCreate
	PermCalendarEdit
	PermCalendarDelete
	PermCalendarArchive
	PermCalendarStatusChange

	PermMemberView
	PermMemberCreate
	PermMemberEdit
	PermMemberDelete
	PermMemberArchive
	PermMemberStatusChange
)
