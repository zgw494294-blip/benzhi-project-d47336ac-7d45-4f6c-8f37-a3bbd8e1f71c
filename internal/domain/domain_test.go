package domain

import (
	"strings"
	"testing"
)

func TestBatchLifecycle(t *testing.T) {
	b, e, err := NewBatch("公园", "上午", "悬铃木", "叶斑", "采集员")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = b.AddEvidence(FieldEvidence{SampleNumber: "S1", Grid: "31,121", PhotoDigest: "x", Environment: map[string]float64{"humidity": 60}}); err != nil {
		t.Fatal(err)
	}
	if _, err = b.Screen(); err != nil {
		t.Fatal(err)
	}
	if _, err = b.ReviewExpert("轻度叶斑", "鉴定员"); err != nil {
		t.Fatal(err)
	}
	if b.Status != StatusReview {
		t.Fatalf("状态=%s", b.Status)
	}
	if _, err = b.Release("隔离", "明天", "复核员"); err != nil {
		t.Fatal(err)
	}
	if b.Certificate == nil {
		t.Fatal("缺少凭据")
	}
	_ = e
}
func TestEvidenceCheckCombinesMissingAndAbnormalWarnings(t *testing.T) {
	evidence := FieldEvidence{Grid: "G", PhotoDigest: "P", Environment: map[string]float64{"humidity": 130}}
	if err := ValidateEvidence(evidence); err != nil {
		t.Fatalf("现场提交阶段不应拒绝待检查读数: %v", err)
	}
	check := CheckEvidence(evidence)
	if check.Complete || check.Score != 55 {
		t.Fatalf("检查结果=%+v", check)
	}
	warnings := strings.Join(check.Warnings, "|")
	if !strings.Contains(warnings, "缺少样本编号") || !strings.Contains(warnings, "湿度超出允许范围") {
		t.Fatalf("警告未合并: %v", check.Warnings)
	}
}

func TestParseDispositionPlanPreservesOrderAndRejectsEmptyStep(t *testing.T) {
	plan, err := ParseDispositionPlan("综合处置", "隔离病株、喷施药剂，三日复查;记录结果", "明天上午", "绿化一组")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"隔离病株", "喷施药剂", "三日复查", "记录结果"}
	if strings.Join(plan.Steps, "|") != strings.Join(want, "|") {
		t.Fatalf("步骤=%v", plan.Steps)
	}
	if _, err = ParseDispositionPlan("综合处置", "隔离病株、、复查", "明天上午", "绿化一组"); err == nil {
		t.Fatal("应拒绝空步骤")
	}
}
