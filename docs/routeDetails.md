# Route Details

## File Purpose
This file is generated from module route registrations and handler contracts. It documents transport contracts, policy chains, RBAC checks, cache behavior, and response examples.

## Generation
- Generator command: go run ./cmd/routedocgen
- Route sources: internal/modules/*/routes.go
- Handler signatures: internal/modules/*/handler.go
- Endpoint IDs: docs/ProjectBookDocs/endpoint-tracker.json
- Output schema fallback: docs/ProjectBookDocs/API-GUIDELINES.md
- Generated at: 2026-04-14T06:09:01Z

## Module Endpoint Counts

| Module | Total Endpoints | Tracked (EP) | Operational (OP) |
| --- | ---: | ---: | ---: |
| activity | 1 | 1 | 0 |
| artifacts | 29 | 29 | 0 |
| auth | 8 | 7 | 1 |
| calendar | 5 | 5 | 0 |
| health | 2 | 0 | 2 |
| home | 13 | 13 | 0 |
| pages | 5 | 5 | 0 |
| project | 7 | 7 | 0 |
| resources | 5 | 5 | 0 |
| sidebar | 3 | 3 | 0 |
| system | 2 | 0 | 2 |
| team | 7 | 7 | 0 |

## Module: activity

Total endpoints: 1

### EP-082 - GET /api/v1/projects/{projectId}/activity

- Status: tested
- Endpoint: ListProjectActivity
- Handler: httpx.Adapter(m.handler.ListProjectActivity)
- Business Logic Source: getProjectActivity()
- Path Params: projectId
- Query Params (inferred): limit

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. ProjectRequired
- Applied Call:
```go
policy.ProjectRequired()
```
3. ProjectMatchFromPath
- Applied Call:
```go
policy.ProjectMatchFromPath("projectId")
```
4. ResolvePermissions
- Applied Call:
```go
policy.ResolvePermissions(resolver)
```
5. RequirePermission
- Applied Call:
```go
policy.RequirePermission(rbac.PermProjectView)
```

#### RBAC Permissions
- rbac.PermProjectView

#### Cache Details
- Auth Status: true
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {},
  "path_params": {
    "projectId": "string"
  },
  "query_params": {
    "limit": "string"
  }
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": [
    {
      "id": "a-1",
      "user": "Alex Morgan",
      "initials": "AM",
      "action": "locked Problem Statement",
      "artifact": "Checkout fields are too dense on mobile",
      "href": "/project/atlas-2026/problem-statement/problem-1",
      "at": "2026-02-09T15:10:00.000Z"
    }
  ]
}
```

## Module: artifacts

Total endpoints: 29

### EP-035 - GET /api/v1/projects/{projectId}/stories

- Status: tested
- Endpoint: ListStories
- Handler: httpx.Adapter(m.handler.ListStories)
- Business Logic Source: getStories()
- Path Params: projectId
- Query Params (inferred): status

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. ProjectRequired
- Applied Call:
```go
policy.ProjectRequired()
```
3. ProjectMatchFromPath
- Applied Call:
```go
policy.ProjectMatchFromPath("projectId")
```
4. ResolvePermissions
- Applied Call:
```go
policy.ResolvePermissions(resolver)
```
5. RequirePermission
- Applied Call:
```go
policy.RequirePermission(rbac.PermStoryView)
```
6. CacheReadOptional
- Applied Call:
```go
policy.CacheReadOptional(cacheMgr, cache.CacheReadConfig{
	TTL:			ttl,
	TagSpecs:		storyReadTags,
	AllowAuthenticated:	true,
	VaryBy: cache.CacheVaryBy{
		ProjectID:	true,
		PathParams:	[]string{"projectId"},
		QueryParams:	[]string{"status", "offset", "limit"},
	},
})
```
7. CacheControlOptional
- Applied Call:
```go
policy.CacheControlOptional(cacheMgr, ttl)
```

#### RBAC Permissions
- rbac.PermStoryView

#### Cache Details
- Auth Status: true
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {},
  "path_params": {
    "projectId": "string"
  },
  "query_params": {
    "status": "string"
  }
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": [
    {
      "id": "streamline-checkout",
      "title": "Streamline checkout for first-time users",
      "personaName": "Avery Patel",
      "painPointsCount": 3,
      "problemHypothesesCount": 2,
      "owner": "Avery Patel",
      "lastUpdated": "2026-02-05",
      "status": "Locked",
      "isOrphan": false
    }
  ]
}
```

### EP-036 - POST /api/v1/projects/{projectId}/stories

- Status: tested
- Endpoint: CreateStory
- Handler: httpx.Adapter(m.handler.CreateStory)
- Business Logic Source: createStory()
- Path Params: projectId
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. ProjectRequired
- Applied Call:
```go
policy.ProjectRequired()
```
3. ProjectMatchFromPath
- Applied Call:
```go
policy.ProjectMatchFromPath("projectId")
```
4. ResolvePermissions
- Applied Call:
```go
policy.ResolvePermissions(resolver)
```
5. RequireJSON
- Applied Call:
```go
policy.RequireJSON()
```
6. RequirePermission
- Applied Call:
```go
policy.RequirePermission(rbac.PermStoryCreate)
```
7. CacheInvalidateOptional
- Applied Call:
```go
policy.CacheInvalidateOptional(cacheMgr, invalidateArtifactTags)
```

#### RBAC Permissions
- rbac.PermStoryCreate

#### Cache Details
- Auth Status: true
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {
    "_type": "createStoryRequest"
  },
  "path_params": {
    "projectId": "string"
  },
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "id": "new-persona-discovery",
    "title": "New persona discovery",
    "personaName": "",
    "painPointsCount": 0,
    "problemHypothesesCount": 0,
    "owner": "Ayush",
    "lastUpdated": "2026-02-14",
    "status": "Draft",
    "isOrphan": true
  }
}
```

### EP-037 - GET /api/v1/projects/{projectId}/stories/{slug}

- Status: tested
- Endpoint: GetStory
- Handler: httpx.Adapter(m.handler.GetStory)
- Business Logic Source: getStoryPageData()
- Path Params: projectId, slug
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. ProjectRequired
- Applied Call:
```go
policy.ProjectRequired()
```
3. ProjectMatchFromPath
- Applied Call:
```go
policy.ProjectMatchFromPath("projectId")
```
4. ResolvePermissions
- Applied Call:
```go
policy.ResolvePermissions(resolver)
```
5. RequirePermission
- Applied Call:
```go
policy.RequirePermission(rbac.PermStoryView)
```
6. CacheReadOptional
- Applied Call:
```go
policy.CacheReadOptional(cacheMgr, cache.CacheReadConfig{
	TTL:			ttl,
	TagSpecs:		storyReadTags,
	AllowAuthenticated:	true,
	VaryBy: cache.CacheVaryBy{
		ProjectID:	true,
		PathParams:	[]string{"projectId", "slug"},
	},
})
```
7. CacheControlOptional
- Applied Call:
```go
policy.CacheControlOptional(cacheMgr, ttl)
```

#### RBAC Permissions
- rbac.PermStoryView

#### Cache Details
- Auth Status: true
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {},
  "path_params": {
    "projectId": "string",
    "slug": "string"
  },
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "story": {
      "id": "streamline-checkout",
      "title": "Streamline checkout for first-time users",
      "status": "Locked",
      "owner": "Avery Patel",
      "lastUpdated": "2026-02-05"
    },
    "detail": {
      "title": "Streamline checkout for first-time users",
      "description": "",
      "status": "draft",
      "persona": {
        "name": "Avery Patel",
        "bio": "",
        "role": "",
        "age": 0,
        "job": "",
        "edu": ""
      },
      "context": "",
      "empathyMap": {
        "says": "",
        "thinks": "",
        "does": "",
        "feels": ""
      },
      "painPoints": ["Point 1"],
      "hypothesis": ["hypothesis 1", "hypothesis 2"],
      "notes": ""
    },
    "addOnCatalog": [
      {
        "type": "goals_success",
        "name": "Goals & Success Criteria",
        "description": "Define what success looks like from the user's perspective.",
        "tag": "Recommended"
      }
    ],
    "addOnSections": [],
    "reference": {
      "permissions": { }
    }
  }
}
```

### EP-038 - PUT /api/v1/projects/{projectId}/stories/{storyId}

- Status: tested
- Endpoint: UpdateStory
- Handler: httpx.Adapter(m.handler.UpdateStory)
- Business Logic Source: updateStory()
- Path Params: projectId, storyId
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. ProjectRequired
- Applied Call:
```go
policy.ProjectRequired()
```
3. ProjectMatchFromPath
- Applied Call:
```go
policy.ProjectMatchFromPath("projectId")
```
4. ResolvePermissions
- Applied Call:
```go
policy.ResolvePermissions(resolver)
```
5. RequireJSON
- Applied Call:
```go
policy.RequireJSON()
```
6. RequirePermission
- Applied Call:
```go
policy.RequirePermission(rbac.PermStoryEdit)
```
7. CacheInvalidateOptional
- Applied Call:
```go
policy.CacheInvalidateOptional(cacheMgr, invalidateArtifactTags)
```

#### RBAC Permissions
- rbac.PermStoryEdit

#### Cache Details
- Auth Status: true
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {
    "_type": "updateStoryRequest"
  },
  "path_params": {
    "projectId": "string",
    "storyId": "string"
  },
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "id": "streamline-checkout",
    "title": "Streamline checkout for first-time users",
    "personaName": "Updated Persona",
    "painPointsCount": 2,
    "problemHypothesesCount": 2,
    "owner": "Avery Patel",
    "lastUpdated": "2026-02-14",
    "status": "Locked",
    "isOrphan": false
  }
}
```

### EP-039 - GET /api/v1/projects/{projectId}/journeys

- Status: tested
- Endpoint: ListJourneys
- Handler: httpx.Adapter(m.handler.ListJourneys)
- Business Logic Source: getJourneys()
- Path Params: projectId
- Query Params (inferred): status

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. ProjectRequired
- Applied Call:
```go
policy.ProjectRequired()
```
3. ProjectMatchFromPath
- Applied Call:
```go
policy.ProjectMatchFromPath("projectId")
```
4. ResolvePermissions
- Applied Call:
```go
policy.ResolvePermissions(resolver)
```
5. RequirePermission
- Applied Call:
```go
policy.RequirePermission(rbac.PermStoryView)
```
6. CacheReadOptional
- Applied Call:
```go
policy.CacheReadOptional(cacheMgr, cache.CacheReadConfig{
	TTL:			ttl,
	TagSpecs:		journeyReadTags,
	AllowAuthenticated:	true,
	VaryBy: cache.CacheVaryBy{
		ProjectID:	true,
		PathParams:	[]string{"projectId"},
		QueryParams:	[]string{"status", "offset", "limit"},
	},
})
```
7. CacheControlOptional
- Applied Call:
```go
policy.CacheControlOptional(cacheMgr, ttl)
```

#### RBAC Permissions
- rbac.PermStoryView

#### Cache Details
- Auth Status: true
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {},
  "path_params": {
    "projectId": "string"
  },
  "query_params": {
    "status": "string"
  }
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": [
    {
      "id": "student-assignment-journey",
      "title": "Student assignment journey",
      "linkedPersonas": ["Avery Patel"],
      "stagesCount": 6,
      "painPointsCount": 3,
      "owner": "Avery Patel",
      "lastUpdated": "2026-02-04",
      "status": "Draft",
      "isOrphan": false
    }
  ]
}
```

### EP-040 - POST /api/v1/projects/{projectId}/journeys

- Status: tested
- Endpoint: CreateJourney
- Handler: httpx.Adapter(m.handler.CreateJourney)
- Business Logic Source: createJourney()
- Path Params: projectId
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. ProjectRequired
- Applied Call:
```go
policy.ProjectRequired()
```
3. ProjectMatchFromPath
- Applied Call:
```go
policy.ProjectMatchFromPath("projectId")
```
4. ResolvePermissions
- Applied Call:
```go
policy.ResolvePermissions(resolver)
```
5. RequireJSON
- Applied Call:
```go
policy.RequireJSON()
```
6. RequirePermission
- Applied Call:
```go
policy.RequirePermission(rbac.PermStoryCreate)
```
7. CacheInvalidateOptional
- Applied Call:
```go
policy.CacheInvalidateOptional(cacheMgr, invalidateArtifactTags)
```

#### RBAC Permissions
- rbac.PermStoryCreate

#### Cache Details
- Auth Status: true
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {
    "_type": "createJourneyRequest"
  },
  "path_params": {
    "projectId": "string"
  },
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "id": "new-onboarding-journey",
    "title": "New onboarding journey",
    "linkedPersonas": [],
    "stagesCount": 0,
    "painPointsCount": 0,
    "owner": "Ayush",
    "lastUpdated": "2026-02-14",
    "status": "Draft",
    "isOrphan": true
  }
}
```

### EP-041 - GET /api/v1/projects/{projectId}/journeys/{slug}

- Status: tested
- Endpoint: GetJourney
- Handler: httpx.Adapter(m.handler.GetJourney)
- Business Logic Source: getJourneyPageData()
- Path Params: projectId, slug
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. ProjectRequired
- Applied Call:
```go
policy.ProjectRequired()
```
3. ProjectMatchFromPath
- Applied Call:
```go
policy.ProjectMatchFromPath("projectId")
```
4. ResolvePermissions
- Applied Call:
```go
policy.ResolvePermissions(resolver)
```
5. RequirePermission
- Applied Call:
```go
policy.RequirePermission(rbac.PermStoryView)
```
6. CacheReadOptional
- Applied Call:
```go
policy.CacheReadOptional(cacheMgr, cache.CacheReadConfig{
	TTL:			ttl,
	TagSpecs:		journeyReadTags,
	AllowAuthenticated:	true,
	VaryBy: cache.CacheVaryBy{
		ProjectID:	true,
		PathParams:	[]string{"projectId", "slug"},
	},
})
```
7. CacheControlOptional
- Applied Call:
```go
policy.CacheControlOptional(cacheMgr, ttl)
```

#### RBAC Permissions
- rbac.PermStoryView

#### Cache Details
- Auth Status: true
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {},
  "path_params": {
    "projectId": "string",
    "slug": "string"
  },
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "journey": {
      "id": "student-assignment-journey",
      "title": "Student assignment journey",
      "status": "Draft",
      "owner": "Avery Patel",
      "lastUpdated": "2026-02-04"
    },
    "detail": {
      "title": "Student assignment journey",
      "description": "",
      "status": "draft",
      "persona": {
        "name": "",
        "bio": "",
        "role": "",
        "age": 0,
        "job": "",
        "edu": ""
      },
      "context": "",
      "stages": [
        {
          "name": "Discovery",
          "actions": ["Sees coffee shop", "Checks line"],
          "emotion": "Neutral",
          "painPoints": []
        },
        {
          "name": "Ordering",
          "actions": ["Waits in line", "Orders drink"],
          "emotion": "Frustrated",
          "painPoints": ["Line is slow"]
        }
      ],
      "notes": ""
    },
    "emotionOptions": ["Neutral", "Frustrated", "Anxious", "Relieved"],
    "reference": {
      "permissions": { }
    }
  }
}
```

### EP-042 - PUT /api/v1/projects/{projectId}/journeys/{journeyId}

- Status: tested
- Endpoint: UpdateJourney
- Handler: httpx.Adapter(m.handler.UpdateJourney)
- Business Logic Source: updateJourney()
- Path Params: journeyId, projectId
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. ProjectRequired
- Applied Call:
```go
policy.ProjectRequired()
```
3. ProjectMatchFromPath
- Applied Call:
```go
policy.ProjectMatchFromPath("projectId")
```
4. ResolvePermissions
- Applied Call:
```go
policy.ResolvePermissions(resolver)
```
5. RequireJSON
- Applied Call:
```go
policy.RequireJSON()
```
6. RequirePermission
- Applied Call:
```go
policy.RequirePermission(rbac.PermStoryEdit)
```
7. CacheInvalidateOptional
- Applied Call:
```go
policy.CacheInvalidateOptional(cacheMgr, invalidateArtifactTags)
```

#### RBAC Permissions
- rbac.PermStoryEdit

