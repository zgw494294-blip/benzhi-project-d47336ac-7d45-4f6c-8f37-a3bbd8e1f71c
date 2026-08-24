package persistence

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

func VerifyEvents(events []map[string]any) error {
	prev := ""
	for i, item := range events {
		if item["prevHash"] != prev {
			return fmt.Errorf("第 %d 条事件前置摘要不匹配", i+1)
		}
		payload, _ := json.Marshal(item["payload"])
		base := fmt.Sprintf("%v|%v|%v|%s|%s", item["sequence"], item["batchId"], item["eventType"], string(payload), prev)
		h := sha256.Sum256([]byte(base))
		got := hex.EncodeToString(h[:])
		if item["hash"] != got {
			return fmt.Errorf("第 %d 条事件摘要不匹配", i+1)
		}
		prev = got
	}
	return nil
}
