package concurrent_audit_head_test

import (
	"encoding/json"
	"sync"
	"testing"

	"github.com/benzhi/city-tree-release/internal/domain"
	"github.com/benzhi/city-tree-release/internal/persistence"
)

type coordinatedPayload struct {
	id      string
	ready   chan<- struct{}
	release <-chan struct{}
}

func (p coordinatedPayload) MarshalJSON() ([]byte, error) {
	p.ready <- struct{}{}
	<-p.release
	return json.Marshal(map[string]string{"id": p.id})
}

func TestConcurrentSavesKeepSingleAuditHead(t *testing.T) {
	dir := t.TempDir()
	store, err := persistence.Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	ready := make(chan struct{}, 2)
	release := make(chan struct{})
	errs := make(chan error, 2)
	var started sync.WaitGroup
	started.Add(2)

	save := func(id string) {
		defer started.Done()
		batch, _, createErr := domain.NewBatch("并发采集点-"+id, "2026-08-25 09:00-11:00", "香樟", "叶斑", "采集员甲")
		if createErr != nil {
			errs <- createErr
			return
		}
		event := domain.Event{Type: "BatchCreated", Payload: coordinatedPayload{id: id, ready: ready, release: release}}
		errs <- store.Save(batch.BatchID, batch, event, "idem-"+id)
	}

	go save("a")
	go save("b")
	<-ready
	<-ready
	close(release)
	started.Wait()
	close(errs)
	for saveErr := range errs {
		if saveErr != nil {
			t.Fatalf("并发保存返回错误: %v", saveErr)
		}
	}

	events := store.AllEvents()
	if len(events) != 2 {
		t.Fatalf("审计事件数量=%d, want 2", len(events))
	}
	if err := persistence.VerifyChain(events); err != nil {
		t.Fatalf("并发保存破坏审计链: %v", err)
	}
	reopened, err := persistence.Open(dir)
	if err != nil {
		t.Fatalf("并发保存后的账本无法重启恢复: %v", err)
	}
	if reopened.LastSequence() != 2 {
		t.Fatalf("重启恢复序号=%d, want 2", reopened.LastSequence())
	}
}
