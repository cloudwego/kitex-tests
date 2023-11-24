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
	"runtime"
	"test/kitex_gen/base"
	obase "test/old_gen/base"
	"testing"

	"github.com/cloudwego/thriftgo/fieldmask"
	"github.com/stretchr/testify/require"
)

func SampleNewBase() *base.Base {
	obj := base.NewBase()
	obj.Addr = "abcd"
	obj.Caller = "abcd"
	obj.LogID = "abcd"
	obj.Meta = base.NewMetaInfo()
	obj.Meta.StrMap = map[string]*base.Val{
		"abcd": base.NewVal(),
		"1234": base.NewVal(),
	}
	obj.Meta.IntMap = map[int64]*base.Val{
		1: base.NewVal(),
		2: base.NewVal(),
	}
	v0 := base.NewVal()
	v0.Id = "a"
	v0.Name = "a"
	v1 := base.NewVal()
	v1.Id = "b"
	v1.Name = "b"
	obj.Meta.List = []*base.Val{v0, v1}
	obj.Meta.Set = []*base.Val{v0, v1}
	obj.Extra = base.NewExtraInfo()
	obj.TrafficEnv = base.NewTrafficEnv()
	obj.TrafficEnv.Code = 1
	obj.TrafficEnv.Env = "abcd"
	obj.TrafficEnv.Name = "abcd"
	obj.TrafficEnv.Open = true
	return obj
}

func SampleOldBase() *obase.Base {
	obj := obase.NewBase()
	obj.Addr = "abcd"
	obj.Caller = "abcd"
	obj.LogID = "abcd"
	obj.Meta = obase.NewMetaInfo()
	obj.Meta.StrMap = map[string]*obase.Val{
		"abcd": obase.NewVal(),
		"1234": obase.NewVal(),
	}
	obj.Meta.IntMap = map[int64]*obase.Val{
		1: obase.NewVal(),
		2: obase.NewVal(),
	}
	v0 := obase.NewVal()
	v0.Id = "a"
	v0.Name = "a"
	v1 := obase.NewVal()
	v1.Id = "b"
	v1.Name = "b"
	obj.Meta.List = []*obase.Val{v0, v1}
	obj.Meta.Set = []*obase.Val{v0, v1}
	obj.Extra = obase.NewExtraInfo()
	obj.TrafficEnv = obase.NewTrafficEnv()
	obj.TrafficEnv.Code = 1
	obj.TrafficEnv.Env = "abcd"
	obj.TrafficEnv.Name = "abcd"
	obj.TrafficEnv.Open = true
	return obj
}

func TestFastWrite(t *testing.T) {
	var obj = SampleNewBase()
	fm, err := fieldmask.NewFieldMask(obj.GetTypeDescriptor(), "$.Addr",
		"$.LogID",
		"$.TrafficEnv.Code",
		"$.Meta.IntMap{1}.id",
		"$.Meta.IntMap{2}.name",
		"$.Meta.StrMap{\"1234\"}",
		"$.Meta.List[1]",
		"$.Meta.Set[0].id",
		"$.Meta.Set[1].name")
	if err != nil {
		t.Fatal(err)
	}
	obj.SetFieldMask(fm)
	out := make([]byte, obj.BLength())
	e := obj.FastWriteNocopy(out, nil)
	var obj2 = base.NewBase()
	n, err := obj2.FastRead(out)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, e, n)
	require.Equal(t, obj.Addr, obj2.Addr)
	require.Equal(t, obj.LogID, obj2.LogID)
	require.Equal(t, "", obj2.Caller)
	require.Equal(t, "", obj2.TrafficEnv.Name)
	require.Equal(t, false, obj2.TrafficEnv.Open)
	require.Equal(t, "", obj2.TrafficEnv.Env)
	require.Equal(t, obj.TrafficEnv.Code, obj2.TrafficEnv.Code)
	require.Equal(t, 2, len(obj2.Meta.IntMap))
	require.Equal(t, obj.Meta.IntMap[1].Id, obj2.Meta.IntMap[1].Id)
	require.Equal(t, "", obj2.Meta.IntMap[1].Name)
	require.Equal(t, "", obj2.Meta.IntMap[2].Id)
	require.Equal(t, obj.Meta.IntMap[2].Name, obj2.Meta.IntMap[2].Name)
	require.Equal(t, obj.Meta.StrMap["1234"].Id, obj2.Meta.StrMap["1234"].Id)
	require.Equal(t, (*base.Val)(nil), obj2.Meta.StrMap["abcd"])
	require.Equal(t, obj.Meta.List[1].Id, obj2.Meta.List[0].Id)
	require.Equal(t, obj.Meta.List[1].Name, obj2.Meta.List[0].Name)
	require.Equal(t, 1, len(obj2.Meta.List))
	require.Equal(t, 2, len(obj2.Meta.Set))
	require.Equal(t, obj.Meta.Set[0].Id, obj2.Meta.Set[0].Id)
	require.Equal(t, "", obj2.Meta.Set[0].Name)
	require.Equal(t, "", obj2.Meta.Set[1].Id)
	require.Equal(t, obj.Meta.Set[1].Name, obj2.Meta.Set[1].Name)
}

