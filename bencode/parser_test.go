package bencode

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func assertString(t *testing.T, str string, obj *BObject) {
	assert.Equal(t, STR, obj.typ)
	val, err := obj.Str()
	assert.Equal(t, nil, err)
	assert.Equal(t, str, val)
}

func assertInt(t *testing.T, val int, obj *BObject) {
	assert.Equal(t, INT, obj.typ)
	a, err := obj.Int()
	assert.Equal(t, nil, err)
	assert.Equal(t, val, a)
}

func TestParseString(t *testing.T) {
	in := bytes.NewBufferString("3:abc")
	obj, _ := Parse(in)
	// 校验解析结果
	assertString(t, "abc", obj)

	out := bytes.NewBufferString("")
	// 校验反向编码结果
	assert.Equal(t, len("3:abc"), obj.Bencode(out))
	assert.Equal(t, "3:abc", out.String())
}

func TestParseInt(t *testing.T) {
	in := bytes.NewBufferString("i123e")
	obj, _ := Parse(in)
	// 校验解析结果
	assertInt(t, 123, obj)

	out := bytes.NewBufferString("")
	// 校验反向编码结果
	assert.Equal(t, len("i123e"), obj.Bencode(out))
	assert.Equal(t, "i123e", out.String())
}

func TestParseList(t *testing.T) {
	// 校验解析
	// [123, "Ryan", 789]
	code := "li123e4:Ryani789ee"
	in := bytes.NewBufferString(code)
	obj, _ := Parse(in)
	assert.Equal(t, LIST, obj.typ)
	list, err := obj.List()
	assert.Equal(t, nil, err)
	assert.Equal(t, len(list), 3)
	assertInt(t, 123, list[0])
	assertString(t, "Ryan", list[1])
	assertInt(t, 789, list[2])

	// 校验反编码
	out := bytes.NewBufferString("")
	assert.Equal(t, len(code), obj.Bencode(out))
	assert.Equal(t, code, out.String())
}

func TestParseDict(t *testing.T) {
	// {name: Ryan; age: 20}
	code := "d4:name4:Ryan3:agei20ee"
	//code := "d4:name6:archer3:agei29ee"
	in := bytes.NewBufferString(code)
	obj, _ := Parse(in)

	assert.Equal(t, DICT, obj.typ)
	mp, err := obj.Dict()
	t.Log(mp["name"], mp["age"])
	assert.Equal(t, nil, err)
	assertString(t, "Ryan", mp["name"])
	assertInt(t, 20, mp["age"])

	// 校验反编码
	out := bytes.NewBufferString("")
	assert.Equal(t, len(code), obj.Bencode(out))
	// 编码后key的顺序可能与原本不一致
	//assert.Equal(t, code, out.String())
}

func TestParseMultiDict(t *testing.T) {
	// {"user": {name:"Ryan"; age:20}; "hobby": [123, "abc", 789]}
	code := "d4:userd4:name4:Ryan3:agei20ee5:hobbyli123e3:abci789eee"
	in := bytes.NewBufferString(code)
	obj, _ := Parse(in)
	assert.Equal(t, DICT, obj.typ)
	mp, err := obj.Dict()
	assert.Equal(t, nil, err)
	assert.Equal(t, DICT, mp["user"].typ)
	assert.Equal(t, LIST, mp["hobby"].typ)
}
