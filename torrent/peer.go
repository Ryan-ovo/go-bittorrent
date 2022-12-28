package torrent

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"time"
)

type MsgID uint8

const (
	MsgChoke       MsgID = 0
	MsgUnchoke     MsgID = 1
	MsgInterested  MsgID = 2
	MsgNotInterest MsgID = 3
	MsgHave        MsgID = 4
	MsgBitfield    MsgID = 5
	MsgRequest     MsgID = 6
	MsgPiece       MsgID = 7
	MsgCancel      MsgID = 8
)

const LenByte = 4

type PeerMsg struct {
	ID      MsgID
	Payload []byte
}

func NewRequestMsg(index, offset, length int) *PeerMsg {
	payload := make([]byte, 12)
	binary.BigEndian.PutUint32(payload[0:4], uint32(index))
	binary.BigEndian.PutUint32(payload[4:8], uint32(offset))
	binary.BigEndian.PutUint32(payload[8:12], uint32(length))
	return &PeerMsg{MsgRequest, payload}
}

type PeerConn struct {
	net.Conn
	Choked  bool
	Field   Bitfield
	peer    PeerInfo
	peerID  [IDLen]byte
	infoSHA [SHALEN]byte
}

func NewPeerConn(peer PeerInfo, infoSHA [SHALEN]byte, peerID [SHALEN]byte) (*PeerConn, error) {
	addr := net.JoinHostPort(peer.IP.String(), strconv.Itoa(int(peer.Port)))
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		log.Println("establish conn error = ", err)
	}
	// 建立p2p连接
	if err = handshake(conn, infoSHA, peerID); err != nil {
		conn.Close()
		log.Println("handshake error = ", err)
		return nil, err
	}

	pc := &PeerConn{
		Conn:    conn,
		Choked:  true,
		peer:    peer,
		peerID:  peerID,
		infoSHA: infoSHA,
	}
	if err = fillBitField(pc); err != nil {
		log.Println("fill bit field error = ", err)
	}
	return pc, nil
}

func (c *PeerConn) ReadMsg() (*PeerMsg, error) {
	lenBuf := make([]byte, LenByte)
	if _, err := io.ReadFull(c, lenBuf); err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint32(lenBuf)
	if length == 0 {
		return nil, nil
	}
	msgBuf := make([]byte, length)
	if _, err := io.ReadFull(c, msgBuf); err != nil {
		return nil, err
	}

	return &PeerMsg{
		ID:      MsgID(msgBuf[0]),
		Payload: msgBuf[1:],
	}, nil
}

// WriteMsg 写入格式：消息长度（id + payload） + id + payload
func (c *PeerConn) WriteMsg(msg *PeerMsg) (int, error) {
	var buf []byte
	if msg == nil {
		buf = make([]byte, LenByte)
	}
	length := 1 + len(msg.Payload)
	buf = make([]byte, LenByte+length)
	binary.BigEndian.PutUint32(buf[0:LenByte], uint32(length))
	buf[LenByte] = byte(msg.ID)
	copy(buf[LenByte+1:], msg.Payload)
	return c.Write(buf)
}

func handshake(conn net.Conn, infoSHA [SHALEN]byte, peerId [IDLen]byte) error {
	conn.SetDeadline(time.Now().Add(15 * time.Second))
	defer conn.SetDeadline(time.Time{})
	// send HandshakeMsg
	req := NewHandShakeMsg(infoSHA, peerId)
	_, err := WriteHandShake(conn, req)
	if err != nil {
		fmt.Println("send handshake failed")
		return err
	}
	// read HandshakeMsg
	res, err := ReadHandShake(conn)
	if err != nil {
		fmt.Println("read handshake failed")
		return err
	}
	// check HandshakeMsg
	if !bytes.Equal(res.InfoSHA[:], infoSHA[:]) {
		fmt.Println("check handshake failed")
		return fmt.Errorf("handshake msg error: " + string(res.InfoSHA[:]))
	}
	return nil
}

func fillBitField(c *PeerConn) error {
	c.SetDeadline(time.Now().Add(5 * time.Second))
	defer c.SetDeadline(time.Time{})

	msg, err := c.ReadMsg()
	if err != nil {
		return err
	}
	if msg == nil {
		return fmt.Errorf("expect bitfield")
	}
	if msg.ID != MsgBitfield {
		return fmt.Errorf("expect bitfield, get %d", msg.ID)
	}
	log.Println("fill bitfield successfully, peer = ", c.peer.IP.String())
	c.Field = msg.Payload
	return nil
}
