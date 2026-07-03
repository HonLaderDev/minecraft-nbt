package nbt

import (
	"fmt"
	"io"
	"reflect"
)

// ExpectTagTypeError 在链式读取器读到的标签类型与期望类型不一致时返回。
//
// ExpectTagTypeError is returned by the chained reader when the actual tag type does not match the expected
// tag type.
type ExpectTagTypeError struct {
	Off      int64
	Expected tagType
	Actual   tagType
}

// Error 返回 ExpectTagTypeError 的错误文本。
//
// Error returns the ExpectTagTypeError message.
func (err ExpectTagTypeError) Error() string {
	return fmt.Sprintf("nbt: expected tag type %v at offset %v, got %v", err.Expected, err.Off, err.Actual)
}

// ExpectTagNameError 在链式读取器读到的标签名与期望名称不一致时返回。
//
// ExpectTagNameError is returned by the chained reader when the actual tag name does not match the expected
// tag name.
type ExpectTagNameError struct {
	Off      int64
	Expected string
	Actual   string
}

// Error 返回 ExpectTagNameError 的错误文本。
//
// Error returns the ExpectTagNameError message.
func (err ExpectTagNameError) Error() string {
	return fmt.Sprintf("nbt: expected tag name %q at offset %v, got %q", err.Expected, err.Off, err.Actual)
}

// InvalidChainStateError 在链式读取顺序与当前标签状态不匹配时返回。
//
// InvalidChainStateError is returned when chained reader calls do not match the current tag state.
type InvalidChainStateError struct {
	Off int64
	Op  string
}

// Error 返回 InvalidChainStateError 的错误文本。
//
// Error returns the InvalidChainStateError message.
func (err InvalidChainStateError) Error() string {
	return fmt.Sprintf("nbt: invalid chained reader state at offset %v during op %q", err.Off, err.Op)
}

// InvalidChainCallbackError 在链式读取回调签名不符合要求时返回。
//
// InvalidChainCallbackError is returned when a chained reader callback does not have the expected signature.
type InvalidChainCallbackError struct {
	Op     string
	Actual reflect.Type
}

// Error 返回 InvalidChainCallbackError 的错误文本。
//
// Error returns the InvalidChainCallbackError message.
func (err InvalidChainCallbackError) Error() string {
	return fmt.Sprintf("nbt: invalid callback for op %q: got %v, want func(T) error", err.Op, err.Actual)
}