#### Cache Details
- Auth Status: true
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {
    "_type": "updateJourneyRequest"
  },
  "path_params": {
    "journeyId": "string",
    "projectId": "string"
  },
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "id": "student-assignment-journey",
    "title": "Student assignment journey",
    "linkedPersonas": ["Avery Patel"],
    "stagesCount": 1,
    "painPointsCount": 0,
    "owner": "Avery Patel",
    "lastUpdated": "2026-02-14",
    "status": "Draft",
    "isOrphan": false
  }
}
```

### EP-043 - GET /api/v1/projects/{projectId}/problems

- Status: tested
- Endpoint: ListProblems
- Handler: httpx.Adapter(m.handler.ListProblems)
- Business Logic Source: getProblems()
- Path Params: projectId
- Query Params (inferred): status

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. ProjectRequired
- Applied Call:
```go
policy.ProjectRequired()
```
3. ProjectMatchFromPath
- Applied Call:
```go
policy.ProjectMatchFromPath("projectId")
```
4. ResolvePermissions
- Applied Call:
```go
policy.ResolvePermissions(resolver)
```
5. RequirePermission
- Applied Call:
```go
policy.RequirePermission(rbac.PermProblemView)
```
6. CacheReadOptional
- Applied Call:
```go
policy.CacheReadOptional(cacheMgr, cache.CacheReadConfig{
	TTL:			ttl,
	TagSpecs:		problemReadTags,
	AllowAuthenticated:	true,
	VaryBy: cache.CacheVaryBy{
		ProjectID:	true,
		PathParams:	[]string{"projectId"},
		QueryParams:	[]string{"status", "offset", "limit"},
	},
})
```
7. CacheControlOptional
- Applied Call:
```go
policy.CacheControlOptional(cacheMgr, ttl)
```

#### RBAC Permissions
- rbac.PermProblemView

#### Cache Details
- Auth Status: true
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {},
  "path_params": {
    "projectId": "string"
  },
  "query_params": {
    "status": "string"
  }
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": [
    {
      "id": "deadline-clarity-students",
      "statement": "Students need a clear way to track assignment deadlines because requirements are fragmented across channels.",
      "linkedSources": [
        "Story: Streamline checkout for first-time users",
        "Journey: Student assignment journey"
      ],
      "painPointsCount": 3,
      "ideasCount": 2,
      "status": "Locked",
      "owner": "Avery Patel",
      "lastUpdated": "2026-02-03",
      "isOrphan": false
    }
  ]
}
```

### EP-044 - POST /api/v1/projects/{projectId}/problems

- Status: tested
- Endpoint: CreateProblem
- Handler: httpx.Adapter(m.handler.CreateProblem)
- Business Logic Source: createProblem()
- Path Params: projectId
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. ProjectRequired
- Applied Call:
```go
policy.ProjectRequired()
```
3. ProjectMatchFromPath
- Applied Call:
```go
policy.ProjectMatchFromPath("projectId")
```
4. ResolvePermissions
- Applied Call:
```go
policy.ResolvePermissions(resolver)
```
5. RequireJSON
- Applied Call:
```go
policy.RequireJSON()
```
6. RequirePermission
- Applied Call:
```go
policy.RequirePermission(rbac.PermProblemCreate)
```
7. CacheInvalidateOptional
- Applied Call:
```go
policy.CacheInvalidateOptional(cacheMgr, invalidateArtifactTags)
```

#### RBAC Permissions
- rbac.PermProblemCreate

#### Cache Details
- Auth Status: true
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {
    "_type": "createProblemRequest"
  },
  "path_params": {
    "projectId": "string"
  },
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "id": "users-struggle-to-find-relevant-resources",
    "statement": "Users struggle to find relevant resources",
    "linkedSources": [],
    "painPointsCount": 0,
    "ideasCount": 0,
    "status": "Draft",
    "owner": "Ayush",
    "lastUpdated": "2026-02-14",
    "isOrphan": true
  }
}
```

### EP-045 - GET /api/v1/projects/{projectId}/problems/{slug}

- Status: tested
- Endpoint: GetProblem
- Handler: httpx.Adapter(m.handler.GetProblem)
- Business Logic Source: getProblemPageData()
- Path Params: projectId, slug
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. ProjectRequired
- Applied Call:
```go
policy.ProjectRequired()
```
3. ProjectMatchFromPath
- Applied Call:
```go
policy.ProjectMatchFromPath("projectId")
```
4. ResolvePermissions
- Applied Call:
```go
policy.ResolvePermissions(resolver)
```
5. RequirePermission
- Applied Call:
```go
policy.RequirePermission(rbac.PermProblemView)
```
6. CacheReadOptional
- Applied Call:
```go
policy.CacheReadOptional(cacheMgr, cache.CacheReadConfig{
	TTL:			ttl,
	TagSpecs:		problemReadTags,
	AllowAuthenticated:	true,
	VaryBy: cache.CacheVaryBy{
		ProjectID:	true,
		PathParams:	[]string{"projectId", "slug"},
	},
})
```
7. CacheControlOptional
- Applied Call:
```go
policy.CacheControlOptional(cacheMgr, ttl)
```

#### RBAC Permissions
- rbac.PermProblemView

#### Cache Details
- Auth Status: true
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {},
  "path_params": {
    "projectId": "string",
    "slug": "string"
  },
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "problem": {
      "id": "deadline-clarity-students",
      "statement": "Students need a clear way to track assignment deadlines...",
      "status": "Locked",
      "owner": "Avery Patel",
      "lastUpdated": "2026-02-03"
    },
    "detail": {
      "title": "",
      "finalStatement": "",
      "selectedPainPoints": [],
      "linkedSources": [],
      "activeModules": [],
      "moduleContent": {},
      "notes": ""
    },
    "reference": {
      "storyOptions": [
        {
          "id": "story-1",
          "title": "Streamline the checkout experience",
          "phase": "Empathize",
          "href": "/project/alpha/stories/streamline-checkout"
        }
      ],
      "journeyOptions": [
        {
          "id": "journey-1",
          "title": "Checkout journey map",
          "phase": "Empathize",
          "href": "/project/alpha/journeys/checkout-journey"
        }
      ],
      "sourcePainPoints": [
        {
          "id": "pain-1",
          "text": "Users abandon checkout when the form asks for repeated information.",
          "sourceLabel": "User Story - Avery Patel"
        }
      ],
      "permissions": { }
    }
  }
}
```

### EP-046 - PUT /api/v1/projects/{projectId}/problems/{problemId}

- Status: tested
- Endpoint: UpdateProblem
- Handler: httpx.Adapter(m.handler.UpdateProblem)
- Business Logic Source: updateProblem()
- Path Params: problemId, projectId
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. ProjectRequired
- Applied Call:
```go
policy.ProjectRequired()
```
3. ProjectMatchFromPath
- Applied Call:
```go
policy.ProjectMatchFromPath("projectId")
```
4. ResolvePermissions
- Applied Call:
```go
policy.ResolvePermissions(resolver)
```
5. RequireJSON
- Applied Call:
```go
policy.RequireJSON()
```
6. RequirePermission
- Applied Call:
```go
policy.RequirePermission(rbac.PermProblemEdit)
```
7. CacheInvalidateOptional
- Applied Call:
```go
policy.CacheInvalidateOptional(cacheMgr, invalidateArtifactTags)
```

#### RBAC Permissions
- rbac.PermProblemEdit

#### Cache Details
- Auth Status: true
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {
    "_type": "updateProblemRequest"
  },
  "path_params": {
    "problemId": "string",
    "projectId": "string"
  },
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "id": "deadline-clarity-students",
    "statement": "Updated problem title",
    "linkedSources": ["Story: ..."],
    "painPointsCount": 2,
    "ideasCount": 2,
    "status": "Draft",
    "owner": "Avery Patel",
    "lastUpdated": "2026-02-14",
    "isOrphan": false
  }
}
```

### EP-047 - POST /api/v1/projects/{projectId}/problems/{problemId}/lock

- Status: tested
- Endpoint: LockProblem
- Handler: httpx.Adapter(m.handler.LockProblem)
- Business Logic Source: lockProblem()
- Path Params: problemId, projectId
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. ProjectRequired
- Applied Call:
```go
policy.ProjectRequired()
```
3. ProjectMatchFromPath
- Applied Call:
```go
policy.ProjectMatchFromPath("projectId")
```
4. ResolvePermissions
- Applied Call:
```go
policy.ResolvePermissions(resolver)
```
5. RequirePermission
- Applied Call:
```go
policy.RequirePermission(rbac.PermProblemStatusChange)
```
6. CacheInvalidateOptional
- Applied Call:
```go
policy.CacheInvalidateOptional(cacheMgr, invalidateArtifactTags)
```

#### RBAC Permissions
- rbac.PermProblemStatusChange

#### Cache Details
- Auth Status: true
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {},
  "path_params": {
    "problemId": "string",
    "projectId": "string"
  },
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "id": "deadline-clarity-students",
    "statement": "Students need a clear way to track...",
    "status": "Locked",
    "lastUpdated": "2026-02-14"
  }
}
```

### EP-048 - PUT /api/v1/projects/{projectId}/problems/{problemId}/status

- Status: tested
- Endpoint: UpdateProblemStatus
- Handler: httpx.Adapter(m.handler.UpdateProblemStatus)
- Business Logic Source: updateProblemStatus()
- Path Params: problemId, projectId
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. ProjectRequired
- Applied Call:
```go
policy.ProjectRequired()
```
3. ProjectMatchFromPath
- Applied Call:
```go
policy.ProjectMatchFromPath("projectId")
```
4. ResolvePermissions
- Applied Call:
```go
policy.ResolvePermissions(resolver)
```
5. RequireJSON
- Applied Call:
```go
policy.RequireJSON()
```
6. RequirePermission
- Applied Call:
```go
policy.RequirePermission(rbac.PermProblemStatusChange)
```
7. CacheInvalidateOptional
- Applied Call:
```go
policy.CacheInvalidateOptional(cacheMgr, invalidateArtifactTags)
```

#### RBAC Permissions
- rbac.PermProblemStatusChange

#### Cache Details
- Auth Status: true
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {
    "_type": "updateProblemStatusRequest"
  },
  "path_params": {
    "problemId": "string",
    "projectId": "string"
  },
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "id": "deadline-clarity-students",
    "status": "Archived",
    "lastUpdated": "2026-02-14"
  }
}
```

### EP-049 - GET /api/v1/projects/{projectId}/ideas

- Status: tested
- Endpoint: ListIdeas
- Handler: httpx.Adapter(m.handler.ListIdeas)
- Business Logic Source: getIdeas()
- Path Params: projectId
- Query Params (inferred): status

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. ProjectRequired
- Applied Call:
```go
policy.ProjectRequired()
```
3. ProjectMatchFromPath
- Applied Call:
```go
policy.ProjectMatchFromPath("projectId")
```
4. ResolvePermissions
- Applied Call:
```go
policy.ResolvePermissions(resolver)
```
5. RequirePermission
- Applied Call:
```go
policy.RequirePermission(rbac.PermIdeaView)
```
6. CacheReadOptional
- Applied Call:
```go
policy.CacheReadOptional(cacheMgr, cache.CacheReadConfig{
	TTL:			ttl,
	TagSpecs:		ideaReadTags,
	AllowAuthenticated:	true,
	VaryBy: cache.CacheVaryBy{
		ProjectID:	true,
		PathParams:	[]string{"projectId"},
		QueryParams:	[]string{"status", "offset", "limit"},
	},
})
```
7. CacheControlOptional
- Applied Call:
```go
policy.CacheControlOptional(cacheMgr, ttl)
```

#### RBAC Permissions
- rbac.PermIdeaView

#### Cache Details
- Auth Status: true
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {},
  "path_params": {
    "projectId": "string"
  },
  "query_params": {
    "status": "string"
  }
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": [
    {
      "id": "deadline-lane-view",
      "title": "Deadline lane view",
      "linkedProblemStatement": "Students need a clear way to track assignment deadlines.",
      "persona": "Avery Patel",
      "status": "Selected",
      "tasksCount": 2,
      "owner": "Avery Patel",
      "lastUpdated": "2026-02-06",
      "linkedProblemLocked": true,
      "isOrphan": false
    }
  ]
}
```

### EP-050 - POST /api/v1/projects/{projectId}/ideas

- Status: tested
- Endpoint: CreateIdea
- Handler: httpx.Adapter(m.handler.CreateIdea)
- Business Logic Source: createIdea()
- Path Params: projectId
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. ProjectRequired
- Applied Call:
```go
policy.ProjectRequired()
```
3. ProjectMatchFromPath
- Applied Call:
```go
policy.ProjectMatchFromPath("projectId")
```
4. ResolvePermissions
- Applied Call:
```go
policy.ResolvePermissions(resolver)
```
5. RequireJSON
- Applied Call:
```go
policy.RequireJSON()
```
6. RequirePermission
- Applied Call:
```go
policy.RequirePermission(rbac.PermIdeaCreate)
```
7. CacheInvalidateOptional
- Applied Call:
```go
policy.CacheInvalidateOptional(cacheMgr, invalidateArtifactTags)
```

#### RBAC Permissions
- rbac.PermIdeaCreate

#### Cache Details
- Auth Status: true
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {
    "_type": "createIdeaRequest"
  },
  "path_params": {
    "projectId": "string"
  },
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "id": "interactive-deadline-calendar",
    "title": "Interactive deadline calendar",
    "linkedProblemStatement": "",
    "persona": "",
    "status": "Considered",
    "tasksCount": 0,
    "owner": "Ayush",
    "lastUpdated": "2026-02-14",
    "linkedProblemLocked": false,
    "isOrphan": true
  }
}
```

### EP-051 - GET /api/v1/projects/{projectId}/ideas/{slug}

- Status: tested
- Endpoint: GetIdea
- Handler: httpx.Adapter(m.handler.GetIdea)
- Business Logic Source: getIdeaPageData()
- Path Params: projectId, slug
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. ProjectRequired
- Applied Call:
```go
policy.ProjectRequired()
```
3. ProjectMatchFromPath
- Applied Call:
```go
policy.ProjectMatchFromPath("projectId")
```
4. ResolvePermissions
- Applied Call:
```go
policy.ResolvePermissions(resolver)
```
5. RequirePermission
- Applied Call:
```go
policy.RequirePermission(rbac.PermIdeaView)
```
6. CacheReadOptional
- Applied Call:
```go
policy.CacheReadOptional(cacheMgr, cache.CacheReadConfig{
	TTL:			ttl,
	TagSpecs:		ideaReadTags,
	AllowAuthenticated:	true,
	VaryBy: cache.CacheVaryBy{
		ProjectID:	true,
		PathParams:	[]string{"projectId", "slug"},
	},
})
```
7. CacheControlOptional
- Applied Call:
```go
policy.CacheControlOptional(cacheMgr, ttl)
```

#### RBAC Permissions
- rbac.PermIdeaView

#### Cache Details
- Auth Status: true
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {},
  "path_params": {
    "projectId": "string",
    "slug": "string"
  },
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "idea": {
      "id": "deadline-lane-view",
      "title": "Deadline lane view",
      "status": "Selected",
      "owner": "Avery Patel",
      "lastUpdated": "2026-02-06"
    },
    "detail": {
      "description": "",
      "status": "Selected",
      "summary": "",
      "notes": "",
      "selectedProblemId": "",
      "activeModules": [],
      "moduleContent": {}
    },
    "reference": {
      "problemOptions": [
        {
          "id": "problem-41",
          "title": "Students miss assignment requirements",
          "phase": "Define",
          "href": "/project/alpha/problem-statement/missed-requirements",
          "status": "Locked"
        }
      ],
      "linkedStories": [
        {
          "id": "story-7",
          "title": "Avery Patel - First-year student",
          "phase": "Empathize",
          "href": "/project/alpha/stories/avery-patel"
        }
      ],
      "derivedPersonas": ["Avery Patel"],
      "permissions": { }
    }
  }
}
```

### EP-052 - PUT /api/v1/projects/{projectId}/ideas/{ideaId}

- Status: tested
- Endpoint: UpdateIdea
- Handler: httpx.Adapter(m.handler.UpdateIdea)
- Business Logic Source: updateIdea()
- Path Params: ideaId, projectId
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. ProjectRequired
- Applied Call:
```go
policy.ProjectRequired()
```
3. ProjectMatchFromPath
- Applied Call:
```go
policy.ProjectMatchFromPath("projectId")
```
4. ResolvePermissions
- Applied Call:
```go
policy.ResolvePermissions(resolver)
```
5. RequireJSON
- Applied Call:
```go
policy.RequireJSON()
```
6. RequirePermission
- Applied Call:
```go
policy.RequirePermission(rbac.PermIdeaEdit)
```
7. CacheInvalidateOptional
- Applied Call:
```go
policy.CacheInvalidateOptional(cacheMgr, invalidateArtifactTags)
```

#### RBAC Permissions
- rbac.PermIdeaEdit

