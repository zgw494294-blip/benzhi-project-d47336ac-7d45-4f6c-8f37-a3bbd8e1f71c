package persistence

import "github.com/benzhi/city-tree-release/internal/domain"

type Repository interface {
	Get(string) (domain.SampleBatch, bool)
	List() []domain.SampleBatch
	Query(BatchQuery) BatchPage
	Events(string) []domain.AuditEvent
	AllEvents() []domain.AuditEvent
	Save(string, domain.SampleBatch, domain.Event, string) error
	Idempotent(string) (IdempotencyRecord, bool)
}
