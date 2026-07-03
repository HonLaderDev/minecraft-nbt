package nbt

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Dump 使用传入的编码解析序列化 NBT 数据，并返回便于阅读的解码文本。
// 类型名称采用 doc.go 中列出的名称，嵌套标签使用单个制表符缩进。
// 由于 NBT 复合标签的映射顺序不固定，TAG_Compound 每次输出的顺序可能不同。
//
// 如果传入的序列化 NBT 数据不能按指定编码解析，会返回错误，并且结果字符串始终为空。
//
// Dump returns a human readable decoded version of a serialised slice of NBT encoded using the encoding that
// is passed.
// Types are printed using the names present in the doc.go file and nested tags are indented using a single
// tab.
// Due to the nature of NBT, TAG_Compounds will not be printed in the same order. A different result is to be
// expected every time Dump is called, due to the random ordering.
//
// If the serialised NBT data passed is not parsable using the encoding that was passed, an error is returned
// and the resulting string will always be empty.
func Dump(data []byte, encoding Encoding) (string, error) {
	var m map[string]any
	if err := UnmarshalEncoding(data, &m, encoding); err != nil {
		return "", fmt.Errorf("decode NBT: %w", err)
	}
	s := &dumpState{}
	return s.encodeTagType(m) + "(" + s.encodeTagValue(m) + ")", nil
}

// dumpState 保存单次 dump 操作中的状态。每次调用 Dump 都会创建新的 dumpState。
//
// dumpState is used to keep track of values used during a single dump operations. A new one is created upon
// every call to Dump().
type dumpState struct {
	// currentIndent 指定 dump 输出标签前应写入的制表符数量。打开复合标签或列表标签时增加，
	// 关闭复合标签或列表标签时减少。
	//
	// currentIndent specifies the amount of tabs that should be present in front of tags in the dump upon
	// writing. The value is increased every time a compound or list tag is opened, and reduced every time
	// a compound or list tag is closed.
	currentIndent int
}

// indent 返回当前嵌套层级所需的缩进。打开列表或复合标签时缩进增加，关闭时减少。
//
// indent returns the indentation required for the current nesting level. It is increased every time a list
// or compound tag is opened, and reduced when it is closed.
func (s *dumpState) indent() string {
	return strings.Repeat("	", s.currentIndent)
}

// encodeTagType 将传入值的类型编码为 NBT 标签名。具体映射关系见 doc.go。
//
// encodeTagType encodes the type of the value passed to an NBT tag name. The way these are translated can be
// found in the doc.go file.
func (s *dumpState) encodeTagType(val any) string {
	if val == nil {
		return "nil"
	}
	switch val.(type) {
	case byte:
		return "TAG_Byte"
	case int16:
		return "TAG_Short"
	case int32:
		return "TAG_Int"
	case int64:
		return "TAG_Long"
	case float32:
		return "TAG_Float"
	case float64:
		return "TAG_Double"
	case string:
		return "TAG_String"
	}
	t := reflect.TypeOf(val)
	switch t.Kind() {
	case reflect.Map:
		return "TAG_Compound"
	case reflect.Slice:
		elemType := reflect.New(t.Elem()).Elem().Interface()

		v := reflect.ValueOf(val)
		if v.Len() != 0 && elemType == nil {
			elemType = v.Index(0).Elem().Interface()
		}
		return "TAG_List<" + s.encodeTagType(elemType) + ">"
	case reflect.Array:
		switch t.Elem().Kind() {
		case reflect.Uint8, reflect.Bool:
			return "TAG_ByteArray"
		case reflect.Int32:
			return "TAG_IntArray"
		case reflect.Int64:
			return "TAG_LongArray"
		}
	}
	panic("should not happen")
}

// encodeTagValue 将传入值编码为 dump 字符串中的显示形式。它会递归处理嵌套的列表和复合标签。
//
// encodeTagValue encodes a value passed to a format in which they are displayed in the dump string.
// encodeTagValue operates recursively: If lists or compounds are nested, encodeTagValue will include all
// nested tags.
func (s *dumpState) encodeTagValue(val any) string {
	// 抑制拼写检查提示。
	//
	//noinspection SpellCheckingInspection
	const hexTable = "0123456789abcdef"

	switch v := val.(type) {
	case byte:
		return "0x" + string([]byte{hexTable[v>>4], hexTable[v&0x0f]})
	case int16:
		return strconv.Itoa(int(v))
	case int32:
		return strconv.Itoa(int(v))
	case int64:
		return strconv.FormatInt(v, 10)
	case float32:
		return strconv.FormatFloat(float64(v), 'g', -1, 32)
	case float64:
		return strconv.FormatFloat(v, 'g', -1, 64)
	case string:
		return v
	}
	t := reflect.TypeOf(val)
	reflectVal := reflect.ValueOf(val)
	switch t.Kind() {
	case reflect.Map:
		b := strings.Builder{}
		b.WriteString("{\n")
		for _, k := range reflectVal.MapKeys() {
			v := reflectVal.MapIndex(k)
			actualVal := v.Interface()

			s.currentIndent++
			b.WriteString(fmt.Sprintf("%v'%v': %v(%v),\n", s.indent(), k.String(), s.encodeTagType(actualVal), s.encodeTagValue(actualVal)))
			s.currentIndent--
		}
		b.WriteString(s.indent() + "}")
		return b.String()
	case reflect.Slice:
		b := strings.Builder{}
		b.WriteString("{\n")
		for i := 0; i < reflectVal.Len(); i++ {
			v := reflectVal.Index(i)
			actualVal := v.Interface()

			s.currentIndent++
			b.WriteString(fmt.Sprintf("%v%v,\n", s.indent(), s.encodeTagValue(actualVal)))
			s.currentIndent--
		}
		b.WriteString(s.indent() + "}")
		return b.String()
	case reflect.Array:
		switch t.Elem().Kind() {
		case reflect.Uint8:
			b := strings.Builder{}
			for i := 0; i < reflectVal.Len(); i++ {
				v := reflectVal.Index(i).Uint()
				b.WriteString("0x")
				b.WriteString(string([]byte{hexTable[v>>4], hexTable[v&0x0f]}))
				if i != reflectVal.Len()-1 {
					b.WriteByte(' ')
				}
			}
			return b.String()
		case reflect.Int32, reflect.Int64:
			b := strings.Builder{}
			for i := 0; i < reflectVal.Len(); i++ {
				v := reflectVal.Index(i).Int()
				b.WriteString(strconv.FormatInt(v, 10))
				if i != reflectVal.Len()-1 {
					b.WriteByte(' ')
				}
			}
			return b.String()
		}
	}
	panic("should not happen")
}