// NBTChain 表示链式读取中的一个步骤。根级 Return 在成功时返回 nil，失败时返回可作为 error 使用的 NBTChain。
//
// NBTChain represents a step in a chained read. The root Return returns nil on success, or an NBTChain that
// can be used as an error on failure.
type NBTChain interface {
	error

	// ReadTag 读取下一个完整标签头，并对该标签执行子链。
	//
	// ReadTag reads the next complete tag header and executes child chains against that tag.
	ReadTag() NBTChain

	// ReadStruct 循环读取当前 TAG_Compound 内的子标签，直到 TAG_End 或出错。
	// 每个未被子链消费的子标签会自动跳过。
	//
	// ReadStruct loops over child tags in the current TAG_Compound until TAG_End or an error. Child tags not
	// consumed by child chains are skipped automatically.
	ReadStruct() NBTChain

	// ReadList 循环读取当前 TAG_List 内的元素。列表元素没有名称，因此通常搭配 ExpectType 与 ExpectIndex 使用。
	// 每个未被子链消费的元素会自动跳过。
	//
	// ReadList loops over elements in the current TAG_List. List elements have no names, so it is usually used
	// with ExpectType and ExpectIndex. Elements not consumed by child chains are skipped automatically.
	ReadList() NBTChain

	// Struct 直接读取完整 TAG_Compound，并根据 func(value T) error 的入参类型自动反序列化为 T。
	//
	// Struct reads a complete TAG_Compound and automatically unmarshals it into the input type of
	// func(value T) error.
	Struct(any) NBTChain

	// List 直接读取完整 TAG_List，并根据 func(value T) error 的入参类型自动反序列化为 T。
	//
	// List reads a complete TAG_List and automatically unmarshals it into the input type of func(value T) error.
	List(any) NBTChain

	// Do 在当前分支命中时执行回调；它不会消费当前 tag 或列表元素。
	//
	// Do runs the callback when the current branch matches. It does not consume the current tag or list element.
	Do(func() error) NBTChain

	// Skip 在当前分支命中时跳过当前 tag 或列表元素，并将其标记为已消费。
	//
	// Skip skips the current tag or list element when the current branch matches and marks it as consumed.
	Skip() NBTChain

	// When 使用当前 NBTReader 状态进行软匹配；返回 false 时跳过当前分支，不报错，也不消费当前 tag 或列表元素。
	//
	// When softly matches using the current NBTReader state. Returning false skips the current branch without
	// returning an error or consuming the current tag or list element.
	When(func(*NBTReader) bool) NBTChain

	// ExpectType 软匹配当前标签类型；不命中时跳过当前分支，不报错，也不消费当前 tag 或列表元素。
	//
	// ExpectType softly matches the current tag type. A mismatch skips the current branch without returning an
	// error or consuming the current tag or list element.
	ExpectType(tagType) NBTChain

	// ExpectMustType 强匹配当前标签类型；不命中时返回错误。
	//
	// ExpectMustType strictly matches the current tag type and returns an error on mismatch.
	ExpectMustType(tagType) NBTChain

	// ExpectName 软匹配当前标签名；不命中时跳过当前分支，不报错，也不消费当前 tag。
	// 列表元素没有名称，因此 ReadList 下通常不使用 ExpectName。
	//
	// ExpectName softly matches the current tag name. A mismatch skips the current branch without returning an
	// error or consuming the current tag. List elements do not have names, so ExpectName is usually not used
	// under ReadList.
	ExpectName(string) NBTChain

	// ExpectMustName 强匹配当前标签名；不命中时返回错误。
	//
	// ExpectMustName strictly matches the current tag name and returns an error on mismatch.
	ExpectMustName(string) NBTChain

	// ExpectIndex 软匹配当前列表元素索引；不命中时跳过当前分支，不报错，也不消费当前元素。
	// 它只在 ReadList 的子链中有意义。
	//
	// ExpectIndex softly matches the current list element index. A mismatch skips the current branch without
	// returning an error or consuming the current element. It is meaningful only under ReadList.
	ExpectIndex(int) NBTChain

	// Byte 读取当前 TAG_Byte 值；类型不匹配会返回错误。
	//
	// Byte reads the current TAG_Byte value and returns an error on type mismatch.
	Byte(func(byte) error) NBTChain

	// Int16 读取当前 TAG_Short 值；类型不匹配会返回错误。
	//
	// Int16 reads the current TAG_Short value and returns an error on type mismatch.
	Int16(func(int16) error) NBTChain

	// Int32 读取当前 TAG_Int 值；类型不匹配会返回错误。
	//
	// Int32 reads the current TAG_Int value and returns an error on type mismatch.
	Int32(func(int32) error) NBTChain

	// Int64 读取当前 TAG_Long 值；类型不匹配会返回错误。
	//
	// Int64 reads the current TAG_Long value and returns an error on type mismatch.
	Int64(func(int64) error) NBTChain

	// Float32 读取当前 TAG_Float 值；类型不匹配会返回错误。
	//
	// Float32 reads the current TAG_Float value and returns an error on type mismatch.
	Float32(func(float32) error) NBTChain

	// Float64 读取当前 TAG_Double 值；类型不匹配会返回错误。
	//
	// Float64 reads the current TAG_Double value and returns an error on type mismatch.
	Float64(func(float64) error) NBTChain

	// String 读取当前 TAG_String 值；类型不匹配会返回错误。
	//
	// String reads the current TAG_String value and returns an error on type mismatch.
	String(func(string) error) NBTChain

	// ByteArray 读取当前 TAG_ByteArray 值；类型不匹配会返回错误。
	//
	// ByteArray reads the current TAG_ByteArray value and returns an error on type mismatch.
	ByteArray(func([]byte) error) NBTChain

	// Int32Array 读取当前 TAG_IntArray 值；类型不匹配会返回错误。
	//
	// Int32Array reads the current TAG_IntArray value and returns an error on type mismatch.
	Int32Array(func([]int32) error) NBTChain

	// Int64Array 读取当前 TAG_LongArray 值；类型不匹配会返回错误。
	//
	// Int64Array reads the current TAG_LongArray value and returns an error on type mismatch.
	Int64Array(func([]int64) error) NBTChain

	// Return 结束当前链式步骤；根级 Return 会执行已构建的规则并返回错误结果。
	//
	// Return ends the current chain step. The root Return executes the built rule tree and returns the result.
	Return() NBTChain
}

