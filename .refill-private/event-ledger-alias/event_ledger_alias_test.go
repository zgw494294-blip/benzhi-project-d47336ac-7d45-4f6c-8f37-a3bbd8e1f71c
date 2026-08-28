package eventledgeralias

import (
	"testing"

	"github.com/benzhi/city-tree-release/internal/application"
	"github.com/benzhi/city-tree-release/internal/persistence"
)

func TestAuditEventsAreSnapshotIsolated(t *testing.T) {
	store, err := persistence.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	app := application.New(store)
	batch, err := app.Create(application.CreateBatchInput{
		Location: "人民公园", CollectionWindow: "上午", Species: "悬铃木",
		Collector: "采集员", Role: "collector", IdempotencyKey: "alias-create",
	})
	if err != nil {
		t.Fatal(err)
	}
	batch, err = app.AddEvidence(batch.BatchID, application.EvidenceInput{
		Role: "collector", ExpectedVersion: batch.Version, SampleNumber: "S-1",
		Grid: "31.2,121.4", PhotoDigest: "sha256:photo",
		Environment: map[string]float64{"humidity": 50}, IdempotencyKey: "alias-evidence",
	})
	if err != nil {
		t.Fatal(err)
	}
	batch, err = app.Screen(batch.BatchID, application.ScreeningInput{
		Role: "expert", ExpectedVersion: batch.Version, IdempotencyKey: "alias-screen",
	})
	if err != nil {
		t.Fatal(err)
	}
	batch, err = app.Review(batch.BatchID, application.ReviewInput{
		Role: "expert", Reviewer: "鉴定员", Conclusion: "轻度叶斑",
		ExpectedVersion: batch.Version, IdempotencyKey: "alias-review",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = app.Release(batch.BatchID, application.ReleaseInput{
		Role: "reviewer", Reviewer: "复核员", Plan: "隔离病株;三日复查",
		ExecutionWindow: "明天上午", ExpectedVersion: batch.Version, IdempotencyKey: "alias-release",
	})
	if err != nil {
		t.Fatal(err)
	}

	view, err := app.Events(batch.BatchID)
	if err != nil {
		t.Fatal(err)
	}
	if len(view) == 0 {
		t.Fatal("审计事件为空")
	}
	view[len(view)-1].EventType = "TamperedRelease"

	verified, err := app.VerifyCertificate(batch.BatchID)
	if err != nil {
		t.Fatal(err)
	}
	if verified["valid"] != true {
		t.Fatalf("修改返回的审计视图不应污染账本: %+v", verified)
	}
}
