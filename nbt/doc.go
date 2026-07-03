// Package nbt 实现了 Minecraft 基岩版与 Java 版所使用的 NBT 格式。这些格式
// 分别是小端序格式、大端序格式，以及使用 varint 的小端序格式（通常用于
// 基岩版网络传输）。
//
// 该包对外提供的序列化与反序列化接口与标准 JSON 库大致一致：
// 操作字节切片时使用 nbt.Marshal() 与 nbt.Unmarshal；
// 操作读写流时使用 nbt.NewEncoder() 与 nbt.NewDecoder()。
//
// 该包按以下对应关系，将 Go 类型编码/解码为对应的 NBT 标签：
//
//	byte/uint8: TAG_Byte
//	bool: TAG_Byte
//	int16: TAG_Short
//	int32: TAG_Int
//	int64: TAG_Long
//	float32: TAG_Float
//	float64: TAG_Double
//	[...]byte: TAG_ByteArray
//	[...]int32: TAG_IntArray
//	[...]int64: TAG_LongArray
//	string: TAG_String
//	[]: TAG_List
//	struct{...}: TAG_Compound
//	map[string]<type/any>: TAG_Compound
//
// 被编码或解码的结构体可像标准 JSON 库一样使用结构体标签。
// “nbt” 结构体标签支持以下写法：
//
//	'-': 编码与解码时完全忽略该字段。
//	',omitempty': 字段值为默认值时不进行编码。
//	'名称(,omitempty)': 以指定名称编码/解码该字段。
//
// 若字段未设置 nbt 标签，则使用字段名进行编码/解码。
// 注意：与标准 JSON 库不同，本包在解码时区分大小写。
//
// Package nbt implements the NBT formats used by Minecraft Bedrock Edition and Minecraft Java Edition. These
// formats are a little endian format, a big endian format and a little endian format using varints (typically
// used over network in Bedrock Edition).
//
// The package exposes serialisation and deserialisation roughly the same way as the JSON standard library
// does, using nbt.Marshal() and nbt.Unmarshal when working with byte slices, and nbt.NewEncoder() and
// nbt.NewDecoder() when working with readers or writers.
//
// The package encodes and decodes the following Go types with the following NBT tags.
//
//	byte/uint8: TAG_Byte
//	bool: TAG_Byte
//	int16: TAG_Short
//	int32: TAG_Int
//	int64: TAG_Long
//	float32: TAG_Float
//	float64: TAG_Double
//	[...]byte: TAG_ByteArray
//	[...]int32: TAG_IntArray
//	[...]int64: TAG_LongArray
//	string: TAG_String
//	[]<type>: TAG_List
//	struct{...}: TAG_Compound
//	map[string]<type/any>: TAG_Compound
//
// Structures decoded or encoded may have struct field tags in a comparable way to the JSON standard library.
// The 'nbt' struct tag may be filled out the following ways:
//
//	'-': Ignores the field completely when encoding and decoding.
//	',omitempty': Doesn't encode the field if its value is the same as the default value.
//	'name(,omitempty)': Encodes/decodes the field with a different name than its usual name.
//
// If no 'nbt' struct tag is present for a field, the name of the field will be used to encode/decode the
// struct. Note that this package, unlike the JSON standard library package, is case sensitive when decoding.
package nbt
