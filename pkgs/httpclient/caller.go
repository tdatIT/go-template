package httpclient

import (
	"net"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

// Caller wraps a resty client and exposes R() to build outbound requests.
type Caller interface {
	MakeRequest() *resty.Request
	GetClient() *resty.Client
}

type Config struct {
	BaseURL        string
	Timeout        time.Duration
	KeepAlive      time.Duration // idle connection keep-alive duration
	RetryCount     int
	RetryWait      time.Duration // wait between retries on error
	RetryCondition resty.RetryConditionFunc
	Debug          bool
}

type caller struct {
	client *resty.Client
}

func New(cfg Config) Caller {
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			KeepAlive: cfg.KeepAlive,
		}).DialContext,
		IdleConnTimeout: cfg.KeepAlive,
	}

	client := resty.New().
		SetBaseURL(cfg.BaseURL).
		SetTimeout(cfg.Timeout).
		SetRetryCount(cfg.RetryCount).
		SetRetryWaitTime(cfg.RetryWait).
		SetTransport(transport).
		SetDebug(cfg.Debug)

	if cfg.RetryCondition != nil {
		client.AddRetryCondition(cfg.RetryCondition)
	}

	return &caller{client: client}
}

func (c *caller) MakeRequest() *resty.Request {
	return c.client.R()
}

func (c *caller) GetClient() *resty.Client {
	return c.client
}
