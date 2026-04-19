package routedocgen

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Options struct {
	ModulesRoot     string
	TrackerPath     string
	GuidelinesPath  string
	OutputPath      string
	GeneratedAtTime time.Time
}

type EndpointTracker struct {
	ID             string `json:"id"`
	Category       string `json:"category"`
	Module         string `json:"module"`
	Method         string `json:"method"`
	Path           string `json:"path"`
	SourceFunction string `json:"sourceFunction"`
	Status         string `json:"status"`
	Owner          string `json:"owner"`
}

type Route struct {
	Module        string
	Method        string
	Path          string
	HandlerExpr   string
	HandlerMethod string
	Policies      []PolicyInfo
	ReqType       string
	RespType      string
	QueryParams   []string
	Tracker       *EndpointTracker
}

type PolicyInfo struct {
	Name            string
	Call            string
	PermissionArgs  []string
	CacheRead       *CacheReadInfo
	CacheInvalidate *CacheInvalidateInfo
	CacheControl    *CacheControlInfo
}

type CacheReadInfo struct {
	TTL                string
	AllowAuthenticated string
	TagSpecs           []string
	VaryBy             []string
}

type CacheInvalidateInfo struct {
	TagSpecs []string
}

type CacheControlInfo struct {
	Directives []string
	Vary       []string
}

type HandlerInfo struct {
	ReqType    string
	RespType   string
	QueryNames []string
}

type ModuleModel struct {
	Name        string
	Routes      []Route
	Tracked     int
	Operational int
}

type StructInfo struct {
	Fields []StructField
}

type StructField struct {
	Name     string
	TypeExpr ast.Expr
}

func Generate(opts Options) error {
	if strings.TrimSpace(opts.ModulesRoot) == "" {
		opts.ModulesRoot = filepath.FromSlash("internal/modules")
	}
	if strings.TrimSpace(opts.TrackerPath) == "" {
		opts.TrackerPath = filepath.FromSlash("docs/ProjectBookDocs/endpoint-tracker.json")
	}
	if strings.TrimSpace(opts.GuidelinesPath) == "" {
		opts.GuidelinesPath = filepath.FromSlash("docs/ProjectBookDocs/API-GUIDELINES.md")
	}
	if strings.TrimSpace(opts.OutputPath) == "" {
		opts.OutputPath = filepath.FromSlash("docs/routeDetails.md")
	}
	if opts.GeneratedAtTime.IsZero() {
		opts.GeneratedAtTime = time.Now().UTC()
	}

	trackerMap, err := loadEndpointTracker(opts.TrackerPath)
	if err != nil {
		return err
	}
	examples, err := parseGuidelineSuccessExamples(opts.GuidelinesPath)
	if err != nil {
		return err
	}

	moduleDirs, err := listModuleDirs(opts.ModulesRoot)
	if err != nil {
		return err
	}

	modules := make([]ModuleModel, 0, len(moduleDirs))
	for _, moduleDir := range moduleDirs {
		moduleName := filepath.Base(moduleDir)
		handlerInfo, structMap, err := parseModuleHandlersAndStructs(moduleDir)
		if err != nil {
			return fmt.Errorf("parse module %s handlers: %w", moduleName, err)
		}
		routes, err := parseModuleRoutes(moduleDir)
		if err != nil {
			return fmt.Errorf("parse module %s routes: %w", moduleName, err)
		}

		for i := range routes {
			routes[i].Module = moduleName
			key := routeKey(routes[i].Method, routes[i].Path)
			if t, ok := trackerMap[key]; ok {
				tCopy := t
				routes[i].Tracker = &tCopy
			}
			if hi, ok := handlerInfo[routes[i].HandlerMethod]; ok {
				routes[i].ReqType = hi.ReqType
				routes[i].RespType = hi.RespType
				routes[i].QueryParams = dedupeSorted(hi.QueryNames)
			}
		}

		tracked := 0
		for _, route := range routes {
			if route.Tracker != nil {
				tracked++
			}
		}

		modules = append(modules, ModuleModel{
			Name:        moduleName,
			Routes:      routes,
			Tracked:     tracked,
			Operational: len(routes) - tracked,
		})

		_ = structMap
	}

	sort.Slice(modules, func(i, j int) bool {
		return modules[i].Name < modules[j].Name
	})

	content, err := buildRouteDetailsMarkdown(modules, examples, opts.GeneratedAtTime)
	if err != nil {
		return err
	}

	if err := os.WriteFile(opts.OutputPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write route details: %w", err)
	}
	return nil
}