// NBTReader 提供基于 TagReader 的链式流式读取 API。
//
// NBTReader provides a chained streaming API on top of TagReader.
type NBTReader struct {
	tagReader *TagReader
	reader    *offsetReader
	err       error

	currentTag *nbtChainTag
	listLen    int
	listIndex  int
}

type nbtChainNodeKind byte

const (
	nbtChainReadTag nbtChainNodeKind = iota
	nbtChainReadStruct
	nbtChainReadList
	nbtChainExpectType
	nbtChainExpectName
	nbtChainExpectIndex
	nbtChainByte
	nbtChainInt16
	nbtChainInt32
	nbtChainInt64
	nbtChainFloat32
	nbtChainFloat64
	nbtChainString
	nbtChainByteArray
	nbtChainInt32Array
	nbtChainInt64Array
	nbtChainStruct
	nbtChainList
	nbtChainDo
	nbtChainSkip
	nbtChainWhen
)

type nbtChainTag struct {
	tagType  tagType
	name     string
	off      int64
	index    int
	hasIndex bool
	consumed bool
}

type nbtChainNode struct {
	reader   *NBTReader
	parent   NBTChain
	children []*nbtChainNode

	kind nbtChainNodeKind
	root bool
	must bool

	expectedType  tagType
	expectedName  string
	expectedIndex int
	byteFn        func(byte) error
	int16Fn       func(int16) error
	int32Fn       func(int32) error
	int64Fn       func(int64) error
	float32Fn     func(float32) error
	float64Fn     func(float64) error
	stringFn      func(string) error
	byteArrayFn   func([]byte) error
	int32ArrayFn  func([]int32) error
	int64ArrayFn  func([]int64) error
	structFn      any
	listFn        any
	doFn          func() error
	whenFn        func(*NBTReader) bool
}

type nbtChainError struct {
	err error
}

// NewNBTReader 使用指定 TagReader 与输入流创建链式 NBT 读取器。
//
// NewNBTReader creates a chained NBT reader using the provided TagReader and input stream.
func NewNBTReader(tagReader *TagReader, reader io.Reader) *NBTReader {
	return &NBTReader{
		tagReader: tagReader,
		reader:    newOffsetReader(reader),
		listLen:   -1,
		listIndex: -1,
	}
}

// ListLen 返回当前 ReadList 循环正在读取的列表长度；不在 ReadList 循环中时返回 -1。
//
// ListLen returns the length of the list currently being read by ReadList. It returns -1 outside a ReadList
// loop.
func (r *NBTReader) ListLen() int {
	return r.listLen
}

// ListIndex 返回当前 ReadList 循环正在读取的元素索引；不在 ReadList 循环中时返回 -1。
//
// ListIndex returns the element index currently being read by ReadList. It returns -1 outside a ReadList
// loop.
func (r *NBTReader) ListIndex() int {
	return r.listIndex
}

// TagName 返回当前链式分支正在处理的标签名；不在标签上下文中或当前为列表元素时返回空字符串。
//
// TagName returns the name of the tag currently being handled by the chained reader. It returns an empty
// string outside a tag context or for list elements.
func (r *NBTReader) TagName() string {
	if r.currentTag == nil {
		return ""
	}
	return r.currentTag.name
}

// TagType 返回当前链式分支正在处理的标签类型；不在标签上下文中时返回 TagEnd。
//
// TagType returns the type of the tag currently being handled by the chained reader. It returns TagEnd outside
// a tag context.
func (r *NBTReader) TagType() tagType {
	if r.currentTag == nil {
		return tagEnd
	}
	return r.currentTag.tagType
}

