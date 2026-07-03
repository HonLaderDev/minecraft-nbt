package nbt

import (
	"errors"
	"fmt"
	"reflect"
)

// InvalidTypeError 在读取到的标签类型与同名结构体字段类型不匹配时返回。
//
// InvalidTypeError is returned when the type of a tag read is not equal to the struct field with the name
// of that tag.
type InvalidTypeError struct {
	Off       int64
	Field     string
	TagType   tagType
	FieldType reflect.Type
}

// Error 返回 InvalidTypeError 的错误文本。
//
// Error returns the InvalidTypeError message.
func (err InvalidTypeError) Error() string {
	return fmt.Sprintf("nbt: invalid type for tag '%v' at offset %v: cannot unmarshal %v into %v", err.Field, err.Off, err.TagType, err.FieldType)
}

// UnknownTagError 在读取到未知标签类型时返回，表示该类型没有在 tag.go 中定义。
//
// UnknownTagError is returned when the type of tag read is not known, meaning it is not found in the tag.go
// file.
type UnknownTagError struct {
	Off     int64
	Op      string
	TagType tagType
}

// Error 返回 UnknownTagError 的错误文本。
//
// Error returns the UnknownTagError message.
func (err UnknownTagError) Error() string {
	return fmt.Sprintf("nbt: unknown tag '%v' at offset %v during op '%v'", byte(err.TagType), err.Off, err.Op)
}

// UnexpectedTagError 在遇到当前上下文不允许的标签类型时返回。
//
// UnexpectedTagError is returned when a tag type encountered was not expected, and thus valid in its context.
type UnexpectedTagError struct {
	Off     int64
	TagType tagType
}

// Error 返回 UnexpectedTagError 的错误文本。
//
// Error returns the UnexpectedTagError message.
func (err UnexpectedTagError) Error() string {
	return fmt.Sprintf("nbt: unexpected tag %v at offset %v: tag is not valid in its context", err.TagType, err.Off)
}

// NonPointerTypeError 在传给 Decoder.Decode 或 Unmarshal 的值不是指针时返回。
//
// NonPointerTypeError is returned when the type of value passed in Decoder.Decode or Unmarshal is not a
// pointer.
type NonPointerTypeError struct {
	ActualType reflect.Type
}

// Error 返回 NonPointerTypeError 的错误文本。
//
// Error returns the NonPointerTypeError message.
func (err NonPointerTypeError) Error() string {
	return fmt.Sprintf("nbt: expected ptr type to decode into, but got '%v'", err.ActualType)
}

// BufferOverrunError 在读取操作越过传入数据缓冲区末尾时返回。
//
// BufferOverrunError is returned when the data buffer passed in when reading is overrun, meaning one of the
// reading operations extended beyond the end of the slice.
type BufferOverrunError struct {
	Op string
}

// Error 返回 BufferOverrunError 的错误文本。
//
// Error returns the BufferOverrunError message.
func (err BufferOverrunError) Error() string {
	return fmt.Sprintf("nbt: unexpected buffer end during op: '%v'", err.Op)
}

// InvalidArraySizeError 在 NBT 中读取到的数组长度与 Go 数组长度不一致时返回。
//
// InvalidArraySizeError is returned when an array read from the NBT (that includes byte arrays, int32 arrays
// and int64 arrays) does not have the same size as the Go representation.
type InvalidArraySizeError struct {
	Off       int64
	Op        string
	GoLength  int
	NBTLength int
}

// Error 返回 InvalidArraySizeError 的错误文本。
//
// Error returns the InvalidArraySizeError message.
func (err InvalidArraySizeError) Error() string {
	return fmt.Sprintf("nbt: mismatched array size at %v during op '%v': expected size %v, found %v in NBT", err.Off, err.Op, err.GoLength, err.NBTLength)
}

// UnexpectedNamedTagError 在复合标签中读取到目标结构体不存在的命名标签时返回。
//
// UnexpectedNamedTagError is returned when a named tag is read from a compound which is not present in the
// struct it is decoded into.
type UnexpectedNamedTagError struct {
	Off     int64
	TagName string
	TagType tagType
}

