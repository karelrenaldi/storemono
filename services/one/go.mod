module github.com/karelrenaldi/storemono/services/one

go 1.16

require (
	github.com/karelrenaldi/storemono/libs/hello v0.0.0
	github.com/labstack/echo/v4 v4.6.3
)

replace github.com/karelrenaldi/storemono/libs/hello v0.0.0 => ../../libs/hello