#### Cache Details
- Auth Status: true
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {
    "_type": "updateIdeaRequest"
  },
  "path_params": {
    "ideaId": "string",
    "projectId": "string"
  },
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "id": "deadline-lane-view",
    "title": "Deadline lane view",
    "linkedProblemStatement": "Students miss assignment requirements",
    "persona": "Avery Patel",
    "status": "Selected",
    "tasksCount": 2,
    "owner": "Avery Patel",
    "lastUpdated": "2026-02-14",
    "linkedProblemLocked": true,
    "isOrphan": false
  }
}
```

### EP-053 - POST /api/v1/projects/{projectId}/ideas/{ideaId}/select

- Status: tested
- Endpoint: SelectIdea
- Handler: httpx.Adapter(m.handler.SelectIdea)
- Business Logic Source: selectIdea()
- Path Params: ideaId, projectId
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. ProjectRequired
- Applied Call:
```go
policy.ProjectRequired()
```
3. ProjectMatchFromPath
- Applied Call:
```go
policy.ProjectMatchFromPath("projectId")
```
4. ResolvePermissions
- Applied Call:
```go
policy.ResolvePermissions(resolver)
```
5. RequirePermission
- Applied Call:
```go
policy.RequirePermission(rbac.PermIdeaStatusChange)
```
6. CacheInvalidateOptional
- Applied Call:
```go
policy.CacheInvalidateOptional(cacheMgr, invalidateArtifactTags)
```

#### RBAC Permissions
- rbac.PermIdeaStatusChange

#### Cache Details
- Auth Status: true
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {},
  "path_params": {
    "ideaId": "string",
    "projectId": "string"
  },
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "id": "deadline-lane-view",
    "status": "Selected",
    "lastUpdated": "2026-02-14"
  }
}
```

### EP-054 - PUT /api/v1/projects/{projectId}/ideas/{ideaId}/status

- Status: tested
- Endpoint: UpdateIdeaStatus
- Handler: httpx.Adapter(m.handler.UpdateIdeaStatus)
- Business Logic Source: updateIdeaStatus()
- Path Params: ideaId, projectId
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. ProjectRequired
- Applied Call:
```go
policy.ProjectRequired()
```
3. ProjectMatchFromPath
- Applied Call:
```go
policy.ProjectMatchFromPath("projectId")
```
4. ResolvePermissions
- Applied Call:
```go
policy.ResolvePermissions(resolver)
```
5. RequireJSON
- Applied Call:
```go
policy.RequireJSON()
```
6. RequirePermission
- Applied Call:
```go
policy.RequirePermission(rbac.PermIdeaStatusChange)
```
7. CacheInvalidateOptional
- Applied Call:
```go
policy.CacheInvalidateOptional(cacheMgr, invalidateArtifactTags)
```

#### RBAC Permissions
- rbac.PermIdeaStatusChange

#### Cache Details
- Auth Status: true
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {
    "_type": "updateIdeaStatusRequest"
  },
  "path_params": {
    "ideaId": "string",
    "projectId": "string"
  },
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "id": "deadline-lane-view",
    "status": "Rejected",
    "lastUpdated": "2026-02-14"
  }
}
```

### EP-055 - GET /api/v1/projects/{projectId}/tasks

- Status: tested
- Endpoint: ListTasks
- Handler: httpx.Adapter(m.handler.ListTasks)
- Business Logic Source: getTasks()
- Path Params: projectId
- Query Params (inferred): status

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. ProjectRequired
- Applied Call:
```go
policy.ProjectRequired()
```
3. ProjectMatchFromPath
- Applied Call:
```go
policy.ProjectMatchFromPath("projectId")
```
4. ResolvePermissions
- Applied Call:
```go
policy.ResolvePermissions(resolver)
```
5. RequirePermission
- Applied Call:
```go
policy.RequirePermission(rbac.PermTaskView)
```
6. CacheReadOptional
- Applied Call:
```go
policy.CacheReadOptional(cacheMgr, cache.CacheReadConfig{
	TTL:			ttl,
	TagSpecs:		taskReadTags,
	AllowAuthenticated:	true,
	VaryBy: cache.CacheVaryBy{
		ProjectID:	true,
		PathParams:	[]string{"projectId"},
		QueryParams:	[]string{"status", "offset", "limit"},
	},
})
```
7. CacheControlOptional
- Applied Call:
```go
policy.CacheControlOptional(cacheMgr, ttl)
```

#### RBAC Permissions
- rbac.PermTaskView

#### Cache Details
- Auth Status: true
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {},
  "path_params": {
    "projectId": "string"
  },
  "query_params": {
    "status": "string"
  }
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": [
    {
      "id": "deadline-lane-prototype",
      "title": "Prototype deadline lane interaction",
      "linkedIdea": "Deadline lane view",
      "linkedProblemStatement": "Students need a clear way to track assignment deadlines.",
      "persona": "Avery Patel",
      "owner": "Avery Patel",
      "deadline": "2026-02-09",
      "lastUpdated": "2026-02-10",
      "status": "In Progress",
      "ideaRejected": false,
      "hasFeedback": false,
      "isOrphan": false
    }
  ]
}
```

### EP-056 - POST /api/v1/projects/{projectId}/tasks

- Status: tested
- Endpoint: CreateTask
- Handler: httpx.Adapter(m.handler.CreateTask)
- Business Logic Source: createTask()
- Path Params: projectId
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. ProjectRequired
- Applied Call:
```go
policy.ProjectRequired()
```
3. ProjectMatchFromPath
- Applied Call:
```go
policy.ProjectMatchFromPath("projectId")
```
4. ResolvePermissions
- Applied Call:
```go
policy.ResolvePermissions(resolver)
```
5. RequireJSON
- Applied Call:
```go
policy.RequireJSON()
```
6. RequirePermission
- Applied Call:
```go
policy.RequirePermission(rbac.PermTaskCreate)
```
7. CacheInvalidateOptional
- Applied Call:
```go
policy.CacheInvalidateOptional(cacheMgr, invalidateArtifactTags)
```

#### RBAC Permissions
- rbac.PermTaskCreate

#### Cache Details
- Auth Status: true
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {
    "_type": "createTaskRequest"
  },
  "path_params": {
    "projectId": "string"
  },
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "id": "build-prototype-for-deadline-view",
    "title": "Build prototype for deadline view",
    "linkedIdea": "",
    "linkedProblemStatement": "",
    "persona": "",
    "owner": "Ayush",
    "deadline": "2026-02-14",
    "lastUpdated": "2026-02-14",
    "status": "Planned",
    "ideaRejected": false,
    "hasFeedback": false,
    "isOrphan": true
  }
}
```

### EP-057 - GET /api/v1/projects/{projectId}/tasks/{slug}

- Status: tested
- Endpoint: GetTask
- Handler: httpx.Adapter(m.handler.GetTask)
- Business Logic Source: getTaskPageData()
- Path Params: projectId, slug
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. ProjectRequired
- Applied Call:
```go
policy.ProjectRequired()
```
3. ProjectMatchFromPath
- Applied Call:
```go
policy.ProjectMatchFromPath("projectId")
```
4. ResolvePermissions
- Applied Call:
```go
policy.ResolvePermissions(resolver)
```
5. RequirePermission
- Applied Call:
```go
policy.RequirePermission(rbac.PermTaskView)
```
6. CacheReadOptional
- Applied Call:
```go
policy.CacheReadOptional(cacheMgr, cache.CacheReadConfig{
	TTL:			ttl,
	TagSpecs:		taskReadTags,
	AllowAuthenticated:	true,
	VaryBy: cache.CacheVaryBy{
		ProjectID:	true,
		PathParams:	[]string{"projectId", "slug"},
	},
})
```
7. CacheControlOptional
- Applied Call:
```go
policy.CacheControlOptional(cacheMgr, ttl)
```

#### RBAC Permissions
- rbac.PermTaskView

#### Cache Details
- Auth Status: true
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {},
  "path_params": {
    "projectId": "string",
    "slug": "string"
  },
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "task": {
      "id": "deadline-lane-prototype",
      "title": "Prototype deadline lane interaction",
      "status": "In Progress",
      "owner": "Avery Patel",
      "deadline": "2026-02-09"
    },
    "detail": {
      "assignedToId": "",
      "selectedIdeaId": "",
      "deadline": "2026-02-09",
      "hypothesis": "",
      "planItems": [],
      "executionLinks": [],
      "notes": "",
      "activeModules": [],
      "abandonReason": ""
    },
    "reference": {
      "assigneeOptions": [
        { "id": "user-1", "name": "Nia Clark", "role": "Designer" },
        { "id": "user-2", "name": "Dr. Ramos", "role": "Product" }
      ],
      "ideaOptions": [
        {
          "id": "idea-31",
          "title": "Visual deadline timeline for assignments",
          "phase": "Ideate",
          "href": "/project/alpha/ideas/deadline-timeline",
          "status": "Active",
          "problem": {
            "id": "problem-7",
            "title": "Students miss assignment requirements",
            "phase": "Define",
            "href": "/project/alpha/problem-statement/missed-requirements",
            "status": "Locked"
          },
          "context": {
            "type": "Persona",
            "title": "Nia Clark",
            "detail": "First-year student balancing coursework and a part-time job.",
            "phase": "Empathize",
            "href": "/project/alpha/personas/nia-clark",
            "status": "Active"
          }
        }
      ],
      "permissions": { }
    }
  }
}
```

### EP-058 - PUT /api/v1/projects/{projectId}/tasks/{taskId}

- Status: tested
- Endpoint: UpdateTask
- Handler: httpx.Adapter(m.handler.UpdateTask)
- Business Logic Source: updateTask()
- Path Params: projectId, taskId
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. ProjectRequired
- Applied Call:
```go
policy.ProjectRequired()
```
3. ProjectMatchFromPath
- Applied Call:
```go
policy.ProjectMatchFromPath("projectId")
```
4. ResolvePermissions
- Applied Call:
```go
policy.ResolvePermissions(resolver)
```
5. RequireJSON
- Applied Call:
```go
policy.RequireJSON()
```
6. RequirePermission
- Applied Call:
```go
policy.RequirePermission(rbac.PermTaskEdit)
```
7. CacheInvalidateOptional
- Applied Call:
```go
policy.CacheInvalidateOptional(cacheMgr, invalidateArtifactTags)
```

#### RBAC Permissions
- rbac.PermTaskEdit

#### Cache Details
- Auth Status: true
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {
    "_type": "updateTaskRequest"
  },
  "path_params": {
    "projectId": "string",
    "taskId": "string"
  },
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "id": "deadline-lane-prototype",
    "title": "Prototype deadline lane interaction",
    "linkedIdea": "Visual deadline timeline for assignments",
    "linkedProblemStatement": "Students miss assignment requirements",
    "persona": "Nia Clark",
    "owner": "Avery Patel",
    "deadline": "2026-02-20",
    "status": "In Progress",
    "ideaRejected": false,
    "hasFeedback": false,
    "isOrphan": false
  }
}
```

### EP-059 - PUT /api/v1/projects/{projectId}/tasks/{taskId}/status

- Status: tested
- Endpoint: UpdateTaskStatus
- Handler: httpx.Adapter(m.handler.UpdateTaskStatus)
- Business Logic Source: updateTaskStatus()
- Path Params: projectId, taskId
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. ProjectRequired
- Applied Call:
```go
policy.ProjectRequired()
```
3. ProjectMatchFromPath
- Applied Call:
```go
policy.ProjectMatchFromPath("projectId")
```
4. ResolvePermissions
- Applied Call:
```go
policy.ResolvePermissions(resolver)
```
5. RequireJSON
- Applied Call:
```go
policy.RequireJSON()
```
6. RequirePermission
- Applied Call:
```go
policy.RequirePermission(rbac.PermTaskStatusChange)
```
7. CacheInvalidateOptional
- Applied Call:
```go
policy.CacheInvalidateOptional(cacheMgr, invalidateArtifactTags)
```

#### RBAC Permissions
- rbac.PermTaskStatusChange

#### Cache Details
- Auth Status: true
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {
    "_type": "updateTaskStatusRequest"
  },
  "path_params": {
    "projectId": "string",
    "taskId": "string"
  },
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "id": "deadline-lane-prototype",
    "status": "Completed",
    "lastUpdated": "2026-02-14"
  }
}
```

### EP-060 - GET /api/v1/projects/{projectId}/feedback

- Status: tested
- Endpoint: ListFeedback
- Handler: httpx.Adapter(m.handler.ListFeedback)
- Business Logic Source: getFeedback()
- Path Params: projectId
- Query Params (inferred): outcome

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. ProjectRequired
- Applied Call:
```go
policy.ProjectRequired()
```
3. ProjectMatchFromPath
- Applied Call:
```go
policy.ProjectMatchFromPath("projectId")
```
4. ResolvePermissions
- Applied Call:
```go
policy.ResolvePermissions(resolver)
```
5. RequirePermission
- Applied Call:
```go
policy.RequirePermission(rbac.PermFeedbackView)
```
6. CacheReadOptional
- Applied Call:
```go
policy.CacheReadOptional(cacheMgr, cache.CacheReadConfig{
	TTL:			ttl,
	TagSpecs:		feedbackReadTags,
	AllowAuthenticated:	true,
	VaryBy: cache.CacheVaryBy{
		ProjectID:	true,
		PathParams:	[]string{"projectId"},
		QueryParams:	[]string{"outcome", "offset", "limit"},
	},
})
```
7. CacheControlOptional
- Applied Call:
```go
policy.CacheControlOptional(cacheMgr, ttl)
```

#### RBAC Permissions
- rbac.PermFeedbackView

#### Cache Details
- Auth Status: true
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {},
  "path_params": {
    "projectId": "string"
  },
  "query_params": {
    "outcome": "string"
  }
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": [
    {
      "id": "deadline-lane-session-1",
      "title": "Deadline lane usability session",
      "linkedArtifacts": ["Task: Deadline lane view", "Idea: Deadline lane view"],
      "outcome": "Validated",
      "linkedTaskOrIdea": "Task: Deadline lane view",
      "owner": "Avery Patel",
      "createdDate": "2026-02-06",
      "hasTaskLink": true,
      "isOrphan": false
    }
  ]
}
```

### EP-061 - POST /api/v1/projects/{projectId}/feedback

- Status: tested
- Endpoint: CreateFeedback
- Handler: httpx.Adapter(m.handler.CreateFeedback)
- Business Logic Source: createFeedback()
- Path Params: projectId
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. ProjectRequired
- Applied Call:
```go
policy.ProjectRequired()
```
3. ProjectMatchFromPath
- Applied Call:
```go
policy.ProjectMatchFromPath("projectId")
```
4. ResolvePermissions
- Applied Call:
```go
policy.ResolvePermissions(resolver)
```
5. RequireJSON
- Applied Call:
```go
policy.RequireJSON()
```
6. RequirePermission
- Applied Call:
```go
policy.RequirePermission(rbac.PermFeedbackCreate)
```
7. CacheInvalidateOptional
- Applied Call:
```go
policy.CacheInvalidateOptional(cacheMgr, invalidateArtifactTags)
```

#### RBAC Permissions
- rbac.PermFeedbackCreate

#### Cache Details
- Auth Status: true
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {
    "_type": "createFeedbackRequest"
  },
  "path_params": {
    "projectId": "string"
  },
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "id": "usability-test-round-1",
    "title": "Usability test round 1",
    "linkedArtifacts": [],
    "outcome": "Needs Iteration",
    "linkedTaskOrIdea": "",
    "owner": "Ayush",
    "createdDate": "2026-02-14",
    "hasTaskLink": false,
    "isOrphan": true
  }
}
```

### EP-062 - GET /api/v1/projects/{projectId}/feedback/{slug}

- Status: tested
- Endpoint: GetFeedback
- Handler: httpx.Adapter(m.handler.GetFeedback)
- Business Logic Source: getFeedbackPageData()
- Path Params: projectId, slug
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. ProjectRequired
- Applied Call:
```go
policy.ProjectRequired()
```
3. ProjectMatchFromPath
- Applied Call:
```go
policy.ProjectMatchFromPath("projectId")
```
4. ResolvePermissions
- Applied Call:
```go
policy.ResolvePermissions(resolver)
```
5. RequirePermission
- Applied Call:
```go
policy.RequirePermission(rbac.PermFeedbackView)
```
6. CacheReadOptional
- Applied Call:
```go
policy.CacheReadOptional(cacheMgr, cache.CacheReadConfig{
	TTL:			ttl,
	TagSpecs:		feedbackReadTags,
	AllowAuthenticated:	true,
	VaryBy: cache.CacheVaryBy{
		ProjectID:	true,
		PathParams:	[]string{"projectId", "slug"},
	},
})
```
7. CacheControlOptional
- Applied Call:
```go
policy.CacheControlOptional(cacheMgr, ttl)
```

#### RBAC Permissions
- rbac.PermFeedbackView

