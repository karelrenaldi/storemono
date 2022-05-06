package config

import (
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/karelrenaldi/storemono/libs/logger"
	"go.uber.org/zap"
)

const (
	readTimeoutDefault        = 2 * time.Second
	writeTimeoutDefault       = 10 * time.Second
	httpClientTimeoutDefault  = 5 * time.Second
	httpRetryAttemptsDefault  = 3
	httpRetryDelayDefault     = 10 * time.Millisecond
	httpRetryMaxDelayDefault  = 1 * time.Second
	httpMaxConcurrencyDefault = 10
)

func New() (*AppConfig, error) {
	serverAddress, err := getServerAddress()
	if err != nil {
		return nil, err
	}

	zapLogger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}

	readTimeout, writeTimeout := getServerTimeout()
	cliTimeout, retryDelay, retryMaxDelay, retryAttempts, concurrency := getHTTPClientConfig()

	return &AppConfig{
		serverAddress:      serverAddress,
		logger:             logger.NewLogger(zapLogger),
		readTimeout:        readTimeout,
		writeTimeout:       writeTimeout,
		dbConfig:           getDBConfig(),
		httpClientTimeout:  cliTimeout,
		httpRetryDelay:     retryDelay,
		httpRetryMaxDelay:  retryMaxDelay,
		httpRetryAttempts:  retryAttempts,
		httpMaxConcurrency: concurrency,
	}, nil
}

func getServerAddress() (string, error) {
	applicationHost := os.Getenv("APPLICATION_HOST")
	applicationPort := os.Getenv("APPLICATION_PORT")

	if applicationHost == "" || applicationPort == "" {
		return "", errors.New("application host or port is empty")
	}

	return os.Getenv("APPLICATION_HOST") + ":" + os.Getenv("APPLICATION_PORT"), nil
}

func getServerTimeout() (readTimeout, writeTimeout time.Duration) {
	readTimeout = readTimeoutDefault

	if os.Getenv("SERVER_READ_TIMEOUT") != "" {
		convertedReadTimeout, err := strconv.Atoi(os.Getenv("SERVER_READ_TIMEOUT"))
		if err == nil {
			readTimeout = time.Duration(convertedReadTimeout) * time.Second
		}
	}

	writeTimeout = writeTimeoutDefault

	if os.Getenv("SERVER_WRITE_TIMEOUT") != "" {
		convertedWriteTimeout, err := strconv.Atoi(os.Getenv("SERVER_WRITE_TIMEOUT"))
		if err == nil {
			writeTimeout = time.Duration(convertedWriteTimeout) * time.Second
		}
	}

	return
}

func getHTTPClientConfig() (cliTimeout, retryDelay, retryMaxDelay time.Duration, retryAttempts, concurrency int) {
	cliTimeout = httpClientTimeoutDefault
	retryDelay = httpRetryDelayDefault
	retryMaxDelay = httpRetryMaxDelayDefault
	retryAttempts = httpRetryAttemptsDefault
	concurrency = httpMaxConcurrencyDefault

	if v, err := strconv.Atoi(os.Getenv("HTTP_CLIENT_TIMEOUT_MS")); err == nil {
		cliTimeout = time.Duration(v) * time.Millisecond
	}

	if v, err := strconv.Atoi(os.Getenv("HTTP_RETRY_DELAY_MS")); err == nil {
		retryDelay = time.Duration(v) * time.Millisecond
	}

	if v, err := strconv.Atoi(os.Getenv("HTTP_RETRY_MAX_DELAY_MS")); err == nil {
		retryMaxDelay = time.Duration(v) * time.Millisecond
	}

	if v, err := strconv.Atoi(os.Getenv("HTTP_RETRY_ATTEMPTS")); err == nil {
		retryAttempts = v
	}

	if v, err := strconv.Atoi(os.Getenv("HTTP_CLIENT_MAX_CONCURRENCY")); err == nil {
		concurrency = v
	}

	return
}

type AppConfig struct {
	serverAddress      string
	logger             *logger.Logger
	readTimeout        time.Duration
	writeTimeout       time.Duration
	dbConfig           *DBConfig
	httpClientTimeout  time.Duration
	httpRetryDelay     time.Duration
	httpRetryMaxDelay  time.Duration
	httpRetryAttempts  int
	httpMaxConcurrency int
}

// ServerAddress returns the server listening address
func (cfg *AppConfig) ServerAddress() string {
	return cfg.serverAddress
}

// Logger returns the logging client
func (cfg *AppConfig) Logger() *logger.Logger {
	return cfg.logger
}

// ReadTimeout returns the server read timeout
func (cfg *AppConfig) ReadTimeout() time.Duration {
	return cfg.readTimeout
}

// WriteTimeout returns the server write timeout
func (cfg *AppConfig) WriteTimeout() time.Duration {
	return cfg.writeTimeout
}

// DBConfig returns the db configuration
func (cfg *AppConfig) DBConfig() *DBConfig {
	return cfg.dbConfig
}

// HTTPClientTimeout returns the timeout for the smarthttp client
func (cfg *AppConfig) HTTPClientTimeout() time.Duration {
	return cfg.httpClientTimeout
}

// HTTPRetryAttempts returns the max attempts for the smarthttp client
func (cfg *AppConfig) HTTPRetryAttempts() int {
	return cfg.httpRetryAttempts
}

// HTTPRetryDelay returns the base delay for the smarthttp client
func (cfg *AppConfig) HTTPRetryDelay() time.Duration {
	return cfg.httpRetryDelay
}

// HTTPRetryMaxDelay returns the max delay for the smarthttp client
func (cfg *AppConfig) HTTPRetryMaxDelay() time.Duration {
	return cfg.httpRetryMaxDelay
}

// HTTPMaxConcurrency returns the max concurrency for the smarthttp client
func (cfg *AppConfig) HTTPMaxConcurrency() int {
	return cfg.httpMaxConcurrency
}
