package transport

import (
	"embed"
	"encoding/json"
	"errors"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/benzhi/city-tree-release/internal/application"
)

//go:embed web/index.html web/styles.css web/app.js
var webFiles embed.FS

type Server struct {
	app    *application.Service
	logger *log.Logger
}

func New(app *application.Service, logger *log.Logger) *Server {
	return &Server{app: app, logger: logger}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/workbench", s.workbench)
	mux.HandleFunc("/", s.redirectWorkbench)
	mux.HandleFunc("/api/batches", s.batches)
	mux.HandleFunc("/api/batches/", s.batchSubresource)
	mux.HandleFunc("/api/certificates/", s.verifyCertificate)
	mux.HandleFunc("/static/", s.static)
	RegisterDiagnostics(mux, s.app)
	RegisterCatalog(mux)
	return requestLog(mux, s.logger)
}

func (s *Server) workbench(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "仅支持 GET", http.StatusMethodNotAllowed)
		return
	}
	data, _ := webFiles.ReadFile("web/index.html")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(data)
}
func (s *Server) redirectWorkbench(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		http.Redirect(w, r, "/workbench", http.StatusTemporaryRedirect)
		return
	}
	http.NotFound(w, r)
}
func (s *Server) static(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/static/")
	data, err := webFiles.ReadFile("web/" + name)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if strings.HasSuffix(name, ".css") {
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
	}
	if strings.HasSuffix(name, ".js") {
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	}
	_, _ = w.Write(data)
}

func (s *Server) batches(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		input, err := parseBatchListQuery(r)
		if err != nil {
			writeError(w, err)
			return
		}
		result, err := s.app.ListBatches(input)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, result)
	case http.MethodPost:
		var in application.CreateBatchInput
		if !decode(w, r, &in) {
			return
		}
		b, err := s.app.Create(in)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, b)
	default:
		http.Error(w, "方法不支持", http.StatusMethodNotAllowed)
	}
}

func (s *Server) batchSubresource(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/batches/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.NotFound(w, r)
		return
	}
	id := parts[0]
	if len(parts) == 1 && r.Method == http.MethodGet {
		b, err := s.app.Get(id)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, b)
		return
	}
	if len(parts) != 2 || r.Method != http.MethodPost {
		http.Error(w, "路径或方法不支持", http.StatusMethodNotAllowed)
		return
	}
	switch parts[1] {
	case "evidence":
		var in application.EvidenceInput
		if !decode(w, r, &in) {
			return
		}
		b, err := s.app.AddEvidence(id, in)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, b)
	case "screening":
		var in application.ScreeningInput
		if !decode(w, r, &in) {
			return
		}
		b, err := s.app.Screen(id, in)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, b)
	case "review":
		var in application.ReviewInput
		if !decode(w, r, &in) {
			return
		}
		b, err := s.app.Review(id, in)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, b)
	case "rectification":
		var in application.RectificationInput
		if !decode(w, r, &in) {
			return
		}
		b, err := s.app.Rectify(id, in)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, b)
	case "release":
		var in application.ReleaseInput
		if !decode(w, r, &in) {
			return
		}
		b, err := s.app.Release(id, in)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, b)
	case "events":
		events, err := s.app.Events(id)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"events": events})
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) verifyCertificate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "仅支持 GET", http.StatusMethodNotAllowed)
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/certificates/")
	id = strings.TrimSuffix(id, "/verify")
	if id == "" {
		http.NotFound(w, r)
		return
	}
	result, err := s.app.VerifyCertificate(id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}
func decode(w http.ResponseWriter, r *http.Request, dst any) bool {
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		writeError(w, err)
		return false
	}
	return true
}
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func writeError(w http.ResponseWriter, err error) {
	status := http.StatusBadRequest
	if errors.Is(err, application.ErrNotFound) {
		status = http.StatusNotFound
	}
	if errors.Is(err, application.ErrConflict) {
		status = http.StatusConflict
	}
	writeJSON(w, status, map[string]any{"error": err.Error()})
}
func requestLog(next http.Handler, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if logger != nil {
			logger.Printf("%s %s", r.Method, r.URL.Path)
		}
		next.ServeHTTP(w, r)
	})
}

var _ = template.HTMLEscapeString
