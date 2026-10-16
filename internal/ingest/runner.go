package ingest

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/stock/ashare/internal/applog"
	"github.com/stock/ashare/internal/dates"
)

type JobRequest struct {
	Job    string   `json:"job"` // calendar | instruments | daily | sync | cb-basic | cb-holders | cb-daily
	Start  string   `json:"start"`
	End    string   `json:"end"`
	Period string   `json:"period"` // 可转债十大持有人报告期 YYYYMMDD
	Date   string   `json:"date"`   // 可转债行情交易日 YYYYMMDD（截面）
	Code   string   `json:"code"`
	Codes  []string `json:"codes"`
}

type JobStatus struct {
	Running    bool   `json:"running"`
	Job        string `json:"job,omitempty"`
	Start      string `json:"start,omitempty"`
	End        string `json:"end,omitempty"`
	Code       string `json:"code,omitempty"`
	Message    string `json:"message,omitempty"`
	Error      string `json:"error,omitempty"`
	StartedAt  string `json:"started_at,omitempty"`
	FinishedAt string `json:"finished_at,omitempty"`
}

type Runner struct {
	svc *Service

	mu     sync.Mutex
	status JobStatus
	cancel context.CancelFunc
}

func NewRunner(svc *Service) *Runner {
	return &Runner{svc: svc}
}

func (r *Runner) Status() JobStatus {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.status
}

