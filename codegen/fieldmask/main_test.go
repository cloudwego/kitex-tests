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
	"fmt"
	"runtime"
	hbase "test/halfway_gen/base"
	nbase "test/kitex_gen/base"
	obase "test/old_gen/base"
	"testing"

	"github.com/cloudwego/thriftgo/fieldmask"
	"github.com/stretchr/testify/require"
)

func SampleNewBase() *nbase.Base {
	obj := nbase.NewBase()
	obj.Addr = "abcd"
	obj.Caller = "abcd"
	obj.LogID = "abcd"
	obj.Meta = nbase.NewMetaInfo()
	obj.Meta.StrMap = map[string]*nbase.Val{
		"abcd": nbase.NewVal(),
		"1234": nbase.NewVal(),
	}
	obj.Meta.IntMap = map[int64]*nbase.Val{
		1: nbase.NewVal(),
		2: nbase.NewVal(),
	}
	v0 := nbase.NewVal()
	v0.Id = "a"
	v0.Name = "a"
	v1 := nbase.NewVal()
	v1.Id = "b"
	v1.Name = "b"
	obj.Meta.List = []*nbase.Val{v0, v1}
	// v0 = nbase.NewVal()
	// v0.ID = "a"
	// v0.Name = "a"
	// v1 = nbase.NewVal()
	// v1.ID = "b"
	// v1.Name = "b"
	obj.Meta.Set = []*nbase.Val{v0, v1}
	// obj.Extra = nbase.NewExtraInfo()
	obj.TrafficEnv = nbase.NewTrafficEnv()
	obj.TrafficEnv.Code = 1
	obj.TrafficEnv.Env = "abcd"
	obj.TrafficEnv.Name = "abcd"
	obj.TrafficEnv.Open = true
	obj.Meta.Base = nbase.NewBase()
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
	// obj.Extra = obase.NewExtraInfo()
	obj.TrafficEnv = obase.NewTrafficEnv()
	obj.TrafficEnv.Code = 1
	obj.TrafficEnv.Env = "abcd"
	obj.TrafficEnv.Name = "abcd"
	obj.TrafficEnv.Open = true
	obj.Meta.Base = obase.NewBase()
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
	obj.Set_FieldMask(fm)
	out := make([]byte, obj.BLength())
	// out := make([]byte, 24000000)
	e := obj.FastWriteNocopy(out, nil)
	var obj2 = nbase.NewBase()
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
	require.Equal(t, (*nbase.Val)(nil), obj2.Meta.StrMap["abcd"])
	require.Equal(t, 1, len(obj2.Meta.List))
	require.Equal(t, obj.Meta.List[1].Id, obj2.Meta.List[0].Id)
	require.Equal(t, obj.Meta.List[1].Name, obj2.Meta.List[0].Name)
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
	obj2 := nbase.NewBase()
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
	obj2.Set_FieldMask(fm)
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
	require.Equal(t, (*nbase.Val)(nil), obj2.Meta.StrMap["abcd"])
	require.Equal(t, obj.Meta.List[1].Id, obj2.Meta.List[0].Id)
	require.Equal(t, obj.Meta.List[1].Name, obj2.Meta.List[0].Name)
	require.Equal(t, 1, len(obj2.Meta.List))
	require.Equal(t, 2, len(obj2.Meta.Set))
	require.Equal(t, obj.Meta.Set[0].Id, obj2.Meta.Set[0].Id)
	require.Equal(t, "", obj2.Meta.Set[0].Name)
	require.Equal(t, "", obj2.Meta.Set[1].Id)
	require.Equal(t, obj.Meta.Set[1].Name, obj2.Meta.Set[1].Name)
}

