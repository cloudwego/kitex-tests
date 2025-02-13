#! /bin/bash
# Copyright 2021 CloudWeGo Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -e

# add time for each command for timing
export PS4='[$(date "+%F %T")] '

bits=$(getconf LONG_BIT)
if [[ $bits != 64 ]]; then
    echo "this script runs on 64-bit architectures only" >&2
    exit 1
fi

set -x

PATH_BIN=$PWD/bin
mkdir -p $PATH_BIN

export PATH=$PATH_BIN:$PATH

# GOBIN need to be changed here as kitex-tests will install different version bin when testing,
# GOPATH remains the same as default,
# coz other runners in the same host may share path like GOMODCACHE, GOCACHE
export GOBIN=$PATH_BIN

PROTOC_VERSION=v3.20.2

install_protoc() {
    echo "installing protoc ... "
    os=`uname -s | sed 's/Darwin/osx/'`
    arch=`uname -m | sed 's/amd64/x86_64/' | sed 's/arm64/aarch_64/'`
    suffix=$(echo ${os}-${arch})
    filename=protoc-${PROTOC_VERSION#v}-${suffix}.zip
    url=https://github.com/protocolbuffers/protobuf/releases/download/${PROTOC_VERSION}/${filename}
    rm -f $filename
    wget -q $url || exit 1
    unzip -o -q $filename -d ./tmp/
    mv ./tmp/bin/protoc $PATH_BIN
    rm -rf ./tmp/
    echo "installing protoc ... done"
}

# use abs path, the env will be passed to and used by other scripts in sub dirs
if [[ -n $1 ]]; then
    LOCAL_REPO=`realpath $1`
fi

# export and reuse the latest version,
# then no need to query the remote develop branch multiple times for it.
export KITEX_LATEST_VERSION=develop # default value if not set
if [[ -z $LOCAL_REPO ]]; then
    export KITEX_LATEST_VERSION=`go list -m github.com/cloudwego/kitex@develop | cut -d" " -f2`
fi

go_install() {
    echo "installing $@ ..."
    go install $@ || go get $@
    echo "installing $@ ... done"
}

kitex_cmd() {
  kitex --no-dependency-check $@
}

echo -e "\ninstalling missing commands\n"

# install protoc
which protoc || install_protoc &

# install protoc-gen-go and protoc-gen-go-kitexgrpc
which protoc-gen-go || go_install google.golang.org/protobuf/cmd/protoc-gen-go@latest &

# install protoc-gen-go and protoc-gen-go-kitexgrpc
which protoc-gen-go-grpc || go_install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest &

# install thriftgo
which thriftgo || go_install github.com/cloudwego/thriftgo@latest &

# install kitex

if [[ -n $LOCAL_REPO ]]; then
    cd ${LOCAL_REPO}
    go_install ${LOCAL_REPO}/tool/cmd/kitex
    cd -
else
    go_install github.com/cloudwego/kitex/tool/cmd/kitex@$KITEX_LATEST_VERSION
fi

wait
echo -e "\ninstalling missing commands ... done\n"

# double check commands,
# set -e may not working since commands run in background
protoc --version
protoc-gen-go --version
protoc-gen-go-grpc --version
thriftgo --version
kitex -version

echo -e "\ngenerating code for testing ...\n"

rm -rf kitex_gen
kitex_cmd -module github.com/cloudwego/kitex-tests ./idl/stability.thrift
kitex_cmd -module github.com/cloudwego/kitex-tests ./idl/http.thrift
kitex_cmd -module github.com/cloudwego/kitex-tests ./idl/tenant.thrift
kitex_cmd -module github.com/cloudwego/kitex-tests -combine-service ./idl/combine_service.thrift
kitex_cmd -module github.com/cloudwego/kitex-tests ./idl/thrift_multi_service.thrift
kitex_cmd -module github.com/cloudwego/kitex-tests ./idl/thrift_multi_service_2.thrift
kitex_cmd -module github.com/cloudwego/kitex-tests -I idl ./idl/stability.proto
kitex_cmd -module github.com/cloudwego/kitex-tests -I idl ./idl/unknown_handler.proto
kitex_cmd -module github.com/cloudwego/kitex-tests -I idl ./idl/grpc_demo.proto
kitex_cmd -module github.com/cloudwego/kitex-tests -I idl -combine-service ./idl/grpc_multi_service.proto
kitex_cmd -module github.com/cloudwego/kitex-tests -I idl ./idl/grpc_multi_service_2.proto
kitex_cmd -module github.com/cloudwego/kitex-tests -I idl ./idl/pb_multi_service.proto
kitex_cmd -module github.com/cloudwego/kitex-tests -I idl -combine-service ./idl/combine_service.proto

rm -rf kitex_gen_slim
kitex_cmd -module github.com/cloudwego/kitex-tests -thrift template=slim -gen-path kitex_gen_slim ./idl/stability.thrift

rm -rf kitex_gen_noDefSerdes
kitex_cmd -module github.com/cloudwego/kitex-tests -thrift no_default_serdes -gen-path kitex_gen_noDefSerdes ./idl/stability.thrift

rm -rf grpc_gen
mkdir grpc_gen
protoc --go_out=grpc_gen/. ./idl/grpc_demo_2.proto
protoc --go-grpc_out=grpc_gen/. ./idl/grpc_demo_2.proto

# generate thrift streaming code
LOCAL_REPO=$LOCAL_REPO ./thrift_streaming/generate.sh
# generate thrift streaming code with streamx mode
cd ./streamx && rm -rf ./kitex_gen && kitex_cmd -streamx -I idl ./idl/echo.thrift && kitex_cmd -streamx -I idl ./idl/echo.proto && cd -

echo -e "\ngenerating code for testing ... done\n"


echo -e "\nupdating dependencies ... \n"

# Init dependencies
rm -f go.mod
go mod init github.com/cloudwego/kitex-tests

if [[ -n $LOCAL_REPO ]]; then
    go mod edit -replace github.com/cloudwego/kitex=${LOCAL_REPO}
    go mod edit -replace github.com/cloudwego/kitex/pkg/protocol/bthrift=${LOCAL_REPO}/pkg/protocol/bthrift
else
    go get github.com/cloudwego/kitex@$KITEX_LATEST_VERSION
fi

fixed_version() {
    go mod edit -replace $1=$1@$2
}

# ...
fixed_version github.com/apache/thrift v0.13.0

# https://github.com/googleapis/go-genproto/issues/1015
fixed_version google.golang.org/genproto v0.0.0-20250227231956-55c901821b1e
fixed_version google.golang.org/genproto/googleapis/rpc v0.0.0-20250227231956-55c901821b1e

# used by this repo, and it takes seconds for `go mod tidy`
# can we get rid of it one day? used in kitexgrpc/normalcall/normalcall_test.go
fixed_version github.com/shirou/gopsutil/v3 v3.24.5

go mod tidy

echo -e "\nupdating dependencies ... done\n"

# run tests
# use less cores for stability
# coz some of them are logical cores
nproc=$(nproc)
export GOMAXPROCS=$(( 3 * $nproc / 4 ))

echo -e "\nrunning tests ... \n"

# skip kitex_gen dirs which have no tests
test_modules=`go list ./... | grep -v kitex_gen | grep -v grpc_gen | grep -v kitexgrpc/abc/`

if [[ -n $LOCAL_REPO && -n $CI ]]; then
    # only generate coverage file in ci env
    # CI=true will be set by Github runner or you can set by yourself
    # It will be slower if generate code coverage
    go test -failfast -covermode=atomic -coverprofile=${LOCAL_REPO}/coverage.txt -coverpkg=github.com/cloudwego/kitex/... $test_modules
   if [[ "$OSTYPE" =~ ^darwin ]]; # remove `mode: atomic` line
   then
       sed -i '' 1d ${LOCAL_REPO}/coverage.txt
   else
       sed -i '1d' ${LOCAL_REPO}/coverage.txt
   fi
else
    go test -failfast $test_modules
fi

echo "PASSED"