func listModuleDirs(modulesRoot string) ([]string, error) {
	entries, err := os.ReadDir(modulesRoot)
	if err != nil {
		return nil, fmt.Errorf("read modules root: %w", err)
	}
	dirs := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		routesPath := filepath.Join(modulesRoot, entry.Name(), "routes.go")
		if _, err := os.Stat(routesPath); err == nil {
			dirs = append(dirs, filepath.Join(modulesRoot, entry.Name()))
		}
	}
	return dirs, nil
}

func loadEndpointTracker(path string) (map[string]EndpointTracker, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read endpoint tracker: %w", err)
	}
	var items []EndpointTracker
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, fmt.Errorf("parse endpoint tracker: %w", err)
	}
	out := make(map[string]EndpointTracker, len(items))
	for _, item := range items {
		out[routeKey(item.Method, item.Path)] = item
	}
	return out, nil
}

var routeHeaderRe = regexp.MustCompile(`^####\s+(GET|POST|PUT|PATCH|DELETE)\s+` + "`" + `([^` + "`" + `]+)` + "`" + `$`)

func parseGuidelineSuccessExamples(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open API guidelines: %w", err)
	}
	defer f.Close()

	type section struct {
		method string
		path   string
		lines  []string
	}

	sections := make([]section, 0, 128)
	current := section{}
	flush := func() {
		if current.method == "" || current.path == "" {
			return
		}
		sections = append(sections, current)
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if m := routeHeaderRe.FindStringSubmatch(line); len(m) == 3 {
			flush()
			current = section{method: strings.TrimSpace(m[1]), path: strings.TrimSpace(m[2])}
			continue
		}
		if current.method != "" {
			current.lines = append(current.lines, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan API guidelines: %w", err)
	}
	flush()

	out := make(map[string]string, len(sections))
	for _, sec := range sections {
		example := extractSuccessJSONBlock(sec.lines)
		if example == "" {
			continue
		}
		out[routeKey(sec.method, sec.path)] = example
	}
	return out, nil
}

func extractSuccessJSONBlock(lines []string) string {
	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if !strings.HasPrefix(line, "**Success response:**") {
			continue
		}
		for j := i + 1; j < len(lines); j++ {
			if strings.TrimSpace(lines[j]) != "```json" {
				continue
			}
			end := -1
			for k := j + 1; k < len(lines); k++ {
				if strings.TrimSpace(lines[k]) == "```" {
					end = k
					break
				}
			}
			if end == -1 {
				return ""
			}
			return strings.TrimSpace(strings.Join(lines[j+1:end], "\n"))
		}
	}
	return ""
}

func parseModuleHandlersAndStructs(moduleDir string) (map[string]HandlerInfo, map[string]StructInfo, error) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, moduleDir, func(info os.FileInfo) bool {
		name := strings.ToLower(info.Name())
		return strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go")
	}, parser.SkipObjectResolution)
	if err != nil {
		return nil, nil, err
	}

	handlers := make(map[string]HandlerInfo)
	structs := make(map[string]StructInfo)

	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			ast.Inspect(file, func(node ast.Node) bool {
				switch n := node.(type) {
				case *ast.TypeSpec:
					structType, ok := n.Type.(*ast.StructType)
					if !ok {
						return true
					}
					fields := make([]StructField, 0, len(structType.Fields.List))
					for _, field := range structType.Fields.List {
						jsonName := jsonFieldName(field)
						if jsonName == "" || jsonName == "-" {
							continue
						}
						fields = append(fields, StructField{Name: jsonName, TypeExpr: field.Type})
					}
					structs[n.Name.Name] = StructInfo{Fields: fields}
				case *ast.FuncDecl:
					if n.Recv == nil || len(n.Recv.List) != 1 {
						return true
					}
					if !isHandlerReceiver(n.Recv.List[0].Type) {
						return true
					}
					if n.Type == nil || n.Type.Params == nil || len(n.Type.Params.List) < 2 {
						return true
					}
					reqType := nodeString(fset, n.Type.Params.List[1].Type)
					respType := ""
					if n.Type.Results != nil && len(n.Type.Results.List) > 0 {
						respType = nodeString(fset, n.Type.Results.List[0].Type)
					}
					h := handlers[n.Name.Name]
					h.ReqType = reqType
					h.RespType = respType
					h.QueryNames = dedupeSorted(append(h.QueryNames, extractQueryNamesFromBody(fset, n.Body)...))
					handlers[n.Name.Name] = h
				}
				return true
			})
		}
	}

	return handlers, structs, nil
}

