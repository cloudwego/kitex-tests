package http1

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/kitex/pkg/endpoint"
	"github.com/cloudwego/kitex/pkg/remote/trans/detection"
	"github.com/cloudwego/kitex/pkg/remote/trans/http1"
	"github.com/cloudwego/kitex/pkg/remote/trans/netpoll"
	"github.com/cloudwego/kitex/server"

	"github.com/cloudwego/kitex-tests/http1/kitex_gen/thrift/stability"
	"github.com/cloudwego/kitex-tests/http1/kitex_gen/thrift/stability/stservice"
	"github.com/cloudwego/kitex-tests/pkg/test"
)

const (
	port = 8001
)

var middlewareExecuted uint32

func serverMiddleware() endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, req, resp interface{}) (err error) {
			err = next(ctx, req, resp)
			atomic.StoreUint32(&middlewareExecuted, 1)
			return err
		}
	}
}

func runServer() {
	address := &net.UnixAddr{Net: "tcp", Name: fmt.Sprintf(":%d", port)}
	svr := stservice.NewServer(new(STServiceHandler), server.WithServiceAddr(address), server.WithMiddleware(serverMiddleware()),
		server.WithTransHandlerFactory(detection.NewSvrTransHandlerFactory(netpoll.NewSvrTransHandlerFactory(), http1.NewSvrTransHandlerFactory())))

	err := svr.Run()
	if err != nil {
		log.Println(err.Error())
	}
}

var cli *http.Client

func TestMain(m *testing.M) {
	go func() {
		runServer()
	}()
	cli = &http.Client{Timeout: 1 * time.Second}
	time.Sleep(1 * time.Second)
	m.Run()
}

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// score: 55
func TestBody(t *testing.T) {
	_, req := CreateSTRequest(context.Background())
	buf, _ := sonic.Marshal(req)
	request, err := http.NewRequest("POST", "http://localhost:8001/api/STService/testSTReq", bytes.NewBuffer(buf))
	test.Assert(t, err == nil)
	response, err := cli.Do(request)
	test.Assert(t, err == nil)
	test.Assert(t, response.StatusCode == 200)
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}
	resp := stability.NewSTResponse()
	rresp := &Response{Data: resp}
	err = sonic.Unmarshal(body, rresp)
	test.Assert(t, err == nil)
	test.Assert(t, *resp.Str == *req.Str)
	test.Assert(t, reflect.DeepEqual(resp.Mp, req.StringMap))
}

// score: 5
func TestTimeout(t *testing.T) {
	waitMS := "2s"
	req := &stability.STRequest{MockCost: &waitMS}
	buf, _ := sonic.Marshal(req)
	request, err := http.NewRequest("POST", "http://localhost:8001/api/STService/testSTReq", bytes.NewBuffer(buf))
	test.Assert(t, err == nil)
	_, err = cli.Do(request)
	test.Assert(t, strings.Contains(err.Error(), "context deadline exceeded"))
}

// score: 25
func TestHeaderAndQuery(t *testing.T) {
	query := make(url.Values)
	query.Add("framework", "kitex")
	request, err := http.NewRequest("POST", fmt.Sprintf("http://localhost:8001/api/STService/testSTReq?%s", query.Encode()), nil)
	request.Header.Set("X-User-Id", "123456")
	test.Assert(t, err == nil)
	response, err := cli.Do(request)
	test.Assert(t, err == nil)
	test.Assert(t, response.StatusCode == 200)
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}
	resp := stability.NewSTResponse()
	rresp := &Response{Data: resp}
	err = sonic.Unmarshal(body, rresp)
	test.Assert(t, err == nil)
	test.Assert(t, *resp.Name == "123456")
	test.Assert(t, *resp.Framework == "kitex")
}

// score: 5
func TestMiddleware(t *testing.T) {
	request, err := http.NewRequest("POST", "http://localhost:8001/api/STService/testSTReq", nil)
	test.Assert(t, err == nil)
	response, err := cli.Do(request)
	test.Assert(t, err == nil)
	test.Assert(t, response.StatusCode == 200)
	response.Body.Close()
	test.Assert(t, atomic.LoadUint32(&middlewareExecuted) == 1)
}

// score: 10
func TestBizError(t *testing.T) {
	str := "biz_err"
	req := &stability.STRequest{
		Str: &str,
	}
	buf, _ := sonic.Marshal(req)
	request, err := http.NewRequest("POST", "http://localhost:8001/api/STService/testSTReq", bytes.NewBuffer(buf))
	test.Assert(t, err == nil)
	response, err := cli.Do(request)
	test.Assert(t, err == nil)
	test.Assert(t, response.StatusCode == 200)
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}
	rresp := &Response{}
	err = sonic.Unmarshal(body, rresp)
	test.Assert(t, err == nil)
	test.Assert(t, rresp.Code == 10001)
	test.Assert(t, rresp.Message == "biz error")
}
