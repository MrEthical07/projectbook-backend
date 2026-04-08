# Plan: Create Comprehensive REST API Blueprint for ProjectBook

## Goal
Create two documents that serve as the definitive API contract between frontend and backend:
1. **`API-GUIDELINES.md`** - Human-readable reference with all routes, payloads, examples, and conventions
2. **`openapi.yaml`** - Machine-readable OpenAPI 3.0 spec with full schema definitions

Both files will be placed in the project root (`d:\Files\League\ProjectBook\Web\`).

---

## Design Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| API style | REST | User selected |
| Auth | JWT Bearer + Session Cookies | Dual support for browsers and API clients |
| Versioning | `/api/v1` prefix | Industry standard, allows future versions |
| Pagination | Offset-based (`?offset=0&limit=25`) | Maps naturally to current array-based data access patterns |
| Date format | ISO 8601 | Already used throughout the codebase |
| Status changes | Separate endpoints (`PUT /:id/status`) | Matches existing codebase pattern with separate permission checks (`statusChange` vs `edit`) |
| Update payloads | Partial updates (only send changed fields) | Matches existing `"field" in state` pattern in remote functions |

---

## Document Structure

### API-GUIDELINES.md (~3500-4000 lines)

```
1. Overview (base URL, versioning, content-type, dates)
2. Authentication
   2.1 Session Cookies (browser flow)
   2.2 JWT Bearer Tokens (API client flow)
   2.3 Session lifecycle (7-day default, 30-day remember-me)
3. Standard Conventions
   3.1 Request format
   3.2 Success response envelope
   3.3 Error response envelope
   3.4 HTTP status codes (200, 201, 204, 400, 401, 403, 404, 409, 429, 500)
   3.5 Pagination (offset/limit with meta)
   3.6 Filtering & sorting query params
   3.7 Rate limiting headers
4. Error Codes Reference (complete table)
5. Permission System
   5.1 Roles (Owner, Admin, Editor, Member, Viewer, Limited Access)
   5.2 Domains (project, story, problem, idea, task, feedback, resource, page, calendar, member)
   5.3 Actions (view, create, edit, delete, archive, statusChange)
   5.4 Default role permission masks (`rolePermissionMasks`)
   5.5 Member-level overrides (`isCustom`, `permissionMask`)
6. API Endpoints (every endpoint with: method, URL, auth, permissions, request body, query params, success response, error responses, examples)
   6.1  Authentication (7 endpoints)
   6.2  Home (12 endpoints)
   6.3  Project (7 endpoints)
   6.4  Team Management (6 endpoints)
   6.5  Stories (4 endpoints)
   6.6  Journeys (4 endpoints)
   6.7  Problem Statements (6 endpoints)
   6.8  Ideas (6 endpoints)
   6.9  Tasks (5 endpoints)
   6.10 Feedback (4 endpoints)
   6.11 Resources (5 endpoints)
   6.12 Pages (5 endpoints)
   6.13 Calendar (5 endpoints)
   6.14 Sidebar Artifacts (3 endpoints)
   6.15 Activity (3 endpoints)
7. Status Transition Rules (with diagrams for Problem, Idea, Task, Story)
8. Data Types Reference (all entity schemas from app.d.ts)
```

### openapi.yaml (~5000-6000 lines)

```
- info, servers, tags, security schemes
- ~80 path operations across all endpoints
- components/schemas: all entity types, enums, request/response bodies
- components/parameters: reusable path/query params
- components/responses: standard error responses
- Extensive use of $ref for deduplication
```

---

## Endpoint Summary (~82 endpoints total)

### Authentication (7)
| Method | Path | Source Function |
|--------|------|----------------|
| POST | /api/v1/auth/signup | `authService.registerUser()` |
| POST | /api/v1/auth/login | `authService.authenticate()` |
| POST | /api/v1/auth/logout | `authService.invalidateSessionByToken()` |
| POST | /api/v1/auth/verify-email | `authService.verifyEmailToken()` |
| POST | /api/v1/auth/resend-verification | `authService.resendVerificationEmail()` |
| POST | /api/v1/auth/forgot-password | `authService.requestPasswordReset()` |
| POST | /api/v1/auth/reset-password | `authService.resetPassword()` |

### Home (13)
| Method | Path | Source Function |
|--------|------|----------------|
| GET | /api/v1/home/dashboard | `getUserDashboard()` |
| GET | /api/v1/home/projects | `getUserProjects()` |
| POST | /api/v1/home/projects | `createProject()` |
| GET | /api/v1/home/projects/reference | `getProjectCreationReference()` |
| GET | /api/v1/home/invites | `getUserInvitesPage()` |
| POST | /api/v1/home/invites/{inviteId}/accept | `acceptProjectInvite()` |
| POST | /api/v1/home/invites/{inviteId}/decline | `declineProjectInvite()` |
| GET | /api/v1/home/notifications | `getUserNotificationsPage()` |
| GET | /api/v1/home/activity | `getUserActivityPage()` |
| GET | /api/v1/home/dashboard-activity | `getUserDashboardActivity()` |
| GET | /api/v1/home/account | `getUserAccountSettings()` |
| PUT | /api/v1/home/account | `updateUserAccountSettings()` |
| GET | /api/v1/home/docs | `getUserDocsSections()` |

### Project (7)
| Method | Path | Source Function |
|--------|------|----------------|
| GET | /api/v1/projects/{projectId}/dashboard | `getProjectDashboard()` |
| GET | /api/v1/projects/{projectId}/access | `getProjectAccess()` |
| GET | /api/v1/projects/{projectId}/sidebar | `getProjectSidebarData()` |
| GET | /api/v1/projects/{projectId}/settings | `getProjectSettings()` |
| PUT | /api/v1/projects/{projectId}/settings | `updateProjectSettings()` |
| POST | /api/v1/projects/{projectId}/archive | `archiveProject()` |
| DELETE | /api/v1/projects/{projectId} | `deleteProject()` |

### Team Management (7)
| Method | Path | Source Function |
|--------|------|----------------|
| GET | /api/v1/projects/{projectId}/team/members | `getProjectTeamMembers()` |
| GET | /api/v1/projects/{projectId}/team/roles | `getProjectTeamRoles()` |
| POST | /api/v1/projects/{projectId}/team/invites | `createProjectInvite()` |
| POST | /api/v1/projects/{projectId}/team/invites/batch | `sendProjectInvites()` |
| DELETE | /api/v1/projects/{projectId}/team/invites/{email} | `cancelProjectInvite()` |
| PUT | /api/v1/projects/{projectId}/team/members/{memberId}/permissions | `updateProjectMemberPermissions()` |
| PUT | /api/v1/projects/{projectId}/team/roles/{role}/permissions | `updateProjectRolePermissions()` |

### Stories (4)
| Method | Path | Source Function |
|--------|------|----------------|
| GET | /api/v1/projects/{projectId}/stories | `getStories()` |
| POST | /api/v1/projects/{projectId}/stories | `createStory()` |
| GET | /api/v1/projects/{projectId}/stories/{slug} | `getStoryPageData()` |
| PUT | /api/v1/projects/{projectId}/stories/{storyId} | `updateStory()` |

### Journeys (4)
| Method | Path | Source Function |
|--------|------|----------------|
| GET | /api/v1/projects/{projectId}/journeys | `getJourneys()` |
| POST | /api/v1/projects/{projectId}/journeys | `createJourney()` |
| GET | /api/v1/projects/{projectId}/journeys/{slug} | `getJourneyPageData()` |
| PUT | /api/v1/projects/{projectId}/journeys/{journeyId} | `updateJourney()` |

### Problem Statements (6)
| Method | Path | Source Function |
|--------|------|----------------|
| GET | /api/v1/projects/{projectId}/problems | `getProblems()` |
| POST | /api/v1/projects/{projectId}/problems | `createProblem()` |
| GET | /api/v1/projects/{projectId}/problems/{slug} | `getProblemPageData()` |
| PUT | /api/v1/projects/{projectId}/problems/{problemId} | `updateProblem()` |
| POST | /api/v1/projects/{projectId}/problems/{problemId}/lock | `lockProblem()` |
| PUT | /api/v1/projects/{projectId}/problems/{problemId}/status | `updateProblemStatus()` |

### Ideas (6)
| Method | Path | Source Function |
|--------|------|----------------|
| GET | /api/v1/projects/{projectId}/ideas | `getIdeas()` |
| POST | /api/v1/projects/{projectId}/ideas | `createIdea()` |
| GET | /api/v1/projects/{projectId}/ideas/{slug} | `getIdeaPageData()` |
| PUT | /api/v1/projects/{projectId}/ideas/{ideaId} | `updateIdea()` |
| POST | /api/v1/projects/{projectId}/ideas/{ideaId}/select | `selectIdea()` |
| PUT | /api/v1/projects/{projectId}/ideas/{ideaId}/status | `updateIdeaStatus()` |

### Tasks (5)
| Method | Path | Source Function |
|--------|------|----------------|
| GET | /api/v1/projects/{projectId}/tasks | `getTasks()` |
| POST | /api/v1/projects/{projectId}/tasks | `createTask()` |
| GET | /api/v1/projects/{projectId}/tasks/{slug} | `getTaskPageData()` |
| PUT | /api/v1/projects/{projectId}/tasks/{taskId} | `updateTask()` |
| PUT | /api/v1/projects/{projectId}/tasks/{taskId}/status | `updateTaskStatus()` |

### Feedback (4)
| Method | Path | Source Function |
|--------|------|----------------|
| GET | /api/v1/projects/{projectId}/feedback | `getFeedback()` |
| POST | /api/v1/projects/{projectId}/feedback | `createFeedback()` |
| GET | /api/v1/projects/{projectId}/feedback/{slug} | `getFeedbackPageData()` |
| PUT | /api/v1/projects/{projectId}/feedback/{feedbackId} | `updateFeedback()` |

### Resources (5)
| Method | Path | Source Function |
|--------|------|----------------|
| GET | /api/v1/projects/{projectId}/resources | `getResources()` |
| POST | /api/v1/projects/{projectId}/resources | `createResource()` |
| GET | /api/v1/projects/{projectId}/resources/{resourceId} | `getResourcePageData()` |
| PUT | /api/v1/projects/{projectId}/resources/{resourceId} | `updateResource()` |
| PUT | /api/v1/projects/{projectId}/resources/{resourceId}/status | `updateResourceStatus()` |

### Pages (5)
| Method | Path | Source Function |
|--------|------|----------------|
| GET | /api/v1/projects/{projectId}/pages | `getPages()` |
| POST | /api/v1/projects/{projectId}/pages | `createPage()` |
| GET | /api/v1/projects/{projectId}/pages/{slug} | `getPageEditorData()` |
| PUT | /api/v1/projects/{projectId}/pages/{pageId} | `updatePageEditor()` |
| PUT | /api/v1/projects/{projectId}/pages/{pageId}/rename | `renamePage()` |

### Calendar (5)
| Method | Path | Source Function |
|--------|------|----------------|
| GET | /api/v1/projects/{projectId}/calendar | `getCalendarData()` |
| POST | /api/v1/projects/{projectId}/calendar | `createCalendarEvent()` |
| GET | /api/v1/projects/{projectId}/calendar/{eventId} | `getCalendarEventData()` |
| PUT | /api/v1/projects/{projectId}/calendar/{eventId} | `updateCalendarEvent()` |
| DELETE | /api/v1/projects/{projectId}/calendar/{eventId} | `deleteCalendarEvent()` |

### Sidebar Artifacts (3)
| Method | Path | Source Function |
|--------|------|----------------|
| POST | /api/v1/projects/{projectId}/sidebar/artifacts | `createSidebarArtifact()` |
| PUT | /api/v1/projects/{projectId}/sidebar/artifacts/{artifactId}/rename | `renameSidebarArtifact()` |
| DELETE | /api/v1/projects/{projectId}/sidebar/artifacts/{artifactId} | `deleteSidebarArtifact()` |

### Activity (3)
| Method | Path | Source Function |
|--------|------|----------------|
| GET | /api/v1/home/activity | `getUserActivity()` |
| GET | /api/v1/home/dashboard-activity | `getUserDashboardActivity()` |
| GET | /api/v1/projects/{projectId}/activity | `getProjectActivity()` |

---

## Implementation Steps

### Step 1: Create `API-GUIDELINES.md` - Foundation sections
Write sections 1-5: Overview, Authentication, Conventions, Error Codes, Permission System.

### Step 2: Create `API-GUIDELINES.md` - Auth & Home endpoints
Write sections 6.1 (Auth - 7 endpoints) and 6.2 (Home - 12 endpoints) with full request/response details.

### Step 3: Create `API-GUIDELINES.md` - Project & Team endpoints
Write sections 6.3 (Project - 7 endpoints) and 6.4 (Team - 6 endpoints).

### Step 4: Create `API-GUIDELINES.md` - Core artifact endpoints
Write sections 6.5-6.8 (Stories, Journeys, Problems, Ideas - 20 endpoints). These are the most complex due to enriched detail views and status transition rules.

### Step 5: Create `API-GUIDELINES.md` - Supporting artifact endpoints
Write sections 6.9-6.12 (Tasks, Feedback, Resources, Pages - 19 endpoints).

### Step 6: Create `API-GUIDELINES.md` - Calendar, Sidebar, Activity + reference sections
Write sections 6.13-6.15 (Calendar, Sidebar, Activity - 11 endpoints), plus sections 7 (Status Transitions) and 8 (Data Types).

### Step 7: Create `openapi.yaml` - Foundation
Write info, servers, tags, security schemes, and all reusable components (schemas for enums, entities, permissions, errors, pagination, request bodies, response wrappers, parameters, standard error responses).

### Step 8: Create `openapi.yaml` - All path operations
Write all ~80 path operations referencing the component schemas, with full request/response definitions.

---

## Critical Source Files

| File | Purpose |
|------|---------|
| `src/app.d.ts` | All TypeScript interfaces and enum types |
| `src/lib/remote/*.remote.ts` (14 files) | All function signatures, Zod schemas, business logic |
| `src/lib/server/auth/service.ts` | Auth operations and error conditions |
| `src/lib/schemas/auth.schema.ts` | Auth validation rules and password policy |
| `src/lib/server/auth/constants.ts` | Session TTLs, rate limit values |
| `src/lib/server/auth/rate-limit.ts` | Rate limiting implementation |
| `src/lib/constants/member-roles.ts` | Role definitions |
| `src/lib/server/data/datastore.ts` | Master data shape |

---

## Verification

1. Cross-reference every remote function in `src/lib/remote/` against the endpoint list to ensure nothing is missed
2. Validate the OpenAPI spec using an online validator (e.g., Swagger Editor)
3. Verify all TypeScript types from `app.d.ts` are represented in the OpenAPI schemas
4. Check that all Zod validation rules from remote functions are captured in request body schemas
5. Confirm status transition rules match the `statusTransitions` maps in the remote files