// ReadTag 读取下一个完整标签头，并开始一条链式读取调用。
//
// ReadTag reads the next complete tag header and starts a chained read call.
func (r *NBTReader) ReadTag() NBTChain {
	return &nbtChainNode{
		reader: r,
		kind:   nbtChainReadTag,
		root:   true,
	}
}

func (r *NBTReader) readTag() *nbtChainTag {
	if r.err != nil {
		return nil
	}
	off := r.reader.GetOffset()
	tagType, name, err := r.tagReader.ReadTag(r.reader)
	if err != nil {
		r.setErr(err)
		return nil
	}
	return &nbtChainTag{tagType: tagType, name: name, off: off}
}

func (r *NBTReader) rootResult() NBTChain {
	if r.err != nil {
		return &nbtChainError{err: r.err}
	}
	return nil
}

func (r *NBTReader) setErr(err error) {
	if err != nil && r.err == nil {
		r.err = err
	}
}

func (r *NBTReader) expectTagType(tag *nbtChainTag, expected tagType) {
	if r.err != nil || tag == nil {
		return
	}
	if tag.tagType != expected {
		r.setErr(ExpectTagTypeError{
			Off:      tag.off,
			Expected: expected,
			Actual:   tag.tagType,
		})
	}
}

func (r *NBTReader) expectTagName(tag *nbtChainTag, expected string) {
	if r.err != nil || tag == nil {
		return
	}
	if tag.name != expected {
		r.setErr(ExpectTagNameError{
			Off:      tag.off,
			Expected: expected,
			Actual:   tag.name,
		})
	}
}

func (r *NBTReader) finishTag(tag *nbtChainTag) {
	if r.err != nil || tag == nil || tag.consumed {
		return
	}
	if err := r.tagReader.SkipTagValue(r.reader, tag.tagType); err != nil {
		r.setErr(err)
		return
	}
	tag.consumed = true
}

func (r *NBTReader) readTagInt16(tag *nbtChainTag, fn func(int16) error) {
	readChainValue(r, tag, tagInt16, "Int16", func() (int16, error) {
		return r.tagReader.ReadTagInt16(r.reader)
	}, fn)
}

func readChainValue[T any](r *NBTReader, tag *nbtChainTag, expected tagType, op string, read func() (T, error), fn func(T) error) {
	r.expectTagType(tag, expected)
	if r.err != nil || tag == nil {
		return
	}
	if tag.consumed {
		r.setErr(InvalidChainStateError{Off: tag.off, Op: op})
		return
	}
	value, err := read()
	if err != nil {
		r.setErr(err)
		return
	}
	tag.consumed = true
	if fn != nil {
		r.setErr(fn(value))
	}
}

func (r *NBTReader) readTagStruct(tag *nbtChainTag, fn any) {
	r.expectTagType(tag, tagStruct)
	if r.err != nil || tag == nil {
		return
	}
	if tag.consumed {
		r.setErr(InvalidChainStateError{Off: tag.off, Op: "Struct"})
		return
	}
	if fn == nil {
		_, err := r.tagReader.ReadTagCompound(r.reader)
		if err != nil {
			r.setErr(err)
			return
		}
		tag.consumed = true
		return
	}

	callback := reflect.ValueOf(fn)
	callbackType := callback.Type()
	errorType := reflect.TypeOf((*error)(nil)).Elem()
	if callback.Kind() != reflect.Func ||
		callback.IsNil() ||
		callbackType.NumIn() != 1 ||
		callbackType.NumOut() != 1 ||
		!callbackType.Out(0).AssignableTo(errorType) {
		r.setErr(InvalidChainCallbackError{Op: "Struct", Actual: callbackType})
		return
	}

	argType := callbackType.In(0)
	var callbackArg reflect.Value
	var decodeTarget reflect.Value
	if argType.Kind() == reflect.Pointer {
		callbackArg = reflect.New(argType.Elem())
		decodeTarget = callbackArg.Elem()
	} else {
		decodePtr := reflect.New(argType)
		callbackArg = decodePtr.Elem()
		decodeTarget = callbackArg
	}

	decoder := &Decoder{Encoding: r.tagReader.endian, r: r.reader}
	if err := decoder.unmarshalTag(decodeTarget, tagStruct, tag.name); err != nil {
		r.setErr(err)
		return
	}
	tag.consumed = true

	results := callback.Call([]reflect.Value{callbackArg})
	if !results[0].IsNil() {
		r.setErr(results[0].Interface().(error))
	}
}

