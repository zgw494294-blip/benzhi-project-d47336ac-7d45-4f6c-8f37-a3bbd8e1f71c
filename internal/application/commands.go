package application

import (
	"fmt"
	"github.com/benzhi/city-tree-release/internal/domain"
	"strings"
)

type CommandResult struct {
	Batch     domain.SampleBatch `json:"batch"`
	EventType string             `json:"eventType"`
	Message   string             `json:"message"`
}

func NormalizeIdempotency(raw string) string { return strings.TrimSpace(raw) }
func RequireExpectedVersion(v int) error {
	if v < 0 {
		return fmt.Errorf("expectedVersion 不能为负数")
	}
	return nil
}
func ValidateCreateCommand(in CreateBatchInput) error {
	if !domain.RoleCan(domain.ParseRole(in.Role), "create") {
		return fmt.Errorf("创建命令需要采集员角色")
	}
	if err := RequireExpectedVersion(valueOrZero(in.ExpectedVersion)); err != nil {
		return err
	}
	return nil
}
func valueOrZero(v *int) int {
	if v == nil {
		return 0
	}
	return *v
}
func DescribeStatus(b domain.SampleBatch) string {
	return fmt.Sprintf("批次 %s 当前处于%s，版本 %d", b.BatchID, domain.StatusLabel(b.Status), b.Version)
}
func ActionForStatus(s domain.BatchStatus) string {
	switch s {
	case domain.StatusDraft:
		return "提交现场证据"
	case domain.StatusEvidence:
		return "执行完整性检查"
	case domain.StatusScreened:
		return "提交专家鉴定"
	case domain.StatusRectification:
		return "上传整改证据"
	case domain.StatusReview:
		return "冻结并签发凭据"
	case domain.StatusReleased:
		return "查看凭据"
	}
	return ""
}