func parseModuleRoutes(moduleDir string) ([]Route, error) {
	routesPath := filepath.Join(moduleDir, "routes.go")
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, routesPath, nil, parser.SkipObjectResolution)
	if err != nil {
		return nil, err
	}

	routes := make([]Route, 0, 64)
	ast.Inspect(file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "Handle" {
			return true
		}
		if len(call.Args) < 3 {
			return true
		}

		method, err := extractMethod(call.Args[0])
		if err != nil {
			return true
		}
		path, err := extractStringLiteral(call.Args[1])
		if err != nil {
			return true
		}
		handlerExpr := nodeString(fset, call.Args[2])
		handlerMethod := extractHandlerMethod(call.Args[2])

		policies := make([]PolicyInfo, 0, len(call.Args)-3)
		for _, arg := range call.Args[3:] {
			policies = append(policies, parsePolicyInfo(fset, arg))
		}

		routes = append(routes, Route{
			Method:        method,
			Path:          path,
			HandlerExpr:   handlerExpr,
			HandlerMethod: handlerMethod,
			Policies:      policies,
		})
		return true
	})

	return dedupeRoutes(routes), nil
}

func dedupeRoutes(routes []Route) []Route {
	type pickedRoute struct {
		route Route
		score int
	}

	order := make([]string, 0, len(routes))
	picked := make(map[string]pickedRoute, len(routes))
	for _, route := range routes {
		key := routeKey(route.Method, route.Path)
		score := routeScore(route)
		if current, ok := picked[key]; ok {
			if score > current.score {
				picked[key] = pickedRoute{route: route, score: score}
			}
			continue
		}
		order = append(order, key)
		picked[key] = pickedRoute{route: route, score: score}
	}

	out := make([]Route, 0, len(order))
	for _, key := range order {
		out = append(out, picked[key].route)
	}
	return out
}

func routeScore(route Route) int {
	score := len(route.Policies) * 10
	if hasPolicy(route.Policies, "CacheRead") {
		score += 200
	}
	if hasPolicy(route.Policies, "CacheControl") {
		score += 150
	}
	if hasPolicy(route.Policies, "CacheInvalidate") {
		score += 120
	}
	if hasPolicy(route.Policies, "RateLimitWithKeyer") || hasPolicy(route.Policies, "RateLimit") {
		score += 80
	}
	if hasPolicy(route.Policies, "RequireJSON") {
		score += 40
	}
	return score
}

