package torrent

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"net"
	"os"
	"testing"
)

func TestPeer(t *testing.T) {
	var peer PeerInfo
	peer.IP = net.ParseIP("36.229.97.194")
	peer.Port = uint16(2131)

	file, _ := os.Open("../testfile/debian-iso.torrent")
	tf, _ := ParseFile(bufio.NewReader(file))

	var peerID [IDLen]byte
	_, _ = rand.Read(peerID[:])

	conn, err := NewPeerConn(peer, tf.InfoSHA, peerID)
	if err != nil {
		t.Error("new peer err : " + err.Error())
	}
	fmt.Println(conn)
}
