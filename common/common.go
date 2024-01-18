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

package common

import (
	"net"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
)

func WaitServer(hostPort string) {
	for begin := time.Now(); time.Since(begin) < time.Second; {
		if _, err := net.Dial("tcp", hostPort); err == nil {
			klog.Infof("server %s is up", hostPort)
			return
		}
		time.Sleep(time.Millisecond * 10)
	}
}