func buildRouteDetailsMarkdown(modules []ModuleModel, examples map[string]string, generatedAt time.Time) (string, error) {
	var b strings.Builder

	b.WriteString("# Route Details\n\n")
	b.WriteString("## File Purpose\n")
	b.WriteString("This file is generated from module route registrations and handler contracts. It documents transport contracts, policy chains, RBAC checks, cache behavior, and response examples.\n\n")
	b.WriteString("## Generation\n")
	b.WriteString("- Generator command: go run ./cmd/routedocgen\n")
	b.WriteString("- Route sources: internal/modules/*/routes.go\n")
	b.WriteString("- Handler signatures: internal/modules/*/handler.go\n")
	b.WriteString("- Endpoint IDs: docs/ProjectBookDocs/endpoint-tracker.json\n")
	b.WriteString("- Output schema fallback: docs/ProjectBookDocs/API-GUIDELINES.md\n")
	b.WriteString("- Generated at: ")
	b.WriteString(generatedAt.UTC().Format(time.RFC3339))
	b.WriteString("\n\n")

	b.WriteString("## Module Endpoint Counts\n\n")
	b.WriteString("| Module | Total Endpoints | Tracked (EP) | Operational (OP) |\n")
	b.WriteString("| --- | ---: | ---: | ---: |\n")
	for _, module := range modules {
		b.WriteString(fmt.Sprintf("| %s | %d | %d | %d |\n", module.Name, len(module.Routes), module.Tracked, module.Operational))
	}
	b.WriteString("\n")

	for _, module := range modules {
		b.WriteString(fmt.Sprintf("## Module: %s\n\n", module.Name))
		b.WriteString(fmt.Sprintf("Total endpoints: %d\n\n", len(module.Routes)))

		opIndex := 1
		for _, route := range module.Routes {
			entryID := ""
			businessSource := "n/a"
			statusLabel := "operational"
			if route.Tracker != nil {
				entryID = route.Tracker.ID
				if strings.TrimSpace(route.Tracker.SourceFunction) != "" {
					businessSource = route.Tracker.SourceFunction
				}
				if strings.TrimSpace(route.Tracker.Status) != "" {
					statusLabel = strings.TrimSpace(route.Tracker.Status)
				}
			} else {
				entryID = fmt.Sprintf("OP-%03d", opIndex)
				opIndex++
			}

			b.WriteString(fmt.Sprintf("### %s - %s %s\n\n", entryID, route.Method, route.Path))
			b.WriteString(fmt.Sprintf("- Status: %s\n", statusLabel))
			b.WriteString(fmt.Sprintf("- Endpoint: %s\n", safeString(route.HandlerMethod, "unknown")))
			b.WriteString(fmt.Sprintf("- Handler: %s\n", safeString(route.HandlerExpr, "unknown")))
			b.WriteString(fmt.Sprintf("- Business Logic Source: %s\n", businessSource))

			pathParams := extractPathParams(route.Path)
			if len(pathParams) == 0 {
				b.WriteString("- Path Params: none\n")
			} else {
				b.WriteString("- Path Params: ")
				b.WriteString(strings.Join(pathParams, ", "))
				b.WriteString("\n")
			}
			if len(route.QueryParams) == 0 {
				b.WriteString("- Query Params (inferred): none\n\n")
			} else {
				b.WriteString("- Query Params (inferred): ")
				b.WriteString(strings.Join(route.QueryParams, ", "))
				b.WriteString("\n\n")
			}

			b.WriteString("#### Policies\n")
			if len(route.Policies) == 0 {
				b.WriteString("- none\n\n")
			} else {
				for i, p := range route.Policies {
					b.WriteString(fmt.Sprintf("%d. %s\n", i+1, p.Name))
					if strings.TrimSpace(p.Call) != "" {
						b.WriteString("- Applied Call:\n")
						b.WriteString("```go\n")
						b.WriteString(p.Call)
						b.WriteString("\n```\n")
					}
				}
				b.WriteString("\n")
			}

			b.WriteString("#### RBAC Permissions\n")
			perms := extractPermissionArgs(route.Policies)
			if len(perms) == 0 {
				b.WriteString("- none\n\n")
			} else {
				for _, perm := range perms {
					b.WriteString("- ")
					b.WriteString(perm)
					b.WriteString("\n")
				}
				b.WriteString("\n")
			}

			b.WriteString("#### Cache Details\n")
			cacheRead := findCacheRead(route.Policies)
			cacheInvalidate := findCacheInvalidate(route.Policies)
			cacheControl := findCacheControl(route.Policies)
			b.WriteString(fmt.Sprintf("- Auth Status: %t\n", hasPolicy(route.Policies, "AuthRequired")))
			if cacheRead == nil {
				b.WriteString("- Read Cache: none\n")
			} else {
				b.WriteString("- Read Cache:\n")
				if cacheRead.TTL != "" {
					b.WriteString(fmt.Sprintf("  - TTL: %s\n", cacheRead.TTL))
				}
				if cacheRead.AllowAuthenticated != "" {
					b.WriteString(fmt.Sprintf("  - AllowAuthenticated: %s\n", cacheRead.AllowAuthenticated))
				}
				if len(cacheRead.TagSpecs) > 0 {
					b.WriteString("  - TagSpecs: ")
					b.WriteString(strings.Join(cacheRead.TagSpecs, ", "))
					b.WriteString("\n")
				}
				if len(cacheRead.VaryBy) > 0 {
					b.WriteString("  - VaryBy: ")
					b.WriteString(strings.Join(cacheRead.VaryBy, ", "))
					b.WriteString("\n")
				}
			}
			if cacheControl == nil {
				b.WriteString("- Cache-Control: none\n")
			} else {
				if len(cacheControl.Directives) > 0 {
					b.WriteString("- Cache-Control Directives: ")
					b.WriteString(strings.Join(cacheControl.Directives, ", "))
					b.WriteString("\n")
				}
				if len(cacheControl.Vary) > 0 {
					b.WriteString("- Cache-Control Vary: ")
					b.WriteString(strings.Join(cacheControl.Vary, ", "))
					b.WriteString("\n")
				}
			}
			if cacheInvalidate == nil || len(cacheInvalidate.TagSpecs) == 0 {
				b.WriteString("- Invalidation: none\n\n")
			} else {
				b.WriteString("- Invalidation Tags: ")
				b.WriteString(strings.Join(cacheInvalidate.TagSpecs, ", "))
				b.WriteString("\n\n")
			}

			bodyObj := buildRequestBodySchema(route)
			pathObj := make(map[string]any)
			for _, param := range pathParams {
				pathObj[param] = "string"
			}
			queryObj := make(map[string]any)
			for _, q := range route.QueryParams {
				queryObj[q] = "string"
			}
			input := map[string]any{
				"body":         bodyObj,
				"path_params":  pathObj,
				"query_params": queryObj,
			}

			b.WriteString("#### Input Structure (JSON)\n")
			inputJSON, err := json.MarshalIndent(input, "", "  ")
			if err != nil {
				return "", err
			}
			b.WriteString("```json\n")
			b.Write(inputJSON)
			b.WriteString("\n```\n\n")

			example := examples[routeKey(route.Method, route.Path)]
			if strings.TrimSpace(example) == "" {
				example = fallbackResponseJSON(route)
			}
			b.WriteString("#### Output Structure (JSON)\n")
			b.WriteString("```json\n")
			b.WriteString(strings.TrimSpace(example))
			b.WriteString("\n```\n\n")
		}
	}

	return b.String(), nil
}

