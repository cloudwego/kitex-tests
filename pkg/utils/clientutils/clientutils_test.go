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

package clientutils

import (
	"testing"
	"unsafe"
)

type clientWithClose struct {
	closed bool
}

func (c *clientWithClose) Close() error {
	c.closed = true
	return nil
}

func TestCallClose(t *testing.T) {
	cli := &clientWithClose{}
	t.Log("*clientWithClose", unsafe.Pointer(cli))

	type TestClient0 struct {
		*clientWithClose
	}
	c0 := TestClient0{cli}
	type TestClient1 struct {
		TestClient0
	}
	c1 := TestClient1{c0}
	type TestClient2 struct {
		cli TestClient1
	}
	c2 := &TestClient2{c1}
	CallClose(c2)
	if !cli.closed {
		t.Fatal("not close")
	}
}
