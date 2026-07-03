package nbt

import (
	"bytes"
	"encoding/binary"
	"errors"
	"math"
	"reflect"
	"strings"
	"testing"
)

func TestNBTReaderChainReadsStructValues(t *testing.T) {
	data := rootCompound(
		namedTag(tagByte, "Byte", []byte{0x07}),
		namedTag(tagInt16, "Short", i16(198)),
		namedTag(tagInt32, "Int", i32(123456)),
		namedTag(tagInt64, "Long", i64(-7)),
		namedTag(tagFloat32, "Float", f32(1.5)),
		namedTag(tagFloat64, "Double", f64(2.25)),
		namedTag(tagString, "String", nbtString("ok")),
		namedTag(tagByteArray, "Bytes", append(i32(3), 1, 2, 3)),
		namedTag(tagInt32Array, "Ints", append(append(i32(2), i32(10)...), i32(20)...)),
		namedTag(tagInt64Array, "Longs", append(i32(1), i64(30)...)),
		namedTag(tagSlice, "List", append(append([]byte{byte(tagInt16)}, i32(2)...), append(i16(5), i16(6)...)...)),
	)

	var byteValue byte
	var shortValue int16
	var intValue int32
	var longValue int64
	var floatValue float32
	var doubleValue float64
	var stringValue string
	var byteArray []byte
	var int32Array []int32
	var int64Array []int64
	var listValues []int16

	reader := NewNBTReader(NewTagReader(BigEndian), bytes.NewReader(data))
	err := reader.ReadTag().
		ExpectMustType(TagStruct).
		ReadStruct().
		ExpectType(TagByte).
		ExpectName("Byte").
		Byte(func(value byte) error {
			byteValue = value
			return nil
		}).
		Return().
		Return().
		ExpectType(TagInt16).
		ExpectName("Short").
		Int16(func(value int16) error {
			shortValue = value
			return nil
		}).
		Return().
		Return().
		ExpectType(TagInt32).
		ExpectName("Int").
		Int32(func(value int32) error {
			intValue = value
			return nil
		}).
		Return().
		Return().
		ExpectType(TagInt64).
		ExpectName("Long").
		Int64(func(value int64) error {
			longValue = value
			return nil
		}).
		Return().
		Return().
		ExpectType(TagFloat32).
		ExpectName("Float").
		Float32(func(value float32) error {
			floatValue = value
			return nil
		}).
		Return().
		Return().
		ExpectType(TagFloat64).
		ExpectName("Double").
		Float64(func(value float64) error {
			doubleValue = value
			return nil
		}).
		Return().
		Return().
		ExpectType(TagString).
		ExpectName("String").
		String(func(value string) error {
			stringValue = value
			return nil
		}).
		Return().
		Return().
		ExpectType(TagByteArray).
		ExpectName("Bytes").
		ByteArray(func(value []byte) error {
			byteArray = value
			return nil
		}).
		Return().
		Return().
		ExpectType(TagInt32Array).
		ExpectName("Ints").
		Int32Array(func(value []int32) error {
			int32Array = value
			return nil
		}).
		Return().
		Return().
		ExpectType(TagInt64Array).
		ExpectName("Longs").
		Int64Array(func(value []int64) error {
			int64Array = value
			return nil
		}).
		Return().
		Return().
		ExpectType(TagSlice).
		ExpectName("List").
		ReadList().
		ExpectType(TagInt16).
		Int16(func(value int16) error {
			listValues = append(listValues, value)
			return nil
		}).
		Return().
		Return().
		Return().
		Return().
		Return().
		Return().
		Return()
	if err != nil {
		t.Fatal(err)
	}

	if byteValue != 7 || shortValue != 198 || intValue != 123456 || longValue != -7 {
		t.Fatalf("integer values = (%d, %d, %d, %d)", byteValue, shortValue, intValue, longValue)
	}
	if floatValue != 1.5 || doubleValue != 2.25 || stringValue != "ok" {
		t.Fatalf("scalar values = (%v, %v, %q)", floatValue, doubleValue, stringValue)
	}
	if !bytes.Equal(byteArray, []byte{1, 2, 3}) {
		t.Fatalf("byteArray = %v", byteArray)
	}
	if !reflect.DeepEqual(int32Array, []int32{10, 20}) {
		t.Fatalf("int32Array = %v", int32Array)
	}
	if !reflect.DeepEqual(int64Array, []int64{30}) {
		t.Fatalf("int64Array = %v", int64Array)
	}
	if !reflect.DeepEqual(listValues, []int16{5, 6}) {
		t.Fatalf("listValues = %v", listValues)
	}
}