func TestFastRead(t *testing.T) {
	obj := SampleNewBase()
	buf := make([]byte, obj.BLength())
	e := obj.FastWriteNocopy(buf, nil)
	obj2 := base.NewBase()
	fm, err := fieldmask.NewFieldMask(obj2.GetTypeDescriptor(), "$.Addr",
		"$.LogID",
		"$.TrafficEnv.Code",
		"$.Meta.IntMap{1}.id",
		"$.Meta.IntMap{2}.name",
		"$.Meta.StrMap{\"1234\"}",
		"$.Meta.List[1]",
		"$.Meta.Set[0].id",
		"$.Meta.Set[1].name")
	if err != nil {
		t.Fatal(err)
	}
	obj2.SetFieldMask(fm)
	n, err := obj2.FastRead(buf)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, e, n)

	require.Equal(t, e, n)
	require.Equal(t, obj.Addr, obj2.Addr)
	require.Equal(t, obj.LogID, obj2.LogID)
	require.Equal(t, "", obj2.Caller)
	require.Equal(t, "", obj2.TrafficEnv.Name)
	require.Equal(t, false, obj2.TrafficEnv.Open)
	require.Equal(t, "", obj2.TrafficEnv.Env)
	require.Equal(t, obj.TrafficEnv.Code, obj2.TrafficEnv.Code)
	require.Equal(t, 2, len(obj2.Meta.IntMap))
	require.Equal(t, obj.Meta.IntMap[1].Id, obj2.Meta.IntMap[1].Id)
	require.Equal(t, "", obj2.Meta.IntMap[1].Name)
	require.Equal(t, "", obj2.Meta.IntMap[2].Id)
	require.Equal(t, obj.Meta.IntMap[2].Name, obj2.Meta.IntMap[2].Name)
	require.Equal(t, obj.Meta.StrMap["1234"].Id, obj2.Meta.StrMap["1234"].Id)
	require.Equal(t, (*base.Val)(nil), obj2.Meta.StrMap["abcd"])
	require.Equal(t, obj.Meta.List[1].Id, obj2.Meta.List[0].Id)
	require.Equal(t, obj.Meta.List[1].Name, obj2.Meta.List[0].Name)
	require.Equal(t, 1, len(obj2.Meta.List))
	require.Equal(t, 2, len(obj2.Meta.Set))
	require.Equal(t, obj.Meta.Set[0].Id, obj2.Meta.Set[0].Id)
	require.Equal(t, "", obj2.Meta.Set[0].Name)
	require.Equal(t, "", obj2.Meta.Set[1].Id)
	require.Equal(t, obj.Meta.Set[1].Name, obj2.Meta.Set[1].Name)
}

func BenchmarkFastWriteWithFieldMask(b *testing.B) {
	b.Run("old", func(b *testing.B) {
		obj := SampleOldBase()
		buf := make([]byte, obj.BLength())
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf := buf[:obj.BLength()]
			if err := obj.FastWriteNocopy(buf, nil); err == 0 {
				b.Fatal(err)
			}
		}
	})

	runtime.GC()

	b.Run("new", func(b *testing.B) {
		obj := SampleNewBase()
		buf := make([]byte, obj.BLength())
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf := buf[:obj.BLength()]
			if err := obj.FastWriteNocopy(buf, nil); err == 0 {
				b.Fatal(err)
			}
		}
	})

	runtime.GC()

	b.Run("new-mask-half", func(b *testing.B) {
		obj := SampleNewBase()
		buf := make([]byte, obj.BLength())
		fm, err := fieldmask.NewFieldMask(obj.GetTypeDescriptor(), "$.Addr", "$.LogID", "$.TrafficEnv.Code", "$.Meta.IntMap{1}", "$.Meta.StrMap{\"1234\"}", "$.Meta.List[1]", "$.Meta.Set[1]")
		if err != nil {
			b.Fatal(err)
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			obj.SetFieldMask(fm)
			buf := buf[:obj.BLength()]
			if err := obj.FastWriteNocopy(buf, nil); err == 0 {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkFastReadWithFieldMask(b *testing.B) {
	b.Run("old", func(b *testing.B) {
		obj := SampleOldBase()
		buf := make([]byte, obj.BLength())
		if err := obj.FastWriteNocopy(buf, nil); err == 0 {
			b.Fatal(err)
		}
		obj = obase.NewBase()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if _, err := obj.FastRead(buf); err != nil {
				b.Fatal(err)
			}
		}
	})

	runtime.GC()

	b.Run("new", func(b *testing.B) {
		obj := SampleNewBase()
		buf := make([]byte, obj.BLength())
		if err := obj.FastWriteNocopy(buf, nil); err == 0 {
			b.Fatal(err)
		}
		obj = base.NewBase()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if _, err := obj.FastRead(buf); err != nil {
				b.Fatal(err)
			}
		}
	})

	runtime.GC()

	b.Run("new-mask-half", func(b *testing.B) {
		obj := SampleNewBase()
		buf := make([]byte, obj.BLength())
		if err := obj.FastWriteNocopy(buf, nil); err == 0 {
			b.Fatal(err)
		}
		obj = base.NewBase()

		fm, err := fieldmask.NewFieldMask(obj.GetTypeDescriptor(), "$.Addr", "$.LogID", "$.TrafficEnv.Code", "$.Meta.IntMap{1}", "$.Meta.StrMap{\"1234\"}", "$.Meta.List[1]", "$.Meta.Set[1]")
		if err != nil {
			b.Fatal(err)
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			obj.SetFieldMask(fm)
			if _, err := obj.FastRead(buf); err != nil {
				b.Fatal(err)
			}
		}
	})
}
