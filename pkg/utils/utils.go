// Copyright 2021 CloudWeGo Authors
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

package utils

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// RandomString generates a random string of the given length.
func RandomString(length int) string {
	chars := "abcdefghijklmnopqrstuvwxyz0123456789"
	bytes := make([]byte, length)
	for i := 0; i < length; i++ {
		bytes[i] = chars[rand.Intn(len(chars))]
	}
	return string(bytes)
}

func Int32Ptr(v int32) *int32    { return &v }
func Int64Ptr(v int64) *int64    { return &v }
func StringPtr(v string) *string { return &v }
func BoolPtr(v bool) *bool       { return &v }
