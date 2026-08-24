package persistence

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/benzhi/city-tree-release/internal/domain"
)

func EventHash(e domain.AuditEvent) string {
	payload := e.HashPayload
	if len(payload) == 0 {
		payload, _ = json.Marshal(e.Payload)
	}
	base := fmt.Sprintf("%d|%s|%s|%s|%s", e.Sequence, e.BatchID, e.EventType, string(payload), e.PrevHash)
	sum := sha256.Sum256([]byte(base))
	return hex.EncodeToString(sum[:])
}
func VerifyEvent(e domain.AuditEvent) error {
	if e.SchemaVersion != 0 && e.SchemaVersion != 1 {
		return fmt.Errorf("事件 schemaVersion 无效")
	}
	if e.Hash == "" {
		return fmt.Errorf("事件摘要为空")
	}
	if EventHash(e) != e.Hash {
		return fmt.Errorf("事件 %s 摘要不匹配", e.EventID)
	}
	return nil
}
func VerifyChain(events []domain.AuditEvent) error {
	prev := ""
	for i, e := range events {
		if e.Sequence != i+1 {
			return fmt.Errorf("事件序号不连续：期望 %d，实际 %d", i+1, e.Sequence)
		}
		if e.PrevHash != prev {
			return fmt.Errorf("事件 %d 前置摘要不匹配", e.Sequence)
		}
		if err := VerifyEvent(e); err != nil {
			return err
		}
		prev = e.Hash
	}
	return nil
}
func ChainDigest(events []domain.AuditEvent) string {
	if len(events) == 0 {
		return ""
	}
	return events[len(events)-1].Hash
}
