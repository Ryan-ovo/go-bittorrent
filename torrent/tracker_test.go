package torrent

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"os"
	"testing"
)

func TestFindPeers(t *testing.T) {
	file, _ := os.Open("../testfile/debian-iso.torrent")
	tf, _ := ParseFile(bufio.NewReader(file))

	var peerID [IDLen]byte
	_, _ = rand.Read(peerID[:])

	peers := FindPeers(tf, peerID)
	//t.Log(peers)
	for i, p := range peers {
		fmt.Printf("Peer %d, Ip: %s, Port: %d\n", i, p.IP, p.Port)
	}
}
