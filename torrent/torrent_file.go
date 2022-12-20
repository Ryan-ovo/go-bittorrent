package torrent

import (
	"bytes"
	"crypto/sha1"
	"github.com/Ryan-ovo/go-bittorrent/bencode"
	"io"
	"log"
)

const SHALEN int = 20

type rawInfo struct {
	Length      int    `bencode:"length"`
	Name        string `bencode:"name"`
	PieceLength int    `bencode:"piece length"`
	Pieces      string `bencode:"pieces"`
}

type rawFile struct {
	Announce string  `bencode:"announce"`
	Info     rawInfo `bencode:"info"`
}

type TorrentFile struct {
	Announce string
	InfoSHA  [SHALEN]byte
	FileName string
	FileLen  int
	PieceLen int
	PieceSHA [][SHALEN]byte
}

func ParseFile(r io.Reader) (*TorrentFile, error) {
	raw := new(rawFile)
	// 将流中的数据反序列化到rawFile结构中
	if err := bencode.Unmarshal(r, raw); err != nil {
		log.Printf("Parse file error, err = [%v]", err)
		return nil, err
	}
	tf := new(TorrentFile)
	tf.Announce = raw.Announce
	tf.FileName = raw.Info.Name
	tf.FileLen = raw.Info.Length
	tf.PieceLen = raw.Info.PieceLength

	buf := bytes.NewBuffer([]byte{})
	// 单独把raw的核心数据序列化到流中
	wLen := bencode.Marshal(buf, raw.Info)
	if wLen == 0 {
		log.Println("marshal raw file info to stream error")
	}
	// 求整个文件的sha1哈希值
	tf.InfoSHA = sha1.Sum(buf.Bytes())
	// 求每个分片的哈希值
	bs := []byte(raw.Info.Pieces)
	hash := make([][SHALEN]byte, len(bs)/SHALEN)
	for i := 0; i < len(bs)/SHALEN; i++ {
		copy(hash[i][:], bs[i*SHALEN:(i+1)*SHALEN])
	}
	tf.PieceSHA = hash
	return nil, nil
}