func (r *NBTReader) readTagList(tag *nbtChainTag, fn any) {
	r.expectTagType(tag, tagSlice)
	if r.err != nil || tag == nil {
		return
	}
	if tag.consumed {
		r.setErr(InvalidChainStateError{Off: tag.off, Op: "List"})
		return
	}
	if fn == nil {
		_, err := r.tagReader.ReadTagList(r.reader)
		if err != nil {
			r.setErr(err)
			return
		}
		tag.consumed = true
		return
	}

	callback := reflect.ValueOf(fn)
	callbackType := callback.Type()
	errorType := reflect.TypeOf((*error)(nil)).Elem()
	if callback.Kind() != reflect.Func ||
		callback.IsNil() ||
		callbackType.NumIn() != 1 ||
		callbackType.NumOut() != 1 ||
		!callbackType.Out(0).AssignableTo(errorType) {
		r.setErr(InvalidChainCallbackError{Op: "List", Actual: callbackType})
		return
	}

	argType := callbackType.In(0)
	var callbackArg reflect.Value
	var decodeTarget reflect.Value
	if argType.Kind() == reflect.Pointer {
		callbackArg = reflect.New(argType.Elem())
		decodeTarget = callbackArg.Elem()
	} else {
		decodePtr := reflect.New(argType)
		callbackArg = decodePtr.Elem()
		decodeTarget = callbackArg
	}

	decoder := &Decoder{Encoding: r.tagReader.endian, r: r.reader}
	if err := decoder.unmarshalTag(decodeTarget, tagSlice, tag.name); err != nil {
		r.setErr(err)
		return
	}
	tag.consumed = true

	results := callback.Call([]reflect.Value{callbackArg})
	if !results[0].IsNil() {
		r.setErr(results[0].Interface().(error))
	}
}

func (r *NBTReader) readListHeader(tag *nbtChainTag) (tagType, int32, bool) {
	r.expectTagType(tag, tagSlice)
	if r.err != nil || tag == nil {
		return tagEnd, 0, false
	}
	if tag.consumed {
		r.setErr(InvalidChainStateError{Off: tag.off, Op: "List"})
		return tagEnd, 0, false
	}
	elementTypeByte, err := r.reader.ReadByte()
	if err != nil {
		r.setErr(BufferOverrunError{Op: "List"})
		return tagEnd, 0, false
	}
	elementType := tagType(elementTypeByte)
	if !elementType.IsValid() {
		r.setErr(UnknownTagError{Off: r.reader.GetOffset(), TagType: elementType, Op: "List"})
		return tagEnd, 0, false
	}
	length, err := r.tagReader.endian.Int32(r.reader)
	if err != nil {
		r.setErr(err)
		return tagEnd, 0, false
	}
	if length < 0 {
		r.setErr(BufferOverrunError{Op: "List"})
		return tagEnd, 0, false
	}
	return elementType, length, true
}

func (n *nbtChainNode) Error() string {
	if n == nil || n.reader == nil || n.reader.err == nil {
		return "nbt: chained reader has not returned to root"
	}
	return n.reader.err.Error()
}

func (n *nbtChainNode) ReadTag() NBTChain {
	return n.addChild(nbtChainReadTag)
}

func (n *nbtChainNode) ReadStruct() NBTChain {
	return n.addChild(nbtChainReadStruct)
}

func (n *nbtChainNode) ReadList() NBTChain {
	return n.addChild(nbtChainReadList)
}

func (n *nbtChainNode) Struct(fn any) NBTChain {
	child := n.addChild(nbtChainStruct).(*nbtChainNode)
	child.structFn = fn
	return n
}

func (n *nbtChainNode) List(fn any) NBTChain {
	child := n.addChild(nbtChainList).(*nbtChainNode)
	child.listFn = fn
	return n
}

func (n *nbtChainNode) Do(fn func() error) NBTChain {
	child := n.addChild(nbtChainDo).(*nbtChainNode)
	child.doFn = fn
	return n
}

