package utils

import (
	"bytes"
	"encoding/binary"
	"math"
	"reflect"
	"slices"
	"strings"
)

func getValueType(value reflect.Value) byte {
	kind := value.Kind()
	if kind == reflect.Interface || kind == reflect.Pointer {
		value = value.Elem()
		kind = value.Kind()
	}

	switch kind {
	case reflect.Uint8:
		return 1
	case reflect.Int16:
		return 2
	case reflect.Int32:
		return 3
	case reflect.Int64:
		return 4
	case reflect.Float32:
		return 5
	case reflect.Float64:
		return 6
	case reflect.Array:
		switch value.Type().Elem().Kind() {
		case reflect.Uint8:
			return 7
		case reflect.Int32:
			return 11
		case reflect.Int64:
			return 12
		}
	case reflect.String:
		return 8
	case reflect.Slice:
		return 9
	case reflect.Map:
		return 10
	}

	return 0
}

func marshalToName(writer *bytes.Buffer, name string) {
	temp := make([]byte, 2)
	binary.LittleEndian.PutUint16(temp, uint16(len(name)))
	writer.Write(temp)
	writer.WriteString(name)
}

func marshalToValue(writer *bytes.Buffer, value any, valueType int) {
	switch valueType {
	case 1:
		writer.WriteByte(value.(byte))
	case 2:
		temp := make([]byte, 2)
		binary.LittleEndian.PutUint16(temp, uint16(value.(int16)))
		writer.Write(temp)
	case 3:
		temp := make([]byte, 4)
		binary.LittleEndian.PutUint32(temp, uint32(value.(int32)))
		writer.Write(temp)
	case 4:
		temp := make([]byte, 8)
		binary.LittleEndian.PutUint64(temp, uint64(value.(int64)))
		writer.Write(temp)
	case 5:
		temp := make([]byte, 4)
		binary.LittleEndian.PutUint32(temp, math.Float32bits(value.(float32)))
		writer.Write(temp)
	case 6:
		temp := make([]byte, 8)
		binary.LittleEndian.PutUint64(temp, math.Float64bits(value.(float64)))
		writer.Write(temp)
	case 7, 11, 12:
		marshalToArray(writer, value, valueType)
	case 8:
		marshalToName(writer, value.(string))
	case 9:
		marshalToList(writer, value)
	case 10:
		marshalToCompound(writer, value.(map[string]any))
	}
}

func marshalToArray(writer *bytes.Buffer, value any, valueType int) {
	val := reflect.ValueOf(value)
	kind := val.Kind()

	if kind == reflect.Interface || kind == reflect.Pointer {
		val = val.Elem()
	}

	n := val.Cap()
	temp := make([]byte, 4)
	binary.LittleEndian.PutUint32(temp, uint32(n))
	writer.Write(temp)

	switch valueType {
	case 7:
		for i := range n {
			v := val.Index(i)
			if v.Kind() == reflect.Pointer {
				v = v.Elem()
			}
			writer.WriteByte(byte(v.Uint()))
		}
	case 11:
		for i := range n {
			v := val.Index(i)
			if v.Kind() == reflect.Pointer {
				v = v.Elem()
			}

			temp := make([]byte, 4)
			binary.LittleEndian.PutUint32(temp, uint32(v.Int()))
			writer.Write(temp)
		}
	case 12:
		for i := range n {
			v := val.Index(i)
			if v.Kind() == reflect.Pointer {
				v = v.Elem()
			}

			temp := make([]byte, 8)
			binary.LittleEndian.PutUint64(temp, uint64(v.Int()))
			writer.Write(temp)
		}
	}
}

func marshalToList(writer *bytes.Buffer, value any) {
	val := reflect.ValueOf(value)
	kind := val.Kind()

	if kind == reflect.Interface || kind == reflect.Pointer {
		val = val.Elem()
	}

	length := val.Len()
	if length == 0 {
		writer.Write([]byte{0, 0, 0, 0, 0})
		return
	}

	writer.WriteByte(getValueType(val.Index(0)))
	temp := make([]byte, 4)
	binary.LittleEndian.PutUint32(temp, uint32(length))
	writer.Write(temp)

	for i := range length {
		nestedValue := val.Index(i)
		if nestedValue.Kind() == reflect.Pointer {
			nestedValue = nestedValue.Elem()
		}
		marshalToValue(writer, nestedValue.Interface(), int(getValueType(nestedValue)))
	}
}

func marshalToCompound(writer *bytes.Buffer, value map[string]any) {
	keys := make([]string, 0)
	for key := range value {
		keys = append(keys, key)
	}
	slices.SortStableFunc(keys, func(a string, b string) int {
		return strings.Compare(a, b)
	})

	for _, key := range keys {
		val := reflect.ValueOf(value[key])
		if val.Kind() == reflect.Pointer {
			val = val.Elem()
		}

		valType := getValueType(val)
		writer.WriteByte(valType)
		marshalToName(writer, key)
		marshalToValue(writer, val.Interface(), int(valType))
	}
	writer.WriteByte(0)
}

// MarshalNBT marshal value as its little endian NBT represents to writer.
// Note that this implements is stable because the key in TAG_Compound is sorted.
// Other features are same to NBT implement on gophertunnel.
func MarshalNBT(writer *bytes.Buffer, value any, name string) {
	val := reflect.ValueOf(value)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}

	valueType := getValueType(val)
	writer.WriteByte(valueType)
	marshalToName(writer, name)
	marshalToValue(writer, val.Interface(), int(valueType))
}
