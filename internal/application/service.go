package application

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/benzhi/city-tree-release/internal/domain"
	"github.com/benzhi/city-tree-release/internal/persistence"
)

var ErrNotFound = errors.New("批次不存在")
var ErrConflict = errors.New("版本冲突")

type Service struct {
	repo        persistence.Repository
	listCacheMu sync.RWMutex
	listCache   map[batchListCacheKey]BatchListResult
}

func New(repo persistence.Repository) *Service {
	return &Service{repo: repo, listCache: make(map[batchListCacheKey]BatchListResult)}
}

type CreateBatchInput struct {
	Location         string `json:"location"`
	CollectionWindow string `json:"collectionWindow"`
	Species          string `json:"species"`
	SuspectedIssue   string `json:"suspectedIssue"`
	Collector        string `json:"collector"`
	Role             string `json:"role"`
	ExpectedVersion  *int   `json:"expectedVersion"`
	IdempotencyKey   string `json:"idempotencyKey"`
}
type EvidenceInput struct {
	SampleNumber    string             `json:"sampleNumber"`
	Grid            string             `json:"grid"`
	PhotoDigest     string             `json:"photoDigest"`
	Environment     map[string]float64 `json:"environment"`
	Notes           string             `json:"notes"`
	Role            string             `json:"role"`
	ExpectedVersion int                `json:"expectedVersion"`
	IdempotencyKey  string             `json:"idempotencyKey"`
}
type ScreeningInput struct {
	Role            string `json:"role"`
	ExpectedVersion int    `json:"expectedVersion"`
	IdempotencyKey  string `json:"idempotencyKey"`
}
type ReviewInput struct {
	Conclusion      string `json:"conclusion"`
	Reviewer        string `json:"reviewer"`
	Role            string `json:"role"`
	ExpectedVersion int    `json:"expectedVersion"`
	IdempotencyKey  string `json:"idempotencyKey"`
}
type RectificationInput struct {
	Notes           string `json:"notes"`
	Collector       string `json:"collector"`
	Role            string `json:"role"`
	ExpectedVersion int    `json:"expectedVersion"`
	IdempotencyKey  string `json:"idempotencyKey"`
}
type ReleaseInput struct {
	Plan            string `json:"plan"`
	PlanName        string `json:"planName"`
	Owner           string `json:"owner"`
	ExecutionWindow string `json:"executionWindow"`
	Reviewer        string `json:"reviewer"`
	Role            string `json:"role"`
	ExpectedVersion int    `json:"expectedVersion"`
	IdempotencyKey  string `json:"idempotencyKey"`
}

func (s *Service) Create(in CreateBatchInput) (domain.SampleBatch, error) {
	if !roleAllowed(in.Role, "collector") {
		return domain.SampleBatch{}, fmt.Errorf("角色无权建档")
	}
	if b, ok, err := s.idempotentBatch("", in.IdempotencyKey); ok || err != nil {
		return b, err
	}
	b, event, err := domain.NewBatch(in.Location, in.CollectionWindow, in.Species, in.SuspectedIssue, in.Collector)
	if err != nil {
		return b, err
	}
	if err := s.repo.Save(b.BatchID, b, event, NormalizeIdempotency(in.IdempotencyKey)); err != nil {
		return b, err
	}
	return b, nil
}

func (s *Service) AddEvidence(id string, in EvidenceInput) (domain.SampleBatch, error) {
	if !roleAllowed(in.Role, "collector") {
		return domain.SampleBatch{}, fmt.Errorf("角色无权提交现场证据")
	}
	if b, ok, err := s.idempotentBatch(id, in.IdempotencyKey); ok || err != nil {
		return b, err
	}
	b, err := s.loadVersion(id, in.ExpectedVersion)
	if err != nil {
		return b, err
	}
	e := domain.FieldEvidence{EvidenceID: domain.NewID("evidence"), SampleNumber: in.SampleNumber, Grid: in.Grid, PhotoDigest: in.PhotoDigest, Environment: in.Environment, Notes: in.Notes, SubmittedAt: timeNow()}
	event, err := b.AddEvidence(e)
	if err != nil {
		return b, err
	}
	if err = s.repo.Save(id, b, event, NormalizeIdempotency(in.IdempotencyKey)); err != nil {
		return b, err
	}
	return b, nil
}

