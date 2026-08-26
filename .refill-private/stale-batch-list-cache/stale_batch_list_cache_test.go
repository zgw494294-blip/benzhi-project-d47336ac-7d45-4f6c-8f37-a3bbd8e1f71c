package stalebatchlistcache_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/benzhi/city-tree-release/internal/application"
	"github.com/benzhi/city-tree-release/internal/persistence"
	"github.com/benzhi/city-tree-release/internal/transport"
)

func TestBatchListCacheInvalidatesAfterStatusTransition(t *testing.T) {
	store, err := persistence.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	app := application.New(store)
	batch, err := app.Create(application.CreateBatchInput{
		Location: "人民公园", CollectionWindow: "上午", Species: "悬铃木",
		Collector: "采集员甲", Role: "collector", IdempotencyKey: "list-cache-create",
	})
	if err != nil {
		t.Fatal(err)
	}
	handler := transport.New(app, log.New(io.Discard, "", 0)).Handler()
	draftURL := "/api/batches?status=" + url.QueryEscape("待采集") + "&page=1&pageSize=20"

	before := requestList(t, handler, draftURL)
	if before.Total != 1 || len(before.Batches) != 1 {
		t.Fatalf("状态转换前的待采集目录异常: %+v", before)
	}

	payload := application.EvidenceInput{
		SampleNumber: "S-001", Grid: "31.2,121.4", PhotoDigest: "sha256:photo",
		Environment: map[string]float64{"humidity": 55}, Role: "collector",
		ExpectedVersion: batch.Version, IdempotencyKey: "list-cache-evidence",
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, "/api/batches/"+batch.BatchID+"/evidence", bytes.NewReader(body))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("证据提交状态码=%d body=%s", response.Code, response.Body.String())
	}

	after := requestList(t, handler, draftURL)
	if after.Total != 0 || len(after.Batches) != 0 {
		t.Fatalf("状态转换后仍返回缓存中的待采集批次: %+v", after)
	}
}

func requestList(t *testing.T, handler http.Handler, target string) application.BatchListResult {
	t.Helper()
	request := httptest.NewRequest(http.MethodGet, target, nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("目录请求状态码=%d body=%s", response.Code, response.Body.String())
	}
	var result application.BatchListResult
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	return result
}
