package application

import (
	"fmt"
	"github.com/benzhi/city-tree-release/internal/domain"
)

type Policy struct {
	RequirePhoto       bool
	MaxEvidence        int
	AllowAdminOverride bool
}

func DefaultPolicy() Policy {
	return Policy{RequirePhoto: true, MaxEvidence: 20, AllowAdminOverride: true}
}
func CheckPolicy(p Policy, b domain.SampleBatch) error {
	if p.MaxEvidence > 0 && len(b.Evidence) > p.MaxEvidence {
		return fmt.Errorf("证据数量超过策略上限")
	}
	if p.RequirePhoto {
		for _, e := range b.Evidence {
			if e.PhotoDigest == "" {
				return fmt.Errorf("策略要求每条证据包含照片摘要")
			}
		}
	}
	return nil
}
func CanOverride(p Policy, role string) bool {
	return p.AllowAdminOverride && domain.ParseRole(role) == domain.RoleAdmin
}
func TransitionPolicy(p Policy, b domain.SampleBatch, to domain.BatchStatus) error {
	if err := domain.ValidateTransition(b.Status, to); err != nil {
		return err
	}
	return CheckPolicy(p, b)
}