func TestNBTReaderChainSoftMatchDoesNotConsumeTag(t *testing.T) {
	data := rootCompound(namedTag(tagInt16, "Width", i16(198)))

	var width int16
	reader := NewNBTReader(NewTagReader(BigEndian), bytes.NewReader(data))
	err := reader.ReadTag().
		ExpectMustType(TagStruct).
		ReadStruct().
		ExpectName("Missing").
		Int16(func(value int16) error {
			t.Fatalf("soft-mismatched branch consumed value %d", value)
			return nil
		}).
		Return().
		ExpectName("Width").
		Int16(func(value int16) error {
			width = value
			return nil
		}).
		Return().
		Return().
		Return().
		Return()
	if err != nil {
		t.Fatal(err)
	}
	if width != 198 {
		t.Fatalf("width = %d, want 198", width)
	}
}

func TestNBTReaderChainWhenMatchesReaderState(t *testing.T) {
	data := rootCompound(
		namedTag(tagInt16, "A1", i16(1)),
		namedTag(tagInt16, "A2", i16(2)),
		namedTag(tagInt16, "B1", i16(3)),
	)

	var names []string
	var values []int16
	reader := NewNBTReader(NewTagReader(BigEndian), bytes.NewReader(data))
	err := reader.ReadTag().
		ExpectMustType(TagStruct).
		ReadStruct().
		When(func(reader *NBTReader) bool {
			return reader.TagType() == TagInt16 && strings.HasPrefix(reader.TagName(), "A")
		}).
		Int16(func(value int16) error {
			names = append(names, reader.TagName())
			values = append(values, value)
			return nil
		}).
		Return().
		Return().
		Return().
		Return()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(names, []string{"A1", "A2"}) {
		t.Fatalf("names = %v", names)
	}
	if !reflect.DeepEqual(values, []int16{1, 2}) {
		t.Fatalf("values = %v", values)
	}
}

func TestNBTReaderChainMustNameMismatchReturnsTagStartOffset(t *testing.T) {
	data := rootCompound(namedTag(tagInt16, "Width", i16(198)))

	reader := NewNBTReader(NewTagReader(BigEndian), bytes.NewReader(data))
	err := reader.ReadTag().
		ExpectMustType(TagStruct).
		ReadStruct().
		ExpectMustName("Height").
		Return().
		Return().
		Return().
		Return()
	if err == nil {
		t.Fatal("expected error")
	}

	var nameErr ExpectTagNameError
	if !errors.As(err, &nameErr) {
		t.Fatalf("error = %T %[1]v, want ExpectTagNameError", err)
	}
	if nameErr.Off != 3 {
		t.Fatalf("offset = %d, want 3", nameErr.Off)
	}
}

func TestNBTReaderChainMustTypeMismatchReturnsTagStartOffset(t *testing.T) {
	data := rootCompound()

	reader := NewNBTReader(NewTagReader(BigEndian), bytes.NewReader(data))
	err := reader.ReadTag().
		ExpectMustType(TagInt16).
		Return().
		Return()
	if err == nil {
		t.Fatal("expected error")
	}

	var typeErr ExpectTagTypeError
	if !errors.As(err, &typeErr) {
		t.Fatalf("error = %T %[1]v, want ExpectTagTypeError", err)
	}
	if typeErr.Off != 0 {
		t.Fatalf("offset = %d, want 0", typeErr.Off)
	}
}

