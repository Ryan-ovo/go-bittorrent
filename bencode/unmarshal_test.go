package bencode

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

type User struct {
	Name string `bencode:"name"`
	Age  int    `bencode:"age"`
}

type Role struct {
	Id   int
	User `bencode:"user"`
}

type Score struct {
	User  `bencode:"user"`
	Value []int `bencode:"value"`
}

type Team struct {
	Name   string `bencode:"name"`
	Size   int    `bencode:"size"`
	Member []User `bencode:"member"`
}

func TestUnmarshalList(t *testing.T) {
	str := "li1ei2ei3ee"
	l := &[]int{}
	Unmarshal(bytes.NewBufferString(str), l)
	assert.Equal(t, []int{1, 2, 3}, *l)

	buf := bytes.NewBuffer([]byte{})
	wLen := Marshal(buf, l)
	assert.Equal(t, len(str), wLen)
	assert.Equal(t, str, buf.String())
}

func TestUnmarshalMultiList(t *testing.T) {
	str := "lli1ei2ei3eeli4ei5ei6eee"
	l := &[][]int{}
	Unmarshal(bytes.NewBufferString(str), l)
	assert.Equal(t, [][]int{{1, 2, 3}, {4, 5, 6}}, *l)

	buf := bytes.NewBuffer([]byte{})
	wLen := Marshal(buf, l)
	assert.Equal(t, len(str), wLen)
	assert.Equal(t, str, buf.String())
}

func TestUnmarshalUser(t *testing.T) {
	str := "d4:name4:Ryan3:agei20ee"
	u := &User{}
	Unmarshal(bytes.NewBufferString(str), u)
	t.Logf("%+v", u)
	assert.Equal(t, "Ryan", u.Name)
	assert.Equal(t, 20, u.Age)

	buf := bytes.NewBuffer([]byte{})
	wLen := Marshal(buf, u)
	assert.Equal(t, len(str), wLen)
	assert.Equal(t, str, buf.String())
}

func TestUnmarshalRole(t *testing.T) {
	str := "d2:idi1e4:userd4:name4:Ryan3:agei20eee"
	r := &Role{}
	Unmarshal(bytes.NewBufferString(str), r)
	assert.Equal(t, 1, r.Id)
	assert.Equal(t, "Ryan", r.Name)
	assert.Equal(t, 20, r.Age)

	buf := bytes.NewBuffer([]byte{})
	wLen := Marshal(buf, r)
	assert.Equal(t, len(str), wLen)
	assert.Equal(t, str, buf.String())
}

func TestUnmarshalScore(t *testing.T) {
	str := "d4:userd4:name4:Ryan3:agei20ee5:valueli1ei2ei3eee"
	s := &Score{}
	Unmarshal(bytes.NewBufferString(str), s)
	assert.Equal(t, "Ryan", s.Name)
	assert.Equal(t, 20, s.Age)
	assert.Equal(t, []int{1, 2, 3}, s.Value)

	buf := bytes.NewBuffer([]byte{})
	wLen := Marshal(buf, s)
	assert.Equal(t, len(str), wLen)
	assert.Equal(t, str, buf.String())
}

func TestUnmarshalTeam(t *testing.T) {
	str := "d4:name6:team014:sizei2e6:memberld4:name4:Ryan3:agei20eed4:name5:nancy3:agei31eeee"
	team := &Team{}
	Unmarshal(bytes.NewBufferString(str), team)
	assert.Equal(t, "team01", team.Name)
	assert.Equal(t, 2, team.Size)

	buf := bytes.NewBuffer([]byte{})
	wLen := Marshal(buf, team)
	assert.Equal(t, len(str), wLen)
	assert.Equal(t, str, buf.String())
}

func TestMarshalBasic(t *testing.T) {
	buf := new(bytes.Buffer)
	str := "abc"
	len := Marshal(buf, str)
	assert.Equal(t, 5, len)
	assert.Equal(t, "3:abc", buf.String())

	buf.Reset()
	val := 199
	len = Marshal(buf, val)
	assert.Equal(t, 5, len)
	assert.Equal(t, "i199e", buf.String())
}
