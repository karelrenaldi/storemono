package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/karelrenaldi/storemono/libs/logger"
	"github.com/karelrenaldi/storemono/services/shop-service/internal/constant"
	"go.uber.org/zap"
)

func NewServer(ctx context.Context) (*Server, error) {
	router := mux.NewRouter()

	cfg, ok := ctx.Value(constant.AppConfig).(ServerConfig)
	if !ok {
		fmt.Fprintf(os.Stderr, "failed to convert context with info type ServerConfig\n")
		return nil, errors.New("no config in ctx")
	}

	return &Server{
		logger: cfg.Logger(),
		server: &http.Server{
			Addr:         cfg.ServerAddress(),
			Handler:      router,
			ReadTimeout:  cfg.ReadTimeout(),
			WriteTimeout: cfg.WriteTimeout(),
		},
	}, nil
}

type Server struct {
	server *http.Server
	logger *logger.Logger
}

func (s *Server) Address() string {
	return s.server.Addr
}

func (s *Server) Listen() {
	fmt.Fprintf(os.Stderr, "starting server, address = %s\n", s.Address())

	s.logger.Info("starting server", zap.String("address", s.Address()))

	s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

type ServerConfig interface {
	ServerAddress() string

	Logger() *logger.Logger

	ReadTimeout() time.Duration

	WriteTimeout() time.Duration
}
