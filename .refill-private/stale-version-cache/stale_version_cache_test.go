package staleversioncache_test

import (
	"errors"
	"testing"

	"github.com/benzhi/city-tree-release/internal/application"
	"github.com/benzhi/city-tree-release/internal/persistence"
)

func TestStaleExpectedVersionCannotOverwriteEvidence(t *testing.T) {
	store, err := persistence.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	app := application.New(store)
	batch, err := app.Create(application.CreateBatchInput{
		Location: "人民公园", CollectionWindow: "上午", Species: "悬铃木",
		Collector: "采集员甲", Role: "collector", IdempotencyKey: "create-cache-case",
	})
	if err != nil {
		t.Fatal(err)
	}

	first, err := app.AddEvidence(batch.BatchID, application.EvidenceInput{
		SampleNumber: "S-001", Grid: "31.2,121.4", PhotoDigest: "sha256:first",
		Environment: map[string]float64{"humidity": 50}, Role: "collector",
		ExpectedVersion: batch.Version, IdempotencyKey: "evidence-first",
	})
	if err != nil {
		t.Fatal(err)
	}

	_, staleErr := app.AddEvidence(batch.BatchID, application.EvidenceInput{
		SampleNumber: "S-002", Grid: "31.2,121.5", PhotoDigest: "sha256:stale",
		Environment: map[string]float64{"humidity": 60}, Role: "collector",
		ExpectedVersion: batch.Version, IdempotencyKey: "evidence-stale",
	})
	persisted, ok := store.Get(batch.BatchID)
	if !ok {
		t.Fatal("批次意外消失")
	}
	if !errors.Is(staleErr, application.ErrConflict) {
		t.Fatalf("旧 expectedVersion 应返回 ErrConflict，实际 err=%v，持久化证据=%+v", staleErr, persisted.Evidence)
	}
	if persisted.Version != first.Version || len(persisted.Evidence) != 1 || persisted.Evidence[0].SampleNumber != "S-001" {
		t.Fatalf("冲突请求后首条证据必须保持不变: %+v", persisted)
	}
}
