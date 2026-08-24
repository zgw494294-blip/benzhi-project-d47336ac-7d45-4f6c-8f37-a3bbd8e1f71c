package persistence

import (
	"github.com/benzhi/city-tree-release/internal/domain"
	"sort"
	"strings"
)

type BatchQuery struct {
	Status   domain.BatchStatus
	Term     string
	Role     domain.Role
	Page     int
	PageSize int
}

type BatchPage struct {
	Batches  []domain.SampleBatch
	Total    int
	Page     int
	PageSize int
	HasNext  bool
}

func (s *Store) Query(query BatchQuery) BatchPage {
	batches := Search(FilterByStatus(s.List(), query.Status), query.Term)
	if query.Role != "" && query.Role != domain.RoleAdmin {
		allowed := domain.PendingStatuses(query.Role)
		filtered := make([]domain.SampleBatch, 0, len(batches))
		for _, batch := range batches {
			for _, status := range allowed {
				if batch.Status == status {
					filtered = append(filtered, batch)
					break
				}
			}
		}
		batches = filtered
	}
	batches = domain.SortBatchesByUpdate(batches)
	page, pageSize := query.Page, query.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	total := len(batches)
	if total == 0 || page-1 > (total-1)/pageSize {
		return BatchPage{Batches: []domain.SampleBatch{}, Total: total, Page: page, PageSize: pageSize, HasNext: false}
	}
	start := (page - 1) * pageSize
	end := start + pageSize
	if end > total {
		end = total
	}
	return BatchPage{Batches: append([]domain.SampleBatch(nil), batches[start:end]...), Total: total, Page: page, PageSize: pageSize, HasNext: end < total}
}

func FilterByStatus(batches []domain.SampleBatch, status domain.BatchStatus) []domain.SampleBatch {
	out := []domain.SampleBatch{}
	for _, b := range batches {
		if status == "" || b.Status == status {
			out = append(out, b)
		}
	}
	return out
}
func Search(batches []domain.SampleBatch, term string) []domain.SampleBatch {
	term = strings.ToLower(strings.TrimSpace(term))
	if term == "" {
		return append([]domain.SampleBatch(nil), batches...)
	}
	out := []domain.SampleBatch{}
	for _, b := range batches {
		hay := strings.ToLower(strings.Join([]string{b.BatchID, b.Location, b.Species, b.Collector, b.SuspectedIssue}, " "))
		if strings.Contains(hay, term) {
			out = append(out, b)
		}
	}
	return out
}
func LatestEvents(events []domain.AuditEvent, limit int) []domain.AuditEvent {
	out := append([]domain.AuditEvent(nil), events...)
	sort.Slice(out, func(i, j int) bool { return out[i].Sequence > out[j].Sequence })
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out
}
func EventCounts(events []domain.AuditEvent) map[string]int {
	out := map[string]int{}
	for _, e := range events {
		out[e.EventType]++
	}
	return out
}
