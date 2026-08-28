package replay_tampered_event

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/benzhi/city-tree-release/internal/application"
	"github.com/benzhi/city-tree-release/internal/persistence"
)

func TestReplayRejectsTamperedEventHash(t *testing.T) {
	dir := t.TempDir()
	store, err := persistence.Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	app := application.New(store)
	if _, err := app.Create(application.CreateBatchInput{
		Location: "人民公园", CollectionWindow: "上午", Species: "悬铃木",
		Collector: "甲", Role: "collector", IdempotencyKey: "tamper-create",
	}); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(dir, "events.jsonl")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	line := bytes.TrimSpace(data)
	var event map[string]any
	if err := json.Unmarshal(line, &event); err != nil {
		t.Fatal(err)
	}
	payload, ok := event["payload"].(map[string]any)
	if !ok {
		t.Fatal("事件载荷格式异常")
	}
	payload["location"] = "被篡改的地点"
	tampered, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(tampered, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := persistence.Open(dir); err == nil {
		t.Fatal("篡改事件未被启动恢复拒绝")
	}
}
