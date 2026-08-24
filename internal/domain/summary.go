package domain

import (
	"sort"
	"time"
)

type BatchSummary struct {
	BatchID        string      `json:"batchId"`
	Location       string      `json:"location"`
	Species        string      `json:"species"`
	Collector      string      `json:"collector"`
	SuspectedIssue string      `json:"suspectedIssue"`
	Status         BatchStatus `json:"status"`
	Version        int         `json:"version"`
	UpdatedAt      time.Time   `json:"updatedAt"`
	EvidenceCount  int         `json:"evidenceCount"`
	EvidenceScore  int         `json:"evidenceScore"`
	RiskLevel      string      `json:"riskLevel"`
	OpenIssueCount int         `json:"openIssueCount"`
	Released       bool        `json:"released"`
}

func Summarize(b SampleBatch) BatchSummary {
	risk := ""
	if b.Review != nil {
		risk = b.Review.RiskLevel
	}
	return BatchSummary{BatchID: b.BatchID, Location: b.Location, Species: b.Species, Collector: b.Collector, SuspectedIssue: b.SuspectedIssue, Status: b.Status, Version: b.Version, UpdatedAt: b.UpdatedAt, EvidenceCount: len(b.Evidence), EvidenceScore: b.EvidenceScore, RiskLevel: risk, OpenIssueCount: len(b.OpenIssues), Released: b.Certificate != nil}
}
func SortBatchesByUpdate(b []SampleBatch) []SampleBatch {
	out := append([]SampleBatch(nil), b...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].UpdatedAt.Equal(out[j].UpdatedAt) {
			return out[i].BatchID < out[j].BatchID
		}
		return out[i].UpdatedAt.After(out[j].UpdatedAt)
	})
	return out
}
func StatusLabel(s BatchStatus) string {
	labels := map[BatchStatus]string{StatusDraft: "待采集", StatusEvidence: "待检查", StatusScreened: "待鉴定", StatusRectification: "整改中", StatusReview: "待复核", StatusReleased: "已放行"}
	if x, ok := labels[s]; ok {
		return x
	}
	return "未知状态"
}
func RiskWeight(r string) int {
	switch r {
	case "高":
		return 3
	case "中":
		return 2
	case "低":
		return 1
	default:
		return 0
	}
}
