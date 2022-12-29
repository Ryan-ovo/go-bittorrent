package main

import (
	"crypto/rand"
	"github.com/Ryan-ovo/go-bittorrent/torrent"
	"log"
	"os"
)

func main() {
	// 1. 打开种子文件
	file, err := os.Open(os.Args[1])
	if err != nil {
		log.Println("open file error = ", err)
		return
	}
	defer file.Close()
	// 2. 解析种子文件
	tf, err := torrent.ParseFile(file)
	if err != nil {
		log.Println("parse file error = ", err)
		return
	}
	// 3. 生成客户端的peer id
	var peerID [torrent.IDLen]byte
	_, _ = rand.Read(peerID[:])
	// 4. 连接tracker，获取peer的信息
	peers := torrent.FindPeers(tf, peerID)
	if len(peers) == 0 {
		log.Println("peers not found")
		return
	}
	// 5. 封装下载对象
	task := &torrent.TorrentTask{
		PeerID:   peerID,
		PeerList: peers,
		InfoSHA:  tf.InfoSHA,
		FileName: tf.FileName,
		FileLen:  tf.FileLen,
		PieceLen: tf.PieceLen,
		PieceSHA: tf.PieceSHA,
	}
	torrent.Download(task)
}
