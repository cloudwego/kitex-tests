/*
 * Copyright 2024 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package serverutils

import (
	"fmt"
	"log"
	"net"
	"runtime"
	"sync/atomic"
	"time"
)

var listenStartPort = int32(10000) // no need `listenEndPort`?

func init() {
	// rand add 0, 300, 600, ... , 4800 for running tests in concurrency
	listenStartPort += int32((time.Now().UnixNano() % 17) * 300)
}

// NextListenAddr returns a local addr that can be used to call net.Listen
func NextListenAddr() string {
	for i := 0; i < 100; i++ {
		n := atomic.AddInt32(&listenStartPort, 1)
		addr := fmt.Sprintf("localhost:%d", n)
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			continue
		}
		ln.Close()
		return addr
	}
	panic("not able to get listen addr")
}

// Wait waits utils it's able to connect to the given addr
func Wait(addr string) {
	time.Sleep(5 * time.Millisecond) // likely it's up
	_, file, no, _ := runtime.Caller(1)
	for i := 0; i < 50; i++ { // 5s
		if ln, err := net.Dial("tcp", addr); err == nil {
			ln.Close()
			log.Printf("server %s is up @ %s:%d", addr, file, no)
			return
		}
		log.Printf("waiting server %s @ %s:%d", addr, file, no)
		time.Sleep(100 * time.Millisecond)
	}
	panic("server " + addr + " not ready")
}