func (s *Service) Screen(id string, in ScreeningInput) (domain.SampleBatch, error) {
	if !domain.RoleCan(domain.ParseRole(in.Role), "screen") {
		return domain.SampleBatch{}, fmt.Errorf("角色无权执行检查")
	}
	if err := requireIdempotencyKey(in.IdempotencyKey); err != nil {
		return domain.SampleBatch{}, err
	}
	if b, ok, err := s.idempotentBatch(id, in.IdempotencyKey); ok || err != nil {
		return b, err
	}
	b, err := s.loadVersion(id, in.ExpectedVersion)
	if err != nil {
		return b, err
	}
	event, err := b.Screen()
	if err != nil {
		return b, err
	}
	if err = s.repo.Save(id, b, event, NormalizeIdempotency(in.IdempotencyKey)); err != nil {
		return b, err
	}
	return b, nil
}

func (s *Service) Review(id string, in ReviewInput) (domain.SampleBatch, error) {
	if !roleAllowed(in.Role, "expert") {
		return domain.SampleBatch{}, fmt.Errorf("角色无权提交专家鉴定")
	}
	if b, ok, err := s.idempotentBatch(id, in.IdempotencyKey); ok || err != nil {
		return b, err
	}
	b, err := s.loadVersion(id, in.ExpectedVersion)
	if err != nil {
		return b, err
	}
	event, err := b.ReviewExpert(in.Conclusion, in.Reviewer)
	if err != nil {
		return b, err
	}
	if err = s.repo.Save(id, b, event, NormalizeIdempotency(in.IdempotencyKey)); err != nil {
		return b, err
	}
	return b, nil
}

func (s *Service) Rectify(id string, in RectificationInput) (domain.SampleBatch, error) {
	if !roleAllowed(in.Role, "collector") {
		return domain.SampleBatch{}, fmt.Errorf("角色无权提交整改")
	}
	if b, ok, err := s.idempotentBatch(id, in.IdempotencyKey); ok || err != nil {
		return b, err
	}
	b, err := s.loadVersion(id, in.ExpectedVersion)
	if err != nil {
		return b, err
	}
	event, err := b.ApplyRectification(in.Notes)
	if err != nil {
		return b, err
	}
	if err = s.repo.Save(id, b, event, NormalizeIdempotency(in.IdempotencyKey)); err != nil {
		return b, err
	}
	return b, nil
}

func (s *Service) Release(id string, in ReleaseInput) (domain.SampleBatch, error) {
	if !domain.RoleCan(domain.ParseRole(in.Role), "release") {
		return domain.SampleBatch{}, fmt.Errorf("角色无权冻结放行")
	}
	if err := requireIdempotencyKey(in.IdempotencyKey); err != nil {
		return domain.SampleBatch{}, err
	}
	if b, ok, err := s.idempotentBatch(id, in.IdempotencyKey); ok || err != nil {
		return b, err
	}
	b, err := s.loadVersion(id, in.ExpectedVersion)
	if err != nil {
		return b, err
	}
	name := in.PlanName
	if name == "" {
		name = "标准病虫害处置"
	}
	owner := in.Owner
	if owner == "" {
		owner = in.Reviewer
	}
	plan, err := domain.ParseDispositionPlan(name, in.Plan, in.ExecutionWindow, owner)
	if err != nil {
		return b, err
	}
	event, err := b.ReleaseWithDisposition(plan, strings.TrimSpace(in.Reviewer))
	if err != nil {
		return b, err
	}
	if err = s.repo.Save(id, b, event, NormalizeIdempotency(in.IdempotencyKey)); err != nil {
		return b, err
	}
	return b, nil
}

