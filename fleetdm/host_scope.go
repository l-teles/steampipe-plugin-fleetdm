package fleetdm

import (
	"context"
	"net/url"
	"strconv"
	"sync"

	"github.com/turbot/steampipe-plugin-sdk/v6/plugin"
)

// perHostConcurrency caps parallel per-host requests so a large fleet does not
// turn one query into thousands of simultaneous API calls.
const perHostConcurrency = 10

// hostRef is the minimal host identity per-host tables need.
type hostRef struct {
	ID       uint   `json:"id"`
	Hostname string `json:"hostname"`
}

// hostScopes returns the hosts a per-host endpoint must be called for: only
// the host in a host_id qual, otherwise every host.
func hostScopes(ctx context.Context, d *plugin.QueryData, client *FleetDMClient) ([]hostRef, error) {
	if q := d.EqualsQuals["host_id"]; q != nil {
		v := q.GetInt64Value()
		if v <= 0 {
			return nil, nil
		}
		return []hostRef{{ID: uint(v)}}, nil
	}
	var hosts []hostRef
	err := paginateCore(ctx, 500, func(ctx context.Context, page, perPage int) ([]hostRef, *ListMeta, error) {
		params := url.Values{}
		params.Add("page", strconv.Itoa(page))
		params.Add("per_page", strconv.Itoa(perPage))
		params.Add("order_key", "id")
		var response struct {
			Hosts []hostRef `json:"hosts"`
		}
		if err := client.Get(ctx, "hosts", params, &response); err != nil {
			return nil, nil, err
		}
		return response.Hosts, nil, nil
	}, func(h hostRef) bool {
		hosts = append(hosts, h)
		return true
	})
	return hosts, err
}

// streamPerHost runs fetch for every host in scope with bounded concurrency
// and streams the rows from the calling goroutine (the list call). A host
// that disappeared mid-query (404) yields no rows.
func streamPerHost[T any](ctx context.Context, d *plugin.QueryData, client *FleetDMClient, fetch func(ctx context.Context, host hostRef) ([]T, error)) error {
	hosts, err := hostScopes(ctx, d, client)
	if err != nil {
		return err
	}
	return fanOut(ctx, hosts, perHostConcurrency, func(ctx context.Context, h hostRef) ([]T, error) {
		// Rows from other hosts may already have streamed, so a transient
		// failure is retried here rather than failing the whole query.
		var rows []T
		err := retryTransient(ctx, func() error {
			var err error
			rows, err = fetch(ctx, h)
			return err
		})
		if is404(err) {
			return nil, nil
		}
		return rows, err
	}, func(row T) bool {
		d.StreamListItem(ctx, row)
		return d.RowsRemaining(ctx) > 0
	})
}

// fanOut calls fetch for each input on at most n goroutines and hands every
// returned row to emit on the calling goroutine. It stops at the first fetch
// error or when emit returns false; workers are cancelled either way.
func fanOut[I, T any](ctx context.Context, inputs []I, n int, fetch func(context.Context, I) ([]T, error), emit func(T) bool) error {
	type result struct {
		rows []T
		err  error
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	jobs := make(chan I)
	results := make(chan result)
	var wg sync.WaitGroup
	for range min(n, len(inputs)) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for in := range jobs {
				rows, err := fetch(ctx, in)
				select {
				case results <- result{rows, err}:
				case <-ctx.Done():
					return
				}
			}
		}()
	}
	go func() {
		defer close(jobs)
		for _, in := range inputs {
			select {
			case jobs <- in:
			case <-ctx.Done():
				return
			}
		}
	}()
	go func() {
		wg.Wait()
		close(results)
	}()

	for r := range results {
		if r.err != nil {
			return r.err
		}
		for _, row := range r.rows {
			if !emit(row) {
				return nil
			}
		}
	}
	return nil
}
