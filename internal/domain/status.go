package domain

import (
	"fmt"
	"strings"
)

type BatchStatus string

const (
	StatusDraft         BatchStatus = "待采集"
	StatusEvidence      BatchStatus = "待检查"
	StatusScreened      BatchStatus = "待鉴定"
	StatusRectification BatchStatus = "整改中"
	StatusReview        BatchStatus = "待复核"
	StatusReleased      BatchStatus = "已放行"
)

func (s BatchStatus) String() string { return string(s) }

func ParseStatus(raw string) (BatchStatus, bool) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "all", "全部":
		return "", true
	case "draft", "待采集":
		return StatusDraft, true
	case "evidence", "待检查":
		return StatusEvidence, true
	case "screened", "待鉴定":
		return StatusScreened, true
	case "rectification", "整改中":
		return StatusRectification, true
	case "review", "待复核":
		return StatusReview, true
	case "released", "已放行":
		return StatusReleased, true
	default:
		return "", false
	}
}

func CanTransition(from, to BatchStatus) bool {
	if from == to {
		return true
	}
	allowed := map[BatchStatus]map[BatchStatus]bool{
		StatusDraft:         {StatusEvidence: true},
		StatusEvidence:      {StatusScreened: true},
		StatusScreened:      {StatusRectification: true, StatusReview: true},
		StatusRectification: {StatusScreened: true, StatusReview: true},
		StatusReview:        {StatusReleased: true, StatusRectification: true},
		StatusReleased:      {},
	}
	return allowed[from][to]
}

func ValidateTransition(from, to BatchStatus) error {
	if !CanTransition(from, to) {
		return fmt.Errorf("状态不能从 %s 迁移到 %s", from, to)
	}
	return nil
}