#### Cache Details
- Auth Status: true
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {},
  "path_params": {
    "projectId": "string",
    "slug": "string"
  },
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "feedback": {
      "id": "deadline-lane-session-1",
      "title": "Deadline lane usability session",
      "outcome": "Validated",
      "owner": "Avery Patel",
      "createdDate": "2026-02-06"
    },
    "detail": {
      "description": "",
      "outcome": "Validated",
      "linkedArtifacts": [],
      "activeModules": [],
      "moduleContent": {},
      "notes": ""
    },
    "reference": {
      "taskOptions": [
        {
          "id": "task-21",
          "title": "Prototype timeline for assignment deadlines",
          "type": "Task",
          "phase": "Prototype",
          "href": "/project/alpha/tasks/deadline-timeline",
          "status": "Active"
        }
      ],
      "ideaOptions": [
        {
          "id": "idea-31",
          "title": "Visual deadline timeline for assignments",
          "type": "Idea",
          "phase": "Ideate",
          "href": "/project/alpha/ideas/deadline-timeline",
          "status": "Active"
        }
      ],
      "problemOptions": [
        {
          "id": "problem-7",
          "title": "Students miss assignment requirements",
          "type": "Problem Statement",
          "phase": "Define",
          "href": "/project/alpha/problem-statement/missed-requirements",
          "status": "Archived"
        }
      ],
      "permissions": { }
    }
  }
}
```

### EP-063 - PUT /api/v1/projects/{projectId}/feedback/{feedbackId}

- Status: tested
- Endpoint: UpdateFeedback
- Handler: httpx.Adapter(m.handler.UpdateFeedback)
- Business Logic Source: updateFeedback()
- Path Params: feedbackId, projectId
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. ProjectRequired
- Applied Call:
```go
policy.ProjectRequired()
```
3. ProjectMatchFromPath
- Applied Call:
```go
policy.ProjectMatchFromPath("projectId")
```
4. ResolvePermissions
- Applied Call:
```go
policy.ResolvePermissions(resolver)
```
5. RequireJSON
- Applied Call:
```go
policy.RequireJSON()
```
6. RequirePermission
- Applied Call:
```go
policy.RequirePermission(rbac.PermFeedbackEdit)
```
7. CacheInvalidateOptional
- Applied Call:
```go
policy.CacheInvalidateOptional(cacheMgr, invalidateArtifactTags)
```

#### RBAC Permissions
- rbac.PermFeedbackEdit

#### Cache Details
- Auth Status: true
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {
    "_type": "updateFeedbackRequest"
  },
  "path_params": {
    "feedbackId": "string",
    "projectId": "string"
  },
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "id": "deadline-lane-session-1",
    "title": "Deadline lane usability session",
    "linkedArtifacts": ["Task: Prototype timeline"],
    "outcome": "Validated",
    "linkedTaskOrIdea": "Task: Prototype timeline",
    "owner": "Avery Patel",
    "createdDate": "2026-02-06",
    "hasTaskLink": true,
    "isOrphan": false
  }
}
```

## Module: auth

Total endpoints: 8

### EP-001 - POST /api/v1/auth/signup

- Status: tested
- Endpoint: Signup
- Handler: httpx.Adapter(m.handler.Signup)
- Business Logic Source: authService.registerUser()
- Path Params: none
- Query Params (inferred): none

#### Policies
1. RequireJSON
- Applied Call:
```go
policy.RequireJSON()
```
2. RateLimitWithKeyer
- Applied Call:
```go
policy.RateLimitWithKeyer(limiter, "auth.signup", signupRule, ratelimit.KeyByIP())
```

#### RBAC Permissions
- none

#### Cache Details
- Auth Status: false
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {
    "_type": "signupRequest"
  },
  "path_params": {},
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "user": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Ayush Kumar",
      "email": "ayush@example.com",
      "isEmailVerified": false
    }
  }
}
```

### EP-002 - POST /api/v1/auth/login

- Status: tested
- Endpoint: Login
- Handler: httpx.Adapter(m.handler.Login)
- Business Logic Source: authService.authenticate()
- Path Params: none
- Query Params (inferred): none

#### Policies
1. RequireJSON
- Applied Call:
```go
policy.RequireJSON()
```
2. RateLimitWithKeyer
- Applied Call:
```go
policy.RateLimitWithKeyer(limiter, "auth.login", loginRule, ratelimit.KeyByIP())
```

#### RBAC Permissions
- none

#### Cache Details
- Auth Status: false
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {
    "_type": "loginRequest"
  },
  "path_params": {},
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "user": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Ayush Kumar",
      "email": "ayush@example.com"
    },
    "session": {
      "token": "aB3c...secure-token...",
      "expiresAt": "2026-02-21T09:30:00.000Z"
    }
  }
}
```

### EP-004 - POST /api/v1/auth/verify-email

- Status: tested
- Endpoint: VerifyEmail
- Handler: httpx.Adapter(m.handler.VerifyEmail)
- Business Logic Source: authService.verifyEmailToken()
- Path Params: none
- Query Params (inferred): none

#### Policies
1. RequireJSON
- Applied Call:
```go
policy.RequireJSON()
```
2. RateLimitWithKeyer
- Applied Call:
```go
policy.RateLimitWithKeyer(limiter, "auth.verify_email", verifyRule, ratelimit.KeyByIP())
```

#### RBAC Permissions
- none

#### Cache Details
- Auth Status: false
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {
    "_type": "verifyEmailRequest"
  },
  "path_params": {},
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "status": "success",
    "email": "ayush@example.com"
  }
}
```

### EP-005 - POST /api/v1/auth/resend-verification

- Status: tested
- Endpoint: ResendVerification
- Handler: httpx.Adapter(m.handler.ResendVerification)
- Business Logic Source: authService.resendVerificationEmail()
- Path Params: none
- Query Params (inferred): none

#### Policies
1. RequireJSON
- Applied Call:
```go
policy.RequireJSON()
```
2. RateLimitWithKeyer
- Applied Call:
```go
policy.RateLimitWithKeyer(limiter, "auth.resend_verification", verifyRule, ratelimit.KeyByIP())
```

#### RBAC Permissions
- none

#### Cache Details
- Auth Status: false
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {
    "_type": "resendVerificationRequest"
  },
  "path_params": {},
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "status": "sent"
  }
}
```

### EP-006 - POST /api/v1/auth/forgot-password

- Status: tested
- Endpoint: ForgotPassword
- Handler: httpx.Adapter(m.handler.ForgotPassword)
- Business Logic Source: authService.requestPasswordReset()
- Path Params: none
- Query Params (inferred): none

#### Policies
1. RequireJSON
- Applied Call:
```go
policy.RequireJSON()
```
2. RateLimitWithKeyer
- Applied Call:
```go
policy.RateLimitWithKeyer(limiter, "auth.forgot_password", passwordRule, ratelimit.KeyByIP())
```

#### RBAC Permissions
- none

#### Cache Details
- Auth Status: false
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {
    "_type": "forgotPasswordRequest"
  },
  "path_params": {},
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "message": "If an account exists for this email, a reset link has been sent."
  }
}
```

### EP-007 - POST /api/v1/auth/reset-password

- Status: tested
- Endpoint: ResetPassword
- Handler: httpx.Adapter(m.handler.ResetPassword)
- Business Logic Source: authService.resetPassword()
- Path Params: none
- Query Params (inferred): none

#### Policies
1. RequireJSON
- Applied Call:
```go
policy.RequireJSON()
```
2. RateLimitWithKeyer
- Applied Call:
```go
policy.RateLimitWithKeyer(limiter, "auth.reset_password", passwordRule, ratelimit.KeyByIP())
```

#### RBAC Permissions
- none

#### Cache Details
- Auth Status: false
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {
    "_type": "resetPasswordRequest"
  },
  "path_params": {},
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "message": "Password has been reset successfully."
  }
}
```

### OP-001 - POST /api/v1/auth/refresh

- Status: operational
- Endpoint: Refresh
- Handler: httpx.Adapter(m.handler.Refresh)
- Business Logic Source: n/a
- Path Params: none
- Query Params (inferred): none

#### Policies
1. RequireJSON
- Applied Call:
```go
policy.RequireJSON()
```
2. RateLimitWithKeyer
- Applied Call:
```go
policy.RateLimitWithKeyer(limiter, "auth.refresh", refreshRule, ratelimit.KeyByIP())
```

#### RBAC Permissions
- none

#### Cache Details
- Auth Status: false
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {
    "_type": "refreshRequest"
  },
  "path_params": {},
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "data": {
    "_type": "authTokenResponse"
  },
  "request_id": "req_1234567890",
  "success": true
}
```

### EP-003 - POST /api/v1/auth/logout

- Status: tested
- Endpoint: Logout
- Handler: httpx.Adapter(m.handler.Logout)
- Business Logic Source: authService.invalidateSessionByToken()
- Path Params: none
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. RateLimitWithKeyer
- Applied Call:
```go
policy.RateLimitWithKeyer(limiter, "auth.logout", logoutRule, ratelimit.KeyByUserOrProjectOrTokenHash(16))
```

#### RBAC Permissions
- none

#### Cache Details
- Auth Status: true
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {},
  "path_params": {},
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": null
}
```

## Module: calendar

Total endpoints: 5

### EP-074 - GET /api/v1/projects/{projectId}/calendar

- Status: tested
- Endpoint: ListCalendarData
- Handler: httpx.Adapter(m.handler.ListCalendarData)
- Business Logic Source: getCalendarData()
- Path Params: projectId
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. ProjectRequired
- Applied Call:
```go
policy.ProjectRequired()
```
3. ProjectMatchFromPath
- Applied Call:
```go
policy.ProjectMatchFromPath("projectId")
```
4. ResolvePermissions
- Applied Call:
```go
policy.ResolvePermissions(resolver)
```
5. RequirePermission
- Applied Call:
```go
policy.RequirePermission(rbac.PermCalendarView)
```

#### RBAC Permissions
- rbac.PermCalendarView

#### Cache Details
- Auth Status: true
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {},
  "path_params": {
    "projectId": "string"
  },
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "events": [
      {
        "id": "evt-1",
        "title": "Prototype deadline",
        "type": "Derived",
        "start": "2026-02-14",
        "end": "2026-02-14",
        "allDay": true,
        "owner": "Avery Patel",
        "phase": "Prototype",
        "artifactType": "Task",
        "sourceTitle": "Task - Timeline prototype",
        "createdAt": "2026-02-14"
      },
      {
        "id": "evt-3",
        "title": "Weekly prototype review",
        "type": "Manual",
        "start": "2026-02-16",
        "end": "2026-02-16",
        "allDay": false,
        "startTime": "10:00",
        "endTime": "11:00",
        "owner": "Jordan Lee",
        "phase": "Prototype",
        "artifactType": "Manual",
        "description": "Share progress and align on next steps.",
        "location": "Project room",
        "eventKind": "Review",
        "linkedArtifacts": ["Task - Prototype timeline"],
        "createdAt": "2026-02-14"
      }
    ],
    "reference": {
      "phaseChoices": ["None", "Empathize", "Define", "Ideate", "Prototype", "Test"],
      "manualKinds": ["Workshop", "Review", "Testing Session", "Meeting", "Other"],
      "linkedArtifactOptions": [
        "User Story - Streamline checkout",
        "Problem Statement - Reduce abandonment",
        "Task - Prototype timeline"
      ]
    }
  }
}
```

### EP-075 - POST /api/v1/projects/{projectId}/calendar

- Status: tested
- Endpoint: CreateCalendarEvent
- Handler: httpx.Adapter(m.handler.CreateCalendarEvent)
- Business Logic Source: createCalendarEvent()
- Path Params: projectId
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. ProjectRequired
- Applied Call:
```go
policy.ProjectRequired()
```
3. ProjectMatchFromPath
- Applied Call:
```go
policy.ProjectMatchFromPath("projectId")
```
4. ResolvePermissions
- Applied Call:
```go
policy.ResolvePermissions(resolver)
```
5. RequireJSON
- Applied Call:
```go
policy.RequireJSON()
```
6. RequirePermission
- Applied Call:
```go
policy.RequirePermission(rbac.PermCalendarCreate)
```

#### RBAC Permissions
- rbac.PermCalendarCreate

#### Cache Details
- Auth Status: true
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {
    "_type": "createCalendarEventRequest"
  },
  "path_params": {
    "projectId": "string"
  },
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "id": "evt-1708000000000-0.12345",
    "title": "Design review meeting",
    "type": "Manual",
    "start": "2026-02-20",
    "end": "2026-02-20",
    "allDay": false,
    "startTime": "14:00",
    "endTime": "15:30",
    "owner": "Ayush",
    "phase": "Prototype",
    "artifactType": "Manual",
    "createdAt": "2026-02-14"
  }
}
```

### EP-076 - GET /api/v1/projects/{projectId}/calendar/{eventId}

- Status: tested
- Endpoint: GetCalendarEvent
- Handler: httpx.Adapter(m.handler.GetCalendarEvent)
- Business Logic Source: getCalendarEventData()
- Path Params: eventId, projectId
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. ProjectRequired
- Applied Call:
```go
policy.ProjectRequired()
```
3. ProjectMatchFromPath
- Applied Call:
```go
policy.ProjectMatchFromPath("projectId")
```
4. ResolvePermissions
- Applied Call:
```go
policy.ResolvePermissions(resolver)
```
5. RequirePermission
- Applied Call:
```go
policy.RequirePermission(rbac.PermCalendarView)
```

#### RBAC Permissions
- rbac.PermCalendarView

#### Cache Details
- Auth Status: true
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {},
  "path_params": {
    "eventId": "string",
    "projectId": "string"
  },
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "event": {
      "id": "evt-3",
      "title": "Weekly prototype review",
      "type": "Manual",
      "date": "2026-02-16",
      "allDay": false,
      "owner": "Jordan Lee",
      "eventKind": "Review",
      "description": "Share progress and align on next steps.",
      "location": "Project room",
      "linkedArtifacts": ["Task - Prototype timeline"],
      "tags": [],
      "createdAt": "2026-02-14",
      "lastEdited": "2026-02-14"
    },
    "reference": {
      "phaseChoices": ["None", "Empathize", "Define", "Ideate", "Prototype", "Test"],
      "manualKinds": ["Workshop", "Review", "Testing Session", "Meeting", "Other"],
      "linkedArtifactOptions": [],
      "permissions": { }
    }
  }
}
```

### EP-077 - PUT /api/v1/projects/{projectId}/calendar/{eventId}

- Status: tested
- Endpoint: UpdateCalendarEvent
- Handler: httpx.Adapter(m.handler.UpdateCalendarEvent)
- Business Logic Source: updateCalendarEvent()
- Path Params: eventId, projectId
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. ProjectRequired
- Applied Call:
```go
policy.ProjectRequired()
```
3. ProjectMatchFromPath
- Applied Call:
```go
policy.ProjectMatchFromPath("projectId")
```
4. ResolvePermissions
- Applied Call:
```go
policy.ResolvePermissions(resolver)
```
5. RequireJSON
- Applied Call:
```go
policy.RequireJSON()
```
6. RequirePermission
- Applied Call:
```go
policy.RequirePermission(rbac.PermCalendarEdit)
```

#### RBAC Permissions
- rbac.PermCalendarEdit

#### Cache Details
- Auth Status: true
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {
    "_type": "updateCalendarEventRequest"
  },
  "path_params": {
    "eventId": "string",
    "projectId": "string"
  },
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "id": "evt-3",
    "title": "Updated review meeting",
    "lastEdited": "2026-02-14"
  }
}
```

### EP-078 - DELETE /api/v1/projects/{projectId}/calendar/{eventId}

- Status: tested
- Endpoint: DeleteCalendarEvent
- Handler: httpx.Adapter(m.handler.DeleteCalendarEvent)
- Business Logic Source: deleteCalendarEvent()
- Path Params: eventId, projectId
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. ProjectRequired
- Applied Call:
```go
policy.ProjectRequired()
```
3. ProjectMatchFromPath
- Applied Call:
```go
policy.ProjectMatchFromPath("projectId")
```
4. ResolvePermissions
- Applied Call:
```go
policy.ResolvePermissions(resolver)
```
5. RequirePermission
- Applied Call:
```go
policy.RequirePermission(rbac.PermCalendarDelete)
```

#### RBAC Permissions
- rbac.PermCalendarDelete

#### Cache Details
- Auth Status: true
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {},
  "path_params": {
    "eventId": "string",
    "projectId": "string"
  },
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "eventId": "evt-3"
  }
}
```

## Module: health

Total endpoints: 2

### OP-001 - GET /healthz

- Status: operational
- Endpoint: healthz
- Handler: httpx.Adapter(m.healthz)
- Business Logic Source: n/a
- Path Params: none
- Query Params (inferred): none

#### Policies
- none

#### RBAC Permissions
- none

#### Cache Details
- Auth Status: false
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {},
  "path_params": {},
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "data": {
    "_type": "object"
  },
  "request_id": "req_1234567890",
  "success": true
}
```

### OP-002 - GET /readyz

- Status: operational
- Endpoint: readyz
- Handler: httpx.Adapter(m.readyz)
- Business Logic Source: n/a
- Path Params: none
- Query Params (inferred): none

#### Policies
- none

#### RBAC Permissions
- none

