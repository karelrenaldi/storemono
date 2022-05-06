module github.com/karelrenaldi/storemono/services/shop-service

go 1.16

require (
	github.com/gorilla/mux v1.8.0
	github.com/joho/godotenv v1.4.0
	github.com/karelrenaldi/storemono/libs/http-utils v0.0.0
	github.com/karelrenaldi/storemono/libs/logger v0.0.0
	github.com/karelrenaldi/storemono/libs/smarthttp v0.0.0
	go.uber.org/zap v1.21.0
)

replace github.com/karelrenaldi/storemono/libs/logger v0.0.0 => ../../libs/logger

replace github.com/karelrenaldi/storemono/libs/http-utils v0.0.0 => ../../libs/http-utils

replace github.com/karelrenaldi/storemono/libs/smarthttp v0.0.0 => ../../libs/smarthttp
