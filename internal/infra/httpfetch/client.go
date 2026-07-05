// SPDX-License-Identifier: AGPL-3.0-only

package httpfetch

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/arnesssr/Spyber/internal/domain"
	"github.com/arnesssr/Spyber/internal/ports"
)

type Client struct {
	httpClient           *http.Client
	UserAgent            string
	MaxResponseBytes     int64
	BlockPrivateNetworks bool
	MinHostDelay         time.Duration
	mu                   sync.Mutex
	lastByHost           map[string]time.Time
}

func New() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 20 * time.Second,
		},
		UserAgent:            "Spyber/0.1",
		MaxResponseBytes:     2 * 1024 * 1024,
		BlockPrivateNetworks: true,
		MinHostDelay:         time.Second,
		lastByHost:           map[string]time.Time{},
	}
}

func (c *Client) Fetch(ctx context.Context, rawURL string) (ports.FetchResult, error) {
	normalizedURL, host, err := domain.NormalizeWebsite(rawURL)
	if err != nil {
		return ports.FetchResult{}, err
	}
	if c.BlockPrivateNetworks {
		if err := rejectPrivateHost(ctx, host); err != nil {
			return ports.FetchResult{}, err
		}
	}
	if err := c.waitForHost(ctx, host); err != nil {
		return ports.FetchResult{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, normalizedURL, nil)
	if err != nil {
		return ports.FetchResult{}, err
	}
	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return ports.FetchResult{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return ports.FetchResult{}, fmt.Errorf("fetch failed with status %d", resp.StatusCode)
	}
	limit := c.MaxResponseBytes
	if limit <= 0 {
		limit = 2 * 1024 * 1024
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return ports.FetchResult{}, err
	}
	if int64(len(body)) > limit {
		return ports.FetchResult{}, fmt.Errorf("response exceeds %d bytes", limit)
	}
	return ports.FetchResult{
		URL:        resp.Request.URL.String(),
		StatusCode: resp.StatusCode,
		Body:       body,
	}, nil
}

func (c *Client) waitForHost(ctx context.Context, host string) error {
	if c.MinHostDelay <= 0 {
		return nil
	}
	c.mu.Lock()
	if c.lastByHost == nil {
		c.lastByHost = map[string]time.Time{}
	}
	last := c.lastByHost[host]
	wait := time.Until(last.Add(c.MinHostDelay))
	if wait <= 0 {
		c.lastByHost[host] = time.Now()
		c.mu.Unlock()
		return nil
	}
	c.mu.Unlock()
	timer := time.NewTimer(wait)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		c.mu.Lock()
		c.lastByHost[host] = time.Now()
		c.mu.Unlock()
		return nil
	}
}

func rejectPrivateHost(ctx context.Context, host string) error {
	ips, err := net.DefaultResolver.LookupIP(ctx, "ip", host)
	if err != nil {
		return fmt.Errorf("resolve host %q: %w", host, err)
	}
	for _, ip := range ips {
		if isPrivateIP(ip) {
			return fmt.Errorf("refusing private or local host: %s", host)
		}
	}
	return nil
}

func isPrivateIP(ip net.IP) bool {
	return ip.IsLoopback() ||
		ip.IsPrivate() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() ||
		ip.IsUnspecified() ||
		ip.IsMulticast()
}
