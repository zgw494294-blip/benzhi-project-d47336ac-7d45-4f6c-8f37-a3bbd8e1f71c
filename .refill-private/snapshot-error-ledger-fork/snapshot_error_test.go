package snapshoterrorledgerfork

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/benzhi/city-tree-release/internal/application"
	"github.com/benzhi/city-tree-release/internal/persistence"
)

func TestRetryAfterSnapshotFailureKeepsLedgerReopenable(t *testing.T) {
	dir := t.TempDir()
	store, err := persistence.Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	tmp := filepath.Join(dir, "snapshot.tmp")
	if err = os.Mkdir(tmp, 0o755); err != nil {
		t.Fatal(err)
	}
	app := application.New(store)
	input := application.CreateBatchInput{Location: "公园", CollectionWindow: "上午", Species: "榆树", Collector: "甲", Role: "collector", IdempotencyKey: "retry-create"}
	_, err = app.Create(input)
	if err == nil {
		t.Fatal("预期快照写入失败")
	}
	if err = os.Remove(tmp); err != nil {
		t.Fatal(err)
	}
	retriedBatch, err := app.Create(input)
	if err != nil {
		t.Fatalf("清除故障后重试失败: %v", err)
	}
	reopened, err := persistence.Open(dir)
	if err != nil {
		t.Fatalf("TestRetryAfterSnapshotFailureKeepsLedgerReopenable: 重试后账本无法重开: %v", err)
	}
	if _, ok := reopened.Get(retriedBatch.BatchID); !ok {
		t.Fatalf("TestRetryAfterSnapshotFailureKeepsLedgerReopenable: 重试报告成功但重启后批次丢失: %s", retriedBatch.BatchID)
	}
}
