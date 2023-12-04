// Copyright 2023 CloudWeGo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package test

import (
	"bytes"
	"math"
	"strconv"
	"strings"
	"testing"

	"test/kitex_gen/baseline"

	"github.com/cloudwego/thriftgo/fieldmask"
)

var (
	bytesCount  int = 2
	stringCount int = 2
	listCount   int = 16
)

func getString() string {
	return strings.Repeat("你好,\b\n\r\t世界", stringCount)
}

func getBytes() []byte {
	return bytes.Repeat([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}, bytesCount)
}

func getSimpleValue() *baseline.Simple {
	return &baseline.Simple{
		ByteField:   math.MaxInt8,
		I64Field:    math.MaxInt64,
		DoubleField: math.MaxFloat64,
		I32Field:    math.MaxInt32,
		StringField: getString(),
		BinaryField: getBytes(),
	}
}

func getNestingValue() *baseline.Nesting {
	var ret = &baseline.Nesting{
		String_:         getString(),
		ListSimple:      []*baseline.Simple{},
		Double:          math.MaxFloat64,
		I32:             math.MaxInt32,
		ListI32:         []int32{},
		I64:             math.MaxInt64,
		MapStringString: map[string]string{},
		SimpleStruct:    getSimpleValue(),
		MapI32I64:       map[int32]int64{},
		ListString:      []string{},
		Binary:          getBytes(),
		MapI64String:    map[int64]string{},
		ListI64:         []int64{},
		Byte:            math.MaxInt8,
		MapStringSimple: map[string]*baseline.Simple{},
	}

	for i := 0; i < listCount; i++ {
		ret.ListSimple = append(ret.ListSimple, getSimpleValue())
		ret.ListI32 = append(ret.ListI32, math.MinInt32)
		ret.ListI64 = append(ret.ListI64, math.MinInt64)
		ret.ListString = append(ret.ListString, getString())
	}

	for i := 0; i < listCount; i++ {
		ret.MapStringString[strconv.Itoa(i)] = getString()
		ret.MapI32I64[int32(i)] = math.MinInt64
		ret.MapI64String[int64(i)] = getString()
		ret.MapStringSimple[strconv.Itoa(i)] = getSimpleValue()
	}

	return ret
}

func BenchmarkFastWriteSimple(b *testing.B) {
	b.Run("full", func(b *testing.B) {
		obj := getSimpleValue()
		data := make([]byte, obj.BLength())
		ret := obj.FastWriteNocopy(data, nil)
		if ret != len(data) {
			b.Fatal(ret)
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = obj.BLength()
			_ = obj.FastWriteNocopy(data, nil)
		}
	})
	b.Run("half", func(b *testing.B) {
		obj := getSimpleValue()
		fm, err := fieldmask.NewFieldMask(obj.GetTypeDescriptor(), "$.ByteField", "$.DoubleField", "$.StringField")
		if err != nil {
			b.Fatal(err)
		}
		obj.Set_FieldMask(fm)
		data := make([]byte, obj.BLength())
		ret := obj.FastWriteNocopy(data, nil)
		if ret != len(data) {
			b.Fatal(ret)
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			obj.Set_FieldMask(fm)
			_ = obj.BLength()
			_ = obj.FastWriteNocopy(data, nil)
		}
	})
}

func BenchmarkFastReadSimple(b *testing.B) {
	b.Run("full", func(b *testing.B) {
		obj := getSimpleValue()
		data := make([]byte, obj.BLength())
		ret := obj.FastWriteNocopy(data, nil)
		if ret != len(data) {
			b.Fatal(ret)
		}
		obj = baseline.NewSimple()
		n, err := obj.FastRead(data)
		if n != len(data) {
			b.Fatal(err)
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = obj.FastRead(data)
		}
	})
	b.Run("half", func(b *testing.B) {
		obj := getSimpleValue()
		fm, err := fieldmask.NewFieldMask(obj.GetTypeDescriptor(),
			"$.ByteField", "$.DoubleField", "$.StringField")
		if err != nil {
			b.Fatal(err)
		}
		data := make([]byte, obj.BLength())
		ret := obj.FastWriteNocopy(data, nil)
		if ret != len(data) {
			b.Fatal(ret)
		}
		obj = baseline.NewSimple()
		obj.Set_FieldMask(fm)
		n, err := obj.FastRead(data)
		if n != len(data) {
			b.Fatal(err)
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			obj.Set_FieldMask(fm)
			_, _ = obj.FastRead(data)
		}
	})
}

