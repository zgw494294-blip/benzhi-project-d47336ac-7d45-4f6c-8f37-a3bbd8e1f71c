package transport

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type ValidationIssue struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func decodeStrict(w http.ResponseWriter, r *http.Request, dst any) bool {
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "请求 JSON 无效", "detail": err.Error()})
		return false
	}
	return true
}
func fieldIssue(field, msg string) ValidationIssue {
	return ValidationIssue{Field: field, Message: msg}
}
func validateText(fields map[string]string) []ValidationIssue {
	out := []ValidationIssue{}
	for k, v := range fields {
		if strings.TrimSpace(v) == "" {
			out = append(out, fieldIssue(k, "不能为空"))
		}
		if len([]rune(v)) > 500 {
			out = append(out, fieldIssue(k, "长度不能超过 500 个字符"))
		}
	}
	return out
}
func writeValidation(w http.ResponseWriter, issues []ValidationIssue) bool {
	if len(issues) == 0 {
		return true
	}
	writeJSON(w, http.StatusBadRequest, map[string]any{"error": "输入校验失败", "issues": issues})
	return false
}
func parsePositiveInt(raw string) (int, error) {
	var n int
	if _, err := fmt.Sscanf(raw, "%d", &n); err != nil || n < 0 {
		return 0, fmt.Errorf("版本号无效")
	}
	return n, nil
}
