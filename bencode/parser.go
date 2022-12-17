package bencode

import (
	"bufio"
	"io"
)

// Parse 解析流中的Bencode编码为BObject
func Parse(r io.Reader) (*BObject, error) {
	br, ok := r.(*bufio.Reader)
	if !ok {
		br = bufio.NewReader(r)
	}
	// 查看流中的第一个字节，但不读取
	b, err := br.Peek(1)
	if err != nil {
		return nil, err
	}
	obj := &BObject{}
	if b[0] >= '0' && b[0] <= '9' { // string
		val, err := DecodeString(br)
		if err != nil {
			return nil, err
		}
		obj.typ = STR
		obj.val = val
	} else if b[0] == 'i' { // int
		val, err := DecodeInt(br)
		if err != nil {
			return nil, err
		}
		obj.typ = INT
		obj.val = val
	} else if b[0] == 'l' { // list
		br.ReadByte()
		objs := make([]*BObject, 0)
		for {
			// 如果读取到e，直接退出
			if a, _ := br.Peek(1); a[0] == 'e' {
				br.ReadByte()
				break
			}
			elem, err := Parse(br)
			if err != nil {
				return nil, err
			}
			objs = append(objs, elem)
		}
		obj.typ = LIST
		obj.val = objs
	} else if b[0] == 'd' { // dict
		br.ReadByte()
		objs := make(map[string]*BObject)
		for {
			if a, _ := br.Peek(1); a[0] == 'e' {
				br.ReadByte()
				break
			}
			key, err := DecodeString(br)
			if err != nil {
				return nil, err
			}
			elem, err := Parse(br)
			if err != nil {
				return nil, err
			}
			objs[key] = elem
		}
		obj.typ = DICT
		obj.val = objs
	} else {
		return nil, TypeError
	}
	// 递归解析
	return obj, nil
}
