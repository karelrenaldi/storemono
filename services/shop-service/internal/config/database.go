package config

import (
	"os"
	"strconv"
	"time"
)

func getDBConfig() *DBConfig {
	enableLog, _ := strconv.ParseBool(os.Getenv("DB_ENABLE_LOG"))
	enableAutoMigrate, _ := strconv.ParseBool(os.Getenv("DB_ENABLE_AUTO_MIGRATE"))
	maxIdleConn, _ := strconv.Atoi(os.Getenv("DB_MAX_IDLE_CONN"))
	maxOpenConn, _ := strconv.Atoi(os.Getenv("DB_MAX_OPEN_CONN"))
	connMaxLifetimeSec, _ := strconv.Atoi(os.Getenv("DB_CONN_MAX_LIFETIME_SEC"))

	return &DBConfig{
		os.Getenv("DB_CONN_MASTER"),
		os.Getenv("DB_CONN_SLAVE"),
		os.Getenv("DB_DIALECT"),
		enableLog,
		enableAutoMigrate,
		maxIdleConn,
		maxOpenConn,
		time.Duration(connMaxLifetimeSec) * time.Second,
	}
}

// DBConfig is the configuration DTO used for DB client
type DBConfig struct {
	connStringMaster  string
	connStringSlave   string
	dialect           string
	enableLog         bool
	enableAutoMigrate bool
	maxIdleConn       int
	maxOpenConn       int
	connMaxLifetime   time.Duration
}

// Host returns the master DB connection string, i.e. user:password@(localhost)/dbname?charset=utf8
func (cfg *DBConfig) ConnStringMaster() string {
	return cfg.connStringMaster
}

// Host returns the slave DB connection string, i.e. user:password@(localhost)/dbname?charset=utf8
func (cfg *DBConfig) ConnStringSlave() string {
	return cfg.connStringSlave
}

// Dialect returns the DB dialect, i.e. mysql, postgres, sqlite, mssql
func (cfg *DBConfig) Dialect() string {
	return cfg.dialect
}

func (cfg *DBConfig) EnableLog() bool {
	return cfg.enableLog
}

func (cfg *DBConfig) EnableAutoMigrate() bool {
	return cfg.enableAutoMigrate
}

func (cfg *DBConfig) MaxIdleConn() int {
	return cfg.maxIdleConn
}

func (cfg *DBConfig) MaxOpenConn() int {
	return cfg.maxOpenConn
}

func (cfg *DBConfig) ConnMaxLifetime() time.Duration {
	return cfg.connMaxLifetime
}
