package application

import (
	"strings"
	"testing"

	"github.com/benzhi/city-tree-release/internal/domain"
	"github.com/benzhi/city-tree-release/internal/persistence"
)

func TestServiceLifecycle(t *testing.T) {
	s, e := persistence.Open(t.TempDir())
	if e != nil {
		t.Fatal(e)
	}
	a := New(s)
	b, e := a.Create(CreateBatchInput{Location: "公园", CollectionWindow: "上午", Species: "榆树", Collector: "甲", Role: "collector"})
	if e != nil {
		t.Fatal(e)
	}
	if _, e = a.AddEvidence(b.BatchID, EvidenceInput{Role: "collector", ExpectedVersion: b.Version, SampleNumber: "S", Grid: "G", PhotoDigest: "P", Environment: map[string]float64{"humidity": 50}}); e != nil {
		t.Fatal(e)
	}
}

func TestScreenPersistsPerEvidenceFailureAndIsIdempotent(t *testing.T) {
	store, err := persistence.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	app := New(store)
	batch, err := app.Create(CreateBatchInput{Location: "人民公园", CollectionWindow: "上午", Species: "悬铃木", Collector: "甲", Role: "collector"})
	if err != nil {
		t.Fatal(err)
	}
	batch, err = app.AddEvidence(batch.BatchID, EvidenceInput{Role: "collector", ExpectedVersion: batch.Version, Grid: "31.2,121.4", PhotoDigest: "sha256:x", Environment: map[string]float64{"humidity": 99}})
	if err != nil {
		t.Fatal(err)
	}
	input := ScreeningInput{Role: "expert", ExpectedVersion: batch.Version, IdempotencyKey: "screen-one"}
	checked, err := app.Screen(batch.BatchID, input)
	if err != nil {
		t.Fatal(err)
	}
	if checked.Status != domain.StatusEvidence || checked.Version != batch.Version+1 {
		t.Fatalf("检查后状态=%s 版本=%d", checked.Status, checked.Version)
	}
	if checked.EvidenceCheckedAt == nil || checked.EvidenceScore != 55 || checked.Evidence[0].Integrity != "不通过" || checked.Evidence[0].Check == nil {
		t.Fatalf("检查结果未完整持久化: %+v", checked)
	}
	warnings := strings.Join(checked.Evidence[0].Check.Warnings, "|")
	if !strings.Contains(warnings, "缺少样本编号") || !strings.Contains(warnings, "湿度超出允许范围") {
		t.Fatalf("警告未合并: %s", warnings)
	}
	replayed, err := app.Screen(batch.BatchID, input)
	if err != nil {
		t.Fatal(err)
	}
	if replayed.Version != checked.Version || len(store.Events(batch.BatchID)) != 3 {
		t.Fatal("重复检查不应产生新版本或新事件")
	}
	if _, err = app.Screen(batch.BatchID, ScreeningInput{Role: "collector", ExpectedVersion: checked.Version, IdempotencyKey: "screen-denied"}); err == nil {
		t.Fatal("采集员不应执行完整性检查")
	}
	persisted, _ := store.Get(batch.BatchID)
	if persisted.Version != checked.Version || persisted.Status != domain.StatusEvidence {
		t.Fatal("失败请求改变了批次")
	}
}

