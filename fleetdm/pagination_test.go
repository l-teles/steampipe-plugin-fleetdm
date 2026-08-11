package fleetdm

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
)

// makeItems returns n ints so tests can hand paginateCore pages of a given size.
func makeItems(n int) []int {
	items := make([]int, n)
	for i := range items {
		items[i] = i
	}
	return items
}

// countingEach returns a forEachItem that counts items and always continues.
func countingEach(seen *int) forEachItem[int] {
	return func(int) bool {
		*seen++
		return true
	}
}

func TestPaginateCoreServerClampsPerPage(t *testing.T) {
	t.Parallel()

	// The client requests per_page=1000 but the server clamps to 100 per page.
	// paginateCore must NOT stop on len(items) < perPage; it must keep going
	// until an empty page.
	calls := 0
	fetch := func(_ context.Context, page, perPage int) ([]int, *ListMeta, error) {
		calls++
		if perPage != 1000 {
			t.Errorf("fetch received perPage = %d, want 1000", perPage)
		}
		if page < 3 {
			return makeItems(100), nil, nil
		}
		return nil, nil, nil
	}

	seen := 0
	if err := paginateCore(testCtx(t), 1000, fetch, countingEach(&seen)); err != nil {
		t.Fatalf("paginateCore unexpected error: %v", err)
	}
	if seen != 300 {
		t.Errorf("items seen = %d, want 300", seen)
	}
	if calls != 4 {
		t.Errorf("fetch calls = %d, want 4 (3 full pages + 1 empty)", calls)
	}
}

func TestPaginateCoreMetaDrivenStop(t *testing.T) {
	t.Parallel()

	calls := 0
	fetch := func(_ context.Context, page, _ int) ([]int, *ListMeta, error) {
		calls++
		if page < 2 {
			return makeItems(50), &ListMeta{HasNextResults: true}, nil
		}
		return makeItems(50), &ListMeta{HasNextResults: false}, nil
	}

	seen := 0
	if err := paginateCore(testCtx(t), 50, fetch, countingEach(&seen)); err != nil {
		t.Fatalf("paginateCore unexpected error: %v", err)
	}
	if seen != 150 {
		t.Errorf("items seen = %d, want 150", seen)
	}
	if calls != 3 {
		t.Errorf("fetch calls = %d, want 3 (meta.has_next_results=false must stop without another fetch)", calls)
	}
}

func TestPaginateCoreMetaFalseOnFullFirstPage(t *testing.T) {
	t.Parallel()

	// len(items) == perPage, but meta says there is nothing more: meta is
	// authoritative and no second fetch may happen.
	calls := 0
	fetch := func(_ context.Context, _, _ int) ([]int, *ListMeta, error) {
		calls++
		return makeItems(50), &ListMeta{HasNextResults: false}, nil
	}

	seen := 0
	if err := paginateCore(testCtx(t), 50, fetch, countingEach(&seen)); err != nil {
		t.Fatalf("paginateCore unexpected error: %v", err)
	}
	if seen != 50 {
		t.Errorf("items seen = %d, want 50", seen)
	}
	if calls != 1 {
		t.Errorf("fetch calls = %d, want 1", calls)
	}
}

func TestPaginateCoreEmptyFirstPageNoMeta(t *testing.T) {
	t.Parallel()

	calls := 0
	fetch := func(_ context.Context, _, _ int) ([]int, *ListMeta, error) {
		calls++
		return nil, nil, nil
	}

	seen := 0
	if err := paginateCore(testCtx(t), 100, fetch, countingEach(&seen)); err != nil {
		t.Fatalf("paginateCore unexpected error: %v", err)
	}
	if seen != 0 {
		t.Errorf("items seen = %d, want 0", seen)
	}
	if calls != 1 {
		t.Errorf("fetch calls = %d, want 1", calls)
	}
}

