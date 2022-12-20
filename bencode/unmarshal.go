package bencode

import (
	"errors"
	"io"
	"log"
	"reflect"
	"strings"
)

/*
	Unmarshal 将TorrentFile中的数据反序列化到结构v中，根据反序列化的结果推断v的类型
	v必须是slice或者struct的指针才能序列化成功
*/
func Unmarshal(r io.Reader, v interface{}) error {
	obj, err := Parse(r)
	if err != nil {
		return err
	}
	rv := reflect.ValueOf(v)
	if rv.Type().Kind() != reflect.Pointer {
		return errors.New("unmarshal structure need to be a pointer")
	}
	switch obj.typ {
	case LIST:
		list, err := obj.List()
		if err != nil {
			return err
		}
		// 如果v的值类型不是slice，报错
		if rv.Elem().Type().Kind() != reflect.Slice {
			return errors.New("dest must be pointer to slice")
		}
		a := reflect.MakeSlice(rv.Elem().Type(), len(list), len(list))
		rv.Elem().Set(a)
		if err = unmarshalList(rv, list); err != nil {
			return err
		}
	case DICT:
		dict, err := obj.Dict()
		if err != nil {
			return err
		}
		if err = unmarshalDict(rv, dict); err != nil {
			return err
		}
	default:
		return errors.New("src code must be struct or slice")
	}
	return nil
}

/*
	反序列化列表，按照第一个元素区分类型，入参（反序列化的值对象, BObject的value）
		1. string, 列表元素全部都是string
		2. int, 列表元素全部都是int
		3. list, 列表元素全部都是list，递归处理
		4. dict, 列表元素全部都是dict，递归处理
*/
func unmarshalList(v reflect.Value, list []*BObject) error {
	if v.Kind() != reflect.Pointer || v.Elem().Type().Kind() != reflect.Slice {
		log.Println("unmarshal list need slice ptr")
		return TypeError
	}
	if len(list) == 0 {
		return nil
	}
	// *[]int -> []int, *[][]int -> [][]int
	e := v.Elem()
	switch list[0].typ {
	case STR:
		for i, obj := range list {
			s, err := obj.Str()
			if err != nil {
				return err
			}
			e.Index(i).SetString(s)
		}
	case INT:
		for i, obj := range list {
			a, err := obj.Int()
			if err != nil {
				return err
			}
			e.Index(i).SetInt(int64(a))
		}
	case LIST:
		for i, obj := range list {
			l, err := obj.List()
			if err != nil {
				return err
			}
			// 获取数组元素的类型
			if e.Type().Elem().Kind() != reflect.Slice {
				return TypeError
			}
			// 创建指向子元素类型的指针
			// 在反射中使用append操作比较麻烦，所以直接开辟一块空间然后让指针指向内存，后续直接在对应索引上set即可
			ptr := reflect.New(e.Type().Elem()) // *[]int
			// 创建子元素类型的数组
			subArray := reflect.MakeSlice(e.Type().Elem(), len(l), len(l))
			ptr.Elem().Set(subArray)
			if err = unmarshalList(ptr, l); err != nil {
				return err
			}
			e.Index(i).Set(ptr.Elem())
		}
	case DICT:
		for i, obj := range list {
			d, err := obj.Dict()
			if err != nil {
				return err
			}
			if e.Type().Elem().Kind() != reflect.Struct {
				return TypeError
			}
			ptr := reflect.New(e.Type().Elem())
			if err = unmarshalDict(ptr, d); err != nil {
				return err
			}
			e.Index(i).Set(ptr.Elem())
		}
	}
	return nil
}

/*
	反序列化列表, 入参（反序列化的值对象, BObject的value）
*/
func unmarshalDict(v reflect.Value, dict map[string]*BObject) error {
	if v.Kind() != reflect.Ptr || v.Elem().Type().Kind() != reflect.Struct {
		return errors.New("unmarshal dict need struct ptr")
	}
	// v是指针，做值修改先转换为Elem
	e := v.Elem()
	for i, n := 0, e.NumField(); i < n; i++ {
		vf := e.Field(i) // 表示v的Field，返回的是Value
		if !vf.CanSet() {
			continue
		}
		tf := e.Type().Field(i)
		key := tf.Tag.Get("bencode") //获取结构体的bencode tag
		if key == "" {
			// 如果没设置tag，就把key设置为字段名
			key = strings.ToLower(tf.Name)
		}
		// 字典中没有，直接跳过
		val := dict[key]
		if val == nil {
			continue
		}
		switch val.typ {
		case STR:
			if vf.Kind() != reflect.String {
				continue
			}
			str, err := val.Str()
			if err != nil {
				return err
			}
			vf.SetString(str)
		case INT:
			if vf.Kind() != reflect.Int {
				continue
			}
			a, err := val.Int()
			if err != nil {
				return err
			}
			vf.SetInt(int64(a))
		case LIST:
			if vf.Kind() != reflect.Slice {
				continue
			}
			list, err := val.List()
			if err != nil {
				return err
			}
			ptr := reflect.New(vf.Type())
			array := reflect.MakeSlice(vf.Type(), len(list), len(list))
			ptr.Elem().Set(array)
			if err = unmarshalList(ptr, list); err != nil {
				continue
			}
			vf.Set(ptr.Elem())
		case DICT:
			if vf.Kind() != reflect.Struct {
				continue
			}
			d, err := val.Dict()
			if err != nil {
				return err
			}
			ptr := reflect.New(vf.Type())
			if err = unmarshalDict(ptr, d); err != nil {
				continue
			}
			vf.Set(ptr.Elem())
		}
	}
	return nil
}