// Error 返回 UnexpectedNamedTagError 的错误文本。
//
// Error returns the UnexpectedNamedTagError message.
func (err UnexpectedNamedTagError) Error() string {
	return fmt.Sprintf("nbt: unexpected named tag '%v' with type %v at offset %v: not present in struct to be decoded into", err.TagName, err.TagType, err.Off)
}

// FailedWriteError 在 offsetWriter 写入失败、部分数据无法写入 io.Writer 时返回。
//
// FailedWriteError is returned if a Write operation failed on an offsetWriter, meaning some of the data could
// not be written to the io.Writer.
type FailedWriteError struct {
	Off int64
	Op  string
	Err error
}

// Error 返回 FailedWriteError 的错误文本。
//
// Error returns the FailedWriteError message.
func (err FailedWriteError) Error() string {
	return fmt.Sprintf("nbt: failed write during op '%v' at offset %v: %v", err.Op, err.Off, err.Err)
}

// IncompatibleTypeError 在尝试写入无法转换为 NBT 标签的值类型时返回。
//
// IncompatibleTypeError is returned if a value is attempted to be written to an io.Writer, but its type can-
// not be translated to an NBT tag.
type IncompatibleTypeError struct {
	ValueName string
	Type      reflect.Type
}

// Error 返回 IncompatibleTypeError 的错误文本。
//
// Error returns the IncompatibleTypeError message.
func (err IncompatibleTypeError) Error() string {
	return fmt.Sprintf("nbt: value type %v (%v) cannot be translated to an NBT tag", err.Type, err.ValueName)
}

var errStringTooLong = errors.New("string length exceeds maximum length")

// InvalidStringError 在读取到的字符串无效时返回，例如不是纯 UTF-8 字符或长度前缀无法容纳。
//
// InvalidStringError is returned if a string read is not valid, meaning it does not exist exclusively out of
// utf8 characters, or if it is longer than the length prefix can carry.
type InvalidStringError struct {
	Off int64
	Err error
	N   uint
}

// Error 返回 InvalidStringError 的错误文本。
//
// Error returns the InvalidStringError message.
func (err InvalidStringError) Error() string {
	return fmt.Sprintf("nbt: string at offset %v is not valid: %v (len=%v)", err.Off, err.Err, err.N)
}

const maximumNestingDepth = 512

// MaximumDepthReachedError 在读写 NBT 时复合标签或列表标签的嵌套深度达到 512 层后返回。
//
// MaximumDepthReachedError is returned if the maximum depth of 512 compound/list tags has been reached while
// reading or writing NBT.
type MaximumDepthReachedError struct {
}

// Error 返回 MaximumDepthReachedError 的错误文本。
//
// Error returns the MaximumDepthReachedError message.
func (err MaximumDepthReachedError) Error() string {
	return fmt.Sprintf("nbt: maximum nesting depth of %v was reached", maximumNestingDepth)
}

const maximumNetworkOffset = 4 * 1024 * 1024

// MaximumBytesReadError 在 NetworkLittleEndian 格式读取字节数达到上限时返回。
//
// MaximumBytesReadError is returned if the maximum amount of bytes has been read for NetworkLittleEndian
// format. It is returned if the offset hits maximumNetworkOffset.
type MaximumBytesReadError struct {
}

// Error 返回 MaximumBytesReadError 的错误文本。
//
// Error returns the MaximumBytesReadError message.
func (err MaximumBytesReadError) Error() string {
	return fmt.Sprintf("nbt: limit of bytes read %v with NetworkLittleEndian format exhausted", maximumNetworkOffset)
}

// InvalidVarintError 在读取到没有在 5 或 10 字节内结束的 varint32/varint64 时返回。
//
// InvalidVarintError is returned if a varint(32/64) is encountered that does
// not end after 5 or 10 bytes respectively.
type InvalidVarintError struct {
	Off int64
	N   int
}

// Error 返回 InvalidVarintError 的错误文本。
//
// Error returns the InvalidVarintError message.
func (err InvalidVarintError) Error() string {
	return fmt.Sprintf("nbt: varint did not terminate after %v bytes at offset %v", err.N, err.Off)
}
