package main

import (
	"bytes"
	"fmt"

	"github.com/TriM-Organization/bedrock-chunk-diff/utils"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
)

func main() {
	buf := bytes.NewBuffer(nil)
	sub := map[string]any{"2": byte(2), "是": map[string]any{"ssss": [3]byte{2, 2, 2}}}
	mm := map[string]any{
		"2":   int32(7),
		"你好":  [17]int32{2, 2, 2, 2, 2, 2, 2, 2, 2},
		"i":   []any{int64(2), int64(9)},
		"0sx": &sub,
	}

	utils.MarshalNBT(buf, &mm, ":)")
	fmt.Printf("%#v\n", buf.String())

	var m map[string]any
	buf = bytes.NewBuffer(buf.Bytes())
	fmt.Println(nbt.NewDecoderWithEncoding(buf, nbt.LittleEndian).Decode(&m))
	fmt.Println(m)
	fmt.Printf("%T\n", m["i"].([]int64)[0])

	buf = bytes.NewBuffer(nil)
	nbt.NewEncoderWithEncoding(buf, nbt.LittleEndian).Encode(m)
	fmt.Printf("%#v\n", buf.String())
}
