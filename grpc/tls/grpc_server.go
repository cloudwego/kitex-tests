// Copyright 2024 CloudWeGo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tls

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	grpc_demo "github.com/cloudwego/kitex-tests/grpc_gen/protobuf/grpc_demo_2"
)

func RunGRPCTLSServer(hostport string) (*grpc.Server, error) {
	tlsCredentials, err := serverLoadTLSCredentials()
	if err != nil {
		return nil, err
	}
	cred := grpc.Creds(tlsCredentials)

	svr := grpc.NewServer(cred)
	ms := &GrpcServiceA{}
	grpc_demo.RegisterServiceAServer(svr, ms)
	listener, err := net.Listen("tcp", hostport)
	if err != nil {
		return nil, err
	}
	go svr.Serve(listener)
	return svr, nil
}

type GrpcServiceA struct {
	grpc_demo.UnimplementedServiceAServer
}

func (s *GrpcServiceA) CallUnary(ctx context.Context, req *grpc_demo.Request) (*grpc_demo.Reply, error) {
	res := &grpc_demo.Reply{Message: req.Name + " Hello!"}
	return res, nil
}

func (s *GrpcServiceA) CallClientStream(stream grpc_demo.ServiceA_CallClientStreamServer) error {
	var msgs []string
	for {
		req, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		msgs = append(msgs, req.Name)
	}
	return stream.SendAndClose(&grpc_demo.Reply{Message: "all message: " + strings.Join(msgs, ", ")})
}
func (s *GrpcServiceA) CallServerStream(req *grpc_demo.Request, stream grpc_demo.ServiceA_CallServerStreamServer) error {
	resp := &grpc_demo.Reply{}
	for i := 0; i < 3; i++ {
		resp.Message = fmt.Sprintf("%v-%d", req.Name, i)
		err := stream.Send(resp)
		if err != nil {
			return err
		}
	}
	return nil
}
func (s *GrpcServiceA) CallBidiStream(stream grpc_demo.ServiceA_CallBidiStreamServer) error {
	for {
		recv, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		resp := &grpc_demo.Reply{}
		resp.Message = recv.Name
		err = stream.Send(resp)
		if err != nil {
			return err
		}
	}
	return nil
}

func serverLoadTLSCredentials() (credentials.TransportCredentials, error) {
	// Load certificate of the CA who signed client's certificate
	pemClientCA, err := ioutil.ReadFile("cert/ca-cert.pem")
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemClientCA) {
		return nil, fmt.Errorf("failed to add client CA's certificate")
	}

	// Load server's certificate and private key
	serverCert, err := tls.LoadX509KeyPair("cert/server-cert.pem", "cert/server-key.pem")
	if err != nil {
		return nil, err
	}

	// Create the credentials and return it
	config := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert, // mTLS
		ClientCAs:    certPool,
	}

	return credentials.NewTLS(config), nil
}
