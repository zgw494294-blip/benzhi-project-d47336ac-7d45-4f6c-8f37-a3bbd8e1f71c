package crossoperationidempotency

import (
	"errors"
	"testing"

	"github.com/benzhi/city-tree-release/internal/application"
	"github.com/benzhi/city-tree-release/internal/domain"
	"github.com/benzhi/city-tree-release/internal/persistence"
)

func TestIdempotencyKeyCannotReplayDifferentOperation(t *testing.T) {
	store, err := persistence.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	app := application.New(store)
	batch, err := app.Create(application.CreateBatchInput{Location: "公园", CollectionWindow: "上午", Species: "榆树", Collector: "甲", Role: "collector"})
	if err != nil {
		t.Fatal(err)
	}
	batch, err = app.AddEvidence(batch.BatchID, application.EvidenceInput{Role: "collector", ExpectedVersion: batch.Version, IdempotencyKey: "shared-key", SampleNumber: "S-1", Grid: "G", PhotoDigest: "P", Environment: map[string]float64{"humidity": 50}})
	if err != nil {
		t.Fatal(err)
	}
	checked, err := app.Screen(batch.BatchID, application.ScreeningInput{Role: "expert", ExpectedVersion: batch.Version, IdempotencyKey: "shared-key"})
	if err != nil && !errors.Is(err, application.ErrConflict) {
		t.Fatalf("跨操作重用 key 返回非冲突错误: %v", err)
	}
	if err == nil && (checked.Status != domain.StatusScreened || checked.EvidenceCheckedAt == nil) {
		t.Fatalf("TestIdempotencyKeyCannotReplayDifferentOperation: screening 被旧 evidence 响应替代: status=%s", checked.Status)
	}
}
