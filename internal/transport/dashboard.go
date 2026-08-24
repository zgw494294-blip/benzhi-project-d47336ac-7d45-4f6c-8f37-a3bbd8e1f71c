package transport

import (
	"github.com/benzhi/city-tree-release/internal/application"
	"net/http"
)

func DashboardHandler(app *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !requireMethod(w, r, http.MethodGet) {
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"metrics": app.Metrics(), "workbench": app.Workbench()})
	}
}
func MetricsHandler(app *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !requireMethod(w, r, http.MethodGet) {
			return
		}
		m := app.Metrics()
		writeJSON(w, http.StatusOK, m)
	}
}
func ReleaseRateHandler(app *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !requireMethod(w, r, http.MethodGet) {
			return
		}
		writeJSON(w, http.StatusOK, map[string]float64{"releaseRate": app.ReleaseRate(), "evidenceCoverage": app.EvidenceCoverage()})
	}
}
func RegisterDiagnostics(mux *http.ServeMux, app *application.Service) {
	mux.HandleFunc("/api/dashboard", DashboardHandler(app))
	mux.HandleFunc("/api/metrics", MetricsHandler(app))
	mux.HandleFunc("/api/metrics/rates", ReleaseRateHandler(app))
}
