package ingest

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/stock/ashare/internal/repo"
	"gorm.io/gorm"
)

func ResolveTSCode(ctx context.Context, instruments *repo.InstrumentRepo, input string) (string, error) {
	input = strings.ToUpper(strings.TrimSpace(input))
	if input == "" {
		return "", fmt.Errorf("证券代码为空")
	}
	if strings.Contains(input, ".") {
		return input, nil
	}

	items, err := instruments.FindBySymbol(ctx, input)
	if err != nil {
		return "", err
	}
	switch len(items) {
	case 0:
		_, err := instruments.GetByTSCode(ctx, input)
		if err == nil {
			return input, nil
		}
		if err == gorm.ErrRecordNotFound {
			return "", fmt.Errorf("找不到证券 %s，请先 ingest instruments", input)
		}
		return "", err
	case 1:
		return items[0].TSCode, nil
	default:
		var codes []string
		for _, it := range items {
			codes = append(codes, it.TSCode)
		}
		return "", fmt.Errorf("代码 %s 对应多只证券 %s，请用带后缀的 ts_code", input, strings.Join(codes, ", "))
	}
}

func ResolveLatestRunParams(ctx context.Context, runs *repo.IngestRunRepo, jobName string) (*JobRequest, error) {
	run, err := runs.Latest(ctx, jobName)
	if err != nil {
		return nil, err
	}
	if run == nil {
		return nil, errors.New("no latest run: " + jobName)
	}
	if run.Params == nil {
		return nil, errors.New("no params: " + jobName)
	}
	var params JobRequest
	err = run.Params.Unmarshal(&params)
	if err != nil {
		return nil, err
	}
	return &params, nil
}