func (r *Runner) Start(parent context.Context, req JobRequest) error {
	job := strings.TrimSpace(strings.ToLower(req.Job))
	job = strings.ReplaceAll(job, "_", "-")
	if job == "" {
		return fmt.Errorf("缺少 job")
	}

	codes := append([]string{}, req.Codes...)
	if c := strings.TrimSpace(req.Code); c != "" {
		for _, part := range strings.Split(c, ",") {
			part = strings.TrimSpace(part)
			if part != "" {
				codes = append(codes, part)
			}
		}
	}

	var start, end, period, tradeDate time.Time
	var err error
	needRange := job == "calendar" || job == "daily" || job == "sync" ||
		job == "index-daily" || job == "sw-daily" ||
		job == "etf-daily" || job == "etf-adj" || job == "etf-share" ||
		job == "etf-sh-cons" || job == "etf-sz-cons"
	if needRange {
		if strings.TrimSpace(req.Start) == "" || strings.TrimSpace(req.End) == "" {
			return fmt.Errorf("%s 需要 start 和 end（YYYYMMDD）", job)
		}
		start, err = dates.Parse(req.Start)
		if err != nil {
			return err
		}
		end, err = dates.Parse(req.End)
		if err != nil {
			return err
		}
		if end.Before(start) {
			return fmt.Errorf("结束日期早于开始日期")
		}
	}
	if job == "cb-holders" {
		if p := strings.TrimSpace(req.Period); p != "" {
			period, err = dates.Parse(p)
			if err != nil {
				return err
			}
		}
		if s := strings.TrimSpace(req.Start); s != "" {
			start, err = dates.Parse(s)
			if err != nil {
				return err
			}
		}
		if e := strings.TrimSpace(req.End); e != "" {
			end, err = dates.Parse(e)
			if err != nil {
				return err
			}
		}
		// 参数：period / start / end / code（code 可空=库内全部转债）
		if period.IsZero() && (start.IsZero() || end.IsZero()) {
			return fmt.Errorf("cb-holders 需要 period（报告期），或同时提供 start 与 end")
		}
		if !start.IsZero() && end.IsZero() {
			return fmt.Errorf("cb-holders 填写了 start 时也需要 end")
		}
		if start.IsZero() && !end.IsZero() {
			return fmt.Errorf("cb-holders 填写了 end 时也需要 start")
		}
		if !start.IsZero() && !end.IsZero() && end.Before(start) {
			return fmt.Errorf("结束日期早于开始日期")
		}
	}
	if job == "cb-daily" {
		if d := strings.TrimSpace(req.Date); d != "" {
			tradeDate, err = dates.Parse(d)
			if err != nil {
				return err
			}
		}
		if s := strings.TrimSpace(req.Start); s != "" {
			start, err = dates.Parse(s)
			if err != nil {
				return err
			}
		}
		if e := strings.TrimSpace(req.End); e != "" {
			end, err = dates.Parse(e)
			if err != nil {
				return err
			}
		}
		if tradeDate.IsZero() && (start.IsZero() || end.IsZero()) {
			return fmt.Errorf("cb-daily 需要 date（交易日截面），或同时提供 start 与 end")
		}
		if !start.IsZero() && !end.IsZero() && end.Before(start) {
			return fmt.Errorf("结束日期早于开始日期")
		}
	}

	r.mu.Lock()
	if r.status.Running {
		r.mu.Unlock()
		return fmt.Errorf("已有采集任务在运行：%s", r.status.Job)
	}
	// 与 HTTP 请求生命周期解耦，避免 POST 返回后任务被取消。
	_ = parent
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Hour)
	r.cancel = cancel
	r.status = JobStatus{
		Running:   true,
		Job:       job,
		Start:     firstNonEmpty(req.Date, req.Period, req.Start),
		End:       req.End,
		Code:      strings.Join(codes, ","),
		Message:   "running",
		StartedAt: time.Now().Format(time.RFC3339),
		Error:     "",
	}
	r.mu.Unlock()

	applog.With("component", "ingest").Info("job started",
		"job", job,
		"start", req.Start,
		"end", req.End,
		"period", req.Period,
		"date", req.Date,
		"codes", codes,
	)
	go r.run(ctx, job, start, end, period, tradeDate, codes)
	return nil
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func (r *Runner) run(ctx context.Context, job string, start, end, period, tradeDate time.Time, codes []string) {
	log := applog.With("component", "ingest", "job", job)
	var err error
	switch job {
	case "instruments":
		err = r.svc.SyncInstruments(ctx)
	case "calendar":
		err = r.svc.SyncCalendar(ctx, start, end)
	case "daily":
		resolved, rerr := resolveCodes(ctx, r.svc, codes)
		if rerr != nil {
			err = rerr
			break
		}
		err = r.svc.SyncDaily(ctx, start, end, resolved)
	case "sync":
		if err = r.svc.SyncCalendar(ctx, start, end); err != nil {
			break
		}
		if err = r.svc.SyncInstruments(ctx); err != nil {
			break
		}
		resolved, rerr := resolveCodes(ctx, r.svc, codes)
		if rerr != nil {
			err = rerr
			break
		}
		err = r.svc.SyncDaily(ctx, start, end, resolved)
	case "cb-basic":
		err = r.svc.SyncCBBasic(ctx)
	case "cb-holders":
		err = r.svc.SyncCBTopHolders(ctx, period, start, end, normalizeCBCodes(codes))
	case "cb-daily":
		err = r.svc.SyncCBDaily(ctx, start, end, tradeDate, normalizeCBCodes(codes))
	case "index-basic":
		err = r.svc.SyncIndexBasic(ctx)
	case "index-classify":
		err = r.svc.SyncIndexClassify(ctx)
	case "index-daily":
		err = r.svc.SyncIndexDaily(ctx, start, end, normalizeCBCodes(codes))
	case "sw-daily":
		err = r.svc.SyncSWDaily(ctx, start, end, normalizeCBCodes(codes))
	case "etf-basic":
		err = r.svc.SyncETFBasic(ctx)
	case "etf-daily":
		err = r.svc.SyncETFDaily(ctx, start, end, normalizeCBCodes(codes))
	case "etf-adj":
		err = r.svc.SyncETFAdj(ctx, start, end, normalizeCBCodes(codes))
	case "etf-share":
		err = r.svc.SyncETFShareSize(ctx, start, end, normalizeCBCodes(codes))
	case "etf-sh-cons":
		err = r.svc.SyncETFSHCons(ctx, start, end, normalizeCBCodes(codes))
	case "etf-sz-cons":
		err = r.svc.SyncETFSZCons(ctx, start, end, normalizeCBCodes(codes))
	default:
		err = fmt.Errorf("未知 job %q", job)
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.status.Running = false
	r.status.FinishedAt = time.Now().Format(time.RFC3339)
	if err != nil {
		r.status.Error = err.Error()
		r.status.Message = "failed"
		log.Error("job failed", "err", err)
	} else {
		r.status.Error = ""
		r.status.Message = "success"
		log.Info("job finished", "codes", len(codes))
	}
	r.cancel = nil
}

func resolveCodes(ctx context.Context, svc *Service, codes []string) ([]string, error) {
	if len(codes) == 0 {
		return nil, nil
	}
	out := make([]string, 0, len(codes))
	for _, c := range codes {
		code, err := ResolveTSCode(ctx, svc.Instruments, c)
		if err != nil {
			return nil, err
		}
		out = append(out, code)
	}
	return out, nil
}

func normalizeCBCodes(codes []string) []string {
	out := make([]string, 0, len(codes))
	for _, c := range codes {
		c = strings.ToUpper(strings.TrimSpace(c))
		if c == "" {
			continue
		}
		out = append(out, c)
	}
	return out
}
