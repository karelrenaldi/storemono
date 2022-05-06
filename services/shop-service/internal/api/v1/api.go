package v1

import (
	"context"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	httputils "github.com/karelrenaldi/storemono/libs/http-utils"
	"github.com/karelrenaldi/storemono/libs/logger"
	"github.com/karelrenaldi/storemono/services/shop-service/internal/constant"
)

func NewAPI(ctx context.Context) (a *APIv1, err error) {
	cfg, ok := ctx.Value(constant.AppConfig).(Config)
	if !ok {
		err = errors.New("no AppConfig in ctx")
		return
	}

	a = &APIv1{
		ctx:    ctx,
		logger: cfg.Logger(),
	}

	return
}

type APIv1 struct {
	ctx    context.Context
	logger *logger.Logger
}

func (p *APIv1) AddRoutes(router *mux.Router) {
	apiV1 := router.PathPrefix("/api/v1").Subrouter()

	// Middlewares.
	apiV1.Use(p.RecoverPanicMiddleware)
	apiV1.Use(p.logger.GorillaMiddleware())

	// Routes.
}

func (p *APIv1) RecoverPanicMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				httputils.HTTPRespondFailed(
					w,
					constant.APIv1,
					http.StatusInternalServerError,
					"internal server error",
					err,
				)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

type Config interface {
	Logger() *logger.Logger
}