func (n *nbtChainNode) Skip() NBTChain {
	n.addChild(nbtChainSkip)
	return n
}

func (n *nbtChainNode) When(fn func(*NBTReader) bool) NBTChain {
	child := n.addChild(nbtChainWhen).(*nbtChainNode)
	child.whenFn = fn
	return child
}

func (n *nbtChainNode) ExpectType(expected tagType) NBTChain {
	child := n.addChild(nbtChainExpectType).(*nbtChainNode)
	child.expectedType = expected
	return child
}

func (n *nbtChainNode) ExpectMustType(expected tagType) NBTChain {
	child := n.ExpectType(expected).(*nbtChainNode)
	child.must = true
	return child
}

func (n *nbtChainNode) ExpectName(expected string) NBTChain {
	child := n.addChild(nbtChainExpectName).(*nbtChainNode)
	child.expectedName = expected
	return child
}

func (n *nbtChainNode) ExpectMustName(expected string) NBTChain {
	child := n.ExpectName(expected).(*nbtChainNode)
	child.must = true
	return child
}

func (n *nbtChainNode) ExpectIndex(expected int) NBTChain {
	child := n.addChild(nbtChainExpectIndex).(*nbtChainNode)
	child.expectedIndex = expected
	return child
}

func (n *nbtChainNode) Byte(fn func(byte) error) NBTChain {
	child := n.addChild(nbtChainByte).(*nbtChainNode)
	child.byteFn = fn
	return n
}

func (n *nbtChainNode) Int16(fn func(int16) error) NBTChain {
	child := n.addChild(nbtChainInt16).(*nbtChainNode)
	child.int16Fn = fn
	return n
}

func (n *nbtChainNode) Int32(fn func(int32) error) NBTChain {
	child := n.addChild(nbtChainInt32).(*nbtChainNode)
	child.int32Fn = fn
	return n
}

func (n *nbtChainNode) Int64(fn func(int64) error) NBTChain {
	child := n.addChild(nbtChainInt64).(*nbtChainNode)
	child.int64Fn = fn
	return n
}

func (n *nbtChainNode) Float32(fn func(float32) error) NBTChain {
	child := n.addChild(nbtChainFloat32).(*nbtChainNode)
	child.float32Fn = fn
	return n
}

func (n *nbtChainNode) Float64(fn func(float64) error) NBTChain {
	child := n.addChild(nbtChainFloat64).(*nbtChainNode)
	child.float64Fn = fn
	return n
}

func (n *nbtChainNode) String(fn func(string) error) NBTChain {
	child := n.addChild(nbtChainString).(*nbtChainNode)
	child.stringFn = fn
	return n
}

func (n *nbtChainNode) ByteArray(fn func([]byte) error) NBTChain {
	child := n.addChild(nbtChainByteArray).(*nbtChainNode)
	child.byteArrayFn = fn
	return n
}

func (n *nbtChainNode) Int32Array(fn func([]int32) error) NBTChain {
	child := n.addChild(nbtChainInt32Array).(*nbtChainNode)
	child.int32ArrayFn = fn
	return n
}

func (n *nbtChainNode) Int64Array(fn func([]int64) error) NBTChain {
	child := n.addChild(nbtChainInt64Array).(*nbtChainNode)
	child.int64ArrayFn = fn
	return n
}

func (n *nbtChainNode) Return() NBTChain {
	if !n.root {
		return n.parent
	}
	n.exec(nil)
	return n.reader.rootResult()
}

func (n *nbtChainNode) addChild(kind nbtChainNodeKind) NBTChain {
	child := &nbtChainNode{
		reader: n.reader,
		parent: n,
		kind:   kind,
	}
	n.children = append(n.children, child)
	return child
}

