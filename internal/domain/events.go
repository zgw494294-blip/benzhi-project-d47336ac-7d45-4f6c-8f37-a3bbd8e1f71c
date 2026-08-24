package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type Event struct {
	Type    string
	Payload any
}

func NewBatch(location, window, species, issue, collector string) (SampleBatch, Event, error) {
	if err := ValidateBatchInput(location, window, species, collector); err != nil {
		return SampleBatch{}, Event{}, err
	}
	now := nowUTC()
	b := SampleBatch{BatchID: NewID("batch"), Location: location, CollectionWindow: window, Species: species, SuspectedIssue: issue, Collector: collector, Status: StatusDraft, Version: 0, CreatedAt: now, UpdatedAt: now}
	return b, Event{Type: "BatchCreated", Payload: b}, nil
}

func (b *SampleBatch) AddEvidence(e FieldEvidence) (Event, error) {
	if b.Status != StatusDraft && b.Status != StatusEvidence && b.Status != StatusScreened && b.Status != StatusRectification {
		return Event{}, ValidateTransition(b.Status, StatusEvidence)
	}
	if err := ValidateEvidence(e); err != nil {
		return Event{}, err
	}
	e.BatchID = b.BatchID
	e.Integrity = "待检查"
	e.Check = nil
	b.Evidence = append(b.Evidence, e)
	b.Status = StatusEvidence
	b.EvidenceScore = 0
	b.EvidenceCheckedAt = nil
	b.Version++
	b.UpdatedAt = nowUTC()
	return Event{Type: "EvidenceSubmitted", Payload: e}, nil
}

func (b *SampleBatch) Screen() (Event, error) {
	if len(b.Evidence) == 0 {
		return Event{}, fmtError("至少提交一条现场证据后才能检查")
	}
	if b.Status != StatusEvidence {
		return Event{}, fmtError("当前状态不允许执行完整性检查")
	}
	checkedAt := nowUTC()
	checks := make([]EvidenceCheck, len(b.Evidence))
	total := 0
	passed := true
	for i, evidence := range b.Evidence {
		checks[i] = CheckEvidenceAt(evidence, checkedAt)
		total += checks[i].Score
		if !checks[i].Complete {
			passed = false
		}
	}
	for i := range b.Evidence {
		check := checks[i]
		b.Evidence[i].Check = &check
		if check.Complete {
			b.Evidence[i].Integrity = "通过"
		} else {
			b.Evidence[i].Integrity = "不通过"
		}
	}
	b.EvidenceScore = total / len(checks)
	b.EvidenceCheckedAt = &checkedAt
	if passed {
		b.Status = StatusScreened
	}
	b.Version++
	b.UpdatedAt = checkedAt
	result := "不通过"
	if passed {
		result = "通过"
	}
	return Event{Type: "BatchScreened", Payload: map[string]any{"result": result, "score": b.EvidenceScore, "checkedAt": checkedAt, "evidence": b.Evidence}}, nil
}

func (b *SampleBatch) ReviewExpert(conclusion, reviewer string) (Event, error) {
	if b.Status != StatusScreened && b.Status != StatusRectification {
		return Event{}, fmtError("当前状态不允许专家鉴定")
	}
	risk, issues := CalculateRisk(conclusion, latestEnvironment(b.Evidence))
	b.OpenIssues = issues
	b.Review = &ExpertReview{ReviewID: NewID("review"), BatchID: b.BatchID, Conclusion: conclusion, RiskLevel: risk, Issues: issues, Reviewer: reviewer, ReviewedAt: nowUTC()}
	if len(issues) > 0 {
		b.Status = StatusRectification
		b.Review.Rectification = "完成隔离、复采并上传整改证据"
	} else {
		b.Status = StatusReview
	}
	b.Version++
	b.UpdatedAt = nowUTC()
	return Event{Type: "ExpertReviewed", Payload: b.Review}, nil
}

func (b *SampleBatch) ApplyRectification(notes string) (Event, error) {
	if b.Status != StatusRectification {
		return Event{}, fmtError("当前没有待整改问题")
	}
	if notes == "" {
		return Event{}, fmtError("整改说明不能为空")
	}
	b.OpenIssues = nil
	if b.Review != nil {
		b.Review.Rectification = notes
	}
	b.Status = StatusReview
	b.Version++
	b.UpdatedAt = nowUTC()
	return Event{Type: "RectificationClosed", Payload: map[string]string{"notes": notes}}, nil
}

func (b *SampleBatch) Release(plan, window, reviewer string) (Event, error) {
	disposition, err := ParseDispositionPlan("标准病虫害处置", plan, window, reviewer)
	if err != nil {
		return Event{}, err
	}
	return b.ReleaseWithDisposition(disposition, reviewer)
}

func (b *SampleBatch) ReleaseWithDisposition(plan DispositionPlan, reviewer string) (Event, error) {
	if b.Status != StatusReview {
		return Event{}, fmtError("整改未闭环或批次尚未复核")
	}
	if len(b.OpenIssues) != 0 {
		return Event{}, fmtError("仍有开放问题，不能放行")
	}
	if err := ValidateDisposition(plan); err != nil {
		return Event{}, err
	}
	reviewer = strings.TrimSpace(reviewer)
	if reviewer == "" {
		return Event{}, fmtError("复核人不能为空")
	}
	if len([]rune(reviewer)) > 100 {
		return Event{}, fmtError("复核人不能超过 100 个字符")
	}
	now := nowUTC()
	b.Status = StatusReleased
	b.Version++
	b.UpdatedAt = now
	digest := FreezeDigest(*b, plan)
	cred := "CERT-" + digest[:16]
	b.Certificate = &ReleaseCertificate{DispositionID: NewID("disp"), BatchID: b.BatchID, Plan: strings.Join(plan.Steps, "；"), Disposition: plan, ExecutionWindow: plan.Window, FreezeDigest: digest, FreezeVersion: "v2", Credential: cred, Issuer: reviewer, IssuedAt: now}
	return Event{Type: "BatchReleased", Payload: b.Certificate}, nil
}

func latestEnvironment(es []FieldEvidence) map[string]float64 {
	if len(es) == 0 {
		return nil
	}
	return es[len(es)-1].Environment
}
func nowUTC() time.Time       { return time.Now().UTC() }
func fmtError(s string) error { return fmt.Errorf("%s", s) }

func FreezeDigest(b SampleBatch, plan DispositionPlan) string {
	b.Certificate = nil
	data, _ := json.Marshal(struct {
		Batch       SampleBatch     `json:"batch"`
		Disposition DispositionPlan `json:"disposition"`
	}{Batch: b, Disposition: plan})
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func LegacyFreezeDigest(b SampleBatch) string {
	b.Certificate = nil
	data, _ := json.Marshal(b)
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}
