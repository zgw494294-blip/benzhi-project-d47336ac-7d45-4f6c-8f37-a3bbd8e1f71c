package persistence

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/benzhi/city-tree-release/internal/domain"
)

type Store struct {
	mu          sync.RWMutex
	dir         string
	batches     map[string]domain.SampleBatch
	events      map[string][]domain.AuditEvent
	ledger      []domain.AuditEvent
	idempotency map[string]json.RawMessage
	seq         int
	lastHash    string
}

type snapshot struct {
	SchemaVersion int                           `json:"schemaVersion"`
	Batches       map[string]domain.SampleBatch `json:"batches"`
	Idempotency   map[string]json.RawMessage    `json:"idempotency"`
	Sequence      int                           `json:"sequence"`
	LastHash      string                        `json:"lastHash"`
}

func Open(dir string) (*Store, error) {
	if dir == "" {
		dir = ".refill-data"
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	s := &Store{dir: dir, batches: map[string]domain.SampleBatch{}, events: map[string][]domain.AuditEvent{}, idempotency: map[string]json.RawMessage{}}
	if err := s.loadSnapshot(); err != nil {
		return nil, err
	}
	if err := s.replayEvents(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) Get(id string) (domain.SampleBatch, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	b, ok := s.batches[id]
	return b, ok
}
func (s *Store) List() []domain.SampleBatch {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]domain.SampleBatch, 0, len(s.batches))
	for _, b := range s.batches {
		out = append(out, b)
	}
	return out
}
func (s *Store) Events(id string) []domain.AuditEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.events[id]
}
func (s *Store) AllEvents() []domain.AuditEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]domain.AuditEvent(nil), s.ledger...)
}

func (s *Store) Save(id string, batch domain.SampleBatch, event domain.Event, idem string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if idem != "" {
		if _, ok := s.idempotency[idem]; ok {
			return nil
		}
	}
	payload, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("事件载荷无法编码: %w", err)
	}
	response, err := json.Marshal(batch)
	if err != nil {
		return fmt.Errorf("批次无法编码: %w", err)
	}
	nextSequence := s.seq + 1
	prev := s.lastHash
	base := fmt.Sprintf("%d|%s|%s|%s|%s", nextSequence, id, event.Type, string(payload), prev)
	h := sha256.Sum256([]byte(base))
	hash := hex.EncodeToString(h[:])
	audit := domain.AuditEvent{EventID: domain.NewID("event"), BatchID: id, Sequence: nextSequence, EventType: event.Type, Payload: event.Payload, PrevHash: prev, Hash: hash, OccurredAt: time.Now().UTC(), SchemaVersion: 1, HashPayload: append(json.RawMessage(nil), payload...)}
	if err := s.appendAudit(audit); err != nil {
		return err
	}
	s.seq = nextSequence
	s.lastHash = hash
	s.batches[id] = batch
	s.events[id] = append(s.events[id], audit)
	s.ledger = append(s.ledger, audit)
	if idem != "" {
		s.idempotency[idem] = json.RawMessage(response)
	}
	return s.writeSnapshot()
}

func (s *Store) Idempotent(idem string) (json.RawMessage, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.idempotency[idem]
	return append(json.RawMessage(nil), v...), ok
}

func (s *Store) appendAudit(a domain.AuditEvent) error {
	f, err := os.OpenFile(filepath.Join(s.dir, "events.jsonl"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	return enc.Encode(a)
}

func (s *Store) writeSnapshot() error {
	tmp := filepath.Join(s.dir, "snapshot.tmp")
	final := filepath.Join(s.dir, "snapshot.json")
	b, err := json.MarshalIndent(snapshot{SchemaVersion: 1, Batches: s.batches, Idempotency: s.idempotency, Sequence: s.seq, LastHash: s.lastHash}, "", "  ")
	if err != nil {
		return err
	}
	if err = os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, final)
}

func (s *Store) loadSnapshot() error {
	b, err := os.ReadFile(filepath.Join(s.dir, "snapshot.json"))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	var snap snapshot
	if err := json.Unmarshal(b, &snap); err != nil {
		return fmt.Errorf("快照损坏: %w", err)
	}
	if snap.SchemaVersion != 0 && snap.SchemaVersion != 1 {
		return fmt.Errorf("不支持的快照版本")
	}
	if snap.Batches != nil {
		s.batches = snap.Batches
	}
	if snap.Idempotency != nil {
		s.idempotency = snap.Idempotency
	}
	s.seq = snap.Sequence
	s.lastHash = snap.LastHash
	return nil
}

func (s *Store) replayEvents() error {
	f, err := os.Open(filepath.Join(s.dir, "events.jsonl"))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	defer f.Close()
	scan := bufio.NewScanner(f)
	expectedPrev := ""
	count := 0
	for scan.Scan() {
		var a domain.AuditEvent
		if err := json.Unmarshal(scan.Bytes(), &a); err != nil {
			return fmt.Errorf("事件账本损坏: %w", err)
		}
		var raw struct {
			Payload json.RawMessage `json:"payload"`
		}
		if err := json.Unmarshal(scan.Bytes(), &raw); err != nil {
			return fmt.Errorf("事件载荷损坏: %w", err)
		}
		a.HashPayload = append(json.RawMessage(nil), raw.Payload...)
		if a.PrevHash != expectedPrev {
			return fmt.Errorf("事件摘要链断裂")
		}
		expectedPrev = a.Hash
		count++
		s.events[a.BatchID] = append(s.events[a.BatchID], a)
		s.ledger = append(s.ledger, a)
	}
	if err := scan.Err(); err != nil {
		return err
	}
	if count > s.seq {
		s.seq = count
	}
	if expectedPrev != "" {
		s.lastHash = expectedPrev
	}
	return nil
}