func extractMethod(expr ast.Expr) (string, error) {
	switch v := expr.(type) {
	case *ast.SelectorExpr:
		switch v.Sel.Name {
		case "MethodGet":
			return "GET", nil
		case "MethodPost":
			return "POST", nil
		case "MethodPut":
			return "PUT", nil
		case "MethodPatch":
			return "PATCH", nil
		case "MethodDelete":
			return "DELETE", nil
		case "MethodOptions":
			return "OPTIONS", nil
		case "MethodHead":
			return "HEAD", nil
		default:
			return "", fmt.Errorf("unsupported method selector")
		}
	case *ast.BasicLit:
		return extractStringLiteral(v)
	default:
		return "", fmt.Errorf("unsupported method expression")
	}
}

func extractStringLiteral(expr ast.Expr) (string, error) {
	lit, ok := expr.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return "", fmt.Errorf("expected string literal")
	}
	v, err := strconv.Unquote(lit.Value)
	if err != nil {
		return "", err
	}
	return v, nil
}

func extractHandlerMethod(expr ast.Expr) string {
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return ""
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "Adapter" {
		return ""
	}
	if len(call.Args) != 1 {
		return ""
	}
	hSel, ok := call.Args[0].(*ast.SelectorExpr)
	if !ok {
		return ""
	}
	return hSel.Sel.Name
}

func parsePolicyInfo(fset *token.FileSet, expr ast.Expr) PolicyInfo {
	info := PolicyInfo{Name: "unknown", Call: nodeString(fset, expr)}
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return info
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return info
	}
	name := sel.Sel.Name
	info.Name = name

	switch name {
	case "RequirePermission", "RequireAnyPermission", "RequireAllPermissions":
		info.PermissionArgs = collectArgs(fset, call.Args)
	case "CacheRead":
		info.CacheRead = parseCacheReadInfo(fset, call)
	case "CacheInvalidate":
		info.CacheInvalidate = parseCacheInvalidateInfo(fset, call)
	case "CacheControl":
		info.CacheControl = parseCacheControlInfo(fset, call)
	}
	return info
}