func (n *nbtChainNode) exec(tag *nbtChainTag) {
	if n.reader.err != nil {
		return
	}
	switch n.kind {
	case nbtChainReadTag:
		n.execReadTag()
	case nbtChainReadStruct:
		n.execStruct(tag)
	case nbtChainReadList:
		n.execList(tag)
	case nbtChainExpectType:
		n.execExpectType(tag)
	case nbtChainExpectName:
		n.execExpectName(tag)
	case nbtChainExpectIndex:
		n.execExpectIndex(tag)
	case nbtChainByte:
		readChainValue(n.reader, tag, tagByte, "Byte", func() (byte, error) {
			return n.reader.tagReader.ReadTagByte(n.reader.reader)
		}, n.byteFn)
	case nbtChainInt16:
		n.reader.readTagInt16(tag, n.int16Fn)
	case nbtChainInt32:
		readChainValue(n.reader, tag, tagInt32, "Int32", func() (int32, error) {
			return n.reader.tagReader.ReadTagInt32(n.reader.reader)
		}, n.int32Fn)
	case nbtChainInt64:
		readChainValue(n.reader, tag, tagInt64, "Int64", func() (int64, error) {
			return n.reader.tagReader.ReadTagInt64(n.reader.reader)
		}, n.int64Fn)
	case nbtChainFloat32:
		readChainValue(n.reader, tag, tagFloat32, "Float32", func() (float32, error) {
			return n.reader.tagReader.ReadTagFloat32(n.reader.reader)
		}, n.float32Fn)
	case nbtChainFloat64:
		readChainValue(n.reader, tag, tagFloat64, "Float64", func() (float64, error) {
			return n.reader.tagReader.ReadTagFloat64(n.reader.reader)
		}, n.float64Fn)
	case nbtChainString:
		readChainValue(n.reader, tag, tagString, "String", func() (string, error) {
			return n.reader.tagReader.ReadTagString(n.reader.reader)
		}, n.stringFn)
	case nbtChainByteArray:
		readChainValue(n.reader, tag, tagByteArray, "ByteArray", func() ([]byte, error) {
			return n.reader.tagReader.ReadTagByteArray(n.reader.reader)
		}, n.byteArrayFn)
	case nbtChainInt32Array:
		readChainValue(n.reader, tag, tagInt32Array, "Int32Array", func() ([]int32, error) {
			return n.reader.tagReader.ReadTagInt32Array(n.reader.reader)
		}, n.int32ArrayFn)
	case nbtChainInt64Array:
		readChainValue(n.reader, tag, tagInt64Array, "Int64Array", func() ([]int64, error) {
			return n.reader.tagReader.ReadTagInt64Array(n.reader.reader)
		}, n.int64ArrayFn)
	case nbtChainStruct:
		n.reader.readTagStruct(tag, n.structFn)
	case nbtChainList:
		n.reader.readTagList(tag, n.listFn)
	case nbtChainDo:
		if n.doFn != nil {
			n.reader.setErr(n.doFn())
		}
	case nbtChainSkip:
		n.reader.finishTag(tag)
	case nbtChainWhen:
		n.execWhen(tag)
	}
}

func (n *nbtChainNode) execReadTag() {
	tag := n.reader.readTag()
	if n.reader.err != nil {
		return
	}
	n.reader.withCurrentTag(tag, func() {
		n.execChildren(tag)
		n.reader.finishTag(tag)
	})
}

func (n *nbtChainNode) execStruct(tag *nbtChainTag) {
	if tag == nil {
		n.reader.setErr(InvalidChainStateError{Off: n.reader.reader.GetOffset(), Op: "Struct"})
		return
	}
	if tag.tagType != tagStruct {
		n.reader.setErr(ExpectTagTypeError{
			Off:      tag.off,
			Expected: tagStruct,
			Actual:   tag.tagType,
		})
		return
	}
	for {
		childTag := n.reader.readTag()
		if n.reader.err != nil {
			return
		}
		if childTag.tagType == tagEnd {
			tag.consumed = true
			return
		}
		n.reader.withCurrentTag(childTag, func() {
			n.execChildren(childTag)
			n.reader.finishTag(childTag)
		})
	}
}

