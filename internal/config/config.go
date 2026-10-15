package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	DefaultHTTPAddr   = "127.0.0.1:8080"
	DefaultTushareURL = "https://api.tushare.pro"
	DefaultTushareRPM = 200
)

type Config struct {
	DatabaseURL  string
	HTTPAddr     string
	TushareToken string
	TushareURL   string
	// TushareRPM 默认每分钟请求上限；未在 APIs 中配置的接口用它。
	TushareRPM int
	// TushareAPIs 按接口名覆盖 rpm，如 top10_cb_holders。
	TushareAPIs map[string]int
	Backtest    Backtest
	Log         Log
}

type Log struct {
	Level  string // debug | info | warn | error
	Format string // text | json
}

func DefaultLog() Log {
	return Log{Level: "info", Format: "text"}
}

type Backtest struct {
	InitCash       float64 `json:"init_cash"`
	Fast           int     `json:"fast"`
	Slow           int     `json:"slow"`
	CommissionRate float64 `json:"commission_rate"`
	CommissionMin  float64 `json:"commission_min"`
	StampTaxRate   float64 `json:"stamp_tax_rate"`
	Slippage       float64 `json:"slippage"`
	LotSize        int64   `json:"lot_size"`
}

func DefaultBacktest() Backtest {
	return Backtest{
		InitCash:       100_000,
		Fast:           5,
		Slow:           20,
		CommissionRate: 0.00025,
		CommissionMin:  5,
		StampTaxRate:   0.0005,
		Slippage:       0.001,
		LotSize:        100,
	}
}

// RPMFor 返回某 Tushare 接口的每分钟上限；未配置则用默认。
func (c Config) RPMFor(apiName string) int {
	apiName = strings.TrimSpace(apiName)
	if apiName != "" && c.TushareAPIs != nil {
		if n, ok := c.TushareAPIs[apiName]; ok && n > 0 {
			return n
		}
	}
	if c.TushareRPM > 0 {
		return c.TushareRPM
	}
	return DefaultTushareRPM
}

type fileConfig struct {
	HTTPAddr string `yaml:"http_addr"`
	Tushare  struct {
		URL string `yaml:"url"`
		RPM struct {
			Default int            `yaml:"default"`
			APIs    map[string]int `yaml:"apis"`
		} `yaml:"rpm"`
	} `yaml:"tushare"`
	Backtest fileBacktest `yaml:"backtest"`
	Log      struct {
		Level  string `yaml:"level"`
		Format string `yaml:"format"`
	} `yaml:"log"`
}

type fileBacktest struct {
	InitCash       float64 `yaml:"init_cash"`
	Fast           int     `yaml:"fast"`
	Slow           int     `yaml:"slow"`
	CommissionRate float64 `yaml:"commission_rate"`
	CommissionMin  float64 `yaml:"commission_min"`
	StampTaxRate   float64 `yaml:"stamp_tax_rate"`
	Slippage       float64 `yaml:"slippage"`
	LotSize        int64   `yaml:"lot_size"`
}

