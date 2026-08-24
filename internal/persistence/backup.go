package persistence

import (
	"io"
	"os"
	"path/filepath"
	"time"
)

func (s *Store) Backup(w io.Writer) error {
	data, err := os.ReadFile(filepath.Join(s.dir, "snapshot.json"))
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}
func (s *Store) BackupPath() string { return filepath.Join(s.dir, "snapshot.json") }
func (s *Store) BackupTimestamp() (time.Time, error) {
	info, err := os.Stat(s.BackupPath())
	if err != nil {
		return time.Time{}, err
	}
	return info.ModTime(), nil
}
func (s *Store) EnsureFiles() error {
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return err
	}
	for _, name := range []string{"events.jsonl"} {
		f, err := os.OpenFile(filepath.Join(s.dir, name), os.O_CREATE, 0o644)
		if err != nil {
			return err
		}
		f.Close()
	}
	return nil
}