#### Cache Details
- Auth Status: false
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {},
  "path_params": {},
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "data": {
    "_type": "object"
  },
  "request_id": "req_1234567890",
  "success": true
}
```

## Module: home

Total endpoints: 13

### EP-008 - GET /api/v1/home/dashboard

- Status: tested
- Endpoint: Dashboard
- Handler: httpx.Adapter(m.handler.Dashboard)
- Business Logic Source: getUserDashboard()
- Path Params: none
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. CacheRead
- Applied Call:
```go
policy.CacheRead(cacheMgr, cache.CacheReadConfig{
	TTL:	30 * time.Second,
	TagSpecs: []cache.CacheTagSpec{
		{Name: "home.dashboard", UserID: true},
	},
	AllowAuthenticated:	true,
	VaryBy:			cache.CacheVaryBy{UserID: true},
})
```
3. CacheControl
- Applied Call:
```go
policy.CacheControl(policy.CacheControlConfig{Private: true, MaxAge: 30 * time.Second, Vary: []string{"Authorization"}})
```

#### RBAC Permissions
- none

#### Cache Details
- Auth Status: true
- Read Cache:
  - TTL: 30 * time.Second
  - AllowAuthenticated: true
  - TagSpecs: home.dashboard[user_id]
  - VaryBy: userid
- Cache-Control Directives: max-age=30 * time.Second, private
- Cache-Control Vary: Authorization
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {},
  "path_params": {},
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "user": {
      "id": "u-1",
      "name": "Ayush",
      "email": "ayush@projectbook.dev"
    },
    "projects": [
      {
        "id": "atlas-2026",
        "name": "Atlas Research",
        "organization": "League Studio",
        "icon": "rocket",
        "description": "Prototype onboarding flows for first-time users.",
        "role": "Member",
        "openTasks": 3,
        "lastVisitedAt": "2026-02-14T09:20:00.000Z",
        "lastUpdatedAt": "2026-02-14T08:15:00.000Z",
        "status": "Active"
      }
    ],
    "invites": [],
    "notifications": [],
    "activity": []
  }
}
```

### EP-009 - GET /api/v1/home/projects

- Status: tested
- Endpoint: ListProjects
- Handler: httpx.Adapter(m.handler.ListProjects)
- Business Logic Source: getUserProjects()
- Path Params: none
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. CacheRead
- Applied Call:
```go
policy.CacheRead(cacheMgr, cache.CacheReadConfig{
	TTL:	30 * time.Second,
	TagSpecs: []cache.CacheTagSpec{
		{Name: "home.projects", UserID: true},
	},
	AllowAuthenticated:	true,
	VaryBy:			cache.CacheVaryBy{UserID: true, QueryParams: []string{"limit", "offset"}},
})
```
3. CacheControl
- Applied Call:
```go
policy.CacheControl(policy.CacheControlConfig{Private: true, MaxAge: 30 * time.Second, Vary: []string{"Authorization"}})
```

#### RBAC Permissions
- none

#### Cache Details
- Auth Status: true
- Read Cache:
  - TTL: 30 * time.Second
  - AllowAuthenticated: true
  - TagSpecs: home.projects[user_id]
  - VaryBy: query:limit,offset, userid
- Cache-Control Directives: max-age=30 * time.Second, private
- Cache-Control Vary: Authorization
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {},
  "path_params": {},
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": [
    {
      "id": "atlas-2026",
      "name": "Atlas Research",
      "organization": "League Studio",
      "icon": "rocket",
      "description": "Prototype onboarding flows for first-time users.",
      "role": "Member",
      "openTasks": 3,
      "lastVisitedAt": "2026-02-14T09:20:00.000Z",
      "lastUpdatedAt": "2026-02-14T08:15:00.000Z",
      "status": "Active"
    }
  ]
}
```

### EP-011 - GET /api/v1/home/projects/reference

- Status: tested
- Endpoint: ProjectReference
- Handler: httpx.Adapter(m.handler.ProjectReference)
- Business Logic Source: getProjectCreationReference()
- Path Params: none
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. CacheRead
- Applied Call:
```go
policy.CacheRead(cacheMgr, cache.CacheReadConfig{
	TTL:	60 * time.Second,
	TagSpecs: []cache.CacheTagSpec{
		{Name: "home.reference", UserID: true},
	},
	AllowAuthenticated:	true,
	VaryBy:			cache.CacheVaryBy{UserID: true},
})
```
3. CacheControl
- Applied Call:
```go
policy.CacheControl(policy.CacheControlConfig{Private: true, MaxAge: 60 * time.Second, Vary: []string{"Authorization"}})
```

#### RBAC Permissions
- none

#### Cache Details
- Auth Status: true
- Read Cache:
  - TTL: 60 * time.Second
  - AllowAuthenticated: true
  - TagSpecs: home.reference[user_id]
  - VaryBy: userid
- Cache-Control Directives: max-age=60 * time.Second, private
- Cache-Control Vary: Authorization
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {},
  "path_params": {},
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "existingProjects": ["Project Atlas", "Research Hub", "Delta Sprint"],
    "existingUsers": ["avery@league.dev", "nia@league.dev", "jordan@league.dev", "mira@league.dev"]
  }
}
```

### EP-012 - GET /api/v1/home/invites

- Status: tested
- Endpoint: ListInvites
- Handler: httpx.Adapter(m.handler.ListInvites)
- Business Logic Source: getUserInvitesPage()
- Path Params: none
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. CacheRead
- Applied Call:
```go
policy.CacheRead(cacheMgr, cache.CacheReadConfig{
	TTL:	20 * time.Second,
	TagSpecs: []cache.CacheTagSpec{
		{Name: "home.invites", UserID: true},
	},
	AllowAuthenticated:	true,
	VaryBy:			cache.CacheVaryBy{UserID: true},
})
```
3. CacheControl
- Applied Call:
```go
policy.CacheControl(policy.CacheControlConfig{Private: true, MaxAge: 20 * time.Second, Vary: []string{"Authorization"}})
```

#### RBAC Permissions
- none

#### Cache Details
- Auth Status: true
- Read Cache:
  - TTL: 20 * time.Second
  - AllowAuthenticated: true
  - TagSpecs: home.invites[user_id]
  - VaryBy: userid
- Cache-Control Directives: max-age=20 * time.Second, private
- Cache-Control Vary: Authorization
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {},
  "path_params": {},
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": [
    {
      "id": "inv-1",
      "projectName": "Atlas Research",
      "projectDescription": "Prototype new onboarding flows.",
      "projectStatus": "Active",
      "projectId": "atlas-2026",
      "organizationName": "League Studio",
      "inviterName": "Maya Singh",
      "inviterRole": "Owner",
      "inviterEmail": "maya@league.dev",
      "assignedRole": "Member",
      "sentAt": "Feb 3, 2026",
      "expiresAt": "Feb 10, 2026",
      "expiresSoon": false,
      "expired": false
    }
  ]
}
```

### EP-015 - GET /api/v1/home/notifications

- Status: tested
- Endpoint: ListNotifications
- Handler: httpx.Adapter(m.handler.ListNotifications)
- Business Logic Source: getUserNotificationsPage()
- Path Params: none
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. CacheRead
- Applied Call:
```go
policy.CacheRead(cacheMgr, cache.CacheReadConfig{
	TTL:	15 * time.Second,
	TagSpecs: []cache.CacheTagSpec{
		{Name: "home.notifications", UserID: true},
	},
	AllowAuthenticated:	true,
	VaryBy:			cache.CacheVaryBy{UserID: true, QueryParams: []string{"limit"}},
})
```
3. CacheControl
- Applied Call:
```go
policy.CacheControl(policy.CacheControlConfig{Private: true, MaxAge: 15 * time.Second, Vary: []string{"Authorization"}})
```

#### RBAC Permissions
- none

#### Cache Details
- Auth Status: true
- Read Cache:
  - TTL: 15 * time.Second
  - AllowAuthenticated: true
  - TagSpecs: home.notifications[user_id]
  - VaryBy: query:limit, userid
- Cache-Control Directives: max-age=15 * time.Second, private
- Cache-Control Vary: Authorization
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {},
  "path_params": {},
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": [
    {
      "id": "n-1",
      "text": "Alex mentioned you on Northwind Revamp",
      "timestamp": "10m ago",
      "url": "/notifications",
      "read": false,
      "sourceType": "Project Activity",
      "dismissed": false
    }
  ]
}
```

### EP-016 - GET /api/v1/home/activity

- Status: tested
- Endpoint: ListActivity
- Handler: httpx.Adapter(m.handler.ListActivity)
- Business Logic Source: getUserActivityPage()
- Path Params: none
- Query Params (inferred): projectId, type

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. CacheRead
- Applied Call:
```go
policy.CacheRead(cacheMgr, cache.CacheReadConfig{
	TTL:	15 * time.Second,
	TagSpecs: []cache.CacheTagSpec{
		{Name: "home.activity", UserID: true},
	},
	AllowAuthenticated:	true,
	VaryBy:			cache.CacheVaryBy{UserID: true, QueryParams: []string{"limit", "type", "projectId"}},
})
```
3. CacheControl
- Applied Call:
```go
policy.CacheControl(policy.CacheControlConfig{Private: true, MaxAge: 15 * time.Second, Vary: []string{"Authorization"}})
```

#### RBAC Permissions
- none

#### Cache Details
- Auth Status: true
- Read Cache:
  - TTL: 15 * time.Second
  - AllowAuthenticated: true
  - TagSpecs: home.activity[user_id]
  - VaryBy: query:limit,projectId,type, userid
- Cache-Control Directives: max-age=15 * time.Second, private
- Cache-Control Vary: Authorization
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {},
  "path_params": {},
  "query_params": {
    "projectId": "string",
    "type": "string"
  }
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": [
    {
      "id": "a-1",
      "userName": "Alex",
      "userInitials": "AL",
      "action": "commented on",
      "artifactName": "Onboarding Flow",
      "artifactUrl": "/project/atlas-2026/stories/user-1",
      "projectId": "atlas-2026",
      "projectName": "Atlas Research",
      "type": "Comments",
      "timestamp": "12m ago",
      "occurredAt": "2026-02-14T09:30:00.000Z"
    }
  ]
}
```

### EP-017 - GET /api/v1/home/dashboard-activity

- Status: tested
- Endpoint: DashboardActivity
- Handler: httpx.Adapter(m.handler.DashboardActivity)
- Business Logic Source: getUserDashboardActivity()
- Path Params: none
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. CacheRead
- Applied Call:
```go
policy.CacheRead(cacheMgr, cache.CacheReadConfig{
	TTL:	15 * time.Second,
	TagSpecs: []cache.CacheTagSpec{
		{Name: "home.dashboard_activity", UserID: true},
	},
	AllowAuthenticated:	true,
	VaryBy:			cache.CacheVaryBy{UserID: true, QueryParams: []string{"limit"}},
})
```
3. CacheControl
- Applied Call:
```go
policy.CacheControl(policy.CacheControlConfig{Private: true, MaxAge: 15 * time.Second, Vary: []string{"Authorization"}})
```

#### RBAC Permissions
- none

#### Cache Details
- Auth Status: true
- Read Cache:
  - TTL: 15 * time.Second
  - AllowAuthenticated: true
  - TagSpecs: home.dashboard_activity[user_id]
  - VaryBy: query:limit, userid
- Cache-Control Directives: max-age=15 * time.Second, private
- Cache-Control Vary: Authorization
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {},
  "path_params": {},
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": [
    {
      "id": "a-1",
      "userName": "Alex",
      "userInitials": "AL",
      "action": "commented on Onboarding Flow",
      "projectName": "Atlas Research",
      "timestamp": "12m ago",
      "occurredAt": "2026-02-14T09:30:00.000Z",
      "involved": true
    }
  ]
}
```

### EP-018 - GET /api/v1/home/account

- Status: tested
- Endpoint: GetAccountSettings
- Handler: httpx.Adapter(m.handler.GetAccountSettings)
- Business Logic Source: getUserAccountSettings()
- Path Params: none
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. CacheRead
- Applied Call:
```go
policy.CacheRead(cacheMgr, cache.CacheReadConfig{
	TTL:	60 * time.Second,
	TagSpecs: []cache.CacheTagSpec{
		{Name: "home.account", UserID: true},
	},
	AllowAuthenticated:	true,
	VaryBy:			cache.CacheVaryBy{UserID: true},
})
```
3. CacheControl
- Applied Call:
```go
policy.CacheControl(policy.CacheControlConfig{Private: true, MaxAge: 60 * time.Second, Vary: []string{"Authorization"}})
```

#### RBAC Permissions
- none

#### Cache Details
- Auth Status: true
- Read Cache:
  - TTL: 60 * time.Second
  - AllowAuthenticated: true
  - TagSpecs: home.account[user_id]
  - VaryBy: userid
- Cache-Control Directives: max-age=60 * time.Second, private
- Cache-Control Vary: Authorization
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {},
  "path_params": {},
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "displayName": "Ayush",
    "email": "ayush@projectbook.dev",
    "bio": "",
    "theme": "System",
    "density": "Comfortable",
    "landing": "Last Project",
    "timeFormat": "12-hour",
    "inAppNotifications": true,
    "emailNotifications": true
  }
}
```

### EP-020 - GET /api/v1/home/docs

- Status: tested
- Endpoint: Docs
- Handler: httpx.Adapter(m.handler.Docs)
- Business Logic Source: getUserDocsSections()
- Path Params: none
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. CacheRead
- Applied Call:
```go
policy.CacheRead(cacheMgr, cache.CacheReadConfig{
	TTL:	5 * time.Minute,
	TagSpecs: []cache.CacheTagSpec{
		{Name: "home.docs", UserID: true},
	},
	AllowAuthenticated:	true,
	VaryBy:			cache.CacheVaryBy{},
})
```
3. CacheControl
- Applied Call:
```go
policy.CacheControl(policy.CacheControlConfig{Private: true, MaxAge: 5 * time.Minute, Vary: []string{"Authorization"}})
```

#### RBAC Permissions
- none

#### Cache Details
- Auth Status: true
- Read Cache:
  - TTL: 5 * time.Minute
  - AllowAuthenticated: true
  - TagSpecs: home.docs[user_id]
- Cache-Control Directives: max-age=5 * time.Minute, private
- Cache-Control Vary: Authorization
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {},
  "path_params": {},
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "sections": []
  }
}
```

### EP-010 - POST /api/v1/home/projects

- Status: tested
- Endpoint: CreateProject
- Handler: httpx.Adapter(m.handler.CreateProject)
- Business Logic Source: createProject()
- Path Params: none
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. RequireJSON
- Applied Call:
```go
policy.RequireJSON()
```
3. RateLimit
- Applied Call:
```go
policy.RateLimit(limiter, createProjectRule)
```
4. CacheInvalidate
- Applied Call:
```go
policy.CacheInvalidate(cacheMgr, cache.CacheInvalidateConfig{TagSpecs: []cache.CacheTagSpec{
	{Name: "home.dashboard", UserID: true},
	{Name: "home.projects", UserID: true},
	{Name: "home.reference", UserID: true},
}})
```

#### RBAC Permissions
- none

#### Cache Details
- Auth Status: true
- Read Cache: none
- Cache-Control: none
- Invalidation Tags: home.dashboard[user_id], home.projects[user_id], home.reference[user_id]

#### Input Structure (JSON)
```json
{
  "body": {
    "_type": "createProjectRequest"
  },
  "path_params": {},
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "projectId": "new-research-project",
    "project": {
      "id": "new-research-project",
      "name": "New Research Project",
      "organization": "League Studio",
      "icon": "rocket",
      "description": "Exploring user onboarding patterns.",
      "role": "Owner",
      "openTasks": 0,
      "lastUpdatedAt": "2026-02-14T09:30:00.000Z",
      "status": "Active"
    }
  }
}
```

### EP-013 - POST /api/v1/home/invites/{inviteId}/accept

- Status: tested
- Endpoint: AcceptInvite
- Handler: httpx.Adapter(m.handler.AcceptInvite)
- Business Logic Source: acceptProjectInvite()
- Path Params: inviteId
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. RateLimit
- Applied Call:
```go
policy.RateLimit(limiter, inviteActionRule)
```
3. CacheInvalidate
- Applied Call:
```go
policy.CacheInvalidate(cacheMgr, cache.CacheInvalidateConfig{TagSpecs: []cache.CacheTagSpec{
	{Name: "home.dashboard", UserID: true},
	{Name: "home.invites", UserID: true},
	{Name: "home.projects", UserID: true},
}})
```

#### RBAC Permissions
- none

#### Cache Details
- Auth Status: true
- Read Cache: none
- Cache-Control: none
- Invalidation Tags: home.dashboard[user_id], home.invites[user_id], home.projects[user_id]

#### Input Structure (JSON)
```json
{
  "body": {},
  "path_params": {
    "inviteId": "string"
  },
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "inviteId": "inv-1",
    "projectId": "atlas-2026"
  }
}
```

### EP-014 - POST /api/v1/home/invites/{inviteId}/decline

- Status: tested
- Endpoint: DeclineInvite
- Handler: httpx.Adapter(m.handler.DeclineInvite)
- Business Logic Source: declineProjectInvite()
- Path Params: inviteId
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. RateLimit
- Applied Call:
```go
policy.RateLimit(limiter, inviteActionRule)
```
3. CacheInvalidate
- Applied Call:
```go
policy.CacheInvalidate(cacheMgr, cache.CacheInvalidateConfig{TagSpecs: []cache.CacheTagSpec{
	{Name: "home.invites", UserID: true},
	{Name: "home.dashboard", UserID: true},
}})
```

