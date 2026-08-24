package domain

import (
	"fmt"
	"strings"
	"time"
)

type DispositionPlan struct {
	Name   string   `json:"name"`
	Steps  []string `json:"steps"`
	Window string   `json:"window"`
	Owner  string   `json:"owner"`
}

const maxDispositionTextLength = 500

func ValidateDisposition(p DispositionPlan) error {
	p.Name = strings.TrimSpace(p.Name)
	p.Window = strings.TrimSpace(p.Window)
	p.Owner = strings.TrimSpace(p.Owner)
	if p.Name == "" {
		return fmt.Errorf("处置方案名称不能为空")
	}
	if len([]rune(p.Name)) > 100 {
		return fmt.Errorf("处置方案名称不能超过 100 个字符")
	}
	if len(p.Steps) == 0 {
		return fmt.Errorf("至少需要一个处置步骤")
	}
	if len(p.Steps) > 20 {
		return fmt.Errorf("处置步骤不能超过 20 条")
	}
	for _, step := range p.Steps {
		if strings.TrimSpace(step) == "" {
			return fmt.Errorf("处置步骤不能为空")
		}
		if len([]rune(step)) > 200 {
			return fmt.Errorf("单条处置步骤不能超过 200 个字符")
		}
	}
	if p.Window == "" {
		return fmt.Errorf("执行窗口不能为空")
	}
	if len([]rune(p.Window)) > 120 {
		return fmt.Errorf("执行窗口不能超过 120 个字符")
	}
	if p.Owner == "" {
		return fmt.Errorf("执行负责人不能为空")
	}
	if len([]rune(p.Owner)) > 100 {
		return fmt.Errorf("执行负责人不能超过 100 个字符")
	}
	return nil
}
func PlanFromText(text, window, owner string) DispositionPlan {
	p, _ := ParseDispositionPlan("标准病虫害处置", text, window, owner)
	return p
}

func ParseDispositionPlan(name, text, window, owner string) (DispositionPlan, error) {
	p := DispositionPlan{Name: strings.TrimSpace(name), Window: strings.TrimSpace(window), Owner: strings.TrimSpace(owner), Steps: []string{}}
	if len([]rune(text)) > maxDispositionTextLength {
		return p, fmt.Errorf("处置方案文本不能超过 %d 个字符", maxDispositionTextLength)
	}
	start := 0
	runes := []rune(text)
	for i, r := range runes {
		if r != '、' && r != ',' && r != '，' && r != ';' && r != '；' {
			continue
		}
		step := strings.TrimSpace(string(runes[start:i]))
		if step == "" {
			return p, fmt.Errorf("处置步骤不能为空")
		}
		p.Steps = append(p.Steps, step)
		start = i + 1
	}
	last := strings.TrimSpace(string(runes[start:]))
	if last == "" {
		return p, fmt.Errorf("处置步骤不能为空")
	}
	p.Steps = append(p.Steps, last)
	if err := ValidateDisposition(p); err != nil {
		return p, err
	}
	return p, nil
}
func PlanAge(p DispositionPlan, issued time.Time) time.Duration {
	if issued.IsZero() {
		return 0
	}
	return time.Since(issued)
}
func IsExecutable(p DispositionPlan) bool { return ValidateDisposition(p) == nil }
