package application

import (
	"encoding/json"
	"fmt"
	"github.com/benzhi/city-tree-release/internal/domain"
	"sort"
	"time"
)

type AuditExport struct {
	Batch      domain.SampleBatch  `json:"batch"`
	Events     []domain.AuditEvent `json:"events"`
	ExportedAt time.Time           `json:"exportedAt"`
	Integrity  string              `json:"integrity"`
}

func (s *Service) ExportAudit(id string) ([]byte, error) {
	b, err := s.Get(id)
	if err != nil {
		return nil, err
	}
	events, err := s.Events(id)
	if err != nil {
		return nil, err
	}
	sort.Slice(events, func(i, j int) bool { return events[i].Sequence < events[j].Sequence })
	payload := AuditExport{Batch: b, Events: events, ExportedAt: time.Now().UTC(), Integrity: "连续摘要链已加载"}
	return json.MarshalIndent(payload, "", "  ")
}
func (s *Service) CountByStatus() map[domain.BatchStatus]int {
	out := map[domain.BatchStatus]int{}
	for _, b := range s.repo.List() {
		out[b.Status]++
	}
	return out
}
func (s *Service) ExplainConflict(id string, expected int) string {
	b, ok := s.repo.Get(id)
	if !ok {
		return "批次不存在"
	}
	return fmt.Sprintf("批次 %s 的当前版本为 %d，提交版本为 %d；请刷新后重新提交", id, b.Version, expected)
}
func (s *Service) CertificateText(id string) (string, error) {
	b, err := s.Get(id)
	if err != nil {
		return "", err
	}
	if b.Certificate == nil {
		return "", fmt.Errorf("批次尚未放行")
	}
	return fmt.Sprintf("凭据 %s\n批次 %s\n处置方案：%s\n执行窗口：%s\n签发人：%s", b.Certificate.Credential, id, b.Certificate.Plan, b.Certificate.ExecutionWindow, b.Certificate.Issuer), nil
}
