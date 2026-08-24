package transport

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/benzhi/city-tree-release/internal/application"
)

func pathID(path, prefix string) (string, bool) {
	if !strings.HasPrefix(path, prefix) {
		return "", false
	}
	id := strings.Trim(strings.TrimPrefix(path, prefix), "/")
	if strings.Contains(id, "/") {
		id = strings.Split(id, "/")[0]
	}
	return id, id != ""
}
func requireMethod(w http.ResponseWriter, r *http.Request, method string) bool {
	if r.Method != method {
		http.Error(w, "方法不支持", http.StatusMethodNotAllowed)
		return false
	}
	return true
}
func requestID(r *http.Request) string {
	if x := r.Header.Get("X-Request-ID"); x != "" {
		return x
	}
	return r.Method + " " + r.URL.Path
}
func contentTypeJSON(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
}

func parseBatchListQuery(r *http.Request) (application.BatchListInput, error) {
	values := r.URL.Query()
	input := application.BatchListInput{
		Status:   strings.TrimSpace(values.Get("status")),
		Query:    strings.TrimSpace(values.Get("q")),
		Role:     strings.TrimSpace(values.Get("role")),
		Page:     1,
		PageSize: 20,
	}
	var err error
	if raw := strings.TrimSpace(values.Get("page")); raw != "" {
		input.Page, err = strconv.Atoi(raw)
		if err != nil || input.Page < 1 {
			return input, fmt.Errorf("page 必须是大于 0 的整数")
		}
	}
	if raw := strings.TrimSpace(values.Get("pageSize")); raw != "" {
		input.PageSize, err = strconv.Atoi(raw)
		if err != nil || input.PageSize < 1 || input.PageSize > 100 {
			return input, fmt.Errorf("pageSize 必须是 1 到 100 之间的整数")
		}
	}
	return input, nil
}