func Load() (Config, error) {
	_ = loadDotEnv(".env")

	cfg := Config{
		HTTPAddr:    DefaultHTTPAddr,
		TushareURL:  DefaultTushareURL,
		TushareRPM:  DefaultTushareRPM,
		TushareAPIs: map[string]int{},
		Backtest:    DefaultBacktest(),
		Log:         DefaultLog(),
	}

	if err := mergeYAML("config.yaml", &cfg); err != nil && !os.IsNotExist(err) {
		return Config{}, err
	}

	if url := os.Getenv("DATABASE_URL"); url != "" {
		cfg.DatabaseURL = url
	}
	if cfg.DatabaseURL == "" {
		cfg.DatabaseURL = "postgres://ashare:ashare@127.0.0.1:55432/ashare?sslmode=disable"
	}

	if v := os.Getenv("HTTP_ADDR"); v != "" {
		cfg.HTTPAddr = v
	}
	if v := os.Getenv("TUSHARE_TOKEN"); v != "" {
		cfg.TushareToken = v
	}
	if v := os.Getenv("TUSHARE_URL"); v != "" {
		cfg.TushareURL = v
	}
	if v := os.Getenv("TUSHARE_RPM"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 {
			return Config{}, fmt.Errorf("TUSHARE_RPM 无效，应为正整数")
		}
		cfg.TushareRPM = n
	} else if v := os.Getenv("TUSHARE_MIN_INTERVAL_MS"); v != "" {
		// 兼容旧配置：间隔毫秒 → 近似每分钟次数
		ms, err := strconv.Atoi(v)
		if err != nil || ms < 0 {
			return Config{}, fmt.Errorf("TUSHARE_MIN_INTERVAL_MS 无效")
		}
		if ms == 0 {
			cfg.TushareRPM = 600
		} else {
			cfg.TushareRPM = max(1, 60_000/ms)
		}
	}

	if cfg.TushareRPM <= 0 {
		cfg.TushareRPM = DefaultTushareRPM
	}
	if cfg.TushareAPIs == nil {
		cfg.TushareAPIs = map[string]int{}
	}
	if cfg.TushareURL == "" {
		cfg.TushareURL = DefaultTushareURL
	}
	if cfg.HTTPAddr == "" {
		cfg.HTTPAddr = DefaultHTTPAddr
	}
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		cfg.Log.Level = v
	}
	if v := os.Getenv("LOG_FORMAT"); v != "" {
		cfg.Log.Format = v
	}
	if cfg.Log.Level == "" {
		cfg.Log.Level = DefaultLog().Level
	}
	if cfg.Log.Format == "" {
		cfg.Log.Format = DefaultLog().Format
	}
	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL 不能为空")
	}
	return cfg, nil
}

func mergeYAML(path string, cfg *Config) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var fc fileConfig
	if err := yaml.Unmarshal(raw, &fc); err != nil {
		return fmt.Errorf("解析 %s: %w", path, err)
	}
	if fc.HTTPAddr != "" {
		cfg.HTTPAddr = fc.HTTPAddr
	}
	if fc.Tushare.URL != "" {
		cfg.TushareURL = fc.Tushare.URL
	}
	if fc.Tushare.RPM.Default > 0 {
		cfg.TushareRPM = fc.Tushare.RPM.Default
	}
	if len(fc.Tushare.RPM.APIs) > 0 {
		cfg.TushareAPIs = make(map[string]int, len(fc.Tushare.RPM.APIs))
		for k, v := range fc.Tushare.RPM.APIs {
			if strings.TrimSpace(k) == "" || v <= 0 {
				continue
			}
			cfg.TushareAPIs[strings.TrimSpace(k)] = v
		}
	}
	mergeBacktest(&cfg.Backtest, fc.Backtest)
	if fc.Log.Level != "" {
		cfg.Log.Level = fc.Log.Level
	}
	if fc.Log.Format != "" {
		cfg.Log.Format = fc.Log.Format
	}
	return nil
}

func mergeBacktest(dst *Backtest, src fileBacktest) {
	if src.InitCash > 0 {
		dst.InitCash = src.InitCash
	}
	if src.Fast > 0 {
		dst.Fast = src.Fast
	}
	if src.Slow > 0 {
		dst.Slow = src.Slow
	}
	if src.CommissionRate > 0 {
		dst.CommissionRate = src.CommissionRate
	}
	if src.CommissionMin > 0 {
		dst.CommissionMin = src.CommissionMin
	}
	if src.StampTaxRate > 0 {
		dst.StampTaxRate = src.StampTaxRate
	}
	if src.Slippage > 0 {
		dst.Slippage = src.Slippage
	}
	if src.LotSize > 0 {
		dst.LotSize = src.LotSize
	}
}

func loadDotEnv(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		val = strings.Trim(val, `"'`)
		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, val)
		}
	}
	return scanner.Err()
}

// MinInterval 便于日志展示：默认 rpm 换算的平均间隔。
func (c Config) MinInterval() time.Duration {
	return time.Minute / time.Duration(c.RPMFor(""))
}