func TestMaskRequired(t *testing.T) {
	fm, err := fieldmask.NewFieldMask(nbase.NewBaseResp().GetTypeDescriptor(), "$.F1", "$.F8")
	if err != nil {
		t.Fatal(err)
	}
	j, err := fm.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	println(string(j))
	nf, ex := fm.Field(111)
	if !ex {
		t.Fatal(nf)
	}

	t.Run("read", func(t *testing.T) {
		obj := nbase.NewBaseResp()
		obj.F1 = map[nbase.Str]nbase.Str{"a": "b"}
		obj.F8 = map[float64][]nbase.Str{1.0: []nbase.Str{"a"}}
		buf := make([]byte, obj.BLength())
		if err := obj.FastWriteNocopy(buf, nil); err != len(buf) {
			t.Fatal(err)
		}
		obj2 := nbase.NewBaseResp()
		obj2.Set_FieldMask(fm)
		if _, err := obj2.FastRead(buf); err != nil {
			t.Fatal(err)
		}
		require.Equal(t, obj.F1, obj2.F1)
		require.Equal(t, obj.F8, obj2.F8)
	})

	t.Run("write", func(t *testing.T) {
		obj := nbase.NewBaseResp()
		obj.F1 = map[nbase.Str]nbase.Str{"a": "b"}
		obj.F8 = map[float64][]nbase.Str{1.0: []nbase.Str{"a"}}
		obj.Set_FieldMask(fm)
		buf := make([]byte, obj.BLength())
		if err := obj.FastWriteNocopy(buf, nil); err != len(buf) {
			t.Fatal(err)
		}
		obj2 := nbase.NewBaseResp()
		if _, err := obj2.FastRead(buf); err != nil {
			t.Fatal(err)
		}
		fmt.Printf("%#v\n", obj2)
	})
}

func TestSetMaskHalfway(t *testing.T) {
	obj := hbase.NewBase()
	obj.Extra = hbase.NewExtraInfo()
	obj.Extra.F1 = map[string]string{"a": "b"}
	obj.Extra.F8 = map[int64][]*hbase.Key{1: []*hbase.Key{hbase.NewKey()}}

	fm, err := fieldmask.NewFieldMask(obj.Extra.GetTypeDescriptor(), "$.F1")
	if err != nil {
		t.Fatal(err)
	}
	obj.Extra.Set_FieldMask(fm)
	buf := make([]byte, obj.BLength())
	if err := obj.FastWriteNocopy(buf, nil); err != len(buf) {
		t.Fatal(err)
	}
	obj2 := hbase.NewBase()
	if _, err := obj2.FastRead(buf); err != nil {
		t.Fatal(err)
	}
	require.Equal(t, obj.Extra.F1, obj2.Extra.F1)
	require.Equal(t, map[int64][]*hbase.Key(nil), obj2.Extra.F8)

	fm, err = fieldmask.NewFieldMask(obj.Extra.GetTypeDescriptor(), "$.F8")
	if err != nil {
		t.Fatal(err)
	}
	obj.Extra.Set_FieldMask(fm)
	buf = make([]byte, obj.BLength())
	if err := obj.FastWriteNocopy(buf, nil); err != len(buf) {
		t.Fatal(err)
	}
	obj2 = hbase.NewBase()
	if _, err := obj2.FastRead(buf); err != nil {
		t.Fatal(err)
	}
	require.Equal(t, map[string]string(nil), obj2.Extra.F1)
	require.Equal(t, obj.Extra.F8, obj2.Extra.F8)
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
		fm, err := fieldmask.NewFieldMask(obj.GetTypeDescriptor(), "$.Addr", "$.LogID", "$.TrafficEnv.Code", "$.Meta.IntMap", "$.Meta.List")
		if err != nil {
			b.Fatal(err)
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			obj.Set_FieldMask(fm)
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
		obj = nbase.NewBase()
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
		obj = nbase.NewBase()

		fm, err := fieldmask.NewFieldMask(obj.GetTypeDescriptor(), "$.Addr", "$.LogID", "$.TrafficEnv.Code", "$.Meta.IntMap", "$.Meta.List")
		if err != nil {
			b.Fatal(err)
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			obj.Set_FieldMask(fm)
			if _, err := obj.FastRead(buf); err != nil {
				b.Fatal(err)
			}
		}
	})
}
