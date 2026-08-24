package application

import (
	"fmt"
	"github.com/benzhi/city-tree-release/internal/domain"
	"github.com/benzhi/city-tree-release/internal/persistence"
	"sort"
	"strings"
	"time"
)

type BatchListInput struct {
	Status   string
	Query    string
	Role     string
	Page     int
	PageSize int
}

type BatchListResult struct {
	Total    int                   `json:"total"`
	Page     int                   `json:"page"`
	PageSize int                   `json:"pageSize"`
	HasNext  bool                  `json:"hasNext"`
	Batches  []domain.BatchSummary `json:"batches"`
}

func (s *Service) ListBatches(input BatchListInput) (BatchListResult, error) {
	status, ok := domain.ParseStatus(input.Status)
	if !ok {
		return BatchListResult{}, fmt.Errorf("status 参数无效")
	}
	query := strings.TrimSpace(input.Query)
	if len([]rune(query)) > 200 {
		return BatchListResult{}, fmt.Errorf("q 参数不能超过 200 个字符")
	}
	var role domain.Role
	if strings.TrimSpace(input.Role) != "" {
		role = domain.ParseRole(input.Role)
		if role == "" {
			return BatchListResult{}, fmt.Errorf("role 参数无效")
		}
	}
	page, pageSize := input.Page, input.PageSize
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 20
	}
	if page < 1 {
		return BatchListResult{}, fmt.Errorf("page 必须大于 0")
	}
	if pageSize < 1 || pageSize > 100 {
		return BatchListResult{}, fmt.Errorf("pageSize 必须在 1 到 100 之间")
	}
	result := s.repo.Query(persistence.BatchQuery{Status: status, Term: query, Role: role, Page: page, PageSize: pageSize})
	summaries := make([]domain.BatchSummary, 0, len(result.Batches))
	for _, batch := range result.Batches {
		summaries = append(summaries, domain.Summarize(batch))
	}
	return BatchListResult{Total: result.Total, Page: result.Page, PageSize: result.PageSize, HasNext: result.HasNext, Batches: summaries}, nil
}

type WorkbenchView struct {
	GeneratedAt time.Time             `json:"generatedAt"`
	Total       int                   `json:"total"`
	Pending     int                   `json:"pending"`
	Released    int                   `json:"released"`
	HighRisk    int                   `json:"highRisk"`
	Batches     []domain.BatchSummary `json:"batches"`
}

func (s *Service) Workbench() WorkbenchView {
	batches := domain.SortBatchesByUpdate(s.repo.List())
	view := WorkbenchView{GeneratedAt: time.Now().UTC(), Total: len(batches)}
	for _, b := range batches {
		sum := domain.Summarize(b)
		view.Batches = append(view.Batches, sum)
		if b.Status != domain.StatusReleased {
			view.Pending++
		} else {
			view.Released++
		}
		if domain.RiskWeight(sum.RiskLevel) >= 3 {
			view.HighRisk++
		}
	}
	sort.Slice(view.Batches, func(i, j int) bool { return view.Batches[i].Status < view.Batches[j].Status })
	return view
}
func (s *Service) Timeline(id string) ([]domain.AuditEvent, error) { return s.Events(id) }
func (s *Service) PendingByRole(role domain.Role) []domain.SampleBatch {
	out := []domain.SampleBatch{}
	for _, b := range s.repo.List() {
		if role == domain.RoleCollector && (b.Status == domain.StatusDraft || b.Status == domain.StatusRectification) {
			out = append(out, b)
		}
		if role == domain.RoleExpert && (b.Status == domain.StatusEvidence || b.Status == domain.StatusScreened) {
			out = append(out, b)
		}
		if role == domain.RoleReviewer && b.Status == domain.StatusReview {
			out = append(out, b)
		}
	}
	return out
}