func collectArgs(fset *token.FileSet, args []ast.Expr) []string {
	out := make([]string, 0, len(args))
	for _, arg := range args {
		out = append(out, nodeString(fset, arg))
	}
	return out
}

func parseCacheReadInfo(fset *token.FileSet, call *ast.CallExpr) *CacheReadInfo {
	if len(call.Args) < 2 {
		return nil
	}
	cfg, ok := call.Args[1].(*ast.CompositeLit)
	if !ok {
		return nil
	}
	info := &CacheReadInfo{}
	for _, elt := range cfg.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		key := keyName(kv.Key)
		switch key {
		case "TTL":
			info.TTL = nodeString(fset, kv.Value)
		case "AllowAuthenticated":
			info.AllowAuthenticated = nodeString(fset, kv.Value)
		case "TagSpecs":
			info.TagSpecs = parseTagSpecs(fset, kv.Value)
		case "VaryBy":
			info.VaryBy = parseVaryBy(fset, kv.Value)
		}
	}
	return info
}

func parseCacheInvalidateInfo(fset *token.FileSet, call *ast.CallExpr) *CacheInvalidateInfo {
	if len(call.Args) < 2 {
		return nil
	}
	cfg, ok := call.Args[1].(*ast.CompositeLit)
	if !ok {
		return nil
	}
	info := &CacheInvalidateInfo{}
	for _, elt := range cfg.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		if keyName(kv.Key) == "TagSpecs" {
			info.TagSpecs = parseTagSpecs(fset, kv.Value)
		}
	}
	return info
}

func parseCacheControlInfo(fset *token.FileSet, call *ast.CallExpr) *CacheControlInfo {
	if len(call.Args) < 1 {
		return nil
	}
	cfg, ok := call.Args[0].(*ast.CompositeLit)
	if !ok {
		return nil
	}
	info := &CacheControlInfo{}
	for _, elt := range cfg.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		key := keyName(kv.Key)
		value := nodeString(fset, kv.Value)
		if key == "Vary" {
			info.Vary = parseStringSliceLiteral(kv.Value)
			continue
		}
		trimmedValue := strings.TrimSpace(value)
		switch key {
		case "Public":
			if trimmedValue == "true" {
				info.Directives = append(info.Directives, "public")
			}
		case "Private":
			if trimmedValue == "true" {
				info.Directives = append(info.Directives, "private")
			}
		case "NoStore":
			if trimmedValue == "true" {
				info.Directives = append(info.Directives, "no-store")
			}
		case "NoCache":
			if trimmedValue == "true" {
				info.Directives = append(info.Directives, "no-cache")
			}
		case "MustRevalidate":
			if trimmedValue == "true" {
				info.Directives = append(info.Directives, "must-revalidate")
			}
		case "Immutable":
			if trimmedValue == "true" {
				info.Directives = append(info.Directives, "immutable")
			}
		case "MaxAge":
			if trimmedValue != "" && trimmedValue != "0" {
				info.Directives = append(info.Directives, fmt.Sprintf("max-age=%s", trimmedValue))
			}
		case "SharedMaxAge":
			if trimmedValue != "" && trimmedValue != "0" {
				info.Directives = append(info.Directives, fmt.Sprintf("s-maxage=%s", trimmedValue))
			}
		case "StaleWhileRevalidate":
			if trimmedValue != "" && trimmedValue != "0" {
				info.Directives = append(info.Directives, fmt.Sprintf("stale-while-revalidate=%s", trimmedValue))
			}
		case "StaleIfError":
			if trimmedValue != "" && trimmedValue != "0" {
				info.Directives = append(info.Directives, fmt.Sprintf("stale-if-error=%s", trimmedValue))
			}
		default:
			if trimmedValue == "" || trimmedValue == "false" {
				continue
			}
			if trimmedValue == "true" {
				info.Directives = append(info.Directives, strings.ToLower(key))
				continue
			}
			info.Directives = append(info.Directives, fmt.Sprintf("%s=%s", strings.ToLower(key), trimmedValue))
		}
	}
	info.Directives = dedupeSorted(info.Directives)
	info.Vary = dedupeSorted(info.Vary)
	return info
}

