package bencode

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestString(t *testing.T) {
	str := "abcdefghijklmnopqrstuvwxyz"
	buf := new(bytes.Buffer)
	wLen := EncodeString(buf, str)
	decodeStr, _ := DecodeString(buf)
	assert.Equal(t, wLen, 29)
	assert.Equal(t, decodeStr, str+" ")
}

func TestInt(t *testing.T) {
	// 正数
	val := 123
	buf := new(bytes.Buffer)
	wLen := EncodeInt(buf, val)
	decodeVal, _ := DecodeInt(buf)
	t.Log(wLen, decodeVal)
	assert.Equal(t, wLen, 5)
	assert.Equal(t, decodeVal, val)

	// 零
	val = 0
	buf.Reset()
	wLen = EncodeInt(buf, val)
	decodeVal, _ = DecodeInt(buf)
	t.Log(wLen, decodeVal)
	assert.Equal(t, wLen, 3)
	assert.Equal(t, decodeVal, val)

	// 负数
	val = -123
	buf.Reset()
	wLen = EncodeInt(buf, val)
	decodeVal, _ = DecodeInt(buf)
	t.Log(wLen, decodeVal)
	assert.Equal(t, wLen, 6)
	assert.Equal(t, decodeVal, val)
}
