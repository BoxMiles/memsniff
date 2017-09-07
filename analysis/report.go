package analysis

import (
	"sort"
	"time"
)

// KeyReport contains activity information for a single cache key.
type KeyReport struct {
	// cache key
	Name string
	// size of the cache value in bytes
	Size int
	// number of hits for this cache key
	GetHits int
	// amount of bandwidth consumed by traffic for this cache key in bytes
	TotalTraffic int
	// true if different sizes have been seen for this key
	VariableSize bool
}

// Report represents key activity submitted to a Pool since the last call to
// Reset.
type Report struct {
	// when this report was generated
	Timestamp time.Time
	// key reports in descending order by TotalTraffic
	Keys []KeyReport
}

// Len implements sort.Interface for Report.
func (r Report) Len() int {
	return len(r.Keys)
}

// Less implements sort.Interface for Report, sorting KeyReports in descending
// order by TotalTraffic.
func (r Report) Less(i, j int) bool {
	return r.Keys[j].TotalTraffic < r.Keys[i].TotalTraffic
}

// Swap implements sort.Interface for Report.
func (r Report) Swap(i, j int) {
	r.Keys[i], r.Keys[j] = r.Keys[j], r.Keys[i]
}

// Report returns a summary of activity recorded in this Pool since the last
// call to Reset.
//
// The returned report does not represent a consistent snapshot since
// information is collected from workers concurrent with new information
// coming in.
//
// If shouldReset is true, then a best effort will be made to clear data
// in the Pool while building the report.  Since clearing data is an
// asynchronous operation across the workers in the pool, some information
// may be carried over between successive reports, and some data may be
// lost entirely.
func (p *Pool) Report(shouldReset bool) Report {
	allEntries := make([]KeyReport, 0, p.reportSize*len(p.workers))
	for _, w := range p.workers {
		workerEntries := w.snapshot()
		if shouldReset {
			w.reset()
		}
		allEntries = append(allEntries, workerEntries...)
	}

	ret := Report{
		Timestamp: time.Now(),
		Keys:      allEntries,
	}

	sort.Sort(ret)

	return ret
}
