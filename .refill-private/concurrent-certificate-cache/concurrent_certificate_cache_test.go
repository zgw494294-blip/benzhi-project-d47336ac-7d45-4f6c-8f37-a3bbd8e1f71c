package concurrent_certificate_cache_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/benzhi/city-tree-release/internal/application"
	"github.com/benzhi/city-tree-release/internal/domain"
	"github.com/benzhi/city-tree-release/internal/persistence"
	"github.com/benzhi/city-tree-release/internal/transport"
)

type verificationBarrierRepository struct {
	persistence.Repository
	arrived chan struct{}
	release chan struct{}
}

func (r *verificationBarrierRepository) AllEvents() []domain.AuditEvent {
	r.arrived <- struct{}{}
	<-r.release
	return r.Repository.AllEvents()
}

func TestConcurrentCertificateVerificationCache(t *testing.T) {
	store, err := persistence.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	batch := releaseBatch(t, application.New(store))
	repo := &verificationBarrierRepository{
		Repository: store,
		arrived:    make(chan struct{}, 2),
		release:    make(chan struct{}),
	}
	handler := transport.New(application.New(repo), nil).Handler()

	var wg sync.WaitGroup
	responses := make(chan *httptest.ResponseRecorder, 2)
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/api/certificates/"+batch.BatchID+"/verify", nil)
			handler.ServeHTTP(recorder, request)
			responses <- recorder
		}()
	}
	<-repo.arrived
	<-repo.arrived
	close(repo.release)
	wg.Wait()
	close(responses)

	for response := range responses {
		if response.Code != http.StatusOK {
			t.Fatalf("并发校验状态码=%d body=%s", response.Code, response.Body.String())
		}
		var result map[string]any
		if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		if result["valid"] != true {
			t.Fatalf("并发校验结果=%v", result)
		}
	}
}

func releaseBatch(t *testing.T, app *application.Service) domain.SampleBatch {
	t.Helper()
	batch, err := app.Create(application.CreateBatchInput{
		Location: "并发验证公园", CollectionWindow: "上午", Species: "榆树",
		Collector: "采集员", Role: "collector", IdempotencyKey: "cache-create",
	})
	if err != nil {
		t.Fatal(err)
	}
	batch, err = app.AddEvidence(batch.BatchID, application.EvidenceInput{
		Role: "collector", ExpectedVersion: batch.Version, IdempotencyKey: "cache-evidence",
		SampleNumber: "CACHE-1", Grid: "31.2,121.4", PhotoDigest: "sha256:cache",
		Environment: map[string]float64{"humidity": 50},
	})
	if err != nil {
		t.Fatal(err)
	}
	batch, err = app.Screen(batch.BatchID, application.ScreeningInput{
		Role: "expert", ExpectedVersion: batch.Version, IdempotencyKey: "cache-screen",
	})
	if err != nil {
		t.Fatal(err)
	}
	batch, err = app.Review(batch.BatchID, application.ReviewInput{
		Role: "expert", Reviewer: "鉴定员", Conclusion: "轻度叶斑",
		ExpectedVersion: batch.Version, IdempotencyKey: "cache-review",
	})
	if err != nil {
		t.Fatal(err)
	}
	batch, err = app.Release(batch.BatchID, application.ReleaseInput{
		Role: "reviewer", Reviewer: "复核员", Plan: "隔离病株；三日复查",
		ExecutionWindow: "明天上午", ExpectedVersion: batch.Version, IdempotencyKey: "cache-release",
	})
	if err != nil {
		t.Fatal(err)
	}
	return batch
}
