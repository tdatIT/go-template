package productsvc

import (
	"context"
	"log/slog"

	"github.com/tdatIT/go-template/config"
	"github.com/tdatIT/go-template/internal/infras/adapter/productsvc/dto"
	"github.com/tdatIT/go-template/pkgs/caller"
	"github.com/tdatIT/go-template/pkgs/ultis/svcerr"
)

type adapter struct {
	caller caller.Caller
}

func NewAdapter(cfg *config.HTTPClient) ServiceAdapter {
	newCaller := caller.New(caller.Config{
		BaseURL:    cfg.BaseURL,
		Timeout:    cfg.Timeout,
		KeepAlive:  cfg.KeepAlive,
		RetryCount: cfg.RetryCount,
		RetryWait:  cfg.RetryWait,
		Debug:      cfg.Debug,
	})

	newCaller.GetClient().SetHeader("X-API-Key", cfg.APIKey)

	return &adapter{
		caller: newCaller,
	}
}

func (a *adapter) GetListOfProducts(ctx context.Context, req *dto.GetListProductReq) (*dto.GetListProductResp, error) {
	var result dto.GetListProductResp

	resp, err := a.caller.MakeRequest().
		SetContext(ctx).
		SetResult(&result.Products).
		Get(req.URL())
	if err != nil {
		slog.Error("mock svc get mock data failed", slog.String("error", err.Error()))
		return nil, svcerr.ErrInternalServer
	}

	if resp.IsError() {
		slog.Error("mock svc get mock data error response",
			slog.Int("status", resp.StatusCode()),
			slog.String("body", resp.String()),
		)
		return nil, svcerr.ErrInternalServer
	}

	return &result, nil
}
