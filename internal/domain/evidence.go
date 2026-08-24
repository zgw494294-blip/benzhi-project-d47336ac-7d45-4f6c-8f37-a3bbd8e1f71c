package domain

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

type EvidenceCheck struct {
	Complete  bool      `json:"complete"`
	Warnings  []string  `json:"warnings"`
	Score     int       `json:"score"`
	CheckedAt time.Time `json:"checkedAt"`
}

func CheckEvidence(e FieldEvidence) EvidenceCheck {
	return CheckEvidenceAt(e, time.Now().UTC())
}

func CheckEvidenceAt(e FieldEvidence, checkedAt time.Time) EvidenceCheck {
	r := EvidenceCheck{Complete: true, Warnings: []string{}, CheckedAt: checkedAt.UTC(), Score: 100}
	if strings.TrimSpace(e.SampleNumber) == "" {
		r.Complete = false
		r.Warnings = append(r.Warnings, "缺少样本编号")
		r.Score -= 30
	}
	if strings.TrimSpace(e.Grid) == "" {
		r.Complete = false
		r.Warnings = append(r.Warnings, "缺少经纬度网格")
		r.Score -= 20
	}
	if strings.TrimSpace(e.PhotoDigest) == "" {
		r.Complete = false
		r.Warnings = append(r.Warnings, "缺少现场照片摘要")
		r.Score -= 25
	}
	if len(e.Environment) == 0 {
		r.Complete = false
		r.Warnings = append(r.Warnings, "缺少环境读数")
		r.Score -= 25
	}
	keys := make([]string, 0, len(e.Environment))
	for k := range e.Environment {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := e.Environment[k]
		if math.IsNaN(v) || math.IsInf(v, 0) {
			r.Complete = false
			r.Warnings = append(r.Warnings, k+"不是有限数值")
			r.Score -= 10
			continue
		}
		if warning := EnvironmentWarning(k, v); warning != "" {
			r.Complete = false
			r.Warnings = append(r.Warnings, warning)
			r.Score -= 15
		}
	}
	if r.Score < 0 {
		r.Score = 0
	}
	return r
}

func EnvironmentWarning(key string, value float64) string {
	switch strings.ToLower(strings.TrimSpace(key)) {
	case "temperature":
		if value < -20 || value > 55 {
			return "温度超出允许范围 [-20, 55]"
		}
	case "humidity":
		if value < 5 || value > 95 {
			return "湿度超出允许范围 [5, 95]"
		}
	case "ph":
		if value < 3 || value > 10 {
			return "pH 超出允许范围 [3, 10]"
		}
	}
	return ""
}
func EvidenceDigest(e FieldEvidence) string {
	return fmt.Sprintf("%s|%s|%s|%s", e.EvidenceID, e.SampleNumber, e.Grid, e.PhotoDigest)
}
func IsAbnormalEnvironment(env map[string]float64) bool {
	for k, v := range env {
		if math.IsNaN(v) || math.IsInf(v, 0) || EnvironmentWarning(k, v) != "" {
			return true
		}
	}
	return false
}
func EvidenceFresh(e FieldEvidence, maxAge time.Duration) bool {
	return !e.SubmittedAt.IsZero() && time.Since(e.SubmittedAt) <= maxAge
}
