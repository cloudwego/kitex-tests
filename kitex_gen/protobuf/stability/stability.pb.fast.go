// Code generated by Fastpb v0.0.2. DO NOT EDIT.

package stability

import (
	fmt "fmt"
	fastpb "github.com/cloudwego/fastpb"
	instparam "github.com/cloudwego/kitex-tests/kitex_gen/protobuf/instparam"
)

var (
	_ = fmt.Errorf
	_ = fastpb.Skip
)

func (x *STRequest) FastRead(buf []byte, _type int8, number int32) (offset int, err error) {
	switch number {
	case 1:
		offset, err = x.fastReadField1(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	case 2:
		offset, err = x.fastReadField2(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	case 7:
		offset, err = x.fastReadField7(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	case 9:
		offset, err = x.fastReadField9(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	case 10:
		offset, err = x.fastReadField10(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	case 11:
		offset, err = x.fastReadField11(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	case 13:
		offset, err = x.fastReadField13(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	case 14:
		offset, err = x.fastReadField14(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	case 15:
		offset, err = x.fastReadField15(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	case 16:
		offset, err = x.fastReadField16(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	default:
		offset, err = fastpb.Skip(buf, _type, number)
		if err != nil {
			goto SkipFieldError
		}
	}
	return offset, nil
SkipFieldError:
	return offset, fmt.Errorf("%T cannot parse invalid wire-format data, error: %s", x, err)
ReadFieldError:
	return offset, fmt.Errorf("%T read field %d '%s' error: %s", x, number, fieldIDToName_STRequest[number], err)
}

func (x *STRequest) fastReadField1(buf []byte, _type int8) (offset int, err error) {
	x.Name, offset, err = fastpb.ReadString(buf, _type)
	return offset, err
}

func (x *STRequest) fastReadField2(buf []byte, _type int8) (offset int, err error) {
	x.On, offset, err = fastpb.ReadBool(buf, _type)
	return offset, err
}

func (x *STRequest) fastReadField7(buf []byte, _type int8) (offset int, err error) {
	x.D, offset, err = fastpb.ReadDouble(buf, _type)
	return offset, err
}

func (x *STRequest) fastReadField9(buf []byte, _type int8) (offset int, err error) {
	x.Str, offset, err = fastpb.ReadString(buf, _type)
	return offset, err
}

func (x *STRequest) fastReadField10(buf []byte, _type int8) (offset int, err error) {
	if x.StringMap == nil {
		x.StringMap = make(map[string]string)
	}
	var key string
	var value string
	offset, err = fastpb.ReadMapEntry(buf, _type,
		func(buf []byte, _type int8) (offset int, err error) {
			key, offset, err = fastpb.ReadString(buf, _type)
			return offset, err
		},
		func(buf []byte, _type int8) (offset int, err error) {
			value, offset, err = fastpb.ReadString(buf, _type)
			return offset, err
		})
	if err != nil {
		return offset, err
	}
	x.StringMap[key] = value
	return offset, nil
}

func (x *STRequest) fastReadField11(buf []byte, _type int8) (offset int, err error) {
	var v string
	v, offset, err = fastpb.ReadString(buf, _type)
	if err != nil {
		return offset, err
	}
	x.StringList = append(x.StringList, v)
	return offset, err
}

func (x *STRequest) fastReadField13(buf []byte, _type int8) (offset int, err error) {
	var v int32
	v, offset, err = fastpb.ReadInt32(buf, _type)
	if err != nil {
		return offset, err
	}
	x.E = TestEnum(v)
	return offset, nil
}

func (x *STRequest) fastReadField14(buf []byte, _type int8) (offset int, err error) {
	x.FlagMsg, offset, err = fastpb.ReadString(buf, _type)
	return offset, err
}

func (x *STRequest) fastReadField15(buf []byte, _type int8) (offset int, err error) {
	x.Framework, offset, err = fastpb.ReadString(buf, _type)
	return offset, err
}

func (x *STRequest) fastReadField16(buf []byte, _type int8) (offset int, err error) {
	x.MockCost, offset, err = fastpb.ReadString(buf, _type)
	return offset, err
}

func (x *STResponse) FastRead(buf []byte, _type int8, number int32) (offset int, err error) {
	switch number {
	case 1:
		offset, err = x.fastReadField1(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	case 2:
		offset, err = x.fastReadField2(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	case 3:
		offset, err = x.fastReadField3(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	default:
		offset, err = fastpb.Skip(buf, _type, number)
		if err != nil {
			goto SkipFieldError
		}
	}
	return offset, nil
SkipFieldError:
	return offset, fmt.Errorf("%T cannot parse invalid wire-format data, error: %s", x, err)
ReadFieldError:
	return offset, fmt.Errorf("%T read field %d '%s' error: %s", x, number, fieldIDToName_STResponse[number], err)
}

func (x *STResponse) fastReadField1(buf []byte, _type int8) (offset int, err error) {
	x.Str, offset, err = fastpb.ReadString(buf, _type)
	return offset, err
}

func (x *STResponse) fastReadField2(buf []byte, _type int8) (offset int, err error) {
	if x.Mp == nil {
		x.Mp = make(map[string]string)
	}
	var key string
	var value string
	offset, err = fastpb.ReadMapEntry(buf, _type,
		func(buf []byte, _type int8) (offset int, err error) {
			key, offset, err = fastpb.ReadString(buf, _type)
			return offset, err
		},
		func(buf []byte, _type int8) (offset int, err error) {
			value, offset, err = fastpb.ReadString(buf, _type)
			return offset, err
		})
	if err != nil {
		return offset, err
	}
	x.Mp[key] = value
	return offset, nil
}

func (x *STResponse) fastReadField3(buf []byte, _type int8) (offset int, err error) {
	x.FlagMsg, offset, err = fastpb.ReadString(buf, _type)
	return offset, err
}

func (x *STRequest) FastWrite(buf []byte) (offset int) {
	if x == nil {
		return offset
	}
	offset += x.fastWriteField1(buf[offset:])
	offset += x.fastWriteField2(buf[offset:])
	offset += x.fastWriteField7(buf[offset:])
	offset += x.fastWriteField9(buf[offset:])
	offset += x.fastWriteField10(buf[offset:])
	offset += x.fastWriteField11(buf[offset:])
	offset += x.fastWriteField13(buf[offset:])
	offset += x.fastWriteField14(buf[offset:])
	offset += x.fastWriteField15(buf[offset:])
	offset += x.fastWriteField16(buf[offset:])
	return offset
}

func (x *STRequest) fastWriteField1(buf []byte) (offset int) {
	if x.Name == "" {
		return offset
	}
	offset += fastpb.WriteString(buf[offset:], 1, x.GetName())
	return offset
}

func (x *STRequest) fastWriteField2(buf []byte) (offset int) {
	if !x.On {
		return offset
	}
	offset += fastpb.WriteBool(buf[offset:], 2, x.GetOn())
	return offset
}

func (x *STRequest) fastWriteField7(buf []byte) (offset int) {
	if x.D == 0 {
		return offset
	}
	offset += fastpb.WriteDouble(buf[offset:], 7, x.GetD())
	return offset
}

func (x *STRequest) fastWriteField9(buf []byte) (offset int) {
	if x.Str == "" {
		return offset
	}
	offset += fastpb.WriteString(buf[offset:], 9, x.GetStr())
	return offset
}

func (x *STRequest) fastWriteField10(buf []byte) (offset int) {
	if x.StringMap == nil {
		return offset
	}
	for k, v := range x.GetStringMap() {
		offset += fastpb.WriteMapEntry(buf[offset:], 10,
			func(buf []byte, numTagOrKey, numIdxOrVal int32) int {
				offset := 0
				offset += fastpb.WriteString(buf[offset:], numTagOrKey, k)
				offset += fastpb.WriteString(buf[offset:], numIdxOrVal, v)
				return offset
			})
	}
	return offset
}

func (x *STRequest) fastWriteField11(buf []byte) (offset int) {
	if len(x.StringList) == 0 {
		return offset
	}
	for i := range x.GetStringList() {
		offset += fastpb.WriteString(buf[offset:], 11, x.GetStringList()[i])
	}
	return offset
}

func (x *STRequest) fastWriteField13(buf []byte) (offset int) {
	if x.E == 0 {
		return offset
	}
	offset += fastpb.WriteInt32(buf[offset:], 13, int32(x.GetE()))
	return offset
}

func (x *STRequest) fastWriteField14(buf []byte) (offset int) {
	if x.FlagMsg == "" {
		return offset
	}
	offset += fastpb.WriteString(buf[offset:], 14, x.GetFlagMsg())
	return offset
}

func (x *STRequest) fastWriteField15(buf []byte) (offset int) {
	if x.Framework == "" {
		return offset
	}
	offset += fastpb.WriteString(buf[offset:], 15, x.GetFramework())
	return offset
}

func (x *STRequest) fastWriteField16(buf []byte) (offset int) {
	if x.MockCost == "" {
		return offset
	}
	offset += fastpb.WriteString(buf[offset:], 16, x.GetMockCost())
	return offset
}

func (x *STResponse) FastWrite(buf []byte) (offset int) {
	if x == nil {
		return offset
	}
	offset += x.fastWriteField1(buf[offset:])
	offset += x.fastWriteField2(buf[offset:])
	offset += x.fastWriteField3(buf[offset:])
	return offset
}

func (x *STResponse) fastWriteField1(buf []byte) (offset int) {
	if x.Str == "" {
		return offset
	}
	offset += fastpb.WriteString(buf[offset:], 1, x.GetStr())
	return offset
}

func (x *STResponse) fastWriteField2(buf []byte) (offset int) {
	if x.Mp == nil {
		return offset
	}
	for k, v := range x.GetMp() {
		offset += fastpb.WriteMapEntry(buf[offset:], 2,
			func(buf []byte, numTagOrKey, numIdxOrVal int32) int {
				offset := 0
				offset += fastpb.WriteString(buf[offset:], numTagOrKey, k)
				offset += fastpb.WriteString(buf[offset:], numIdxOrVal, v)
				return offset
			})
	}
	return offset
}

func (x *STResponse) fastWriteField3(buf []byte) (offset int) {
	if x.FlagMsg == "" {
		return offset
	}
	offset += fastpb.WriteString(buf[offset:], 3, x.GetFlagMsg())
	return offset
}

func (x *STRequest) Size() (n int) {
	if x == nil {
		return n
	}
	n += x.sizeField1()
	n += x.sizeField2()
	n += x.sizeField7()
	n += x.sizeField9()
	n += x.sizeField10()
	n += x.sizeField11()
	n += x.sizeField13()
	n += x.sizeField14()
	n += x.sizeField15()
	n += x.sizeField16()
	return n
}

func (x *STRequest) sizeField1() (n int) {
	if x.Name == "" {
		return n
	}
	n += fastpb.SizeString(1, x.GetName())
	return n
}

func (x *STRequest) sizeField2() (n int) {
	if !x.On {
		return n
	}
	n += fastpb.SizeBool(2, x.GetOn())
	return n
}

func (x *STRequest) sizeField7() (n int) {
	if x.D == 0 {
		return n
	}
	n += fastpb.SizeDouble(7, x.GetD())
	return n
}

func (x *STRequest) sizeField9() (n int) {
	if x.Str == "" {
		return n
	}
	n += fastpb.SizeString(9, x.GetStr())
	return n
}

func (x *STRequest) sizeField10() (n int) {
	if x.StringMap == nil {
		return n
	}
	for k, v := range x.GetStringMap() {
		n += fastpb.SizeMapEntry(10,
			func(numTagOrKey, numIdxOrVal int32) int {
				n := 0
				n += fastpb.SizeString(numTagOrKey, k)
				n += fastpb.SizeString(numIdxOrVal, v)
				return n
			})
	}
	return n
}

func (x *STRequest) sizeField11() (n int) {
	if len(x.StringList) == 0 {
		return n
	}
	for i := range x.GetStringList() {
		n += fastpb.SizeString(11, x.GetStringList()[i])
	}
	return n
}

func (x *STRequest) sizeField13() (n int) {
	if x.E == 0 {
		return n
	}
	n += fastpb.SizeInt32(13, int32(x.GetE()))
	return n
}

func (x *STRequest) sizeField14() (n int) {
	if x.FlagMsg == "" {
		return n
	}
	n += fastpb.SizeString(14, x.GetFlagMsg())
	return n
}

func (x *STRequest) sizeField15() (n int) {
	if x.Framework == "" {
		return n
	}
	n += fastpb.SizeString(15, x.GetFramework())
	return n
}

func (x *STRequest) sizeField16() (n int) {
	if x.MockCost == "" {
		return n
	}
	n += fastpb.SizeString(16, x.GetMockCost())
	return n
}

func (x *STResponse) Size() (n int) {
	if x == nil {
		return n
	}
	n += x.sizeField1()
	n += x.sizeField2()
	n += x.sizeField3()
	return n
}

func (x *STResponse) sizeField1() (n int) {
	if x.Str == "" {
		return n
	}
	n += fastpb.SizeString(1, x.GetStr())
	return n
}

func (x *STResponse) sizeField2() (n int) {
	if x.Mp == nil {
		return n
	}
	for k, v := range x.GetMp() {
		n += fastpb.SizeMapEntry(2,
			func(numTagOrKey, numIdxOrVal int32) int {
				n := 0
				n += fastpb.SizeString(numTagOrKey, k)
				n += fastpb.SizeString(numIdxOrVal, v)
				return n
			})
	}
	return n
}

func (x *STResponse) sizeField3() (n int) {
	if x.FlagMsg == "" {
		return n
	}
	n += fastpb.SizeString(3, x.GetFlagMsg())
	return n
}

var fieldIDToName_STRequest = map[int32]string{
	1:  "Name",
	2:  "On",
	7:  "D",
	9:  "Str",
	10: "StringMap",
	11: "StringList",
	13: "E",
	14: "FlagMsg",
	15: "Framework",
	16: "MockCost",
}

var fieldIDToName_STResponse = map[int32]string{
	1: "Str",
	2: "Mp",
	3: "FlagMsg",
}

var _ = instparam.File_instparam_proto
