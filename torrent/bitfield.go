package torrent

import "strconv"

type Bitfield []byte

func (b Bitfield) HasPiece(index int) bool {
	idx, offset := index/8, index%8
	if idx < 0 || idx >= len(b) {
		return false
	}
	return b[idx]>>uint(7-offset)&1 == 1
}

func (b Bitfield) SetPiece(index int) {
	idx, offset := index/8, index%8
	if idx < 0 || idx >= len(b) {
		return
	}
	b[idx] |= 1 << uint(7-offset)
}

func (b Bitfield) String() string {
	str := "piece# "
	for i := 0; i < len(b)*8; i++ {
		if b.HasPiece(i) {
			str += strconv.Itoa(i) + " "
		}
	}
	return str
}
