package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/benzhi/city-tree-release/internal/application"
	"github.com/benzhi/city-tree-release/internal/persistence"
	"github.com/benzhi/city-tree-release/internal/transport"
)

func Run(args []string) error {
	cfg := ParseConfig(args)
	if err := ValidateConfig(cfg); err != nil {
		return err
	}
	store, err := persistence.Open(cfg.DataDir)
	if err != nil {
		return err
	}
	app := application.New(store)
	logger := log.Default()
	srv := &http.Server{Addr: cfg.Addr, Handler: transport.New(app, logger).Handler(), ReadHeaderTimeout: 5 * time.Second}
	if cfg.SelfCheck {
		return selfCheck(app)
	}
	logger.Printf("树木病虫害放行工作台监听 %s", cfg.Addr)
	return srv.ListenAndServe()
}
func selfCheck(app *application.Service) error {
	suffix := strconv.FormatInt(time.Now().UnixNano(), 10)
	b, err := app.Create(application.CreateBatchInput{Location: "静安公园 A 区", CollectionWindow: "2026-08-25 08:00-10:00", Species: "悬铃木", SuspectedIssue: "叶斑", Collector: "自检采集员", Role: "collector", IdempotencyKey: "selfcheck-create-" + suffix})
	if err != nil {
		return err
	}
	b, err = app.AddEvidence(b.BatchID, application.EvidenceInput{SampleNumber: "SC-001", Grid: "31.23,121.47", PhotoDigest: "sha256:selfcheck", Environment: map[string]float64{"temperature": 23, "humidity": 65}, Notes: "自检证据", Role: "collector", ExpectedVersion: b.Version, IdempotencyKey: "selfcheck-evidence-" + suffix})
	if err != nil {
		return err
	}
	b, err = app.Screen(b.BatchID, application.ScreeningInput{Role: "expert", ExpectedVersion: b.Version, IdempotencyKey: "selfcheck-screen-" + suffix})
	if err != nil {
		return err
	}
	if b.EvidenceScore != 100 || b.EvidenceCheckedAt == nil || b.Evidence[0].Check == nil || !b.Evidence[0].Check.Complete {
		return fmt.Errorf("逐条证据检查未生成完整结果")
	}
	b, err = app.Review(b.BatchID, application.ReviewInput{Role: "expert", Reviewer: "自检鉴定员", Conclusion: "轻度叶斑，风险低", ExpectedVersion: b.Version, IdempotencyKey: "selfcheck-review-" + suffix})
	if err != nil {
		return err
	}
	if b.Status == "整改中" {
		b, err = app.Rectify(b.BatchID, application.RectificationInput{Role: "collector", Notes: "已完成整改", ExpectedVersion: b.Version, IdempotencyKey: "selfcheck-rectify-" + suffix})
		if err != nil {
			return err
		}
	}
	b, err = app.Release(b.BatchID, application.ReleaseInput{Role: "reviewer", Reviewer: "自检复核员", PlanName: "叶斑处置", Plan: "隔离病株、持续观察", Owner: "自检绿化组", ExecutionWindow: "2026-08-25 14:00-16:00", ExpectedVersion: b.Version, IdempotencyKey: "selfcheck-release-" + suffix})
	if err != nil {
		return err
	}
	if b.Certificate == nil || len(b.Certificate.Disposition.Steps) != 2 {
		return fmt.Errorf("冻结凭据缺少结构化处置方案")
	}
	catalog, err := app.ListBatches(application.BatchListInput{Query: b.BatchID, Role: "管理员", Page: 1, PageSize: 10})
	if err != nil || catalog.Total != 1 || len(catalog.Batches) != 1 {
		return fmt.Errorf("批次目录查询自检失败")
	}
	result, err := app.VerifyCertificate(b.BatchID)
	if err != nil {
		return err
	}
	if ok, _ := result["valid"].(bool); !ok {
		return fmt.Errorf("凭据校验失败")
	}
	fmt.Printf("自检通过：批次 %s，凭据 %s\n", b.BatchID, b.Certificate.Credential)
	return nil
}
func Shutdown(ctx context.Context, srv *http.Server) error { return srv.Shutdown(ctx) }
