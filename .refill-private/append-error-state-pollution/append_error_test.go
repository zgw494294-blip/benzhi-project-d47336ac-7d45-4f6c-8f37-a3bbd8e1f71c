package appenderrorstatepollution

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/benzhi/city-tree-release/internal/application"
	"github.com/benzhi/city-tree-release/internal/persistence"
)

func TestFailedScreenDoesNotPolluteStoredEvidence(t *testing.T) {
	dir := t.TempDir()
	store, err := persistence.Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	app := application.New(store)
	batch, err := app.Create(application.CreateBatchInput{Location: "公园", CollectionWindow: "上午", Species: "榆树", Collector: "甲", Role: "collector"})
	if err != nil {
		t.Fatal(err)
	}
	batch, err = app.AddEvidence(batch.BatchID, application.EvidenceInput{Role: "collector", ExpectedVersion: batch.Version, SampleNumber: "S-1", Grid: "G", PhotoDigest: "P", Environment: map[string]float64{"humidity": 50}})
	if err != nil {
		t.Fatal(err)
	}
	ledger := filepath.Join(dir, "events.jsonl")
	if err = os.Remove(ledger); err != nil {
		t.Fatal(err)
	}
	if err = os.Mkdir(ledger, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err = app.Screen(batch.BatchID, application.ScreeningInput{Role: "expert", ExpectedVersion: batch.Version, IdempotencyKey: "screen-fails"}); err == nil {
		t.Fatal("预期审计追加失败")
	}
	persisted, ok := store.Get(batch.BatchID)
	if !ok {
		t.Fatal("批次意外消失")
	}
	if persisted.Evidence[0].Check != nil || persisted.Evidence[0].Integrity != "待检查" {
		t.Fatalf("TestFailedScreenDoesNotPolluteStoredEvidence: 失败请求污染已存证据: integrity=%s check=%+v", persisted.Evidence[0].Integrity, persisted.Evidence[0].Check)
	}
}