func TestNBTReaderChainStructSkipsUnhandledTags(t *testing.T) {
	data := rootCompound(
		namedTag(tagString, "Ignored", nbtString("skip me")),
		namedTag(tagInt16, "Width", i16(198)),
	)

	var width int16
	reader := NewNBTReader(NewTagReader(BigEndian), bytes.NewReader(data))
	err := reader.ReadTag().
		ExpectMustType(TagStruct).
		ReadStruct().
		ExpectName("Width").
		Int16(func(value int16) error {
			width = value
			return nil
		}).
		Return().
		Return().
		Return().
		Return()
	if err != nil {
		t.Fatal(err)
	}
	if width != 198 {
		t.Fatalf("width = %d, want 198", width)
	}
}

func TestNBTReaderChainReadsCompleteStructAndListValues(t *testing.T) {
	data := rootCompound(
		namedTag(tagStruct, "Nested", append(namedTag(tagInt16, "Value", i16(9)), byte(tagEnd))),
		namedTag(tagSlice, "Values", append(append([]byte{byte(tagInt16)}, i32(2)...), append(i16(3), i16(4)...)...)),
	)

	type nestedCompound struct {
		Value int16
	}

	var nested nestedCompound
	var values []int16
	reader := NewNBTReader(NewTagReader(BigEndian), bytes.NewReader(data))
	err := reader.ReadTag().
		ExpectMustType(TagStruct).
		ReadStruct().
		ExpectName("Nested").
		Struct(func(value nestedCompound) error {
			nested = value
			return nil
		}).
		Return().
		ExpectName("Values").
		List(func(value []int16) error {
			values = value
			return nil
		}).
		Return().
		Return().
		Return().
		Return()
	if err != nil {
		t.Fatal(err)
	}
	if nested.Value != 9 {
		t.Fatalf("nested = %#v", nested)
	}
	if !reflect.DeepEqual(values, []int16{3, 4}) {
		t.Fatalf("values = %#v", values)
	}
}

func TestNBTReaderChainExpectIndexMatchesReadListElement(t *testing.T) {
	data := rootCompound(
		namedTag(tagSlice, "Values", append(append([]byte{byte(tagInt16)}, i32(3)...), append(append(i16(3), i16(4)...), i16(5)...)...)),
	)

	var value int16
	var missingRan bool
	reader := NewNBTReader(NewTagReader(BigEndian), bytes.NewReader(data))
	err := reader.ReadTag().
		ExpectMustType(TagStruct).
		ReadStruct().
		ExpectName("Values").
		ReadList().
		ExpectIndex(9).
		Do(func() error {
			missingRan = true
			return nil
		}).
		Return().
		ExpectIndex(1).
		Int16(func(v int16) error {
			value = v
			return nil
		}).
		Return().
		Return().
		Return().
		Return().
		Return().
		Return()
	if err != nil {
		t.Fatal(err)
	}
	if missingRan {
		t.Fatal("unexpected index branch ran")
	}
	if value != 4 {
		t.Fatalf("value = %d, want 4", value)
	}
}

func TestNBTReaderListLenAndListIndex(t *testing.T) {
	data := rootCompound(
		namedTag(tagSlice, "Values", append(append([]byte{byte(tagInt16)}, i32(3)...), append(append(i16(3), i16(4)...), i16(5)...)...)),
	)

	var lengths []int
	var indices []int
	reader := NewNBTReader(NewTagReader(BigEndian), bytes.NewReader(data))
	err := reader.ReadTag().
		ExpectMustType(TagStruct).
		ReadStruct().
		ExpectName("Values").
		ReadList().
		ExpectType(TagInt16).
		Int16(func(value int16) error {
			lengths = append(lengths, reader.ListLen())
			indices = append(indices, reader.ListIndex())
			return nil
		}).
		Return().
		Return().
		Return().
		Return().
		Return().
		Return()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(lengths, []int{3, 3, 3}) {
		t.Fatalf("lengths = %v", lengths)
	}
	if !reflect.DeepEqual(indices, []int{0, 1, 2}) {
		t.Fatalf("indices = %v", indices)
	}
	if reader.ListLen() != -1 || reader.ListIndex() != -1 {
		t.Fatalf("list state after read = (%d, %d), want (-1, -1)", reader.ListLen(), reader.ListIndex())
	}
}

