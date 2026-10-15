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
const defaultRPM = 200

type Client struct {
	token      string
	endpoint   string
	httpClient *http.Client
	defaultRPM int
	apis       map[string]int

	mu   sync.Mutex
	last map[string]time.Time // 按 api_name 分桶
}

func NewClient(token, endpoint string, defaultRPM int, apis map[string]int) *Client {
	if endpoint == "" {
		endpoint = defaultURL
	}
	if defaultRPM <= 0 {
		defaultRPM = defaultRPMFallback()
	}
	copied := map[string]int{}
	for k, v := range apis {
		if k == "" || v <= 0 {
			continue
		}
		copied[k] = v
	}
	return &Client{
		token:    token,
		endpoint: endpoint,
		httpClient: &http.Client{
			Timeout: 45 * time.Second,
		},
		defaultRPM: defaultRPM,
		apis:       copied,
		last:       map[string]time.Time{},
	}
}

func defaultRPMFallback() int { return defaultRPM }

// RPM 返回默认每分钟上限。
func (c *Client) RPM() int { return c.defaultRPM }

// RPMFor 返回指定接口的每分钟上限；未配置则用默认。
func (c *Client) RPMFor(apiName string) int {
	if apiName != "" {
		if n, ok := c.apis[apiName]; ok && n > 0 {
			return n
		}
	}
	if c.defaultRPM > 0 {
		return c.defaultRPM
	}
	return defaultRPM
}

// MinGap 默认接口间隔。
func (c *Client) MinGap() time.Duration {
	return time.Minute / time.Duration(c.RPMFor(""))
}

// MinGapFor 指定接口间隔。
func (c *Client) MinGapFor(apiName string) time.Duration {
	return time.Minute / time.Duration(c.RPMFor(apiName))
}

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
	if err := c.wait(ctx, apiName); err != nil {
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
			if err := c.wait(ctx, apiName); err != nil {
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

// wait 按接口 rpm 均匀限流：同一 api_name 两次请求至少间隔 60s/rpm。
func (c *Client) wait(ctx context.Context, apiName string) error {
	rpm := c.RPMFor(apiName)
	minGap := time.Minute / time.Duration(rpm)

	c.mu.Lock()
	prev := c.last[apiName]
	wait := minGap - time.Since(prev)
	c.mu.Unlock()

	if wait > 0 {
		slog.Debug("tushare rate limit wait",
			"component", "tushare",
			"api", apiName,
			"wait", wait.Round(time.Millisecond).String(),
			"rpm", rpm,
			"gap", minGap.String(),
		)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait):
		}
	}

	c.mu.Lock()
	c.last[apiName] = time.Now()
	c.mu.Unlock()
	return nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