func (s *Service) Get(id string) (domain.SampleBatch, error) {
	b, ok := s.repo.Get(id)
	if !ok {
		return b, ErrNotFound
	}
	return b, nil
}
func (s *Service) List() []domain.SampleBatch { return s.repo.List() }
func (s *Service) Events(id string) ([]domain.AuditEvent, error) {
	if _, err := s.Get(id); err != nil {
		return nil, err
	}
	return s.repo.Events(id), nil
}
func (s *Service) VerifyCertificate(id string) (map[string]any, error) {
	b, err := s.Get(id)
	if err != nil {
		return nil, err
	}
	if b.Certificate == nil {
		return map[string]any{"valid": false, "reason": "批次尚未签发凭据"}, nil
	}
	result := map[string]any{"valid": false, "credential": b.Certificate.Credential, "batchId": id, "issuedAt": b.Certificate.IssuedAt}
	digest := domain.LegacyFreezeDigest(b)
	if b.Certificate.FreezeVersion == "v2" {
		digest = domain.FreezeDigest(b, b.Certificate.Disposition)
	}
	if digest != b.Certificate.FreezeDigest {
		result["reason"] = "冻结摘要不匹配"
		return result, nil
	}
	if err := persistence.VerifyChain(s.repo.AllEvents()); err != nil {
		result["reason"] = "审计事件链校验失败：" + err.Error()
		return result, nil
	}
	eventCertificate, ok := releasedCertificate(s.repo.Events(id))
	if !ok || !sameCertificate(*b.Certificate, eventCertificate) {
		result["reason"] = "放行事件与冻结凭据不一致"
		return result, nil
	}
	result["valid"] = true
	result["reason"] = "凭据有效"
	return result, nil
}

func (s *Service) loadVersion(id string, expected int) (domain.SampleBatch, error) {
	b, err := s.Get(id)
	if err != nil {
		return b, err
	}
	if b.Version != expected {
		return b, fmt.Errorf("%w：期望 %d，当前 %d", ErrConflict, expected, b.Version)
	}
	return b, nil
}
func roleAllowed(role string, roles ...string) bool {
	if role == "" {
		return false
	}
	for _, r := range roles {
		if strings.EqualFold(role, r) || role == "管理员" {
			return true
		}
	}
	return false
}
func timeNow() time.Time { return time.Now().UTC() }

func requireIdempotencyKey(key string) error {
	if NormalizeIdempotency(key) == "" {
		return fmt.Errorf("idempotencyKey 不能为空")
	}
	return nil
}

func (s *Service) idempotentBatch(id, key string) (domain.SampleBatch, bool, error) {
	key = NormalizeIdempotency(key)
	if key == "" {
		return domain.SampleBatch{}, false, nil
	}
	raw, ok := s.repo.Idempotent(key)
	if !ok {
		return domain.SampleBatch{}, false, nil
	}
	var batch domain.SampleBatch
	if err := json.Unmarshal(raw, &batch); err != nil || batch.BatchID == "" {
		return domain.SampleBatch{}, false, fmt.Errorf("%w：idempotencyKey 已被旧请求占用", ErrConflict)
	}
	if id != "" && batch.BatchID != id {
		return domain.SampleBatch{}, false, fmt.Errorf("%w：idempotencyKey 已用于其他批次", ErrConflict)
	}
	return batch, true, nil
}

func releasedCertificate(events []domain.AuditEvent) (domain.ReleaseCertificate, bool) {
	for i := len(events) - 1; i >= 0; i-- {
		if events[i].EventType != "BatchReleased" {
			continue
		}
		data, err := json.Marshal(events[i].Payload)
		if err != nil {
			return domain.ReleaseCertificate{}, false
		}
		var certificate domain.ReleaseCertificate
		if json.Unmarshal(data, &certificate) != nil {
			return domain.ReleaseCertificate{}, false
		}
		return certificate, true
	}
	return domain.ReleaseCertificate{}, false
}

func sameCertificate(a, b domain.ReleaseCertificate) bool {
	left, errLeft := json.Marshal(a)
	right, errRight := json.Marshal(b)
	return errLeft == nil && errRight == nil && bytes.Equal(left, right)
}