func TestBatchListCombinesRoleSearchStatusAndPagination(t *testing.T) {
	store, err := persistence.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	app := New(store)
	first, err := app.Create(CreateBatchInput{Location: "滨江公园", CollectionWindow: "上午", Species: "香樟", SuspectedIssue: "叶斑", Collector: "甲", Role: "collector"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = app.Create(CreateBatchInput{Location: "中央公园", CollectionWindow: "下午", Species: "银杏", SuspectedIssue: "虫孔", Collector: "乙", Role: "collector"}); err != nil {
		t.Fatal(err)
	}
	third, err := app.Create(CreateBatchInput{Location: "街心花园", CollectionWindow: "上午", Species: "梧桐", SuspectedIssue: "叶斑", Collector: "丙", Role: "collector"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = app.AddEvidence(third.BatchID, EvidenceInput{Role: "collector", ExpectedVersion: third.Version, SampleNumber: "S-3", Grid: "G", PhotoDigest: "P", Environment: map[string]float64{"humidity": 60}}); err != nil {
		t.Fatal(err)
	}
	page, err := app.ListBatches(BatchListInput{Role: "collector", Page: 1, PageSize: 1})
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 2 || len(page.Batches) != 1 || !page.HasNext || page.Batches[0].Status != domain.StatusDraft {
		t.Fatalf("采集员分页=%+v", page)
	}
	expert, err := app.ListBatches(BatchListInput{Status: "待检查", Query: "梧桐", Role: "鉴定员", Page: 1, PageSize: 20})
	if err != nil {
		t.Fatal(err)
	}
	if expert.Total != 1 || expert.Batches[0].BatchID != third.BatchID {
		t.Fatalf("组合筛选=%+v", expert)
	}
	empty, err := app.ListBatches(BatchListInput{Query: first.BatchID, Page: 2, PageSize: 20})
	if err != nil {
		t.Fatal(err)
	}
	if empty.Total != 1 || empty.Batches == nil || len(empty.Batches) != 0 {
		t.Fatalf("空页结构=%+v", empty)
	}
}

func TestReleaseFreezesStructuredPlanAndVerifiesAfterRestart(t *testing.T) {
	dir := t.TempDir()
	store, err := persistence.Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	app := New(store)
	batch, err := app.Create(CreateBatchInput{Location: "社区公园", CollectionWindow: "上午", Species: "榆树", Collector: "甲", Role: "collector", IdempotencyKey: "release-create"})
	if err != nil {
		t.Fatal(err)
	}
	batch, err = app.AddEvidence(batch.BatchID, EvidenceInput{Role: "collector", ExpectedVersion: batch.Version, SampleNumber: "S", Grid: "G", PhotoDigest: "P", Environment: map[string]float64{"humidity": 50}, IdempotencyKey: "release-evidence"})
	if err != nil {
		t.Fatal(err)
	}
	batch, err = app.Screen(batch.BatchID, ScreeningInput{Role: "expert", ExpectedVersion: batch.Version, IdempotencyKey: "release-screen"})
	if err != nil {
		t.Fatal(err)
	}
	batch, err = app.Review(batch.BatchID, ReviewInput{Role: "expert", Reviewer: "鉴定员", Conclusion: "轻度叶斑", ExpectedVersion: batch.Version, IdempotencyKey: "release-review"})
	if err != nil {
		t.Fatal(err)
	}
	input := ReleaseInput{Role: "reviewer", Reviewer: "复核员", PlanName: "叶斑处置", Plan: "隔离病株、喷施药剂，三日复查;记录结果", Owner: "绿化一组", ExecutionWindow: "明天上午", ExpectedVersion: batch.Version, IdempotencyKey: "release-final"}
	released, err := app.Release(batch.BatchID, input)
	if err != nil {
		t.Fatal(err)
	}
	if released.Certificate == nil || released.Certificate.FreezeVersion != "v2" || len(released.Certificate.Disposition.Steps) != 4 || released.Certificate.Disposition.Owner != "绿化一组" {
		t.Fatalf("结构化凭据=%+v", released.Certificate)
	}
	replayed, err := app.Release(batch.BatchID, input)
	if err != nil || replayed.Version != released.Version || len(store.Events(batch.BatchID)) != 5 {
		t.Fatalf("重复放行产生变化: batch=%+v err=%v", replayed, err)
	}
	if _, err = app.Create(CreateBatchInput{Location: "另一公园", CollectionWindow: "下午", Species: "松树", Collector: "乙", Role: "collector", IdempotencyKey: "unrelated-create"}); err != nil {
		t.Fatal(err)
	}
	result, err := app.VerifyCertificate(batch.BatchID)
	if err != nil || result["valid"] != true {
		t.Fatalf("在线校验=%+v err=%v", result, err)
	}
	reopened, err := persistence.Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	result, err = New(reopened).VerifyCertificate(batch.BatchID)
	if err != nil || result["valid"] != true {
		t.Fatalf("重启后校验=%+v err=%v", result, err)
	}
}
