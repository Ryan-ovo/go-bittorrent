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
		return nil, err
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
// Peer约定消息格式：前4字节是消息长度，后面1字节是消息id，再往后是消息内容
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
	conn.SetDeadline(time.Now().Add(3 * time.Second))
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

// GetIndex 获取消息中的信息：分片序号
func GetIndex(msg *PeerMsg) (int, error) {
	if msg.ID != MsgHave {
		return 0, fmt.Errorf("expect msg id have, get %d", msg.ID)
	}
	if len(msg.Payload) != 4 {
		return 0, fmt.Errorf("expect payload length 4, get %d", len(msg.Payload))
	}
	index := binary.BigEndian.Uint32(msg.Payload)
	return int(index), nil
}

// CopyPieceData 把通信消息中对应分片的子分片内容拷贝到内存buf中
func CopyPieceData(index int, buf []byte, msg *PeerMsg) (int, error) {
	if msg.ID != MsgPiece {
		return 0, fmt.Errorf("expect msg id piece, get %d", msg.ID)
	}
	if len(msg.Payload) < 8 {
		return 0, fmt.Errorf("payload too short, expect 8, get %d", len(msg.Payload))
	}
	parseIndex := int(binary.BigEndian.Uint32(msg.Payload[0:4]))
	if parseIndex != index {
		return 0, fmt.Errorf("expect index %d, get %d", index, parseIndex)
	}
	parseOffset := int(binary.BigEndian.Uint32(msg.Payload[4:8]))
	if parseOffset >= len(buf) {
		return 0, fmt.Errorf("offset too big, offset %d >= bufLen %d", parseOffset, len(buf))
	}
	parseData := msg.Payload[8:]
	if parseOffset+len(parseData) >= len(buf) {
		return 0, fmt.Errorf("data too big, offset %d, dataLen %d, bufLen %d", parseOffset, len(parseData), len(buf))
	}
	// 拷贝消息内容到对应位置
	copy(buf[parseOffset:], parseData)
	return len(parseData), nil
}
