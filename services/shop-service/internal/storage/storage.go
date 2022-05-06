package storage

import (
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"gitlab.com/ovoeng/lendingmono/funding-service/internal/database/model"
)

type Config interface {
	ConnStringMaster() string
	ConnStringSlave() string
	Dialect() string
	EnableLog() bool
	EnableAutoMigrate() bool
	MaxIdleConn() int
	MaxOpenConn() int
	ConnMaxLifetime() time.Duration
}

type DB struct {
	ormMaster *gorm.DB

	// this is for future extension, it will be easier if we keep what can be read from slave in mind
	// we may point the slave to the master when we don't have a slave instance
	ormSlave *gorm.DB
}

type TransactionFunc func(tx *gorm.DB) (err error)

func New(cfg Config) (*DB, error) {
	ormMaster, err := gorm.Open(cfg.Dialect(), cfg.ConnStringMaster())
	if err != nil {
		return nil, err
	}

	ormSlave, err := gorm.Open(cfg.Dialect(), cfg.ConnStringSlave())
	if err != nil {
		return nil, err
	}

	db := &DB{ormMaster, ormSlave}
	db.configORM(cfg)

	return db, nil
}

func (db *DB) Close() {
	db.ormMaster.Close()
	db.ormSlave.Close()
}

func (db *DB) Master() *gorm.DB {
	return db.ormMaster
}

func (db *DB) Slave() *gorm.DB {
	return db.ormSlave
}

func (db *DB) Transaction(fn TransactionFunc) (err error) {
	tx := db.Master().Begin()

	if err = fn(tx); err != nil {
		tx.Rollback()
		return
	}

	tx.Commit()

	return
}

func (db *DB) WithORM(orm *gorm.DB) DataService {
	return &DB{ormMaster: orm, ormSlave: orm}
}

func (db *DB) configORM(cfg Config) {
	db.ormMaster.SingularTable(true)
	db.ormMaster.LogMode(cfg.EnableLog())
	db.ormMaster.DB().SetMaxIdleConns(cfg.MaxIdleConn())
	db.ormMaster.DB().SetMaxOpenConns(cfg.MaxOpenConn())
	db.ormMaster.DB().SetConnMaxLifetime(cfg.ConnMaxLifetime())

	db.ormSlave.SingularTable(true)
	db.ormSlave.LogMode(cfg.EnableLog())
	db.ormSlave.DB().SetMaxIdleConns(cfg.MaxIdleConn())
	db.ormSlave.DB().SetMaxOpenConns(cfg.MaxOpenConn())
	db.ormSlave.DB().SetConnMaxLifetime(cfg.ConnMaxLifetime())

	// auto-migration should be only used by dev on local
	if cfg.EnableAutoMigrate() {
		db.ormMaster.AutoMigrate(&model.Application{})
		db.ormMaster.AutoMigrate(&model.LenderApplicationRelation{})
		db.ormMaster.AutoMigrate(&model.LenderProfile{})
		db.ormMaster.AutoMigrate(&model.LenderAccount{})
		db.ormMaster.AutoMigrate(&model.LenderAccountActivity{})
		db.ormMaster.AutoMigrate(&model.FundingRuleSet{})
		db.ormMaster.AutoMigrate(&model.FundingProportion{})
		db.ormMaster.AutoMigrate(&model.LenderBorrowerDetail{})
	}
}
