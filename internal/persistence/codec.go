package persistence

import (
	"bufio"
	"encoding/json"
	"github.com/benzhi/city-tree-release/internal/domain"
	"io"
)

func EncodeEvent(w io.Writer, e domain.AuditEvent) error { return json.NewEncoder(w).Encode(e) }
func DecodeEvents(r io.Reader) ([]domain.AuditEvent, error) {
	scan := bufio.NewScanner(r)
	out := []domain.AuditEvent{}
	for scan.Scan() {
		var e domain.AuditEvent
		if err := json.Unmarshal(scan.Bytes(), &e); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, scan.Err()
}
func EncodeSnapshot(v snapshot) ([]byte, error) { return json.MarshalIndent(v, "", "  ") }
func DecodeSnapshot(data []byte) (snapshot, error) {
	var s snapshot
	err := json.Unmarshal(data, &s)
	return s, err
}
func EventSchemaVersion(e domain.AuditEvent) int {
	if e.SchemaVersion == 0 {
		return 1
	}
	return e.SchemaVersion
}
