package transport

import "net/http"

type RouteInfo struct {
	Method  string `json:"method"`
	Path    string `json:"path"`
	Purpose string `json:"purpose"`
}

func RouteCatalog() []RouteInfo {
	return []RouteInfo{{"GET", "/workbench", "浏览器工作台"}, {"GET", "/api/batches", "查询批次目录"}, {"POST", "/api/batches", "建立样本批次"}, {"POST", "/api/batches/{id}/evidence", "提交现场证据"}, {"POST", "/api/batches/{id}/screening", "执行完整性检查"}, {"POST", "/api/batches/{id}/review", "提交专家鉴定"}, {"POST", "/api/batches/{id}/rectification", "提交整改闭环"}, {"POST", "/api/batches/{id}/release", "冻结并签发凭据"}, {"GET", "/api/certificates/{id}/verify", "校验放行凭据"}, {"GET", "/api/dashboard", "工作台汇总"}, {"GET", "/api/metrics", "运行指标"}}
}
func CatalogHandler(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"routes": RouteCatalog()})
}
func RegisterCatalog(mux *http.ServeMux) { mux.HandleFunc("/api/catalog", CatalogHandler) }
func RouteCount() int                    { return len(RouteCatalog()) }
func RouteExists(method, path string) bool {
	for _, x := range RouteCatalog() {
		if x.Method == method && x.Path == path {
			return true
		}
	}
	return false
}
func PublicPaths() []string {
	out := make([]string, 0, len(RouteCatalog()))
	for _, x := range RouteCatalog() {
		out = append(out, x.Path)
	}
	return out
}
func NormalizePath(path string) string {
	if path == "" {
		return "/"
	}
	if path[0] != '/' {
		return "/" + path
	}
	return path
}
func IsJSONRoute(path string) bool {
	path = NormalizePath(path)
	return len(path) >= 5 && path[:5] == "/api/"
}
func IsBrowserRoute(path string) bool { return NormalizePath(path) == "/workbench" }
func RoutePurpose(method, path string) string {
	for _, x := range RouteCatalog() {
		if x.Method == method && x.Path == path {
			return x.Purpose
		}
	}
	return "未知路由"
}
func MethodAllowed(path, method string) bool {
	for _, x := range RouteCatalog() {
		if x.Path == path && x.Method == method {
			return true
		}
	}
	return false
}
func CatalogSummary() map[string]any {
	items := RouteCatalog()
	jsonCount := 0
	for _, x := range items {
		if IsJSONRoute(x.Path) {
			jsonCount++
		}
	}
	return map[string]any{"total": len(items), "json": jsonCount, "browser": len(items) - jsonCount, "paths": PublicPaths()}
}
func RouteMethods(path string) []string {
	out := []string{}
	for _, x := range RouteCatalog() {
		if x.Path == path {
			out = append(out, x.Method)
		}
	}
	return out
}
func RouteInfoFor(method, path string) (RouteInfo, bool) {
	for _, x := range RouteCatalog() {
		if x.Method == method && x.Path == path {
			return x, true
		}
	}
	return RouteInfo{}, false
}
func CatalogVersion() string { return "workbench-routes-v1" }
