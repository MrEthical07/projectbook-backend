package policy

import (
	"fmt"
	"strings"
)

func validateRouteRules(method, pattern string, metas []Metadata) error {
	_ = method
	if err := validatePolicyOrdering(metas); err != nil {
		return err
	}
	if err := validateAuthDependencies(metas); err != nil {
		return err
	}
	if err := validateProjectRules(pattern, metas); err != nil {
		return err
	}
	if err := validateResolverRules(metas); err != nil {
		return err
	}
	if err := validateCacheSafety(metas); err != nil {
		return err
	}
	return nil
}

func validatePolicyOrdering(metas []Metadata) error {
	previousStage := 0
	previousType := PolicyTypeUnknown
	projectRequiredIndex := -1
	projectMatchIndex := -1

	for i, meta := range metas {
		stage := policyOrderStage(meta.Type)
		if stage > 0 {
			if stage < previousStage {
				return fmt.Errorf("policy %s cannot appear after %s", meta.Name, previousType)
			}
			previousStage = stage
			previousType = meta.Type
		}

		switch meta.Type {
		case PolicyTypeProjectRequired:
			if projectRequiredIndex == -1 {
				projectRequiredIndex = i
			}
		case PolicyTypeProjectMatchFromPath:
			if projectMatchIndex == -1 {
				projectMatchIndex = i
			}
		}
	}

	if projectRequiredIndex >= 0 && projectMatchIndex >= 0 && projectMatchIndex < projectRequiredIndex {
		return fmt.Errorf("policy %s must appear after %s", PolicyTypeProjectMatchFromPath, PolicyTypeProjectRequired)
	}

	return nil
}

func validateAuthDependencies(metas []Metadata) error {
	hasAuthRequired := hasPolicyType(metas, PolicyTypeAuthRequired)
	if hasAuthRequired {
		return nil
	}

	if hasPolicyType(metas, PolicyTypeRequirePermission) ||
		hasPolicyType(metas, PolicyTypeRequireAnyPermission) ||
		hasPolicyType(metas, PolicyTypeRequireAllPermissions) ||
		hasPolicyType(metas, PolicyTypeProjectRequired) ||
		hasPolicyType(metas, PolicyTypeResolvePermissions) {
		return fmt.Errorf("%s is required when RBAC or project policies are configured", PolicyTypeAuthRequired)
	}

	return nil
}

func validateProjectRules(pattern string, metas []Metadata) error {
	hasProjectRequired := hasPolicyType(metas, PolicyTypeProjectRequired)
	projectMatchPolicies := findPolicies(metas, PolicyTypeProjectMatchFromPath)
	hasProjectPath := patternContainsProjectID(pattern)
	expectedProjectPathParam := routeProjectPathParam(pattern)
	expectedProjectPathParamDisplay := displayProjectPathParam(expectedProjectPathParam)

	if hasProjectRequired {
		if !hasProjectPath {
			return fmt.Errorf("route %s requires path parameter {project_id} or {projectId} when %s is configured", pattern, PolicyTypeProjectRequired)
		}
		if len(projectMatchPolicies) == 0 {
			return fmt.Errorf("%s requires %s", PolicyTypeProjectRequired, PolicyTypeProjectMatchFromPath)
		}
	}

	if len(projectMatchPolicies) > 0 && !hasProjectRequired {
		return fmt.Errorf("%s requires %s", PolicyTypeProjectMatchFromPath, PolicyTypeProjectRequired)
	}

	if hasProjectPath {
		if !hasProjectRequired {
			return fmt.Errorf("route %s requires %s", pattern, PolicyTypeProjectRequired)
		}
		if len(projectMatchPolicies) == 0 {
			return fmt.Errorf("route %s requires %s", pattern, PolicyTypeProjectMatchFromPath)
		}
		for _, projectMatch := range projectMatchPolicies {
			normalizedProjectPathParam := normalizePathParam(projectMatch.ProjectPathParam)
			if !isProjectPathParam(normalizedProjectPathParam) {
				return fmt.Errorf("%s for route %s must use path param %q or %q", PolicyTypeProjectMatchFromPath, pattern, projectIDParam, projectIDParamAlias)
			}
			if expectedProjectPathParam != "" && normalizedProjectPathParam != expectedProjectPathParam {
				return fmt.Errorf("%s for route %s must use path param %q", PolicyTypeProjectMatchFromPath, pattern, expectedProjectPathParamDisplay)
			}
		}
	}

	return nil
}

