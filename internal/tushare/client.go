package tushare

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

const defaultURL = "https://api.tushare.pro"

type Client struct {
	token      string
	endpoint   string
	httpClient *http.Client
	rpm        int
	minGap     time.Duration

	mu   sync.Mutex
	last time.Time
}

func NewClient(token, endpoint string, rpm int) *Client {
	if endpoint == "" {
		endpoint = defaultURL
	}
	if rpm <= 0 {
		rpm = 200
	}
	return &Client{
		token:    token,
		endpoint: endpoint,
		httpClient: &http.Client{
			Timeout: 45 * time.Second,
		},
		rpm:    rpm,
		minGap: time.Minute / time.Duration(rpm),
	}
}

func (c *Client) RPM() int { return c.rpm }

func (c *Client) MinGap() time.Duration { return c.minGap }

type apiRequest struct {
	APIName string            `json:"api_name"`
	Token   string            `json:"token"`
	Params  map[string]string `json:"params"`
	Fields  string            `json:"fields"`
}

type apiResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Fields []string `json:"fields"`
		Items  [][]any  `json:"items"`
	} `json:"data"`
}

func (c *Client) Query(ctx context.Context, apiName string, params map[string]string, fields string) ([]Row, error) {
	if c.token == "" {
		return nil, fmt.Errorf("未配置 TUSHARE_TOKEN，请写入 .env")
	}
	if err := c.wait(ctx); err != nil {
		return nil, err
	}

	body, err := json.Marshal(apiRequest{
		APIName: apiName,
		Token:   c.token,
		Params:  params,
		Fields:  fields,
	})
	if err != nil {
		return nil, err
	}

	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Duration(attempt) * time.Second):
			}
			if err := c.wait(ctx); err != nil {
				return nil, err
			}
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		raw, readErr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if readErr != nil {
			lastErr = readErr
			continue
		}
		if resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("tushare HTTP %d", resp.StatusCode)
			continue
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("tushare HTTP %d: %s", resp.StatusCode, truncate(string(raw), 200))
		}

		var out apiResponse
		if err := json.Unmarshal(raw, &out); err != nil {
			return nil, fmt.Errorf("解析 tushare 响应: %w", err)
		}
		if out.Code != 0 {
			lastErr = fmt.Errorf("tushare %s: [%d] %s", apiName, out.Code, out.Msg)
			if out.Code == 40203 || out.Code == -2001 {
				continue
			}
			return nil, lastErr
		}
		return newRows(out.Data.Fields, out.Data.Items), nil
	}
	return nil, lastErr
}

// wait 按配置的 rpm 均匀限流：两次请求至少间隔 60s/rpm。
func (c *Client) wait(ctx context.Context) error {
	c.mu.Lock()
	wait := c.minGap - time.Since(c.last)
	c.mu.Unlock()

	if wait > 0 {
		slog.Debug("tushare rate limit wait",
			"component", "tushare",
			"wait", wait.Round(time.Millisecond).String(),
			"rpm", c.rpm,
			"gap", c.minGap.String(),
		)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait):
		}
	}

	c.mu.Lock()
	c.last = time.Now()
	c.mu.Unlock()
	return nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
