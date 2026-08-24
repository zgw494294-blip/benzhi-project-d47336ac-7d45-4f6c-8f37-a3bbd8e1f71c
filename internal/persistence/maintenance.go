package persistence

import (
	"os"
	"path/filepath"
	"time"
)

func (s *Store) DataDir() string    { return s.dir }
func (s *Store) LastSequence() int  { s.mu.RLock(); defer s.mu.RUnlock(); return s.seq }
func (s *Store) LastDigest() string { s.mu.RLock(); defer s.mu.RUnlock(); return s.lastHash }
func (s *Store) Healthy() bool {
	_, a := os.Stat(filepath.Join(s.dir, "snapshot.json"))
	_, b := os.Stat(filepath.Join(s.dir, "events.jsonl"))
	return a == nil || b == nil
}
func (s *Store) SnapshotAge() (time.Duration, bool) {
	info, err := os.Stat(filepath.Join(s.dir, "snapshot.json"))
	if err != nil {
		return 0, false
	}
	return time.Since(info.ModTime()), true
}
