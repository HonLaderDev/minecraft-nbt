package nbt

import (
	"math"
	"unsafe"
)

// Encoding 表示一种 NBT 编码变体。通常 NBT 有三种编码，它们只在基础类型的写入方式上有所不同。
//
// Encoding is an encoding variant of NBT. In general, there are three different encodings of NBT, which are
// all the same except for the way basic types are written.
type Encoding interface {
	Int16(r *offsetReader) (int16, error)
	Int32(r *offsetReader) (int32, error)
	Int64(r *offsetReader) (int64, error)
	Float32(r *offsetReader) (float32, error)
	Float64(r *offsetReader) (float64, error)
	String(r *offsetReader) (string, error)
	Int32Slice(r *offsetReader) ([]int32, error)
	Int64Slice(r *offsetReader) ([]int64, error)

	WriteInt16(w *offsetWriter, x int16) error
	WriteInt32(w *offsetWriter, x int32) error
	WriteInt64(w *offsetWriter, x int64) error
	WriteFloat32(w *offsetWriter, x float32) error
	WriteFloat64(w *offsetWriter, x float64) error
	WriteString(w *offsetWriter, x string) error
}

var (
	// NetworkLittleEndian 是使用变长整数的 NBT 实现，其余部分与普通小端 NBT 相同。
	// NetworkLittleEndian 会限制可读取的 NBT 总字节数；达到限制时读取操作会立即失败。
	// 它通常用于 Bedrock Edition 协议中通过网络发送的 NBT。
	//
	// NetworkLittleEndian is the variable sized integer implementation of NBT. It is otherwise the same as the
	// normal little endian NBT. The NetworkLittleEndian format limits the total bytes of NBT that may be read. If
	// the limit is hit, the reading operation will fail immediately. NetworkLittleEndian is generally used for NBT
	// sent over network in the Bedrock Edition protocol.
	NetworkLittleEndian networkLittleEndian

	// LittleEndian 是使用固定长度整数的小端 NBT 实现，通常用于写入 Minecraft Bedrock Edition 的世界存档。
	//
	// LittleEndian is the fixed size little endian implementation of NBT. It is the format typically used for
	// writing Minecraft (Bedrock Edition) world saves.
	LittleEndian littleEndian

	// NetworkBigEndian 是 1.20.2 引入的 BigEndian 变体，不会写入根复合标签名称。
	// 与 BigEndian 一样，它只用于 Minecraft Java Edition，通常用于网络传输的 NBT。
	//
	// NetworkBigEndian is a version of BigEndian introduced in 1.20.2 where the name of the root compound tag is
	// not written. Similarly to BigEndian, it is only used on Minecraft Java Edition and generally used for NBT
	// sent over the network.
	NetworkBigEndian networkBigEndian

	// BigEndian 是使用固定长度整数的大端 NBT 实现，也是最初的 NBT 实现，只用于 Minecraft Java Edition。
	//
	// BigEndian is the fixed size big endian implementation of NBT. It is the original implementation, and is
	// used only on Minecraft Java Edition.
	BigEndian bigEndian

	_ Encoding = NetworkLittleEndian
	_ Encoding = LittleEndian
	_ Encoding = NetworkBigEndian
	_ Encoding = BigEndian
)

const maxStringSize = math.MaxInt16

type networkLittleEndian struct{ littleEndian }

// WriteInt32 将 int32 按 zigzag varint 格式写入。
//
// WriteInt32 writes an int32 in zigzag varint format.
func (networkLittleEndian) WriteInt32(w *offsetWriter, x int32) error {
	ux := uint32(x) << 1
	if x < 0 {
		ux = ^ux
	}
	for ux >= 0x80 {
		if err := w.WriteByte(byte(ux) | 0x80); err != nil {
			return FailedWriteError{Op: "WriteInt32", Off: w.off}
		}
		ux >>= 7
	}
	if err := w.WriteByte(byte(ux)); err != nil {
		return FailedWriteError{Op: "WriteInt32", Off: w.off}
	}
	return nil
}

// WriteInt64 将 int64 按 zigzag varint 格式写入。
//
// WriteInt64 writes an int64 in zigzag varint format.
func (networkLittleEndian) WriteInt64(w *offsetWriter, x int64) error {
	ux := uint64(x) << 1
	if x < 0 {
		ux = ^ux
	}
	for ux >= 0x80 {
		if err := w.WriteByte(byte(ux) | 0x80); err != nil {
			return FailedWriteError{Op: "WriteInt64", Off: w.off}
		}
		ux >>= 7
	}
	if err := w.WriteByte(byte(ux)); err != nil {
		return FailedWriteError{Op: "WriteInt64", Off: w.off}
	}
	return nil
}