func TestPaginateCoreEmptyPageWithHasNextResultsStops(t *testing.T) {
	t.Parallel()

	// A lying server: empty page but has_next_results=true. The empty-page
	// rule wins (no data progress) so pagination must stop anyway.
	calls := 0
	fetch := func(_ context.Context, _, _ int) ([]int, *ListMeta, error) {
		calls++
		return nil, &ListMeta{HasNextResults: true}, nil
	}

	seen := 0
	if err := paginateCore(testCtx(t), 100, fetch, countingEach(&seen)); err != nil {
		t.Fatalf("paginateCore unexpected error: %v", err)
	}
	if seen != 0 {
		t.Errorf("items seen = %d, want 0", seen)
	}
	if calls != 1 {
		t.Errorf("fetch calls = %d, want 1 (empty page must terminate)", calls)
	}
}

func TestPaginateCoreEachStopsMidPage(t *testing.T) {
	t.Parallel()

	calls := 0
	fetch := func(_ context.Context, _, _ int) ([]int, *ListMeta, error) {
		calls++
		return makeItems(100), &ListMeta{HasNextResults: true}, nil
	}

	seen := 0
	each := func(int) bool {
		seen++
		return seen < 5 // stop on the 5th item
	}

	if err := paginateCore(testCtx(t), 100, fetch, each); err != nil {
		t.Fatalf("paginateCore unexpected error: %v", err)
	}
	if seen != 5 {
		t.Errorf("items seen = %d, want 5", seen)
	}
	if calls != 1 {
		t.Errorf("fetch calls = %d, want 1 (each returning false must stop further fetches)", calls)
	}
}

func TestPaginateCoreFetchErrorPropagates(t *testing.T) {
	t.Parallel()

	errBoom := errors.New("boom")
	calls := 0
	fetch := func(_ context.Context, page, _ int) ([]int, *ListMeta, error) {
		calls++
		if page == 0 {
			return makeItems(10), nil, nil
		}
		return nil, nil, errBoom
	}

	seen := 0
	err := paginateCore(testCtx(t), 10, fetch, countingEach(&seen))
	if !errors.Is(err, errBoom) {
		t.Fatalf("paginateCore error = %v, want %v", err, errBoom)
	}
	if seen != 10 {
		t.Errorf("items seen before error = %d, want 10", seen)
	}
	if calls != 2 {
		t.Errorf("fetch calls = %d, want 2", calls)
	}
}

func TestShouldRetryError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "429 retries", err: &FleetAPIError{StatusCode: http.StatusTooManyRequests}, want: true},
		{name: "500 retries", err: &FleetAPIError{StatusCode: http.StatusInternalServerError}, want: true},
		{name: "503 retries", err: &FleetAPIError{StatusCode: http.StatusServiceUnavailable}, want: true},
		{name: "400 does not retry", err: &FleetAPIError{StatusCode: http.StatusBadRequest}, want: false},
		{name: "404 does not retry", err: &FleetAPIError{StatusCode: http.StatusNotFound}, want: false},
		{name: "plain error does not retry", err: errors.New("plain"), want: false},
		{name: "wrapped 429 retries", err: fmt.Errorf("x: %w", &FleetAPIError{StatusCode: http.StatusTooManyRequests}), want: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := shouldRetryError(context.Background(), nil, nil, tc.err); got != tc.want {
				t.Errorf("shouldRetryError(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}

func TestShouldIgnoreError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "404 ignored", err: &FleetAPIError{StatusCode: http.StatusNotFound}, want: true},
		{name: "403 not ignored", err: &FleetAPIError{StatusCode: http.StatusForbidden}, want: false},
		{name: "500 not ignored", err: &FleetAPIError{StatusCode: http.StatusInternalServerError}, want: false},
		{name: "plain error not ignored", err: errors.New("plain"), want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := shouldIgnoreError(context.Background(), nil, nil, tc.err); got != tc.want {
				t.Errorf("shouldIgnoreError(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}
