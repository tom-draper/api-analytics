package routes

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
)

// fakeRows is a minimal pgx.Rows that yields a fixed number of rows and can fail
// the Scan for chosen (1-based) row indices. buildRequestData and
// buildRequestDataCompact only call Next and Scan, so the embedded interface's
// other methods are never used. It exercises the skip-accounting that decides
// how many rows made it into a response.
type fakeRows struct {
	pgx.Rows
	total  int
	failAt map[int]bool
	idx    int
}

func (f *fakeRows) Next() bool {
	f.idx++
	return f.idx <= f.total
}

func (f *fakeRows) Scan(dest ...any) error {
	if f.failAt[f.idx] {
		return errors.New("scan failed")
	}
	return nil
}

func TestBuildRequestDataSkipsScanErrors(t *testing.T) {
	rows := &fakeRows{total: 5, failAt: map[int]bool{2: true, 4: true}}

	requests, skipped := buildRequestData(rows)

	if len(requests) != 3 {
		t.Errorf("got %d requests, want 3 (5 rows minus 2 scan failures)", len(requests))
	}
	if skipped != 2 {
		t.Errorf("skipped = %d, want 2", skipped)
	}
}

func TestBuildRequestDataNoRows(t *testing.T) {
	requests, skipped := buildRequestData(&fakeRows{total: 0})
	if len(requests) != 0 || skipped != 0 {
		t.Errorf("empty result set gave %d requests / %d skipped, want 0/0", len(requests), skipped)
	}
}

func TestBuildRequestDataCompactSkipsAndHeaders(t *testing.T) {
	var cols [compactCols]any
	for i := range cols {
		cols[i] = "col"
	}
	rows := &fakeRows{total: 5, failAt: map[int]bool{3: true}}

	out, skipped := buildRequestDataCompact(rows, cols)

	// One header row plus four successfully scanned data rows.
	if len(out) != 5 {
		t.Errorf("got %d rows including the header, want 5", len(out))
	}
	if skipped != 1 {
		t.Errorf("skipped = %d, want 1", skipped)
	}
	// The header is always the first row.
	if out[0] != cols {
		t.Errorf("first row = %v, want the column header", out[0])
	}
}