#### RBAC Permissions
- none

#### Cache Details
- Auth Status: true
- Read Cache: none
- Cache-Control: none
- Invalidation Tags: home.dashboard[user_id], home.invites[user_id]

#### Input Structure (JSON)
```json
{
  "body": {},
  "path_params": {
    "inviteId": "string"
  },
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "inviteId": "inv-1"
  }
}
```

### EP-019 - PUT /api/v1/home/account

- Status: tested
- Endpoint: UpdateAccountSettings
- Handler: httpx.Adapter(m.handler.UpdateAccountSettings)
- Business Logic Source: updateUserAccountSettings()
- Path Params: none
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. RequireJSON
- Applied Call:
```go
policy.RequireJSON()
```
3. RateLimit
- Applied Call:
```go
policy.RateLimit(limiter, accountUpdateRule)
```
4. CacheInvalidate
- Applied Call:
```go
policy.CacheInvalidate(cacheMgr, cache.CacheInvalidateConfig{TagSpecs: []cache.CacheTagSpec{
	{Name: "home.account", UserID: true},
	{Name: "home.dashboard", UserID: true},
}})
```

#### RBAC Permissions
- none

#### Cache Details
- Auth Status: true
- Read Cache: none
- Cache-Control: none
- Invalidation Tags: home.account[user_id], home.dashboard[user_id]

#### Input Structure (JSON)
```json
{
  "body": {
    "_type": "updateAccountRequest"
  },
  "path_params": {},
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "updatedAt": "2026-02-14T09:30:00.000Z"
  }
}
```

## Module: pages

Total endpoints: 5

### EP-069 - GET /api/v1/projects/{projectId}/pages

- Status: tested
- Endpoint: ListPages
- Handler: httpx.Adapter(m.handler.ListPages)
- Business Logic Source: getPages()
- Path Params: projectId
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. ProjectRequired
- Applied Call:
```go
policy.ProjectRequired()
```
3. ProjectMatchFromPath
- Applied Call:
```go
policy.ProjectMatchFromPath("projectId")
```
4. ResolvePermissions
- Applied Call:
```go
policy.ResolvePermissions(resolver)
```
5. RequirePermission
- Applied Call:
```go
policy.RequirePermission(rbac.PermPageView)
```
6. CacheReadOptional
- Applied Call:
```go
policy.CacheReadOptional(cacheMgr, cache.CacheReadConfig{
	TTL:			30 * time.Second,
	TagSpecs:		pageReadTags,
	AllowAuthenticated:	true,
	VaryBy: cache.CacheVaryBy{
		ProjectID:	true,
		PathParams:	[]string{"projectId"},
		QueryParams:	[]string{"status", "offset", "limit"},
	},
})
```
7. CacheControlOptional
- Applied Call:
```go
policy.CacheControlOptional(cacheMgr, 30*time.Second)
```

#### RBAC Permissions
- rbac.PermPageView

#### Cache Details
- Auth Status: true
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {},
  "path_params": {
    "projectId": "string"
  },
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": [
    {
      "id": "research-notes",
      "title": "Research notes",
      "owner": "Avery Patel",
      "lastEdited": "2026-02-06",
      "linkedArtifactsCount": 2,
      "status": "Draft",
      "isOrphan": false
    }
  ]
}
```

### EP-070 - POST /api/v1/projects/{projectId}/pages

- Status: tested
- Endpoint: CreatePage
- Handler: httpx.Adapter(m.handler.CreatePage)
- Business Logic Source: createPage()
- Path Params: projectId
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. ProjectRequired
- Applied Call:
```go
policy.ProjectRequired()
```
3. ProjectMatchFromPath
- Applied Call:
```go
policy.ProjectMatchFromPath("projectId")
```
4. ResolvePermissions
- Applied Call:
```go
policy.ResolvePermissions(resolver)
```
5. RequireJSON
- Applied Call:
```go
policy.RequireJSON()
```
6. RequirePermission
- Applied Call:
```go
policy.RequirePermission(rbac.PermPageCreate)
```
7. CacheInvalidateOptional
- Applied Call:
```go
policy.CacheInvalidateOptional(cacheMgr, pageInvalidate)
```

#### RBAC Permissions
- rbac.PermPageCreate

#### Cache Details
- Auth Status: true
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {
    "_type": "createPageRequest"
  },
  "path_params": {
    "projectId": "string"
  },
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "id": "sprint-retrospective-notes",
    "title": "Sprint retrospective notes",
    "owner": "Ayush",
    "lastEdited": "2026-02-14",
    "linkedArtifactsCount": 0,
    "status": "Draft",
    "isOrphan": true
  }
}
```

### EP-071 - GET /api/v1/projects/{projectId}/pages/{slug}

- Status: tested
- Endpoint: GetPage
- Handler: httpx.Adapter(m.handler.GetPage)
- Business Logic Source: getPageEditorData()
- Path Params: projectId, slug
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. ProjectRequired
- Applied Call:
```go
policy.ProjectRequired()
```
3. ProjectMatchFromPath
- Applied Call:
```go
policy.ProjectMatchFromPath("projectId")
```
4. ResolvePermissions
- Applied Call:
```go
policy.ResolvePermissions(resolver)
```
5. RequirePermission
- Applied Call:
```go
policy.RequirePermission(rbac.PermPageView)
```
6. CacheReadOptional
- Applied Call:
```go
policy.CacheReadOptional(cacheMgr, cache.CacheReadConfig{
	TTL:			30 * time.Second,
	TagSpecs:		pageReadTags,
	AllowAuthenticated:	true,
	VaryBy: cache.CacheVaryBy{
		ProjectID:	true,
		PathParams:	[]string{"projectId", "slug"},
	},
})
```
7. CacheControlOptional
- Applied Call:
```go
policy.CacheControlOptional(cacheMgr, 30*time.Second)
```

#### RBAC Permissions
- rbac.PermPageView

#### Cache Details
- Auth Status: true
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {},
  "path_params": {
    "projectId": "string",
    "slug": "string"
  },
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "page": {
      "id": "research-notes",
      "title": "Research notes",
      "status": "Draft",
      "owner": "Avery Patel",
      "lastEdited": "2026-02-06"
    },
    "detail": {
      "description": "",
      "tags": [],
      "linkedArtifacts": [],
      "docHeading": "",
      "docBody": "",
      "views": [
        { "id": "view-doc", "name": "Document", "type": "Document" },
        { "id": "view-table", "name": "Table", "type": "Table" },
        { "id": "view-board", "name": "Board", "type": "Board" }
      ],
      "activeViewId": "view-doc",
      "tableData": []
    },
    "reference": {
      "tagOptions": ["Research", "Alignment", "Notes", "Strategy"],
      "linkedArtifactOptions": [
        "User Story - Streamline checkout",
        "Problem Statement - Reduce abandonment",
        "Idea - Timeline reminders"
      ],
      "permissions": { }
    }
  }
}
```

### EP-072 - PUT /api/v1/projects/{projectId}/pages/{pageId}

- Status: tested
- Endpoint: UpdatePage
- Handler: httpx.Adapter(m.handler.UpdatePage)
- Business Logic Source: updatePageEditor()
- Path Params: pageId, projectId
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. ProjectRequired
- Applied Call:
```go
policy.ProjectRequired()
```
3. ProjectMatchFromPath
- Applied Call:
```go
policy.ProjectMatchFromPath("projectId")
```
4. ResolvePermissions
- Applied Call:
```go
policy.ResolvePermissions(resolver)
```
5. RequireJSON
- Applied Call:
```go
policy.RequireJSON()
```
6. RequirePermission
- Applied Call:
```go
policy.RequirePermission(rbac.PermPageEdit)
```
7. CacheInvalidateOptional
- Applied Call:
```go
policy.CacheInvalidateOptional(cacheMgr, pageInvalidate)
```

#### RBAC Permissions
- rbac.PermPageEdit

#### Cache Details
- Auth Status: true
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {
    "_type": "updatePageRequest"
  },
  "path_params": {
    "pageId": "string",
    "projectId": "string"
  },
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "id": "research-notes",
    "title": "Research notes",
    "owner": "Avery Patel",
    "lastEdited": "2026-02-14",
    "linkedArtifactsCount": 1,
    "status": "Draft",
    "isOrphan": false
  }
}
```

### EP-073 - PUT /api/v1/projects/{projectId}/pages/{pageId}/rename

- Status: tested
- Endpoint: RenamePage
- Handler: httpx.Adapter(m.handler.RenamePage)
- Business Logic Source: renamePage()
- Path Params: pageId, projectId
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. ProjectRequired
- Applied Call:
```go
policy.ProjectRequired()
```
3. ProjectMatchFromPath
- Applied Call:
```go
policy.ProjectMatchFromPath("projectId")
```
4. ResolvePermissions
- Applied Call:
```go
policy.ResolvePermissions(resolver)
```
5. RequireJSON
- Applied Call:
```go
policy.RequireJSON()
```
6. RequirePermission
- Applied Call:
```go
policy.RequirePermission(rbac.PermPageEdit)
```
7. CacheInvalidateOptional
- Applied Call:
```go
policy.CacheInvalidateOptional(cacheMgr, pageInvalidate)
```

#### RBAC Permissions
- rbac.PermPageEdit

#### Cache Details
- Auth Status: true
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {
    "_type": "renamePageRequest"
  },
  "path_params": {
    "pageId": "string",
    "projectId": "string"
  },
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "id": "research-notes",
    "title": "Updated page title",
    "lastEdited": "2026-02-14"
  }
}
```

## Module: project

Total endpoints: 7

### EP-021 - GET /api/v1/projects/{projectId}/dashboard

- Status: tested
- Endpoint: Dashboard
- Handler: httpx.Adapter(m.handler.Dashboard)
- Business Logic Source: getProjectDashboard()
- Path Params: projectId
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. ProjectRequired
- Applied Call:
```go
policy.ProjectRequired()
```
3. ProjectMatchFromPath
- Applied Call:
```go
policy.ProjectMatchFromPath("projectId")
```
4. ResolvePermissions
- Applied Call:
```go
policy.ResolvePermissions(resolver)
```
5. RequirePermission
- Applied Call:
```go
policy.RequirePermission(rbac.PermProjectView)
```
6. CacheRead
- Applied Call:
```go
policy.CacheRead(cacheMgr, cache.CacheReadConfig{
	TTL:	30 * time.Second,
	TagSpecs: []cache.CacheTagSpec{
		{Name: "project.dashboard", ProjectID: true},
	},
	AllowAuthenticated:	true,
	VaryBy:			cache.CacheVaryBy{ProjectID: true, UserID: true},
})
```
7. CacheControl
- Applied Call:
```go
policy.CacheControl(policy.CacheControlConfig{Private: true, MaxAge: 30 * time.Second, Vary: []string{"Authorization"}})
```

#### RBAC Permissions
- rbac.PermProjectView

#### Cache Details
- Auth Status: true
- Read Cache:
  - TTL: 30 * time.Second
  - AllowAuthenticated: true
  - TagSpecs: project.dashboard[project_id]
  - VaryBy: projectid, userid
- Cache-Control Directives: max-age=30 * time.Second, private
- Cache-Control Vary: Authorization
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {},
  "path_params": {
    "projectId": "string"
  },
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "project": {
      "id": "atlas-2026",
      "name": "Northstar Checkout Revamp",
      "status": "Active"
    },
    "me": {
      "id": "u-1",
      "name": "Alex Morgan",
      "initials": "AM"
    },
    "events": [
      {
        "id": "event-1",
        "title": "Prototype critique",
        "type": "Review",
        "startAt": "2026-02-10T16:00:00.000Z",
        "creator": "Priya Shah",
        "initials": "PS"
      }
    ],
    "activity": [
      {
        "id": "a-1",
        "user": "Alex Morgan",
        "initials": "AM",
        "action": "locked Problem Statement",
        "artifact": "Checkout fields are too dense on mobile",
        "href": "/project/atlas-2026/problem-statement/problem-1",
        "at": "2026-02-09T15:10:00.000Z"
      }
    ],
    "recentEdits": [
      {
        "id": "e-1",
        "type": "Story",
        "title": "First-time buyers need payment confidence",
        "href": "/project/atlas-2026/stories/story-2",
        "at": "2026-02-09T10:30:00.000Z"
      }
    ]
  }
}
```

### EP-022 - GET /api/v1/projects/{projectId}/access

- Status: tested
- Endpoint: Access
- Handler: httpx.Adapter(m.handler.Access)
- Business Logic Source: getProjectAccess()
- Path Params: projectId
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. ProjectRequired
- Applied Call:
```go
policy.ProjectRequired()
```
3. ProjectMatchFromPath
- Applied Call:
```go
policy.ProjectMatchFromPath("projectId")
```
4. ResolvePermissions
- Applied Call:
```go
policy.ResolvePermissions(resolver)
```
5. CacheRead
- Applied Call:
```go
policy.CacheRead(cacheMgr, cache.CacheReadConfig{
	TTL:	30 * time.Second,
	TagSpecs: []cache.CacheTagSpec{
		{Name: "project.access", ProjectID: true},
	},
	AllowAuthenticated:	true,
	VaryBy:			cache.CacheVaryBy{ProjectID: true, UserID: true},
})
```
6. CacheControl
- Applied Call:
```go
policy.CacheControl(policy.CacheControlConfig{Private: true, MaxAge: 30 * time.Second, Vary: []string{"Authorization"}})
```

#### RBAC Permissions
- none

#### Cache Details
- Auth Status: true
- Read Cache:
  - TTL: 30 * time.Second
  - AllowAuthenticated: true
  - TagSpecs: project.access[project_id]
  - VaryBy: projectid, userid
- Cache-Control Directives: max-age=30 * time.Second, private
- Cache-Control Vary: Authorization
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {},
  "path_params": {
    "projectId": "string"
  },
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "user": {
      "id": "u-1",
      "name": "Ayush",
      "email": "ayush@projectbook.dev"
    },
    "role": "Admin",
    "permissions": {
      "project": { "view": true, "create": false, "edit": true, "delete": false, "archive": true, "statusChange": true },
      "story": { "view": true, "create": true, "edit": true, "delete": true, "archive": true, "statusChange": true },
      "problem": { "view": true, "create": true, "edit": true, "delete": true, "archive": true, "statusChange": true },
      "idea": { "view": true, "create": true, "edit": true, "delete": true, "archive": true, "statusChange": true },
      "task": { "view": true, "create": true, "edit": true, "delete": true, "archive": true, "statusChange": true },
      "feedback": { "view": true, "create": true, "edit": true, "delete": true, "archive": true, "statusChange": true },
      "resource": { "view": true, "create": true, "edit": true, "delete": true, "archive": true, "statusChange": true },
      "page": { "view": true, "create": true, "edit": true, "delete": true, "archive": true, "statusChange": true },
      "calendar": { "view": true, "create": true, "edit": true, "delete": true, "archive": true, "statusChange": true },
      "member": { "view": true, "create": true, "edit": true, "delete": true, "archive": false, "statusChange": true }
    }
  }
}
```

### EP-023 - GET /api/v1/projects/{projectId}/sidebar

- Status: tested
- Endpoint: Sidebar
- Handler: httpx.Adapter(m.handler.Sidebar)
- Business Logic Source: getProjectSidebarData()
- Path Params: projectId
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. ProjectRequired
- Applied Call:
```go
policy.ProjectRequired()
```
3. ProjectMatchFromPath
- Applied Call:
```go
policy.ProjectMatchFromPath("projectId")
```
4. ResolvePermissions
- Applied Call:
```go
policy.ResolvePermissions(resolver)
```
5. RequirePermission
- Applied Call:
```go
policy.RequirePermission(rbac.PermProjectView)
```
6. CacheRead
- Applied Call:
```go
policy.CacheRead(cacheMgr, cache.CacheReadConfig{
	TTL:	30 * time.Second,
	TagSpecs: []cache.CacheTagSpec{
		{Name: "project.sidebar", ProjectID: true},
	},
	AllowAuthenticated:	true,
	VaryBy:			cache.CacheVaryBy{ProjectID: true, UserID: true},
})
```
7. CacheControl
- Applied Call:
```go
policy.CacheControl(policy.CacheControlConfig{Private: true, MaxAge: 30 * time.Second, Vary: []string{"Authorization"}})
```

#### RBAC Permissions
- rbac.PermProjectView

#### Cache Details
- Auth Status: true
- Read Cache:
  - TTL: 30 * time.Second
  - AllowAuthenticated: true
  - TagSpecs: project.sidebar[project_id]
  - VaryBy: projectid, userid
- Cache-Control Directives: max-age=30 * time.Second, private
- Cache-Control Vary: Authorization
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {},
  "path_params": {
    "projectId": "string"
  },
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "user": {
      "id": "u-1",
      "name": "Ayush",
      "email": "ayush@projectbook.dev"
    },
    "projects": [
      {
        "id": "atlas-2026",
        "name": "Atlas Research",
        "icon": "rocket",
        "status": "Active"
      }
    ],
    "artifacts": {
      "stories": [{ "id": "streamline-checkout", "title": "Streamline checkout for first-time users" }],
      "journeys": [{ "id": "student-assignment-journey", "title": "Student assignment journey" }],
      "problems": [{ "id": "deadline-clarity-students", "title": "Students need a clear way to track..." }],
      "ideas": [{ "id": "deadline-lane-view", "title": "Deadline lane view" }],
      "tasks": [{ "id": "deadline-lane-prototype", "title": "Prototype deadline lane interaction" }],
      "feedback": [{ "id": "deadline-lane-session-1", "title": "Deadline lane usability session" }],
      "pages": [{ "id": "research-notes", "title": "Research notes" }]
    }
  }
}
```

