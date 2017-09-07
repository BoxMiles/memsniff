package analysis

import (
	"errors"
	"github.com/box/memsniff/protocol/model"
)

// worker accumulates usage data for a set of cache keys.
type worker struct {
	// channel for reports of cache key activity
	events chan []model.Event
	// channel for requests for the current contents of the hotlist
	snapshotRequest chan struct{}
	// channel for results of snapshot() requests
	snapshotReply chan []KeyReport
	// channel for requests to reset the hotlist to an empty state
	resetRequest chan bool

	keyCounts map[string]*KeyReport
}

// errQueueFull is returned by handleGetResponse if the worker cannot keep
// up with incoming calls.
var errQueueFull = errors.New("analysis worker queue full")

func newWorker() worker {
	w := worker{
		events:          make(chan []model.Event, 1024),
		snapshotRequest: make(chan struct{}),
		snapshotReply:   make(chan []KeyReport),
		resetRequest:    make(chan bool),
		keyCounts:       make(map[string]*KeyReport),
	}
	go w.loop()
	return w
}

// handleEvents asynchronously processes events.
// handleEvents is threadsafe.
// When handleEvents returns, all relevant data from rs has been copied
// and is safe for the caller to discard.
func (w *worker) handleEvents(evts []model.Event) error {
	select {
	case w.events <- evts:
		return nil
	default:
		return errQueueFull
	}
}

// snapshot returns the current contents of the hotlist for this worker.
// snapshot is threadsafe.
func (w *worker) snapshot() []KeyReport {
	w.snapshotRequest <- struct{}{}
	return <-w.snapshotReply
}

// reset clear the contents of the hotlist for this worker.
// Some data may be lost if there is no external coordination of calls
// to snapshot and handleGetResponse.
func (w *worker) reset() {
	w.resetRequest <- true
}

// close exits this worker. Calls to handleGetResponse after calling close
// will panic.
func (w *worker) close() {
	close(w.events)
}

func (w *worker) loop() {
	for {
		select {
		case evts, ok := <-w.events:
			if !ok {
				return
			}
			for _, e := range evts {
				if e.Type != model.EventGetHit {
					continue
				}

				before, ok := w.keyCounts[e.Key]
				if ok {
					before.coalesceActivity(e)
				} else {
					w.keyCounts[e.Key] = &KeyReport{
						Name:         e.Key,
						Size:         e.Size,
						GetHits:      1,
						TotalTraffic: e.Size,
					}
				}
			}

		case <-w.snapshotRequest:
			w.snapshotReply <- w.copyKeyCounts()

		case <-w.resetRequest:
			for k := range w.keyCounts {
				delete(w.keyCounts, k)
			}
		}
	}
}

func (w *worker) copyKeyCounts() []KeyReport {
	results := make([]KeyReport, 0, len(w.keyCounts))
	for _, ka := range w.keyCounts {
		// make copy of each KeyActivity since we export results out of this goroutine
		results = append(results, *ka)
	}
	return results
}

func (kr *KeyReport) coalesceActivity(e model.Event) {
	if e.Size != kr.Size {
		kr.VariableSize = true
	}
	if e.Size > kr.Size {
		kr.Size = e.Size
	}
	kr.GetHits += 1
	kr.TotalTraffic += e.Size
	return
}
