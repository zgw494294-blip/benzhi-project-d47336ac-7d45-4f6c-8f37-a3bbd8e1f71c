package application

import (
	"github.com/benzhi/city-tree-release/internal/domain"
	"sort"
	"time"
)

type Metrics struct {
	GeneratedAt   time.Time                  `json:"generatedAt"`
	ByStatus      map[domain.BatchStatus]int `json:"byStatus"`
	ByRisk        map[string]int             `json:"byRisk"`
	EvidenceTotal int                        `json:"evidenceTotal"`
	EventTotal    int                        `json:"eventTotal"`
	OldestPending *time.Time                 `json:"oldestPending,omitempty"`
}

func (s *Service) Metrics() Metrics {
	m := Metrics{GeneratedAt: time.Now().UTC(), ByStatus: map[domain.BatchStatus]int{}, ByRisk: map[string]int{}}
	for _, b := range s.repo.List() {
		m.ByStatus[b.Status]++
		m.EvidenceTotal += len(b.Evidence)
		if b.Review != nil {
			m.ByRisk[b.Review.RiskLevel]++
		}
		if b.Status != domain.StatusReleased && (m.OldestPending == nil || b.CreatedAt.Before(*m.OldestPending)) {
			v := b.CreatedAt
			m.OldestPending = &v
		}
		m.EventTotal += len(s.repo.Events(b.BatchID))
	}
	return m
}
func (s *Service) RiskRanking() []domain.SampleBatch {
	out := append([]domain.SampleBatch(nil), s.repo.List()...)
	sort.SliceStable(out, func(i, j int) bool {
		ri, rj := 0, 0
		if out[i].Review != nil {
			ri = domain.RiskWeight(out[i].Review.RiskLevel)
		}
		if out[j].Review != nil {
			rj = domain.RiskWeight(out[j].Review.RiskLevel)
		}
		return ri > rj
	})
	return out
}
func (s *Service) ReleaseRate() float64 {
	b := s.repo.List()
	if len(b) == 0 {
		return 0
	}
	n := 0
	for _, x := range b {
		if x.Status == domain.StatusReleased {
			n++
		}
	}
	return float64(n) / float64(len(b))
}
func (s *Service) EvidenceCoverage() float64 {
	b := s.repo.List()
	if len(b) == 0 {
		return 0
	}
	with := 0
	for _, x := range b {
		if len(x.Evidence) > 0 {
			with++
		}
	}
	return float64(with) / float64(len(b))
}
