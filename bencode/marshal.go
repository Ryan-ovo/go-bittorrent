package bencode

import (
	"io"
	"reflect"
)

func Marshal(w io.Writer, x interface{}) int {
	return 0
}

func marshalValue(w io.Writer, v reflect.Value) int {
	return 0
}

func marshalList(w io.Writer, vl reflect.Value) int {
	return 0
}

func marshalDict(w io.Writer, vd reflect.Value) int {
	return 0
}
