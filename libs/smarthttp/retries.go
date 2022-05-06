package smarthttp

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/corsc/go-commons/resilience/retry"
)

const (
	defaultMaxAttempts    = 3
	defaultBaseRetryDelay = 10 * time.Millisecond
	defaultMaxRetryDelay  = 1 * time.Second
)

var (
	// This indicates that we CANNOT retry this request due to the response code
	// This error should not be surfaced callers of this package.
	errRetryImpossible = errors.New("cannot retry due to HTTP status code")

	// This indicates that we CAN retry this request due to the response code
	// This error should not be surfaced callers of this package.
	errRetryAllowed = errors.New("according to the HTTP status we can retry the request")
)

// Retries defines the retry configuration
type Retries struct {
	// MaxAttempts is the maximum number of retry attempts before giving up. (default: 3)
	MaxAttempts int

	// BaseDelay is the base amount of time between attempts (default: 10 ms)
	BaseDelay time.Duration

	// MaxDelay is the maximum possible delay (default: 1 second)
	MaxDelay time.Duration

	retrier *retry.Client

	instrumentation Instrumentation
}

func (r *Retries) getMaxAttempts() int {
	if r.MaxAttempts > 0 {
		return r.MaxAttempts
	}

	r.instrumentation.InitWarning("using default 'max attempts' setting for retries")

	return defaultMaxAttempts
}

func (r *Retries) getBaseDelay() time.Duration {
	if r.BaseDelay > 0 {
		return r.BaseDelay
	}

	r.instrumentation.InitWarning("using default 'base retry delay' setting for retries")

	return defaultBaseRetryDelay
}

func (r *Retries) getMaxDelay() time.Duration {
	if r.MaxDelay > 0 {
		return r.MaxDelay
	}

	r.instrumentation.InitWarning("using default 'max retry delay' setting for retries")

	return defaultMaxRetryDelay
}

// nolint: gocognit,funlen
func (r *Retries) buildMiddleware(doFunc requestClosure) requestClosure {
	return func(req *http.Request) (*http.Response, error) {
		var resp *http.Response
		var innerErr error

		reqClone, err := cloneRequest(req)
		if err != nil {
			return nil, err
		}

		isFirstTry := true

		//nolint:bodyclose
		err = r.retrier.Do(req.Context(), "", func() error {
			if isFirstTry {
				isFirstTry = false
			} else {
				req = reqClone
				reqClone, err = cloneRequest(req)
				if err != nil {
					return err
				}
			}

			resp, innerErr = doFunc(req)
			if innerErr != nil {
				if errors.Is(innerErr, ErrTimeout) {
					// allow timeouts to retry
					r.instrumentation.RetryRetriable(req, 666)
					return errRetryAllowed
				}

				return innerErr
			}

			// process HTTP response codes (and trigger retries)
			switch resp.StatusCode {
			case http.StatusBadRequest, http.StatusUnauthorized, http.StatusPaymentRequired, http.StatusForbidden,
				http.StatusNotFound, http.StatusMethodNotAllowed, http.StatusNotAcceptable, http.StatusProxyAuthRequired,
				http.StatusGone, http.StatusLengthRequired, http.StatusPreconditionFailed, http.StatusRequestEntityTooLarge,
				http.StatusRequestURITooLong, http.StatusUnsupportedMediaType, http.StatusRequestedRangeNotSatisfiable,
				http.StatusExpectationFailed, http.StatusTeapot, http.StatusMisdirectedRequest, http.StatusUnprocessableEntity,
				http.StatusLocked, http.StatusFailedDependency, http.StatusTooEarly, http.StatusUpgradeRequired,
				http.StatusPreconditionRequired, http.StatusTooManyRequests, http.StatusRequestHeaderFieldsTooLarge,
				http.StatusUnavailableForLegalReasons, http.StatusNotImplemented, http.StatusBadGateway,
				http.StatusHTTPVersionNotSupported, http.StatusVariantAlsoNegotiates, http.StatusInsufficientStorage,
				http.StatusLoopDetected, http.StatusNotExtended, http.StatusNetworkAuthenticationRequired:
				// non-retriable status codes

				r.instrumentation.RetryNonRetriable(req, resp.StatusCode)

				return errRetryImpossible

			case http.StatusRequestTimeout,
				http.StatusInternalServerError,
				http.StatusServiceUnavailable,
				http.StatusGatewayTimeout:
				// retriable errors

				r.instrumentation.RetryRetriable(req, resp.StatusCode)

				return errRetryAllowed

			default:
				// happy path - do nothing
				return nil
			}
		})

		switch {
		case errors.Is(err, ErrTimeout) || errors.Is(err, context.DeadlineExceeded):
			// return nil response to avoid data race between retrier goroutine and this one.
			return nil, err

		case errors.Is(err, errRetryImpossible), errors.Is(err, errRetryAllowed), errors.Is(err, retry.ErrAttemptsExceeded):
			return resp, innerErr

		default:
			return resp, err
		}
	}
}

func cloneRequest(req *http.Request) (*http.Request, error) {
	reqClone := req.Clone(req.Context())

	if req.Body != nil {
		var b bytes.Buffer

		_, err := b.ReadFrom(req.Body)
		if err != nil {
			return nil, err
		}

		req.Body = ioutil.NopCloser(&b)
		reqClone.Body = ioutil.NopCloser(bytes.NewReader(b.Bytes()))
	}

	return reqClone, nil
}

func (r *Retries) addMiddleware(doFunc requestClosure) requestClosure {
	if r == nil {
		return doFunc
	}

	return r.buildMiddleware(doFunc)
}

func (r *Retries) doInitOnce(instrumentation Instrumentation) {
	if r == nil {
		return
	}

	r.instrumentation = instrumentation

	r.retrier = &retry.Client{
		MaxAttempts: r.getMaxAttempts(),
		BaseDelay:   r.getBaseDelay(),
		MaxDelay:    r.getMaxDelay(),
		CanRetry: func(err error) bool {
			return errors.Is(err, errRetryAllowed)
		},
	}
}