func parseTagSpecs(fset *token.FileSet, expr ast.Expr) []string {
	list, ok := expr.(*ast.CompositeLit)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(list.Elts))
	for _, elt := range list.Elts {
		spec, ok := elt.(*ast.CompositeLit)
		if !ok {
			continue
		}
		name := ""
		dims := make([]string, 0, 2)
		for _, specElt := range spec.Elts {
			kv, ok := specElt.(*ast.KeyValueExpr)
			if !ok {
				continue
			}
			key := keyName(kv.Key)
			value := strings.TrimSpace(nodeString(fset, kv.Value))
			switch key {
			case "Name":
				name = trimQuotes(value)
			case "UserID":
				if value == "true" {
					dims = append(dims, "user_id")
				}
			case "ProjectID":
				if value == "true" {
					dims = append(dims, "project_id")
				}
			}
		}
		if name == "" {
			name = nodeString(fset, spec)
		}
		if len(dims) > 0 {
			out = append(out, fmt.Sprintf("%s[%s]", name, strings.Join(dims, ",")))
		} else {
			out = append(out, name)
		}
	}
	return dedupeSorted(out)
}

func parseVaryBy(fset *token.FileSet, expr ast.Expr) []string {
	cfg, ok := expr.(*ast.CompositeLit)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(cfg.Elts))
	for _, elt := range cfg.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		key := keyName(kv.Key)
		value := strings.TrimSpace(nodeString(fset, kv.Value))
		switch key {
		case "QueryParams":
			params := parseStringSliceLiteral(kv.Value)
			if len(params) > 0 {
				out = append(out, "query:"+strings.Join(params, ","))
			}
		default:
			if value == "true" {
				out = append(out, strings.ToLower(key))
			}
		}
	}
	return dedupeSorted(out)
}

func parseStringSliceLiteral(expr ast.Expr) []string {
	list, ok := expr.(*ast.CompositeLit)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(list.Elts))
	for _, elt := range list.Elts {
		lit, ok := elt.(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			continue
		}
		raw, err := strconv.Unquote(lit.Value)
		if err != nil {
			continue
		}
		if strings.TrimSpace(raw) == "" {
			continue
		}
		out = append(out, raw)
	}
	return dedupeSorted(out)
}

func extractPermissionArgs(policies []PolicyInfo) []string {
	out := make([]string, 0, 8)
	for _, p := range policies {
		if p.Name != "RequirePermission" && p.Name != "RequireAnyPermission" && p.Name != "RequireAllPermissions" {
			continue
		}
		for _, arg := range p.PermissionArgs {
			if strings.Contains(arg, "rbac.") {
				out = append(out, strings.TrimSpace(arg))
			}
		}
	}
	return dedupeSorted(out)
}

func findCacheRead(policies []PolicyInfo) *CacheReadInfo {
	for _, p := range policies {
		if p.CacheRead != nil {
			return p.CacheRead
		}
	}
	return nil
}

func findCacheInvalidate(policies []PolicyInfo) *CacheInvalidateInfo {
	for _, p := range policies {
		if p.CacheInvalidate != nil {
			return p.CacheInvalidate
		}
	}
	return nil
}

func findCacheControl(policies []PolicyInfo) *CacheControlInfo {
	for _, p := range policies {
		if p.CacheControl != nil {
			return p.CacheControl
		}
	}
	return nil
}

func hasPolicy(policies []PolicyInfo, name string) bool {
	for _, p := range policies {
		if p.Name == name {
			return true
		}
	}
	return false
}

func extractPathParams(path string) []string {
	segments := strings.Split(path, "/")
	out := make([]string, 0, 4)
	for _, seg := range segments {
		if strings.HasPrefix(seg, "{") && strings.HasSuffix(seg, "}") && len(seg) > 2 {
			out = append(out, seg[1:len(seg)-1])
		}
	}
	return dedupeSorted(out)
}

func buildRequestBodySchema(route Route) map[string]any {
	req := strings.TrimSpace(route.ReqType)
	if req == "" || strings.HasSuffix(req, ".NoBody") || req == "NoBody" {
		return map[string]any{}
	}
	// Non-empty request types are emitted as object placeholders because handlers
	// may use unexported DTOs and type aliases.
	return map[string]any{"_type": req}
}

