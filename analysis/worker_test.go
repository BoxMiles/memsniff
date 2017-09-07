package analysis

import (
	"github.com/box/memsniff/protocol/model"
	"testing"
	"time"
)

func TestCoalesce(t *testing.T) {
	w := newWorker()
	w.handleEvents([]model.Event{
		{
			Type: model.EventGetHit,
			Key:  "key1",
			Size: 10,
		},
		{
			Type: model.EventGetHit,
			Key:  "key1",
			Size: 50,
		},
	})
	time.Sleep(time.Millisecond)
	ss := w.snapshot()
	if len(ss) != 1 {
		t.Error("got", len(ss), "entries in snapshot, expected 1")
	}
	kr := ss[0]
	expected := KeyReport{
		Name:         "key1",
		Size:         50,
		GetHits:      2,
		TotalTraffic: 60,
		VariableSize: true,
	}
	if kr != expected {
		t.Error("actual", kr, "expected", expected)
	}
}