### EP-024 - GET /api/v1/projects/{projectId}/settings

- Status: tested
- Endpoint: GetSettings
- Handler: httpx.Adapter(m.handler.GetSettings)
- Business Logic Source: getProjectSettings()
- Path Params: projectId
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. ProjectRequired
- Applied Call:
```go
policy.ProjectRequired()
```
3. ProjectMatchFromPath
- Applied Call:
```go
policy.ProjectMatchFromPath("projectId")
```
4. ResolvePermissions
- Applied Call:
```go
policy.ResolvePermissions(resolver)
```
5. RequirePermission
- Applied Call:
```go
policy.RequirePermission(rbac.PermProjectView)
```
6. CacheRead
- Applied Call:
```go
policy.CacheRead(cacheMgr, cache.CacheReadConfig{
	TTL:	60 * time.Second,
	TagSpecs: []cache.CacheTagSpec{
		{Name: "project.settings", ProjectID: true},
	},
	AllowAuthenticated:	true,
	VaryBy:			cache.CacheVaryBy{ProjectID: true},
})
```
7. CacheControl
- Applied Call:
```go
policy.CacheControl(policy.CacheControlConfig{Private: true, MaxAge: 60 * time.Second, Vary: []string{"Authorization"}})
```

#### RBAC Permissions
- rbac.PermProjectView

#### Cache Details
- Auth Status: true
- Read Cache:
  - TTL: 60 * time.Second
  - AllowAuthenticated: true
  - TagSpecs: project.settings[project_id]
  - VaryBy: projectid
- Cache-Control Directives: max-age=60 * time.Second, private
- Cache-Control Vary: Authorization
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {},
  "path_params": {
    "projectId": "string"
  },
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "projectName": "Project Atlas",
    "projectDescription": "Core product research and prototype delivery.",
    "projectStatus": "Active",
    "whiteboardsEnabled": true,
    "advancedDatabasesEnabled": true,
    "calendarManualEventsEnabled": true,
    "resourceVersioningEnabled": true,
    "feedbackAggregationEnabled": true,
    "notifyArtifactCreated": true,
    "notifyArtifactLocked": true,
    "notifyFeedbackAdded": true,
    "notifyResourceUpdated": true,
    "deliveryChannel": "In-app"
  }
}
```

### EP-025 - PUT /api/v1/projects/{projectId}/settings

- Status: tested
- Endpoint: UpdateSettings
- Handler: httpx.Adapter(m.handler.UpdateSettings)
- Business Logic Source: updateProjectSettings()
- Path Params: projectId
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. ProjectRequired
- Applied Call:
```go
policy.ProjectRequired()
```
3. ProjectMatchFromPath
- Applied Call:
```go
policy.ProjectMatchFromPath("projectId")
```
4. ResolvePermissions
- Applied Call:
```go
policy.ResolvePermissions(resolver)
```
5. RequirePermission
- Applied Call:
```go
policy.RequirePermission(rbac.PermProjectEdit)
```
6. RequireJSON
- Applied Call:
```go
policy.RequireJSON()
```
7. RateLimit
- Applied Call:
```go
policy.RateLimit(limiter, settingsRule)
```
8. CacheInvalidate
- Applied Call:
```go
policy.CacheInvalidate(cacheMgr, invalidateProjectTags)
```

#### RBAC Permissions
- rbac.PermProjectEdit

#### Cache Details
- Auth Status: true
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {
    "_type": "updateProjectSettingsRequest"
  },
  "path_params": {
    "projectId": "string"
  },
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "projectId": "atlas-2026"
  }
}
```

### EP-026 - POST /api/v1/projects/{projectId}/archive

- Status: tested
- Endpoint: Archive
- Handler: httpx.Adapter(m.handler.Archive)
- Business Logic Source: archiveProject()
- Path Params: projectId
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. ProjectRequired
- Applied Call:
```go
policy.ProjectRequired()
```
3. ProjectMatchFromPath
- Applied Call:
```go
policy.ProjectMatchFromPath("projectId")
```
4. ResolvePermissions
- Applied Call:
```go
policy.ResolvePermissions(resolver)
```
5. RequirePermission
- Applied Call:
```go
policy.RequirePermission(rbac.PermProjectArchive)
```
6. RateLimit
- Applied Call:
```go
policy.RateLimit(limiter, archiveRule)
```
7. CacheInvalidate
- Applied Call:
```go
policy.CacheInvalidate(cacheMgr, invalidateProjectTags)
```

#### RBAC Permissions
- rbac.PermProjectArchive

#### Cache Details
- Auth Status: true
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {},
  "path_params": {
    "projectId": "string"
  },
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "projectId": "atlas-2026",
    "status": "Archived"
  }
}
```

### EP-027 - DELETE /api/v1/projects/{projectId}

- Status: tested
- Endpoint: Delete
- Handler: httpx.Adapter(m.handler.Delete)
- Business Logic Source: deleteProject()
- Path Params: projectId
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. ProjectRequired
- Applied Call:
```go
policy.ProjectRequired()
```
3. ProjectMatchFromPath
- Applied Call:
```go
policy.ProjectMatchFromPath("projectId")
```
4. ResolvePermissions
- Applied Call:
```go
policy.ResolvePermissions(resolver)
```
5. RequirePermission
- Applied Call:
```go
policy.RequirePermission(rbac.PermProjectDelete)
```
6. RateLimit
- Applied Call:
```go
policy.RateLimit(limiter, deleteRule)
```
7. CacheInvalidate
- Applied Call:
```go
policy.CacheInvalidate(cacheMgr, invalidateProjectTags)
```

#### RBAC Permissions
- rbac.PermProjectDelete

#### Cache Details
- Auth Status: true
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {},
  "path_params": {
    "projectId": "string"
  },
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "projectId": "atlas-2026",
    "status": "Deleted"
  }
}
```

## Module: resources

Total endpoints: 5

### EP-064 - GET /api/v1/projects/{projectId}/resources

- Status: tested
- Endpoint: ListResources
- Handler: httpx.Adapter(m.handler.ListResources)
- Business Logic Source: getResources()
- Path Params: projectId
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. ProjectRequired
- Applied Call:
```go
policy.ProjectRequired()
```
3. ProjectMatchFromPath
- Applied Call:
```go
policy.ProjectMatchFromPath("projectId")
```
4. ResolvePermissions
- Applied Call:
```go
policy.ResolvePermissions(resolver)
```
5. RequirePermission
- Applied Call:
```go
policy.RequirePermission(rbac.PermResourceView)
```
6. CacheReadOptional
- Applied Call:
```go
policy.CacheReadOptional(cacheMgr, cache.CacheReadConfig{
	TTL:			30 * time.Second,
	TagSpecs:		resourceReadTags,
	AllowAuthenticated:	true,
	VaryBy: cache.CacheVaryBy{
		ProjectID:	true,
		PathParams:	[]string{"projectId"},
		QueryParams:	[]string{"status", "offset", "limit"},
	},
})
```
7. CacheControlOptional
- Applied Call:
```go
policy.CacheControlOptional(cacheMgr, 30*time.Second)
```

#### RBAC Permissions
- rbac.PermResourceView

#### Cache Details
- Auth Status: true
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {},
  "path_params": {
    "projectId": "string"
  },
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "rows": [
      {
        "id": "res-1",
        "name": "Student survey synthesis",
        "fileType": "PDF",
        "docType": "Research Paper",
        "owner": "Avery Patel",
        "version": "v3",
        "lastUpdated": "Jan 12, 2026",
        "linkedCount": 3,
        "status": "Active"
      }
    ],
    "reference": {
      "docTypes": ["Pitch Deck", "Research Paper", "Specification", "Design File", "Other"],
      "fileTypes": ["PDF", "PPTX", "DOCX"],
      "owners": ["Avery Patel", "Nia Clark", "Dr. Ramos"],
      "sortOptions": ["Last Updated", "Name", "Upload Date"]
    }
  }
}
```

### EP-065 - POST /api/v1/projects/{projectId}/resources

- Status: tested
- Endpoint: CreateResource
- Handler: httpx.Adapter(m.handler.CreateResource)
- Business Logic Source: createResource()
- Path Params: projectId
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. ProjectRequired
- Applied Call:
```go
policy.ProjectRequired()
```
3. ProjectMatchFromPath
- Applied Call:
```go
policy.ProjectMatchFromPath("projectId")
```
4. ResolvePermissions
- Applied Call:
```go
policy.ResolvePermissions(resolver)
```
5. RequireJSON
- Applied Call:
```go
policy.RequireJSON()
```
6. RequirePermission
- Applied Call:
```go
policy.RequirePermission(rbac.PermResourceCreate)
```
7. CacheInvalidateOptional
- Applied Call:
```go
policy.CacheInvalidateOptional(cacheMgr, resourceInvalidate)
```

#### RBAC Permissions
- rbac.PermResourceCreate

#### Cache Details
- Auth Status: true
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {
    "_type": "createResourceRequest"
  },
  "path_params": {
    "projectId": "string"
  },
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "id": "res-1708000000000-0.12345",
    "name": "Interview transcript batch 2",
    "fileType": "PDF",
    "docType": "Research Paper",
    "owner": "Ayush",
    "version": "v1",
    "lastUpdated": "2026-02-14",
    "linkedCount": 0,
    "status": "Active"
  }
}
```

### EP-066 - GET /api/v1/projects/{projectId}/resources/{resourceId}

- Status: tested
- Endpoint: GetResource
- Handler: httpx.Adapter(m.handler.GetResource)
- Business Logic Source: getResourcePageData()
- Path Params: projectId, resourceId
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. ProjectRequired
- Applied Call:
```go
policy.ProjectRequired()
```
3. ProjectMatchFromPath
- Applied Call:
```go
policy.ProjectMatchFromPath("projectId")
```
4. ResolvePermissions
- Applied Call:
```go
policy.ResolvePermissions(resolver)
```
5. RequirePermission
- Applied Call:
```go
policy.RequirePermission(rbac.PermResourceView)
```
6. CacheReadOptional
- Applied Call:
```go
policy.CacheReadOptional(cacheMgr, cache.CacheReadConfig{
	TTL:			30 * time.Second,
	TagSpecs:		resourceReadTags,
	AllowAuthenticated:	true,
	VaryBy: cache.CacheVaryBy{
		ProjectID:	true,
		PathParams:	[]string{"projectId", "resourceId"},
	},
})
```
7. CacheControlOptional
- Applied Call:
```go
policy.CacheControlOptional(cacheMgr, 30*time.Second)
```

#### RBAC Permissions
- rbac.PermResourceView

#### Cache Details
- Auth Status: true
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {},
  "path_params": {
    "projectId": "string",
    "resourceId": "string"
  },
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "resource": {
      "id": "res-1",
      "name": "Student survey synthesis",
      "fileType": "PDF",
      "docType": "Research Paper",
      "status": "Active",
      "owner": "Avery Patel"
    },
    "detail": {
      "name": "Student survey synthesis",
      "fileType": "PDF",
      "docType": "Research Paper",
      "status": "Active",
      "description": "Survey synthesis across 24 student interviews.",
      "owner": "Avery Patel",
      "createdAt": "Jan 2, 2026",
      "updatedAt": "Jan 12, 2026",
      "fileSize": "4.2 MB",
      "linkedArtifacts": [
        {
          "id": "story-7",
          "title": "Avery Patel - First-year student",
          "type": "User Story",
          "phase": "Empathize",
          "href": "/project/alpha/stories/avery-patel"
        }
      ],
      "versions": [
        {
          "version": "v3",
          "uploadedBy": "Avery Patel",
          "uploadDate": "Jan 12, 2026",
          "description": "Added follow-up insights"
        }
      ],
      "notes": ""
    },
    "reference": {
      "storyOptions": ["Avery Patel - First-year student"],
      "problemOptions": ["Students miss assignment requirements"],
      "ideaOptions": ["Visual deadline timeline for assignments"],
      "taskOptions": ["Prototype timeline for assignment deadlines"],
      "permissions": { }
    }
  }
}
```

### EP-067 - PUT /api/v1/projects/{projectId}/resources/{resourceId}

- Status: tested
- Endpoint: UpdateResource
- Handler: httpx.Adapter(m.handler.UpdateResource)
- Business Logic Source: updateResource()
- Path Params: projectId, resourceId
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. ProjectRequired
- Applied Call:
```go
policy.ProjectRequired()
```
3. ProjectMatchFromPath
- Applied Call:
```go
policy.ProjectMatchFromPath("projectId")
```
4. ResolvePermissions
- Applied Call:
```go
policy.ResolvePermissions(resolver)
```
5. RequireJSON
- Applied Call:
```go
policy.RequireJSON()
```
6. RequirePermission
- Applied Call:
```go
policy.RequirePermission(rbac.PermResourceEdit)
```
7. CacheInvalidateOptional
- Applied Call:
```go
policy.CacheInvalidateOptional(cacheMgr, resourceInvalidate)
```

#### RBAC Permissions
- rbac.PermResourceEdit

#### Cache Details
- Auth Status: true
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {
    "_type": "updateResourceRequest"
  },
  "path_params": {
    "projectId": "string",
    "resourceId": "string"
  },
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "id": "res-1",
    "name": "Updated resource name",
    "fileType": "PDF",
    "docType": "Research Paper",
    "owner": "Avery Patel",
    "version": "v4",
    "lastUpdated": "2026-02-14",
    "linkedCount": 1,
    "status": "Active"
  }
}
```

### EP-068 - PUT /api/v1/projects/{projectId}/resources/{resourceId}/status

- Status: tested
- Endpoint: UpdateResourceStatus
- Handler: httpx.Adapter(m.handler.UpdateResourceStatus)
- Business Logic Source: updateResourceStatus()
- Path Params: projectId, resourceId
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. ProjectRequired
- Applied Call:
```go
policy.ProjectRequired()
```
3. ProjectMatchFromPath
- Applied Call:
```go
policy.ProjectMatchFromPath("projectId")
```
4. ResolvePermissions
- Applied Call:
```go
policy.ResolvePermissions(resolver)
```
5. RequireJSON
- Applied Call:
```go
policy.RequireJSON()
```
6. RequirePermission
- Applied Call:
```go
policy.RequirePermission(rbac.PermResourceStatusChange)
```
7. CacheInvalidateOptional
- Applied Call:
```go
policy.CacheInvalidateOptional(cacheMgr, resourceInvalidate)
```

#### RBAC Permissions
- rbac.PermResourceStatusChange

#### Cache Details
- Auth Status: true
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {
    "_type": "updateResourceStatusRequest"
  },
  "path_params": {
    "projectId": "string",
    "resourceId": "string"
  },
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "id": "res-1",
    "status": "Archived",
    "lastUpdated": "2026-02-14"
  }
}
```

## Module: sidebar

Total endpoints: 3

### EP-079 - POST /api/v1/projects/{projectId}/sidebar/artifacts

- Status: tested
- Endpoint: CreateSidebarArtifact
- Handler: httpx.Adapter(m.handler.CreateSidebarArtifact)
- Business Logic Source: createSidebarArtifact()
- Path Params: projectId
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. ProjectRequired
- Applied Call:
```go
policy.ProjectRequired()
```
3. ProjectMatchFromPath
- Applied Call:
```go
policy.ProjectMatchFromPath("projectId")
```
4. ResolvePermissions
- Applied Call:
```go
policy.ResolvePermissions(resolver)
```
5. RequireJSON
- Applied Call:
```go
policy.RequireJSON()
```
6. RequireAnyPermission
- Applied Call:
```go
policy.RequireAnyPermission(
	rbac.PermStoryCreate,
	rbac.PermProblemCreate,
	rbac.PermIdeaCreate,
	rbac.PermTaskCreate,
	rbac.PermFeedbackCreate,
	rbac.PermPageCreate,
)
```
7. CacheInvalidateOptional
- Applied Call:
```go
policy.CacheInvalidateOptional(cacheMgr, invalidateTags)
```

#### RBAC Permissions
- rbac.PermFeedbackCreate
- rbac.PermIdeaCreate
- rbac.PermPageCreate
- rbac.PermProblemCreate
- rbac.PermStoryCreate
- rbac.PermTaskCreate

#### Cache Details
- Auth Status: true
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {
    "_type": "createSidebarArtifactRequest"
  },
  "path_params": {
    "projectId": "string"
  },
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "id": "new-user-story-from-sidebar",
    "title": "New user story from sidebar"
  }
}
```

