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
set -x

export GO111MODULE=on
export GOBIN=$(pwd)/bin
export PATH=${GOBIN}:$PATH
mkdir -p ${GOBIN}

bits=$(getconf LONG_BIT)
if [[ $bits != 64 ]]; then
    echo "this script runs on 64-bit architectures only" >&2
    exit 1
fi

# Install protoc

get_protoc() {
    os=$1
    arch=$2
    out=$3
    suffix=$(echo ${os}-${arch} | sed 's/darwin/osx/' | sed 's/amd64/x86_64/' | sed 's/arm64/aarch_64/')
    release=protoc-${PROTOC_VERSION#v}-${suffix}.zip
    url=https://github.com/protocolbuffers/protobuf/releases/download/${PROTOC_VERSION}/${release}
    wget -q $url || exit 1
    python -m zipfile -e $release $os || exit 1
    chmod +x $os/bin/protoc
    mv $os/bin/protoc $out/protoc-${os}-${arch} && rm -rf $os $release
}

install_protoc() {
    PROTOC_VERSION=v3.13.0
    OUT=./bin
    export PATH=$OUT:$PATH
    mkdir -p $OUT

    get_protoc darwin amd64 $OUT
    get_protoc linux amd64 $OUT
    get_protoc linux arm64 $OUT
    for p in $OUT/protoc-*; do
        "$p" --version 2>/dev/null && ln -s $(basename $p) $OUT/protoc || true
    done
}

go_install() {
    go install $@ || go get $@
}

which protoc || install_protoc

# Install thriftgo
which thriftgo || go_install github.com/cloudwego/thriftgo@latest

# Install kitex and generate codes
LOCAL_REPO=$1

if [[ -n $LOCAL_REPO ]]; then
    cd ${LOCAL_REPO}
    go_install ${LOCAL_REPO}/tool/cmd/kitex
    cd -
else
    go_install github.com/cloudwego/kitex/tool/cmd/kitex@latest
fi

test -d kitex_gen && rm -rf kitex_gen
kitex -module github.com/cloudwego/kitex-tests ./idl/stability.thrift
kitex -module github.com/cloudwego/kitex-tests ./idl/http.thrift
kitex -module github.com/cloudwego/kitex-tests ./idl/tenant.thrift
kitex -module github.com/cloudwego/kitex-tests -type protobuf -I idl ./idl/stability.proto
kitex -module github.com/cloudwego/kitex-tests -type protobuf -I idl ./idl/unknown_handler.proto


# Init dependencies
go get github.com/apache/thrift@v0.13.0
go get github.com/cloudwego/kitex@develop

if [[ -n $LOCAL_REPO ]]; then
    go mod edit -replace github.com/cloudwego/kitex=${LOCAL_REPO}
fi

go mod tidy

# static check
go vet -stdmethods=false $(go list ./...)
#go_install mvdan.cc/gofumpt@v0.2.0
#test -z "$(gofumpt -l -extra .)"

# run tests
packages=(
./thriftrpc/normalcall/...
./thriftrpc/muxcall/...
./thriftrpc/retrycall/...
./thriftrpc/failedcall/...
./thriftrpc/failedmux/...
./thriftrpc/abctest/...
./thriftrpc/backupctx/...
./pbrpc/normalcall/...
./pbrpc/muxcall/...
./pbrpc/failedcall/...
./generic/http/...
./generic/map/...
./grpc/...
)

for pkg in ${packages[@]}
do
    if [[ -n $LOCAL_REPO ]]; then
        go test -covermode=atomic -coverprofile=${LOCAL_REPO}/coverage.txt.tmp -coverpkg=github.com/cloudwego/kitex/... $pkg
        if [[ "$OSTYPE" =~ ^darwin ]];
        then
            sed -i '' 1d ${LOCAL_REPO}/coverage.txt.tmp
        else
            sed -i '1d' ${LOCAL_REPO}/coverage.txt.tmp
        fi
        cat ${LOCAL_REPO}/coverage.txt.tmp >> ${LOCAL_REPO}/coverage.txt
        rm ${LOCAL_REPO}/coverage.txt.tmp
    else
        go test $pkg
    fi
done
