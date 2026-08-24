package transport

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/benzhi/city-tree-release/internal/application"
	"github.com/benzhi/city-tree-release/internal/persistence"
)

func TestBatchCatalogHTTPQueryAndValidation(t *testing.T) {
	store, err := persistence.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	app := application.New(store)
	if _, err = app.Create(application.CreateBatchInput{Location: "人民公园", CollectionWindow: "上午", Species: "悬铃木", Collector: "甲", Role: "collector"}); err != nil {
		t.Fatal(err)
	}
	handler := New(app, log.New(io.Discard, "", 0)).Handler()
	request := httptest.NewRequest(http.MethodGet, "/api/batches?status=%E5%BE%85%E9%87%87%E9%9B%86&q=%E5%85%AC%E5%9B%AD&role=collector&page=1&pageSize=1", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("状态码=%d body=%s", response.Code, response.Body.String())
	}
	var result application.BatchListResult
	if err = json.NewDecoder(response.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if result.Total != 1 || result.Page != 1 || result.PageSize != 1 || result.HasNext || len(result.Batches) != 1 {
		t.Fatalf("目录响应=%+v", result)
	}
	bad := httptest.NewRecorder()
	handler.ServeHTTP(bad, httptest.NewRequest(http.MethodGet, "/api/batches?page=0", nil))
	if bad.Code != http.StatusBadRequest {
		t.Fatalf("无效页码状态码=%d", bad.Code)
	}
}
