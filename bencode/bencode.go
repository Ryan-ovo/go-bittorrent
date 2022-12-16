package bencode

import (
	"bufio"
	"errors"
	"io"
)

type BType uint8

var (
	TypeError  = errors.New("wrong type")
	NumError   = errors.New("expect num")
	ColonError = errors.New("expect colon")
	CharIError = errors.New("expect char i")
	CharEError = errors.New("expect char e")
)

const (
	STR BType = iota + 1
	INT
	LIST
	DICT
)

type BValue interface{}

type BObject struct {
	typ BType
	val BValue
}

func (b *BObject) Str() (string, error) {
	if b.typ != STR {
		return "", TypeError
	}
	return b.val.(string), nil
}

func (b *BObject) Int() (int, error) {
	if b.typ != INT {
		return 0, TypeError
	}
	return b.val.(int), nil
}

func (b *BObject) List() ([]*BObject, error) {
	if b.typ != LIST {
		return nil, TypeError
	}
	return b.val.([]*BObject), nil
}

func (b *BObject) Dict() (map[string]*BObject, error) {
	if b.typ != DICT {
		return nil, TypeError
	}
	return b.val.(map[string]*BObject), nil
}

// Bencode 核心函数：往流中写入编码好的BObject，返回编码的字节长度
func (b *BObject) Bencode(w io.Writer) (wLen int) {
	bw, ok := w.(*bufio.Writer)
	if !ok {
		bw = bufio.NewWriter(w)
	}
	switch b.typ {
	case STR:
		str, _ := b.Str()
		wLen += EncodeString(bw, str)
	case INT:
		val, _ := b.Int()
		wLen += EncodeInt(bw, val)
	case LIST:
		// 写入l
		bw.WriteByte('l')
		wLen++
		// 写入列表内容
		list, _ := b.List()
		for _, v := range list {
			// 递归处理列表内容部分
			wLen += v.Bencode(bw)
		}
		// 写入e
		bw.WriteByte('e')
		wLen++
	case DICT:
		// 写入d
		bw.WriteByte('d')
		wLen++
		// 写入列表内容
		dict, _ := b.Dict()
		for k, v := range dict {
			// 字典的key必须是字符串
			wLen += EncodeString(bw, k)
			// 递归处理字典v部分
			wLen += v.Bencode(bw)
		}
		// 写入e
		bw.WriteByte('e')
		wLen++
	}
	bw.Flush()
	return
}

// EncodeString 编码字符串，把编码结果放到流中，返回编码的字节数
func EncodeString(w io.Writer, val string) (wLen int) {
	bw := bufio.NewWriter(w)
	// 写入字符串的字节长度
	wLen += writeInteger(bw, len(val))
	// 写入冒号
	bw.WriteByte(':')
	wLen++
	// 写入字符串内容
	bw.WriteString(val)
	wLen += len(val)
	if err := bw.Flush(); err != nil {
		return 0
	}
	return
}

// DecodeString 解码字符串，从流中读取字节进行解码，返回解码结果
func DecodeString(r io.Reader) (val string, err error) {
	// 新建Reader
	br, ok := r.(*bufio.Reader)
	if !ok {
		br = bufio.NewReader(r)
	}
	// 读取字符串的字节长度
	num, rLen := readInteger(br)
	if rLen == 0 {
		return "", NumError
	}
	// 读取冒号
	b, err := br.ReadByte()
	if b != ':' {
		return "", ColonError
	}
	// 读取字符串内容
	buf := make([]byte, num)
	_, err = io.ReadFull(br, buf)
	val = string(buf)
	return
}

// EncodeInt 编码整数
func EncodeInt(w io.Writer, val int) (wLen int) {
	bw := bufio.NewWriter(w)
	// 写入i
	bw.WriteByte('i')
	wLen++
	// 写入编码后的整数
	wLen += writeInteger(bw, val)
	// 写入e
	bw.WriteByte('e')
	wLen++
	if err := bw.Flush(); err != nil {
		return 0
	}
	return
}

// DecodeInt 解码整数
func DecodeInt(r io.Reader) (val int, err error) {
	br, ok := r.(*bufio.Reader)
	if !ok {
		br = bufio.NewReader(r)
	}
	b, err := br.ReadByte()
	if b != 'i' {
		return 0, CharIError
	}
	val, _ = readInteger(br)
	b, err = br.ReadByte()
	if b != 'e' {
		return 0, CharEError
	}
	return
}

// 将整数转成ASCII码放入流中，返回写入的字节数
func writeInteger(w *bufio.Writer, val int) (wLen int) {
	if val == 0 {
		w.WriteByte('0')
		wLen++
		return
	}
	if val < 0 {
		w.WriteByte('-')
		wLen++
		val = -val
	}
	digits := make([]int, 0)
	for val > 0 {
		digits = append(digits, val%10)
		val /= 10
	}
	//fmt.Println(digits)
	for i := len(digits) - 1; i >= 0; i-- {
		//fmt.Println(byte(digits[i] + '0'))
		w.WriteByte(byte(digits[i] + '0'))
		wLen++
	}
	return
}

// 从流中读取一个数字，返回这个数字和占用的字节数
func readInteger(r *bufio.Reader) (int, int) {
	// 读第一个字节判断是不是负数
	val, rLen := 0, 0
	sign := 1
	b, _ := r.ReadByte()
	rLen++
	if b == '-' {
		sign = -1
		b, _ = r.ReadByte()
		rLen++
	}

	for {
		if !isNum(b) {
			r.UnreadByte()
			rLen--
			return val * sign, rLen
		}
		val = val*10 + int(b-'0')
		b, _ = r.ReadByte()
		rLen++
	}
}

func isNum(b byte) bool {
	return b >= '0' && b <= '9'
}
