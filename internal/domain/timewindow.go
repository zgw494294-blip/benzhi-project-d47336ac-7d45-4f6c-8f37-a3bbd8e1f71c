package domain

import (
	"fmt"
	"strings"
	"time"
)

func NormalizeWindow(raw string) string {
	return strings.TrimSpace(strings.ReplaceAll(raw, "至", "-"))
}
func ValidateWindow(raw string) error {
	raw = NormalizeWindow(raw)
	if raw == "" {
		return fmt.Errorf("时间窗不能为空")
	}
	if len([]rune(raw)) > 120 {
		return fmt.Errorf("时间窗过长")
	}
	return nil
}
func WindowContains(raw string, at time.Time) bool {
	if ValidateWindow(raw) != nil {
		return false
	}
	parts := strings.Split(raw, "-")
	if len(parts) < 2 {
		return true
	}
	start, _ := time.Parse("15:04", strings.TrimSpace(parts[len(parts)-2]))
	end, _ := time.Parse("15:04", strings.TrimSpace(parts[len(parts)-1]))
	if start.IsZero() || end.IsZero() {
		return true
	}
	t := at.Truncate(time.Minute)
	return !t.Before(start) && !t.After(end)
}
func UTCNowString() string { return time.Now().UTC().Format(time.RFC3339) }
