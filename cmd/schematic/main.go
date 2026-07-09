package main

import (
	"compress/gzip"
	"fmt"
	"os"

	"github.com/HonLaderDev/minecraft-nbt/nbt"
)

func main() {
	file, err := os.Open("cmd/schematic/青云阁(1).schematic")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		panic(err)
	}
	defer gzipReader.Close()

	tr := nbt.NewTagReader(nbt.BigEndian)
	nr := nbt.NewNBTReader(tr, gzipReader)
	err = nr.
		// Schematic 的根标签必定是集合
		ReadTag().ExpectMustType(nbt.TagStruct).
			// 处理根标签
			// Schematic 的根标签名为 Schematic，
			// 但有些时候会套一层空名字标签
			ExpectName("").
				ReadTag().ExpectMustType(nbt.TagStruct).
					Return().Return(). // 跳出 ReadTag 与 ExpectMustType
				Return(). // 跳出 ExpectName

			// 处理 Schematic 的集合数据
			ReadStruct().
				// 解析 Schematic 的三维（长宽高），
				// 三维皆是 Short 类型，因此只需要
				// 期盼 Int16 类型即可
				ExpectType(nbt.TagInt16).
					// Schematic(Struct) -> Length(Int16)
					// 处理 Schematic 的长
					ExpectName("Length").
						Int16(func(value int16) error {
							fmt.Printf("Length: %d\n", value)
							return nil
						}).
					Return(). // 跳出 ExpectName

					// Schematic(Struct) -> Width(Int16)
					// 处理 Schematic 的宽
					ExpectName("Width").
						Int16(func(value int16) error {
							fmt.Printf("Width: %d\n", value)
							return nil
						}).
					Return(). // 跳出 ExpectName

					// Schematic(Struct) -> Height(Int16)
					// 处理 Schematic 的高
					ExpectName("Height").
						Int16(func(value int16) error {
							fmt.Printf("Height: %d\n", value)
							return nil
						}).
					Return(). // 跳出 ExpectName
				Return(). // 跳出 ExpectType
			Return(). // 跳出 ReadStruct
		Return().Return() // 跳出 ReadTag 与 ExpectMustType
	if err != nil {
		panic(err)
	}
}