// WriteString 写入以 varuint32 长度为前缀的字符串。
//
// WriteString writes a string prefixed with a varuint32 length.
func (networkLittleEndian) WriteString(w *offsetWriter, x string) error {
	if len(x) > maxStringSize {
		return InvalidStringError{Off: w.off, N: uint(len(x)), Err: errStringTooLong}
	}
	ux := uint32(len(x))
	for ux >= 0x80 {
		if err := w.WriteByte(byte(ux) | 0x80); err != nil {
			return FailedWriteError{Op: "WriteString", Off: w.off}
		}
		ux >>= 7
	}
	if err := w.WriteByte(byte(ux)); err != nil {
		return FailedWriteError{Op: "WriteString", Off: w.off}
	}
	// 使用 unsafe 将 string 转为字节切片，避免额外拷贝。
	//
	// Use unsafe conversion from a string to a byte slice to prevent copying.
	if _, err := w.Write(*(*[]byte)(unsafe.Pointer(&x))); err != nil {
		return FailedWriteError{Op: "WriteString", Off: w.off}
	}
	return nil
}

// Int32 读取 zigzag varint 格式的 int32。
//
// Int32 reads an int32 in zigzag varint format.
func (networkLittleEndian) Int32(r *offsetReader) (int32, error) {
	var ux uint32
	for i := uint(0); i < 35; i += 7 {
		b, err := r.ReadByte()
		if err != nil {
			return 0, BufferOverrunError{Op: "Int32"}
		}
		ux |= uint32(b&0x7f) << i
		if b&0x80 == 0 {
			x := int32(ux >> 1)
			if ux&1 != 0 {
				x = ^x
			}
			return x, nil
		}
	}
	return 0, InvalidVarintError{N: 5, Off: r.off}
}

// Int64 读取 zigzag varint 格式的 int64。
//
// Int64 reads an int64 in zigzag varint format.
func (networkLittleEndian) Int64(r *offsetReader) (int64, error) {
	var ux uint64
	for i := uint(0); i < 70; i += 7 {
		b, err := r.ReadByte()
		if err != nil {
			return 0, BufferOverrunError{Op: "Int64"}
		}
		ux |= uint64(b&0x7f) << i
		if b&0x80 == 0 {
			x := int64(ux >> 1)
			if ux&1 != 0 {
				x = ^x
			}
			return x, nil
		}
	}
	return 0, InvalidVarintError{N: 10, Off: r.off}
}

// String 读取以 varuint32 长度为前缀的字符串。
//
// String reads a string prefixed with a varuint32 length.
func (e networkLittleEndian) String(r *offsetReader) (string, error) {
	length, err := e.stringLength(r)
	if err != nil {
		return "", err
	}
	if length > maxStringSize {
		return "", InvalidStringError{N: uint(length), Off: r.off, Err: errStringTooLong}
	}
	data := make([]byte, length)
	if _, err := r.Read(data); err != nil {
		return "", BufferOverrunError{Op: "String"}
	}
	return *(*string)(unsafe.Pointer(&data)), nil
}

// stringLength 按 varuint32 读取字符串长度。
//
// stringLength reads the length of a string as a varuint32.
func (networkLittleEndian) stringLength(r *offsetReader) (uint32, error) {
	var ux uint32
	for i := uint(0); i < 35; i += 7 {
		b, err := r.ReadByte()
		if err != nil {
			return 0, BufferOverrunError{Op: "StringLength"}
		}
		ux |= uint32(b&0x7f) << i
		if b&0x80 == 0 {
			return ux, nil
		}
	}
	return 0, InvalidVarintError{N: 5, Off: r.off}
}

// Int32Slice 读取以 int32 数量为前缀的 int32 切片。
//
// Int32Slice reads an int32 slice prefixed with an int32 count.
func (e networkLittleEndian) Int32Slice(r *offsetReader) ([]int32, error) {
	n, err := e.Int32(r)
	if err != nil {
		return nil, BufferOverrunError{Op: "Int32Slice"}
	}
	m := make([]int32, n)
	for i := int32(0); i < n; i++ {
		m[i], err = e.Int32(r)
		if err != nil {
			return nil, BufferOverrunError{Op: "Int32Slice"}
		}
	}
	return m, nil
}

// Int64Slice 读取以 int32 数量为前缀的 int64 切片。
//
// Int64Slice reads an int64 slice prefixed with an int32 count.
func (e networkLittleEndian) Int64Slice(r *offsetReader) ([]int64, error) {
	n, err := e.Int32(r)
	if err != nil {
		return nil, BufferOverrunError{Op: "Int64Slice"}
	}
	m := make([]int64, n)
	for i := int32(0); i < n; i++ {
		m[i], err = e.Int64(r)
		if err != nil {
			return nil, BufferOverrunError{Op: "Int64Slice"}
		}
	}
	return m, nil
}

type networkBigEndian struct{ bigEndian }
