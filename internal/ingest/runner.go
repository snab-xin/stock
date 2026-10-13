package ingest

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/stock/ashare/internal/dates"
)

type JobRequest struct {
	Job   string   `json:"job"` // calendar | instruments | daily | sync
	Start string   `json:"start"`
	End   string   `json:"end"`
	Code  string   `json:"code"`
	Codes []string `json:"codes"`
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
	if job == "" {
		return fmt.Errorf("缺少 job，可选 calendar / instruments / daily / sync")
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

	var start, end time.Time
	var err error
	needRange := job == "calendar" || job == "daily" || job == "sync"
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
		Start:     req.Start,
		End:       req.End,
		Code:      strings.Join(codes, ","),
		Message:   "running",
		StartedAt: time.Now().Format(time.RFC3339),
		Error:     "",
	}
	r.mu.Unlock()

	go r.run(ctx, job, start, end, codes)
	return nil
}

func (r *Runner) run(ctx context.Context, job string, start, end time.Time, codes []string) {
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
	} else {
		r.status.Error = ""
		r.status.Message = "success"
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
