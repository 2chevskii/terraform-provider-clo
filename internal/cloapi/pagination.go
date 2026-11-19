package cloapi

import (
	"context"
	"fmt"
	"maps"
	"reflect"
	"strconv"
)

// ListOptions controls collection reads and reports the API's total count.
type ListOptions struct {
	Query map[string]string
	Total int
}

// CollectPages uses the server's initial page size and advances by the number
// actually received. Explicit limits or offsets request just that page.
func CollectPages[T any](ctx context.Context, opts *ListOptions, fetch func(map[string]string) ([]T, int, error)) ([]T, error) {
	query := map[string]string{}
	if opts != nil {
		maps.Copy(query, opts.Query)
	}
	_, limit := query["limit"]
	_, offset := query["offset"]
	var all, previous []T
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		page, total, err := fetch(query)
		if err != nil {
			return nil, err
		}
		if previous != nil && reflect.DeepEqual(page, previous) {
			return nil, fmt.Errorf("CLO returned the preceding page again at offset %d", len(all))
		}
		all = append(all, page...)
		if limit || offset || len(all) >= total {
			if opts != nil {
				opts.Total = total
			}
			return all, nil
		}
		if len(page) == 0 {
			return nil, fmt.Errorf("CLO returned an empty page after %d of %d results", len(all), total)
		}
		previous = page
		query["offset"] = strconv.Itoa(len(all))
		query["limit"] = strconv.Itoa(len(page))
	}
}
