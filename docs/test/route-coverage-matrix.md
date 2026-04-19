# Route Coverage Matrix

Total unique routes: 51

Scenario packs are defined in docs/test/integration-test-plan.md.

| Module | Method | Path | Scenario Pack | Registration Count |
| --- | --- | --- | --- | --- |
| activity | GET | /api/v1/projects/{projectId}/activity | PROTECTED-READ | 1 |
| artifacts | GET | /api/v1/projects/{projectId}/stories | PROTECTED-READ | 1 |
| artifacts | POST | /api/v1/projects/{projectId}/stories | PROTECTED-WRITE | 1 |
| artifacts | GET | /api/v1/projects/{projectId}/journeys | PROTECTED-READ | 1 |
| artifacts | POST | /api/v1/projects/{projectId}/journeys | PROTECTED-WRITE | 1 |
| artifacts | GET | /api/v1/projects/{projectId}/problems | PROTECTED-READ | 1 |
| artifacts | POST | /api/v1/projects/{projectId}/problems | PROTECTED-WRITE | 1 |
| artifacts | GET | /api/v1/projects/{projectId}/ideas | PROTECTED-READ | 1 |
| artifacts | POST | /api/v1/projects/{projectId}/ideas | PROTECTED-WRITE | 1 |
| artifacts | GET | /api/v1/projects/{projectId}/tasks | PROTECTED-READ | 1 |
| artifacts | POST | /api/v1/projects/{projectId}/tasks | PROTECTED-WRITE | 1 |
| artifacts | GET | /api/v1/projects/{projectId}/feedback | PROTECTED-READ | 1 |
| artifacts | POST | /api/v1/projects/{projectId}/feedback | PROTECTED-WRITE | 1 |
| auth | POST | /api/v1/auth/signup | AUTH-PUBLIC | 1 |
| auth | POST | /api/v1/auth/login | AUTH-PUBLIC | 1 |
| auth | POST | /api/v1/auth/verify-email | AUTH-PUBLIC | 1 |
| auth | POST | /api/v1/auth/resend-verification | AUTH-PUBLIC | 1 |
| auth | POST | /api/v1/auth/forgot-password | AUTH-PUBLIC | 1 |
| auth | POST | /api/v1/auth/reset-password | AUTH-PUBLIC | 1 |
| auth | POST | /api/v1/auth/logout | PROTECTED-WRITE | 1 |
| calendar | GET | /api/v1/projects/{projectId}/calendar | PROTECTED-READ | 1 |
| calendar | POST | /api/v1/projects/{projectId}/calendar | PROTECTED-WRITE | 1 |
| calendar | GET | /api/v1/projects/{projectId}/calendar/{eventId} | PROTECTED-READ | 1 |
| calendar | DELETE | /api/v1/projects/{projectId}/calendar/{eventId} | PROTECTED-WRITE | 1 |
| home | GET | /api/v1/home/dashboard | PROTECTED-READ | 1 |
| home | GET | /api/v1/home/projects | PROTECTED-READ | 1 |
| home | GET | /api/v1/home/projects/reference | PROTECTED-READ | 1 |
| home | GET | /api/v1/home/invites | PROTECTED-READ | 1 |
| home | GET | /api/v1/home/notifications | PROTECTED-READ | 1 |
| home | GET | /api/v1/home/activity | PROTECTED-READ | 1 |
| home | GET | /api/v1/home/dashboard-activity | PROTECTED-READ | 1 |
| home | GET | /api/v1/home/account | PROTECTED-READ | 1 |
| home | GET | /api/v1/home/docs | PROTECTED-READ | 1 |
| home | POST | /api/v1/home/projects | PROTECTED-WRITE | 1 |
| home | POST | /api/v1/home/invites/{inviteId}/accept | PROTECTED-WRITE | 1 |
| home | POST | /api/v1/home/invites/{inviteId}/decline | PROTECTED-WRITE | 1 |
| pages | GET | /api/v1/projects/{projectId}/pages | PROTECTED-READ | 1 |
| pages | POST | /api/v1/projects/{projectId}/pages | PROTECTED-WRITE | 1 |
| project | GET | /api/v1/projects/{projectId}/dashboard | PROTECTED-READ | 1 |
| project | GET | /api/v1/projects/{projectId}/access | PROTECTED-READ | 1 |
| project | GET | /api/v1/projects/{projectId}/settings | PROTECTED-READ | 1 |
| project | POST | /api/v1/projects/{projectId}/archive | PROTECTED-WRITE | 1 |
| project | DELETE | /api/v1/projects/{projectId} | PROTECTED-WRITE | 1 |
| resources | GET | /api/v1/projects/{projectId}/resources | PROTECTED-READ | 1 |
| resources | POST | /api/v1/projects/{projectId}/resources | PROTECTED-WRITE | 1 |
| resources | GET | /api/v1/projects/{projectId}/resources/{resourceId} | PROTECTED-READ | 1 |
| team | GET | /api/v1/projects/{projectId}/team/members | PROTECTED-READ | 1 |
| team | GET | /api/v1/projects/{projectId}/team/roles | PROTECTED-READ | 1 |
| team | POST | /api/v1/projects/{projectId}/team/invites | PROTECTED-WRITE | 1 |
| team | POST | /api/v1/projects/{projectId}/team/invites/batch | PROTECTED-WRITE | 1 |
| team | DELETE | /api/v1/projects/{projectId}/team/invites/{email} | PROTECTED-WRITE | 1 |
