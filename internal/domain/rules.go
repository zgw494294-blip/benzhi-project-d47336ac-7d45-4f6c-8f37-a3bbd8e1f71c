package domain

import (
	"fmt"
	"strings"
)

func ValidateBatchInput(location, window, species, collector string) error {
	if location == "" || window == "" || species == "" || collector == "" {
		return fmt.Errorf("采集地点、时间窗、树种和采集员均不能为空")
	}
	if len(location) > 200 || len(species) > 100 || len(window) > 100 {
		return fmt.Errorf("批次字段长度超出限制")
	}
	return nil
}

func ValidateEvidence(e FieldEvidence) error {
	if len([]rune(e.SampleNumber)) > 100 || len([]rune(e.Grid)) > 100 || len([]rune(e.PhotoDigest)) > 200 || len([]rune(e.Notes)) > 1000 {
		return fmt.Errorf("现场证据字段长度超出限制")
	}
	if len(e.Environment) > 50 {
		return fmt.Errorf("环境读数项目过多")
	}
	for key := range e.Environment {
		if strings.TrimSpace(key) == "" || len([]rune(key)) > 100 {
			return fmt.Errorf("环境读数名称无效")
		}
	}
	return nil
}

func CalculateRisk(conclusion string, env map[string]float64) (string, []string) {
	text := conclusion
	if text == "" {
		return "高", []string{"缺少鉴定结论"}
	}
	issues := make([]string, 0)
	risk := "低"
	if containsAny(text, []string{"高", "严重", "疫病", "虫害"}) {
		risk = "高"
		issues = append(issues, "需要隔离与专项处置")
	}
	if h, ok := env["humidity"]; ok && h > 85 {
		if risk == "低" {
			risk = "中"
		}
		issues = append(issues, "现场湿度超过阈值")
	}
	if risk == "中" {
		issues = append(issues, "补充复采证据并观察传播范围")
	}
	return risk, issues
}

func containsAny(text string, words []string) bool {
	for _, w := range words {
		if len(w) > 0 && index(text, w) >= 0 {
			return true
		}
	}
	return false
}
func index(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
