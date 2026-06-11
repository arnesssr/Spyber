// SPDX-License-Identifier: AGPL-3.0-only

package ports

import "context"

type FetchResult struct {
	URL        string
	StatusCode int
	Body       []byte
}

type Fetcher interface {
	Fetch(ctx context.Context, rawURL string) (FetchResult, error)
}
