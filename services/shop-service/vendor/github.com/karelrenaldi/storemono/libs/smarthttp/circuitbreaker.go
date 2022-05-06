package smarthttp

import (
	"errors"
	"net/http"
	"time"

	"github.com/afex/hystrix-go/hystrix"
)

const (
	defaultErrorThreshold = 80

	minErrorThreshold = 50
)

var (
	defaultMaxConcurrentRequests = hystrix.DefaultMaxConcurrent

	// see `getTimeout()` for more details
	defaultCircuitBreakerTimeout = 1 * time.Hour

	// This indicates a HTTP response code that should be tracked by the circuit
	errTrackableStatusCodeError = errors.New("response code is tracked by the circuit")

	// ErrCircuitIsOpen indicates that the circuit is open and any available fallback should be used
	ErrCircuitIsOpen = errors.New("the circuit is open")

	// ErrCircuitMaxConcurrencyReached indicates that there are more concurrent requests than configured going through
	// the circuit
	ErrCircuitMaxConcurrencyReached = errors.New("the circuit's max concurrency is reached")

	// ErrCircuitTimeout indicates that the circuit timed-out the request
	ErrCircuitTimeout = errors.New("the circuit timed out the request")
)

// CircuitBreaker defines the circuit breaker configuration
type CircuitBreaker struct {
	// Default value is 80 (cannot be set below 50)
	ErrorPercentThreshold int

	// Default value is 10 (setting above 100 is not advisable)
	MaxConcurrentRequests int

	name            string
	instrumentation Instrumentation

	// used for testing only
	trackError         func(cb *CircuitBreaker)
	totalTrackedErrors int
}

func (b *CircuitBreaker) getTimeout() int {
	// Set a timeout that is so long that all other timeouts will trigger first
	// We are essentially disabling this timeout
	return int(defaultCircuitBreakerTimeout.Milliseconds())
}

func (b *CircuitBreaker) getMaxConcurrent() int {
	if b.MaxConcurrentRequests > 0 {
		return b.MaxConcurrentRequests
	}

	b.instrumentation.InitWarning("using default 'max concurrent requests' setting for circuit breaker")

	return defaultMaxConcurrentRequests
}

func (b *CircuitBreaker) getErrorPercent() int {
	if b.ErrorPercentThreshold > minErrorThreshold {
		return b.ErrorPercentThreshold
	}

	b.instrumentation.InitWarning("using default 'error threshold' setting for circuit breaker")

	return defaultErrorThreshold
}

//nolint:bodyclose
func (b *CircuitBreaker) buildMiddleware(doFunc requestClosure) requestClosure {
	return func(req *http.Request) (*http.Response, error) {
		var resp *http.Response

		err := hystrix.Do(b.name, func() error {
			var innerErr error

			resp, innerErr = doFunc(req)
			if innerErr != nil {
				return innerErr
			}

			return b.outErrorBasedOnResponseCode(req, resp)
		}, nil)

		switch err {
		case hystrix.ErrCircuitOpen:
			b.instrumentation.CBCircuitOpen(req)
			return resp, ErrCircuitIsOpen

		case hystrix.ErrMaxConcurrency:
			return resp, ErrCircuitMaxConcurrencyReached

		case hystrix.ErrTimeout:
			return resp, ErrCircuitTimeout

		case nil, errTrackableStatusCodeError:
			return resp, nil

		default:
			return resp, err
		}
	}
}

func (b *CircuitBreaker) outErrorBasedOnResponseCode(req *http.Request, resp *http.Response) error {
	// process HTTP response codes (and throw errors that we should track)
	switch resp.StatusCode {
	case http.StatusRequestTimeout,
		http.StatusTooManyRequests,
		http.StatusInternalServerError,
		http.StatusNotImplemented,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout,
		http.StatusHTTPVersionNotSupported,
		http.StatusVariantAlsoNegotiates,
		http.StatusInsufficientStorage,
		http.StatusLoopDetected,
		http.StatusNotExtended,
		http.StatusNetworkAuthenticationRequired:
		// these HTTP response codes should be tracked by the circuit breaker
		b.trackError(b)

		b.instrumentation.CBTrackedStatusCode(req, resp.StatusCode)

		return errTrackableStatusCodeError

	default:
		// do not track these HTTP response codes (they are success codes or user errors)
		return nil
	}
}

func (b *CircuitBreaker) addMiddleware(doFunc requestClosure) requestClosure {
	if b == nil {
		return doFunc
	}

	return b.buildMiddleware(doFunc)
}

func (b *CircuitBreaker) doInitOnce(instrumentation Instrumentation, name string) {
	if b == nil {
		instrumentation.InitWarning("no circuit breaker has been configured.  CB use is strongly recommended")

		return
	}

	b.name = name
	b.instrumentation = instrumentation

	hystrix.ConfigureCommand(b.name, hystrix.CommandConfig{
		Timeout:               b.getTimeout(),
		MaxConcurrentRequests: b.getMaxConcurrent(),
		ErrorPercentThreshold: b.getErrorPercent(),
	})

	if b.trackError == nil {
		b.trackError = func(_ *CircuitBreaker) {
			// noop
		}
	}
}
