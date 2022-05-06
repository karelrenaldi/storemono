package smarthttp

import (
	"net/http"
	"time"
)

// Instrumentation allows users to generated stats or logs from Smart HTTP events
//go:generate mockery -name=Instrumentation -case underscore -testonly -inpkg
type Instrumentation interface {
	// Init is called once during initialization
	Init(name string)

	// InitWarning is called during init for warnings
	InitWarning(message string)

	// SanitizePath sanitizes the url path that can be sent to DataDog as a tag
	SanitizePath(urlPath string) string

	// DoDuration is the total time taken to complete the request (includes retries)
	DoDuration(start time.Time, endpointTag string)

	// BaseDoDuration is the time taken to make a single http.Client.Do() request
	BaseDoDuration(start time.Time, statusCode int, endpointTag string)

	// BaseDoErr is called when the underlying http.Client.Do() request returns an error
	BaseDoErr(err error, endpointTag, errTag string)

	// CBCircuitOpen is called when the circuit breaker circuit is open
	CBCircuitOpen(req *http.Request)

	// CBTrackedStatusCode is called when the response code is tracked by the circuit breaker as an error
	CBTrackedStatusCode(req *http.Request, code int)

	// RetryNonRetriable is called when a non-retriable HTTP status code or error has been returned
	RetryNonRetriable(req *http.Request, code int)

	// RetryRetriable is called when a retriable HTTP status code or error has been returned
	// NOTE: when errors occur status code is set to 666
	RetryRetriable(req *http.Request, code int)

	// SingleflightErr is called when singleflight returns an error
	SingleflightErr(req *http.Request, err error)
}

type noopInstrumentation struct{}

func (n *noopInstrumentation) Init(_ string) {}

func (n *noopInstrumentation) InitWarning(_ string) {}

func (n *noopInstrumentation) SanitizePath(_ string) string { return "" }

func (n *noopInstrumentation) DoDuration(_ time.Time, _ string) {}

func (n *noopInstrumentation) BaseDoDuration(_ time.Time, _ int, _ string) {}

func (n *noopInstrumentation) BaseDoErr(_ error, _, _ string) {}

func (n *noopInstrumentation) CBCircuitOpen(_ *http.Request) {}

func (n *noopInstrumentation) CBTrackedStatusCode(_ *http.Request, _ int) {}

func (n *noopInstrumentation) RetryNonRetriable(_ *http.Request, _ int) {}

func (n *noopInstrumentation) RetryRetriable(_ *http.Request, _ int) {}

func (n *noopInstrumentation) SingleflightErr(_ *http.Request, _ error) {}
