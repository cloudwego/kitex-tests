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

package utils

import (
	"io"
	"reflect"
	"unsafe"
)

const maxdepath = 10

// CallClose calls underlying Close method which provided by kitex/client.
// XXX: kitex tool doesn't generate Close method for client, and we need it for testing
// or Server.Stop will block till timeout
func CallClose(v interface{}) bool {
	// likely ok=false, or we should call Close directly
	if f, ok := v.(io.Closer); ok {
		f.Close()
		return true
	}
	rv := reflect.ValueOf(v)
	p := rv.UnsafePointer()
	ret := callClose(rv, p, maxdepath)
	return ret
}

// callClose returns true if it can stop
func callClose(rv reflect.Value, p unsafe.Pointer, depth int) bool {
	if depth == 0 {
		return false
	}
	if rv.Kind() == reflect.Pointer {
		if tryCallClose(rv, p) {
			return true
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Interface && rv.Kind() != reflect.Struct {
		return false
	}
	if rv.Kind() == reflect.Interface {
		rv = rv.Elem()
		p = unsafe.Add(p, unsafe.Sizeof(uintptr(0)))
		p = *(*unsafe.Pointer)(p) // get data pointer of eface
		if tryCallClose(rv, p) {
			return true
		}
		if rv.Kind() == reflect.Pointer {
			rv = rv.Elem()
		}
		if rv.Kind() != reflect.Struct {
			panic("unknown type: " + rv.Type().String())
		}
	}
	rv = rv.Field(0) // for generated code, always 1 field
	if rv.Kind() == reflect.Pointer {
		p = *(*unsafe.Pointer)(p)
	}
	return callClose(rv, p, depth-1)
}

func tryCallClose(rv reflect.Value, p unsafe.Pointer) bool {
	// check pointer to struct
	if rv.Kind() != reflect.Pointer {
		return false
	}
	if rv.Elem().Kind() != reflect.Struct {
		return false
	}

	// use reflect.NewAt to bypass unexported field issue
	v := reflect.NewAt(rv.Type().Elem(), p)
	f, ok := v.Interface().(io.Closer)
	if ok {
		f.Close()
		return true
	}
	return false
}