func fallbackResponseJSON(route Route) string {
	respType := unwrapResultType(route.RespType)
	var data map[string]any
	trimmed := strings.TrimSpace(respType)
	switch {
	case strings.HasPrefix(trimmed, "[]"):
		data = map[string]any{"items": []any{map[string]any{}}}
	case strings.HasPrefix(trimmed, "map["):
		data = map[string]any{}
	case trimmed == "string":
		data = map[string]any{"value": "string"}
	case trimmed == "bool":
		data = map[string]any{"value": true}
	case strings.Contains(trimmed, "NoBody"):
		data = map[string]any{}
	default:
		data = map[string]any{"_type": safeString(trimmed, "object")}
	}

	envelope := map[string]any{
		"success":    true,
		"data":       data,
		"request_id": "req_1234567890",
	}
	raw, err := json.MarshalIndent(envelope, "", "  ")
	if err != nil {
		return "{\n  \"success\": true,\n  \"data\": {},\n  \"request_id\": \"req_1234567890\"\n}"
	}
	return string(raw)
}

func unwrapResultType(respType string) string {
	trimmed := strings.TrimSpace(respType)
	if strings.HasPrefix(trimmed, "httpx.Result[") && strings.HasSuffix(trimmed, "]") {
		return strings.TrimSuffix(strings.TrimPrefix(trimmed, "httpx.Result["), "]")
	}
	return trimmed
}

func extractQueryNamesFromBody(fset *token.FileSet, body *ast.BlockStmt) []string {
	if body == nil {
		return nil
	}
	out := make([]string, 0, 8)
	ast.Inspect(body, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if !strings.HasPrefix(sel.Sel.Name, "Query") {
			return true
		}
		if len(call.Args) == 0 {
			return true
		}
		param, err := extractStringLiteral(call.Args[0])
		if err != nil || strings.TrimSpace(param) == "" {
			return true
		}
		out = append(out, param)
		_ = nodeString(fset, call)
		return true
	})
	return out
}

func isHandlerReceiver(expr ast.Expr) bool {
	star, ok := expr.(*ast.StarExpr)
	if !ok {
		return false
	}
	id, ok := star.X.(*ast.Ident)
	return ok && id.Name == "Handler"
}

func jsonFieldName(field *ast.Field) string {
	if field == nil {
		return ""
	}
	if field.Tag != nil {
		tagValue, err := strconv.Unquote(field.Tag.Value)
		if err == nil {
			jsonTag := reflectJSONTag(tagValue)
			if jsonTag != "" {
				return jsonTag
			}
		}
	}
	if len(field.Names) > 0 {
		return lowerFirst(field.Names[0].Name)
	}
	return ""
}

func reflectJSONTag(tag string) string {
	parts := strings.Split(tag, " ")
	for _, part := range parts {
		if !strings.HasPrefix(part, "json:") {
			continue
		}
		value := strings.TrimPrefix(part, "json:")
		value = strings.Trim(value, "\"")
		if value == "" {
			return ""
		}
		name := strings.Split(value, ",")[0]
		return name
	}
	return ""
}

func keyName(expr ast.Expr) string {
	switch v := expr.(type) {
	case *ast.Ident:
		return v.Name
	case *ast.SelectorExpr:
		return v.Sel.Name
	default:
		return ""
	}
}

func nodeString(fset *token.FileSet, node ast.Node) string {
	if node == nil {
		return ""
	}
	var buf bytes.Buffer
	_ = printer.Fprint(&buf, fset, node)
	return strings.TrimSpace(buf.String())
}

func routeKey(method, path string) string {
	return strings.ToUpper(strings.TrimSpace(method)) + " " + strings.TrimSpace(path)
}

func dedupeSorted(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	set := make(map[string]struct{}, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		set[trimmed] = struct{}{}
	}
	out := make([]string, 0, len(set))
	for item := range set {
		out = append(out, item)
	}
	sort.Strings(out)
	return out
}

func trimQuotes(value string) string {
	trimmed := strings.TrimSpace(value)
	trimmed = strings.Trim(trimmed, "\"")
	return trimmed
}

func lowerFirst(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToLower(s[:1]) + s[1:]
}

func safeString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
