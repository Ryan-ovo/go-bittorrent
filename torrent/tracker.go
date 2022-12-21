package torrent

import (
	"encoding/binary"
	"github.com/Ryan-ovo/go-bittorrent/bencode"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	PeerPort = 6666
	IpLen    = 4
	PortLen  = 2
	PeerLen  = IpLen + PortLen
	IDLen    = 20
)

type PeerInfo struct {
	IP   net.IP
	Port uint16
}

type TrackerResp struct {
	Interval int    `bencode:"interval"`
	Peers    string `bencode:"peers"`
}

// 利用net/url库封装带参数的Get请求
func buildURL(tf *TorrentFile, peerID [IDLen]byte) (string, error) {
	base, err := url.Parse(tf.Announce)
	if err != nil {
		log.Println("Announce error", err)
		return "", err
	}
	params := url.Values{
		"info_hash":  []string{string(tf.InfoSHA[:])},
		"peer_id":    []string{string(peerID[:])},
		"port":       []string{strconv.Itoa(PeerPort)},
		"left":       []string{strconv.Itoa(tf.FileLen)},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"compact":    []string{"1"},
	}
	base.RawQuery = params.Encode()
	return base.String(), nil
}

// 将tracker返回的结果解析成(ip, port)列表
func buildPeerInfo(peers []byte) []PeerInfo {
	// 数据格式错误
	if len(peers)%PeerLen != 0 {
		log.Println("irregular peer data")
		return nil
	}
	ps := make([]PeerInfo, len(peers)/PeerLen)
	for i := 0; i < len(peers)/PeerLen; i++ {
		offset := i * PeerLen
		ps[i].IP = peers[offset : offset+IpLen]
		ps[i].Port = binary.BigEndian.Uint16(peers[offset+IpLen : offset+PeerLen])
	}
	return ps
}

func FindPeers(tf *TorrentFile, peerID [IDLen]byte) []PeerInfo {
	url, err := buildURL(tf, peerID)
	if err != nil {
		log.Println("build tracker url error = ", err)
		return nil
	}

	// 发送http请求
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		log.Println("connect to tracker error = ", err)
		return nil
	}
	defer resp.Body.Close()
	trackerResp := new(TrackerResp)
	// 将返回的结果反序列化到TrackerResp中
	if err = bencode.Unmarshal(resp.Body, trackerResp); err != nil {
		log.Println("unmarshal tracker resp error = ", err)
		return nil
	}
	// 解析出ip, port列表
	return buildPeerInfo([]byte(trackerResp.Peers))
}
