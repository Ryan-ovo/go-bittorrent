package bencode

import (
	"io"
	"reflect"
	"strings"
)

// Marshal 把x这个结构序列化到TorrentFile中，返回写入的字节数
func Marshal(w io.Writer, x interface{}) int {
	v := reflect.ValueOf(x)
	// 如果是指针，则取一下值
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	return marshalValue(w, v)
}

func marshalValue(w io.Writer, v reflect.Value) int {
	wLen := 0
	switch v.Kind() {
	case reflect.String:
		wLen += EncodeString(w, v.String())
	case reflect.Int:
		wLen += EncodeInt(w, int(v.Int()))
	case reflect.Slice:
		wLen += marshalList(w, v)
	case reflect.Struct:
		wLen += marshalDict(w, v)
	}
	return wLen
}

func marshalList(w io.Writer, vl reflect.Value) int {
	wLen := 2
	w.Write([]byte{'l'})
	for i := 0; i < vl.Len(); i++ {
		elem := vl.Index(i)
		wLen += marshalValue(w, elem)
	}
	w.Write([]byte{'e'})
	return wLen
}

func marshalDict(w io.Writer, vd reflect.Value) int {
	wLen := 2
	w.Write([]byte{'d'})
	for i := 0; i < vd.NumField(); i++ {
		tf := vd.Type().Field(i)
		vf := vd.Field(i)
		key := tf.Tag.Get("bencode")
		if key == "" {
			key = strings.ToLower(tf.Name)
		}
		wLen += EncodeString(w, key)
		wLen += marshalValue(w, vf)
	}
	w.Write([]byte{'e'})
	return wLen
}