func TestNBTReaderChainDoRunsOnlyMatchedBranch(t *testing.T) {
	data := rootCompound(namedTag(tagInt16, "Width", i16(198)))

	var missingRan bool
	var widthRan bool
	reader := NewNBTReader(NewTagReader(BigEndian), bytes.NewReader(data))
	err := reader.ReadTag().
		ExpectMustType(TagStruct).
		ReadStruct().
		ExpectName("Missing").
		Do(func() error {
			missingRan = true
			return nil
		}).
		Return().
		ExpectName("Width").
		Do(func() error {
			widthRan = true
			return nil
		}).
		Return().
		Return().
		Return().
		Return()
	if err != nil {
		t.Fatal(err)
	}
	if missingRan {
		t.Fatal("soft-mismatched Do branch ran")
	}
	if !widthRan {
		t.Fatal("matched Do branch did not run")
	}
}

func TestNBTReaderChainDoErrorStopsReading(t *testing.T) {
	doErr := errors.New("do failed")
	data := rootCompound(
		namedTag(tagInt16, "Width", i16(198)),
		namedTag(tagInt16, "Height", i16(108)),
	)

	var heightRan bool
	reader := NewNBTReader(NewTagReader(BigEndian), bytes.NewReader(data))
	err := reader.ReadTag().
		ExpectMustType(TagStruct).
		ReadStruct().
		ExpectName("Width").
		Do(func() error {
			return doErr
		}).
		Return().
		ExpectName("Height").
		Do(func() error {
			heightRan = true
			return nil
		}).
		Return().
		Return().
		Return().
		Return()
	if !errors.Is(err, doErr) {
		t.Fatalf("error = %v, want Do error", err)
	}
	if heightRan {
		t.Fatal("height branch ran after Do error")
	}
}

func TestNBTReaderChainCallbackErrorStopsReading(t *testing.T) {
	callbackErr := errors.New("stop")
	data := rootCompound(
		namedTag(tagInt16, "Width", i16(198)),
		namedTag(tagInt16, "Height", i16(108)),
	)

	var heightRead bool
	reader := NewNBTReader(NewTagReader(BigEndian), bytes.NewReader(data))
	err := reader.ReadTag().
		ExpectMustType(TagStruct).
		ReadStruct().
		ExpectName("Width").
		Int16(func(value int16) error {
			return callbackErr
		}).
		Return().
		ExpectName("Height").
		Int16(func(value int16) error {
			heightRead = true
			return nil
		}).
		Return().
		Return().
		Return().
		Return().
		Return()
	if !errors.Is(err, callbackErr) {
		t.Fatalf("error = %v, want callback error", err)
	}
	if heightRead {
		t.Fatal("height branch ran after callback error")
	}
}

func rootCompound(tags ...[]byte) []byte {
	data := append([]byte{byte(tagStruct)}, nbtString("")...)
	for _, tag := range tags {
		data = append(data, tag...)
	}
	return append(data, byte(tagEnd))
}

func namedTag(tagType tagType, name string, payload []byte) []byte {
	data := []byte{byte(tagType)}
	data = append(data, nbtString(name)...)
	return append(data, payload...)
}

func nbtString(value string) []byte {
	data := i16(int16(len(value)))
	return append(data, []byte(value)...)
}

func i16(value int16) []byte {
	data := make([]byte, 2)
	binary.BigEndian.PutUint16(data, uint16(value))
	return data
}

func i32(value int32) []byte {
	data := make([]byte, 4)
	binary.BigEndian.PutUint32(data, uint32(value))
	return data
}

func i64(value int64) []byte {
	data := make([]byte, 8)
	binary.BigEndian.PutUint64(data, uint64(value))
	return data
}

func f32(value float32) []byte {
	return i32(int32(math.Float32bits(value)))
}

func f64(value float64) []byte {
	return i64(int64(math.Float64bits(value)))
}
