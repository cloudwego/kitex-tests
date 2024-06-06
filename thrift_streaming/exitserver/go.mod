module github.com/cloudwego/kitex-tests/thrift_streaming/exitserver

go 1.16

replace github.com/apache/thrift => github.com/apache/thrift v0.13.0

replace github.com/cloudwego/kitex => ../../../kitex

require (
	github.com/apache/thrift v0.13.0
	github.com/cloudwego/kitex v0.9.3-rc
)