func BenchmarkFastWriteNesting(b *testing.B) {
	b.Run("full", func(b *testing.B) {
		obj := getNestingValue()
		data := make([]byte, obj.BLength())
		ret := obj.FastWriteNocopy(data, nil)
		if ret != len(data) {
			b.Fatal(ret)
		}
		// println("full data size: ", len(data))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = obj.BLength()
			_ = obj.FastWriteNocopy(data, nil)
		}
	})
	b.Run("half", func(b *testing.B) {
		obj := getNestingValue()
		// ins := []string{}
		// for i := 0; i < listCount/2; i++ {
		// 	ins = append(ins, strconv.Itoa(i))
		// }
		// is := strings.Join(ins, ",")
		// ss := strings.Join(ins, `","`)
		fm, err := fieldmask.NewFieldMask(obj.GetTypeDescriptor(),
			// "$.ListSimple["+is+"]", "$.I32", "$.ListI32["+is+"]", `$.MapStringString{"`+ss+`"}`,
			// "$.MapI32I64{"+is+"}", "$.Binary", "$.ListI64["+is+"]", `$.MapStringSimple{"`+ss+`"}`,
			"$.ListSimple", "$.I32", "$.ListI32", `$.MapStringString`,
			"$.MapI32I64", "$.Binary",
		)
		if err != nil {
			b.Fatal(err)
		}
		obj.Set_FieldMask(fm)
		data := make([]byte, obj.BLength())
		ret := obj.FastWriteNocopy(data, nil)
		if ret != len(data) {
			b.Fatal(ret)
		}
		// println("half data size: ", len(data))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			obj.Set_FieldMask(fm)
			_ = obj.BLength()
			_ = obj.FastWriteNocopy(data, nil)
		}
	})
}

func BenchmarkFastReadNesting(b *testing.B) {
	b.Run("full", func(b *testing.B) {
		obj := getNestingValue()
		data := make([]byte, obj.BLength())
		ret := obj.FastWriteNocopy(data, nil)
		if ret != len(data) {
			b.Fatal(ret)
		}
		obj = baseline.NewNesting()
		n, err := obj.FastRead(data)
		if n != len(data) {
			b.Fatal(err)
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = obj.FastRead(data)
		}
	})
	b.Run("half", func(b *testing.B) {
		obj := getNestingValue()
		// ins := []string{}
		// for i := 0; i < listCount/2; i++ {
		// 	ins = append(ins, strconv.Itoa(i))
		// }
		// is := strings.Join(ins, ",")
		// ss := strings.Join(ins, `","`)
		fm, err := fieldmask.NewFieldMask(obj.GetTypeDescriptor(),
			// "$.ListSimple["+is+"]", "$.I32", "$.ListI32["+is+"]", `$.MapStringString{"`+ss+`"}`,
			// "$.MapI32I64{"+is+"}", "$.Binary", "$.ListI64["+is+"]", `$.MapStringSimple{"`+ss+`"}`,
			"$.ListSimple", "$.I32", "$.ListI32", `$.MapStringString`,
			"$.MapI32I64", "$.Binary",
		)
		if err != nil {
			b.Fatal(err)
		}
		data := make([]byte, obj.BLength())
		ret := obj.FastWriteNocopy(data, nil)
		if ret != len(data) {
			b.Fatal(ret)
		}
		obj = baseline.NewNesting()
		obj.Set_FieldMask(fm)
		n, err := obj.FastRead(data)
		if n != len(data) {
			b.Fatal(err)
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			obj.Set_FieldMask(fm)
			_, _ = obj.FastRead(data)
		}
	})
}