func (n *nbtChainNode) execList(tag *nbtChainTag) {
	elementType, length, ok := n.reader.readListHeader(tag)
	if !ok {
		return
	}
	previousLen := n.reader.listLen
	previousIndex := n.reader.listIndex
	n.reader.listLen = int(length)
	defer func() {
		n.reader.listLen = previousLen
		n.reader.listIndex = previousIndex
	}()
	for i := int32(0); i < length; i++ {
		n.reader.listIndex = int(i)
		elementTag := &nbtChainTag{
			tagType:  elementType,
			off:      n.reader.reader.GetOffset(),
			index:    int(i),
			hasIndex: true,
		}
		n.reader.withCurrentTag(elementTag, func() {
			n.execChildren(elementTag)
			n.reader.finishTag(elementTag)
		})
		if n.reader.err != nil {
			return
		}
	}
	tag.consumed = true
}

func (r *NBTReader) withCurrentTag(tag *nbtChainTag, fn func()) {
	previousTag := r.currentTag
	r.currentTag = tag
	defer func() {
		r.currentTag = previousTag
	}()
	fn()
}

func (n *nbtChainNode) execExpectType(tag *nbtChainTag) {
	if tag == nil {
		return
	}
	if tag.tagType != n.expectedType {
		if n.must {
			n.reader.expectTagType(tag, n.expectedType)
		}
		return
	}
	n.execChildren(tag)
}

func (n *nbtChainNode) execExpectName(tag *nbtChainTag) {
	if tag == nil {
		return
	}
	if tag.name != n.expectedName {
		if n.must {
			n.reader.expectTagName(tag, n.expectedName)
		}
		return
	}
	n.execChildren(tag)
}

func (n *nbtChainNode) execExpectIndex(tag *nbtChainTag) {
	if tag == nil || !tag.hasIndex {
		return
	}
	if tag.index != n.expectedIndex {
		return
	}
	n.execChildren(tag)
}

func (n *nbtChainNode) execWhen(tag *nbtChainTag) {
	if tag == nil || n.whenFn == nil || !n.whenFn(n.reader) {
		return
	}
	n.execChildren(tag)
}

func (n *nbtChainNode) execChildren(tag *nbtChainTag) {
	for _, child := range n.children {
		child.exec(tag)
		if n.reader.err != nil {
			return
		}
	}
}

func (err *nbtChainError) Error() string {
	if err == nil || err.err == nil {
		return ""
	}
	return err.err.Error()
}

func (err *nbtChainError) Unwrap() error {
	if err == nil {
		return nil
	}
	return err.err
}

func (err *nbtChainError) ReadTag() NBTChain {
	return err
}

func (err *nbtChainError) ReadStruct() NBTChain {
	return err
}

func (err *nbtChainError) ReadList() NBTChain {
	return err
}

func (err *nbtChainError) Struct(any) NBTChain {
	return err
}

func (err *nbtChainError) List(any) NBTChain {
	return err
}

func (err *nbtChainError) Do(func() error) NBTChain {
	return err
}

func (err *nbtChainError) Skip() NBTChain {
	return err
}

func (err *nbtChainError) When(func(*NBTReader) bool) NBTChain {
	return err
}

func (err *nbtChainError) ExpectType(tagType) NBTChain {
	return err
}

func (err *nbtChainError) ExpectMustType(tagType) NBTChain {
	return err
}

func (err *nbtChainError) ExpectName(string) NBTChain {
	return err
}

func (err *nbtChainError) ExpectMustName(string) NBTChain {
	return err
}

func (err *nbtChainError) ExpectIndex(int) NBTChain {
	return err
}

func (err *nbtChainError) Byte(func(byte) error) NBTChain {
	return err
}

func (err *nbtChainError) Int16(func(int16) error) NBTChain {
	return err
}

func (err *nbtChainError) Int32(func(int32) error) NBTChain {
	return err
}

func (err *nbtChainError) Int64(func(int64) error) NBTChain {
	return err
}

func (err *nbtChainError) Float32(func(float32) error) NBTChain {
	return err
}

func (err *nbtChainError) Float64(func(float64) error) NBTChain {
	return err
}

func (err *nbtChainError) String(func(string) error) NBTChain {
	return err
}

func (err *nbtChainError) ByteArray(func([]byte) error) NBTChain {
	return err
}

func (err *nbtChainError) Int32Array(func([]int32) error) NBTChain {
	return err
}

func (err *nbtChainError) Int64Array(func([]int64) error) NBTChain {
	return err
}

func (err *nbtChainError) Return() NBTChain {
	return err
}
