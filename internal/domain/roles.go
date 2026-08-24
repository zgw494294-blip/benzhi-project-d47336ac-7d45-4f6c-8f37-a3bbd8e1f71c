package domain

import "strings"

type Role string

const (
	RoleCollector Role = "collector"
	RoleExpert    Role = "expert"
	RoleReviewer  Role = "reviewer"
	RoleAdmin     Role = "管理员"
)

func ParseRole(raw string) Role {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "collector", "采集员":
		return RoleCollector
	case "expert", "鉴定员":
		return RoleExpert
	case "reviewer", "复核员":
		return RoleReviewer
	case "管理员", "admin":
		return RoleAdmin
	}
	return ""
}
func RoleName(r Role) string {
	switch r {
	case RoleCollector:
		return "采集员"
	case RoleExpert:
		return "鉴定员"
	case RoleReviewer:
		return "生态管理复核员"
	case RoleAdmin:
		return "管理员"
	}
	return "未知角色"
}
func RoleCan(r Role, action string) bool {
	if r == RoleAdmin {
		return true
	}
	permissions := map[Role]map[string]bool{RoleCollector: {"create": true, "evidence": true, "rectification": true}, RoleExpert: {"screen": true, "review": true}, RoleReviewer: {"release": true}}
	return permissions[r][action]
}

func PendingStatuses(r Role) []BatchStatus {
	switch r {
	case RoleCollector:
		return []BatchStatus{StatusDraft, StatusRectification}
	case RoleExpert:
		return []BatchStatus{StatusEvidence, StatusScreened}
	case RoleReviewer:
		return []BatchStatus{StatusReview}
	case RoleAdmin:
		return nil
	default:
		return []BatchStatus{}
	}
}
