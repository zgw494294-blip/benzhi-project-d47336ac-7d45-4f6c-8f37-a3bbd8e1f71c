package domain

import (
	"fmt"
	"strings"
	"time"
)

type ReviewDecision struct {
	Risk                  string   `json:"risk"`
	RequiresRectification bool     `json:"requiresRectification"`
	Issues                []string `json:"issues"`
	Guidance              []string `json:"guidance"`
}

func DecideReview(conclusion string, risk string, issues []string) ReviewDecision {
	d := ReviewDecision{Risk: risk, Issues: append([]string(nil), issues...)}
	if risk == "高" {
		d.RequiresRectification = true
		d.Guidance = append(d.Guidance, "隔离疑似病株", "扩大巡查范围", "复采并留存影像")
	}
	if risk == "中" {
		d.RequiresRectification = true
		d.Guidance = append(d.Guidance, "补充环境读数", "三日内复查")
	}
	if len(d.Issues) > 0 {
		d.RequiresRectification = true
	}
	if strings.TrimSpace(conclusion) == "" {
		d.RequiresRectification = true
		d.Issues = append(d.Issues, "鉴定结论为空")
	}
	return d
}
func ValidateReview(r ExpertReview) error {
	if strings.TrimSpace(r.Conclusion) == "" {
		return fmt.Errorf("鉴定结论不能为空")
	}
	if r.Reviewer == "" {
		return fmt.Errorf("鉴定员不能为空")
	}
	if r.RiskLevel != "低" && r.RiskLevel != "中" && r.RiskLevel != "高" {
		return fmt.Errorf("风险等级无效")
	}
	return nil
}
func ReviewAge(r ExpertReview) time.Duration {
	if r.ReviewedAt.IsZero() {
		return 0
	}
	return time.Since(r.ReviewedAt)
}
func IssuesClosed(b SampleBatch) bool {
	return len(b.OpenIssues) == 0 && (b.Review == nil || len(b.Review.Issues) == 0 || b.Status == StatusReview || b.Status == StatusReleased)
}
