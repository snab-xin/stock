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
	// TushareRPM 为每分钟最大请求次数（滑动窗口）。
	TushareRPM int
}

type fileConfig struct {
	HTTPAddr string `yaml:"http_addr"`
	Tushare  struct {
		URL string `yaml:"url"`
		RPM int    `yaml:"rpm"`
	} `yaml:"tushare"`
}

func Load() (Config, error) {
	_ = loadDotEnv(".env")

	cfg := Config{
		HTTPAddr:   DefaultHTTPAddr,
		TushareURL: DefaultTushareURL,
		TushareRPM: DefaultTushareRPM,
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
	if cfg.TushareURL == "" {
		cfg.TushareURL = DefaultTushareURL
	}
	if cfg.HTTPAddr == "" {
		cfg.HTTPAddr = DefaultHTTPAddr
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
	if fc.Tushare.RPM > 0 {
		cfg.TushareRPM = fc.Tushare.RPM
	}
	return nil
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

// MinInterval 便于日志展示：按 RPM 换算的平均间隔。
func (c Config) MinInterval() time.Duration {
	if c.TushareRPM <= 0 {
		return time.Minute / DefaultTushareRPM
	}
	return time.Minute / time.Duration(c.TushareRPM)
}