func validateResolverRules(metas []Metadata) error {
	hasResolver := hasPolicyType(metas, PolicyTypeResolvePermissions)
	hasProjectRequired := hasPolicyType(metas, PolicyTypeProjectRequired)
	hasRBAC := hasPolicyType(metas, PolicyTypeRequirePermission) ||
		hasPolicyType(metas, PolicyTypeRequireAnyPermission) ||
		hasPolicyType(metas, PolicyTypeRequireAllPermissions)

	if hasResolver && !hasProjectRequired {
		return fmt.Errorf("%s requires %s", PolicyTypeResolvePermissions, PolicyTypeProjectRequired)
	}

	if hasRBAC && !hasResolver {
		return fmt.Errorf("%s is required when RBAC policies are configured", PolicyTypeResolvePermissions)
	}

	return nil
}

func validateCacheSafety(metas []Metadata) error {
	if !hasPolicyType(metas, PolicyTypeAuthRequired) {
		return nil
	}

	cacheReadPolicies := findPolicies(metas, PolicyTypeCacheRead)
	for _, cacheRead := range cacheReadPolicies {
		if cacheRead.CacheRead.SharedAuthenticated {
			continue
		}
		if !cacheRead.CacheRead.VaryByUserID && !cacheRead.CacheRead.VaryByProjectID {
			return fmt.Errorf("%s on authenticated routes requires VaryBy.UserID or VaryBy.ProjectID unless SharedAuthenticated is true", PolicyTypeCacheRead)
		}
	}

	return nil
}

func hasPolicyType(metas []Metadata, policyType PolicyType) bool {
	for _, meta := range metas {
		if meta.Type == policyType {
			return true
		}
	}
	return false
}

func findPolicies(metas []Metadata, policyType PolicyType) []Metadata {
	matches := make([]Metadata, 0, len(metas))
	for _, meta := range metas {
		if meta.Type == policyType {
			matches = append(matches, meta)
		}
	}
	return matches
}

func policyOrderStage(policyType PolicyType) int {
	switch policyType {
	case PolicyTypeAuthRequired:
		return 1
	case PolicyTypeProjectRequired, PolicyTypeProjectMatchFromPath:
		return 2
	case PolicyTypeResolvePermissions:
		return 3
	case PolicyTypeRequirePermission, PolicyTypeRequireAnyPermission, PolicyTypeRequireAllPermissions:
		return 4
	case PolicyTypeRateLimit:
		return 5
	case PolicyTypeCacheRead, PolicyTypeCacheInvalidate:
		return 6
	case PolicyTypeCacheControl:
		return 7
	default:
		return 0
	}
}

const (
	projectIDParam      = "project_id"
	projectIDParamAlias = "projectId"
	projectIDParamCamel = "projectid"
)

func patternContainsProjectID(pattern string) bool {
	loweredPattern := strings.ToLower(strings.TrimSpace(pattern))
	return strings.Contains(loweredPattern, "{"+projectIDParam+"}") || strings.Contains(loweredPattern, "{"+projectIDParamCamel+"}")
}

func normalizePathParam(param string) string {
	return strings.ToLower(strings.TrimSpace(param))
}

func routeProjectPathParam(pattern string) string {
	loweredPattern := strings.ToLower(strings.TrimSpace(pattern))
	switch {
	case strings.Contains(loweredPattern, "{"+projectIDParam+"}"):
		return projectIDParam
	case strings.Contains(loweredPattern, "{"+projectIDParamCamel+"}"):
		return projectIDParamCamel
	default:
		return ""
	}
}

func isProjectPathParam(pathParam string) bool {
	normalizedPathParam := normalizePathParam(pathParam)
	return normalizedPathParam == projectIDParam || normalizedPathParam == projectIDParamCamel
}

func displayProjectPathParam(pathParam string) string {
	if pathParam == projectIDParamCamel {
		return projectIDParamAlias
	}
	return projectIDParam
}