### EP-080 - PUT /api/v1/projects/{projectId}/sidebar/artifacts/{artifactId}/rename

- Status: tested
- Endpoint: RenameSidebarArtifact
- Handler: httpx.Adapter(m.handler.RenameSidebarArtifact)
- Business Logic Source: renameSidebarArtifact()
- Path Params: artifactId, projectId
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. ProjectRequired
- Applied Call:
```go
policy.ProjectRequired()
```
3. ProjectMatchFromPath
- Applied Call:
```go
policy.ProjectMatchFromPath("projectId")
```
4. ResolvePermissions
- Applied Call:
```go
policy.ResolvePermissions(resolver)
```
5. RequireJSON
- Applied Call:
```go
policy.RequireJSON()
```
6. RequireAnyPermission
- Applied Call:
```go
policy.RequireAnyPermission(
	rbac.PermStoryEdit,
	rbac.PermProblemEdit,
	rbac.PermIdeaEdit,
	rbac.PermTaskEdit,
	rbac.PermFeedbackEdit,
	rbac.PermPageEdit,
)
```
7. CacheInvalidateOptional
- Applied Call:
```go
policy.CacheInvalidateOptional(cacheMgr, invalidateTags)
```

#### RBAC Permissions
- rbac.PermFeedbackEdit
- rbac.PermIdeaEdit
- rbac.PermPageEdit
- rbac.PermProblemEdit
- rbac.PermStoryEdit
- rbac.PermTaskEdit

#### Cache Details
- Auth Status: true
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {
    "_type": "renameSidebarArtifactRequest"
  },
  "path_params": {
    "artifactId": "string",
    "projectId": "string"
  },
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "id": "streamline-checkout",
    "title": "Renamed story title"
  }
}
```

### EP-081 - DELETE /api/v1/projects/{projectId}/sidebar/artifacts/{artifactId}

- Status: tested
- Endpoint: DeleteSidebarArtifact
- Handler: httpx.Adapter(m.handler.DeleteSidebarArtifact)
- Business Logic Source: deleteSidebarArtifact()
- Path Params: artifactId, projectId
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. ProjectRequired
- Applied Call:
```go
policy.ProjectRequired()
```
3. ProjectMatchFromPath
- Applied Call:
```go
policy.ProjectMatchFromPath("projectId")
```
4. ResolvePermissions
- Applied Call:
```go
policy.ResolvePermissions(resolver)
```
5. RequireJSON
- Applied Call:
```go
policy.RequireJSON()
```
6. RequireAnyPermission
- Applied Call:
```go
policy.RequireAnyPermission(
	rbac.PermStoryDelete,
	rbac.PermProblemDelete,
	rbac.PermIdeaDelete,
	rbac.PermTaskDelete,
	rbac.PermFeedbackDelete,
	rbac.PermPageDelete,
)
```
7. CacheInvalidateOptional
- Applied Call:
```go
policy.CacheInvalidateOptional(cacheMgr, invalidateTags)
```

#### RBAC Permissions
- rbac.PermFeedbackDelete
- rbac.PermIdeaDelete
- rbac.PermPageDelete
- rbac.PermProblemDelete
- rbac.PermStoryDelete
- rbac.PermTaskDelete

#### Cache Details
- Auth Status: true
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {
    "_type": "deleteSidebarArtifactRequest"
  },
  "path_params": {
    "artifactId": "string",
    "projectId": "string"
  },
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "id": "streamline-checkout"
  }
}
```

## Module: system

Total endpoints: 2

### OP-001 - POST /system/parse-duration

- Status: operational
- Endpoint: parseDuration
- Handler: httpx.Adapter(m.parseDuration)
- Business Logic Source: n/a
- Path Params: none
- Query Params (inferred): none

#### Policies
1. RequireJSON
- Applied Call:
```go
policy.RequireJSON()
```

#### RBAC Permissions
- none

#### Cache Details
- Auth Status: false
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {},
  "path_params": {},
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "data": {
    "_type": "object"
  },
  "request_id": "req_1234567890",
  "success": true
}
```

### OP-002 - GET /api/v1/system/whoami

- Status: operational
- Endpoint: whoami
- Handler: httpx.Adapter(m.whoami)
- Business Logic Source: n/a
- Path Params: none
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. RateLimitWithKeyer
- Applied Call:
```go
policy.RateLimitWithKeyer(limiter, "system.whoami", m.rateRule, ratelimit.KeyByUserOrProjectOrTokenHash(16))
```

#### RBAC Permissions
- none

#### Cache Details
- Auth Status: true
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {},
  "path_params": {},
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "data": {
    "_type": "object"
  },
  "request_id": "req_1234567890",
  "success": true
}
```

## Module: team

Total endpoints: 7

### EP-028 - GET /api/v1/projects/{projectId}/team/members

- Status: tested
- Endpoint: ListMembers
- Handler: httpx.Adapter(m.handler.ListMembers)
- Business Logic Source: getProjectTeamMembers()
- Path Params: projectId
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. ProjectRequired
- Applied Call:
```go
policy.ProjectRequired()
```
3. ProjectMatchFromPath
- Applied Call:
```go
policy.ProjectMatchFromPath("projectId")
```
4. ResolvePermissions
- Applied Call:
```go
policy.ResolvePermissions(resolver)
```
5. RequirePermission
- Applied Call:
```go
policy.RequirePermission(rbac.PermMemberView)
```
6. CacheRead
- Applied Call:
```go
policy.CacheRead(cacheMgr, cache.CacheReadConfig{
	TTL:	30 * time.Second,
	TagSpecs: []cache.CacheTagSpec{
		{Name: "team.members", ProjectID: true},
	},
	AllowAuthenticated:	true,
	VaryBy:			cache.CacheVaryBy{ProjectID: true},
})
```
7. CacheControl
- Applied Call:
```go
policy.CacheControl(policy.CacheControlConfig{Private: true, MaxAge: 30 * time.Second, Vary: []string{"Authorization"}})
```

#### RBAC Permissions
- rbac.PermMemberView

#### Cache Details
- Auth Status: true
- Read Cache:
  - TTL: 30 * time.Second
  - AllowAuthenticated: true
  - TagSpecs: team.members[project_id]
  - VaryBy: projectid
- Cache-Control Directives: max-age=30 * time.Second, private
- Cache-Control Vary: Authorization
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {},
  "path_params": {
    "projectId": "string"
  },
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "members": [
      {
        "id": "mem-1",
        "name": "Avery Patel",
        "email": "avery@league.dev",
        "role": "Owner",
        "status": "Active",
        "joinedAt": "2026-02-04"
      },
      {
        "id": "mem-2",
        "name": "Nia Clark",
        "email": "nia@league.dev",
        "role": "Admin",
        "status": "Active",
        "joinedAt": "2026-02-05"
      },
      {
        "id": "mem-3",
        "name": "Jordan Lee",
        "email": "jordan@league.dev",
        "role": "Viewer",
        "status": "Invited",
        "joinedAt": "2026-02-06"
      }
    ],
    "invites": [
      {
        "email": "newuser@example.com",
        "role": "Editor",
        "sentDate": "2026-02-10",
        "status": "pending"
      }
    ]
  }
}
```

### EP-029 - GET /api/v1/projects/{projectId}/team/roles

- Status: tested
- Endpoint: ListRoles
- Handler: httpx.Adapter(m.handler.ListRoles)
- Business Logic Source: getProjectTeamRoles()
- Path Params: projectId
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. ProjectRequired
- Applied Call:
```go
policy.ProjectRequired()
```
3. ProjectMatchFromPath
- Applied Call:
```go
policy.ProjectMatchFromPath("projectId")
```
4. ResolvePermissions
- Applied Call:
```go
policy.ResolvePermissions(resolver)
```
5. RequirePermission
- Applied Call:
```go
policy.RequirePermission(rbac.PermMemberView)
```
6. CacheRead
- Applied Call:
```go
policy.CacheRead(cacheMgr, cache.CacheReadConfig{
	TTL:	30 * time.Second,
	TagSpecs: []cache.CacheTagSpec{
		{Name: "team.roles", ProjectID: true},
	},
	AllowAuthenticated:	true,
	VaryBy:			cache.CacheVaryBy{ProjectID: true},
})
```
7. CacheControl
- Applied Call:
```go
policy.CacheControl(policy.CacheControlConfig{Private: true, MaxAge: 30 * time.Second, Vary: []string{"Authorization"}})
```

#### RBAC Permissions
- rbac.PermMemberView

#### Cache Details
- Auth Status: true
- Read Cache:
  - TTL: 30 * time.Second
  - AllowAuthenticated: true
  - TagSpecs: team.roles[project_id]
  - VaryBy: projectid
- Cache-Control Directives: max-age=30 * time.Second, private
- Cache-Control Vary: Authorization
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {},
  "path_params": {
    "projectId": "string"
  },
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "rolePermissionMasks": {
      "Owner": "1152921504606846975",
      "Admin": "864691128455135229",
      "Editor": "20016033248999873",
      "Member": "875734824153537",
      "Viewer": "18300341342965825",
      "Limited Access": "0"
    },
    "members": [
      {
        "id": "mem-1",
        "name": "Avery Patel",
        "email": "avery@league.dev",
        "role": "Owner",
        "status": "Active",
        "joinedAt": "2026-02-04",
        "isCustom": false,
        "permissionMask": "1152921504606846975"
      },
      {
        "id": "mem-2",
        "name": "Nia Clark",
        "email": "nia@league.dev",
        "role": "Editor",
        "status": "Active",
        "joinedAt": "2026-02-05",
        "isCustom": true,
        "permissionMask": "20016033248999873"
      }
    ]
  }
}
```

### EP-030 - POST /api/v1/projects/{projectId}/team/invites

- Status: tested
- Endpoint: CreateInvite
- Handler: httpx.Adapter(m.handler.CreateInvite)
- Business Logic Source: createProjectInvite()
- Path Params: projectId
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. ProjectRequired
- Applied Call:
```go
policy.ProjectRequired()
```
3. ProjectMatchFromPath
- Applied Call:
```go
policy.ProjectMatchFromPath("projectId")
```
4. ResolvePermissions
- Applied Call:
```go
policy.ResolvePermissions(resolver)
```
5. RequirePermission
- Applied Call:
```go
policy.RequirePermission(rbac.PermMemberCreate)
```
6. RequireJSON
- Applied Call:
```go
policy.RequireJSON()
```
7. RateLimit
- Applied Call:
```go
policy.RateLimit(limiter, inviteRule)
```
8. CacheInvalidate
- Applied Call:
```go
policy.CacheInvalidate(cacheMgr, invalidateTeamTags)
```

#### RBAC Permissions
- rbac.PermMemberCreate

#### Cache Details
- Auth Status: true
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {
    "_type": "createInviteRequest"
  },
  "path_params": {
    "projectId": "string"
  },
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "email": "newuser@example.com",
    "role": "Editor",
    "sentDate": "2026-02-14",
    "status": "pending"
  }
}
```

### EP-031 - POST /api/v1/projects/{projectId}/team/invites/batch

- Status: tested
- Endpoint: BatchInvites
- Handler: httpx.Adapter(m.handler.BatchInvites)
- Business Logic Source: sendProjectInvites()
- Path Params: projectId
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. ProjectRequired
- Applied Call:
```go
policy.ProjectRequired()
```
3. ProjectMatchFromPath
- Applied Call:
```go
policy.ProjectMatchFromPath("projectId")
```
4. ResolvePermissions
- Applied Call:
```go
policy.ResolvePermissions(resolver)
```
5. RequirePermission
- Applied Call:
```go
policy.RequirePermission(rbac.PermMemberCreate)
```
6. RequireJSON
- Applied Call:
```go
policy.RequireJSON()
```
7. RateLimit
- Applied Call:
```go
policy.RateLimit(limiter, batchInviteRule)
```
8. CacheInvalidate
- Applied Call:
```go
policy.CacheInvalidate(cacheMgr, invalidateTeamTags)
```

#### RBAC Permissions
- rbac.PermMemberCreate

#### Cache Details
- Auth Status: true
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {
    "_type": "batchInviteRequest"
  },
  "path_params": {
    "projectId": "string"
  },
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "projectId": "new-research-project",
    "invited": [
      { "email": "avery@league.dev", "role": "Editor" },
      { "email": "nia@league.dev", "role": "Viewer" }
    ]
  }
}
```

### EP-032 - DELETE /api/v1/projects/{projectId}/team/invites/{email}

- Status: tested
- Endpoint: CancelInvite
- Handler: httpx.Adapter(m.handler.CancelInvite)
- Business Logic Source: cancelProjectInvite()
- Path Params: email, projectId
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. ProjectRequired
- Applied Call:
```go
policy.ProjectRequired()
```
3. ProjectMatchFromPath
- Applied Call:
```go
policy.ProjectMatchFromPath("projectId")
```
4. ResolvePermissions
- Applied Call:
```go
policy.ResolvePermissions(resolver)
```
5. RequirePermission
- Applied Call:
```go
policy.RequirePermission(rbac.PermMemberDelete)
```
6. RateLimit
- Applied Call:
```go
policy.RateLimit(limiter, cancelInviteRule)
```
7. CacheInvalidate
- Applied Call:
```go
policy.CacheInvalidate(cacheMgr, invalidateTeamTags)
```

#### RBAC Permissions
- rbac.PermMemberDelete

#### Cache Details
- Auth Status: true
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {},
  "path_params": {
    "email": "string",
    "projectId": "string"
  },
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "email": "newuser@example.com"
  }
}
```

### EP-033 - PUT /api/v1/projects/{projectId}/team/members/{memberId}/permissions

- Status: tested
- Endpoint: UpdateMemberPermissions
- Handler: httpx.Adapter(m.handler.UpdateMemberPermissions)
- Business Logic Source: updateProjectMemberPermissions()
- Path Params: memberId, projectId
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. ProjectRequired
- Applied Call:
```go
policy.ProjectRequired()
```
3. ProjectMatchFromPath
- Applied Call:
```go
policy.ProjectMatchFromPath("projectId")
```
4. ResolvePermissions
- Applied Call:
```go
policy.ResolvePermissions(resolver)
```
5. RequirePermission
- Applied Call:
```go
policy.RequirePermission(rbac.PermMemberEdit)
```
6. RequireJSON
- Applied Call:
```go
policy.RequireJSON()
```
7. RateLimit
- Applied Call:
```go
policy.RateLimit(limiter, memberUpdateRule)
```
8. CacheInvalidate
- Applied Call:
```go
policy.CacheInvalidate(cacheMgr, invalidateTeamTags)
```

#### RBAC Permissions
- rbac.PermMemberEdit

#### Cache Details
- Auth Status: true
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {
    "_type": "updateMemberPermissionsRequest"
  },
  "path_params": {
    "memberId": "string",
    "projectId": "string"
  },
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "memberId": "mem-2",
    "role": "Editor",
    "isCustom": true,
    "permissionMask": 987654321
  }
}
```

### EP-034 - PUT /api/v1/projects/{projectId}/team/roles/{role}/permissions

- Status: tested
- Endpoint: UpdateRolePermissions
- Handler: httpx.Adapter(m.handler.UpdateRolePermissions)
- Business Logic Source: updateProjectRolePermissions()
- Path Params: projectId, role
- Query Params (inferred): none

#### Policies
1. AuthRequired
- Applied Call:
```go
policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode())
```
2. ProjectRequired
- Applied Call:
```go
policy.ProjectRequired()
```
3. ProjectMatchFromPath
- Applied Call:
```go
policy.ProjectMatchFromPath("projectId")
```
4. ResolvePermissions
- Applied Call:
```go
policy.ResolvePermissions(resolver)
```
5. RequirePermission
- Applied Call:
```go
policy.RequirePermission(rbac.PermMemberEdit)
```
6. RequireJSON
- Applied Call:
```go
policy.RequireJSON()
```
7. RateLimit
- Applied Call:
```go
policy.RateLimit(limiter, roleUpdateRule)
```
8. CacheInvalidate
- Applied Call:
```go
policy.CacheInvalidate(cacheMgr, invalidateTeamTags)
```

#### RBAC Permissions
- rbac.PermMemberEdit

#### Cache Details
- Auth Status: true
- Read Cache: none
- Cache-Control: none
- Invalidation: none

#### Input Structure (JSON)
```json
{
  "body": {
    "_type": "updateRolePermissionsRequest"
  },
  "path_params": {
    "projectId": "string",
    "role": "string"
  },
  "query_params": {}
}
```

#### Output Structure (JSON)
```json
{
  "success": true,
  "data": {
    "role": "Admin",
    "permissionMask": 123456789,
    "customMembersUnaffected": 2
  }
}
```

