package resource_save_state_pollution

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/benzhi/city-tree-release/internal/application"
	"github.com/benzhi/city-tree-release/internal/domain"
	"github.com/benzhi/city-tree-release/internal/persistence"
)

func TestResourceFailureDoesNotPolluteStoredBatch(t *testing.T) {
	dir := t.TempDir()
	store, err := persistence.Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	app := application.New(store)
	batch, err := app.Create(application.CreateBatchInput{
		Location:         "人民公园",
		CollectionWindow: "上午",
		Species:          "悬铃木",
		Collector:        "甲",
		Role:             "collector",
		IdempotencyKey:   "resource-create",
	})
	if err != nil {
		t.Fatal(err)
	}
	before, ok := store.Get(batch.BatchID)
	if !ok {
		t.Fatal("创建后的批次不存在")
	}
	beforeEvents := len(store.Events(batch.BatchID))
	beforeSequence := store.LastSequence()

	eventsPath := filepath.Join(store.DataDir(), "events.jsonl")
	if err := os.Remove(eventsPath); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(eventsPath, 0o755); err != nil {
		t.Fatal(err)
	}

	_, err = app.AddEvidence(batch.BatchID, application.EvidenceInput{
		Role: "collector", ExpectedVersion: batch.Version,
		SampleNumber: "S-1", Grid: "31.2,121.4", PhotoDigest: "sha256:sample",
		Environment: map[string]float64{"humidity": 60}, IdempotencyKey: "resource-evidence",
	})
	if err == nil {
		t.Fatal("事件账本不可写时应返回错误")
	}

	after, ok := store.Get(batch.BatchID)
	if !ok {
		t.Fatal("资源错误后批次丢失")
	}
	if after.Version != before.Version || after.Status != domain.StatusDraft || len(after.Evidence) != len(before.Evidence) {
		t.Fatalf("账本追加失败却污染了批次: before=%+v after=%+v", before, after)
	}
	if got := len(store.Events(batch.BatchID)); got != beforeEvents {
		t.Fatalf("账本追加失败却增加了事件: before=%d after=%d", beforeEvents, got)
	}
	if got := store.LastSequence(); got != beforeSequence {
		t.Fatalf("账本追加失败却推进了序号: before=%d after=%d", beforeSequence, got)
	}
	if _, ok := store.Idempotent("resource-evidence"); ok {
		t.Fatal("账本追加失败却写入了幂等响应")
	}
}
