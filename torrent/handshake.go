package torrent

import "io"

const (
	Reserved = 8
	HsMsgLen = Reserved + SHALEN + IDLen
)

// HandShakeMsg 握手消息格式：协议长度 + 协议名 + SHA-1哈希值 + peer_id
type HandShakeMsg struct {
	PreStr  string
	InfoSHA [SHALEN]byte
	PeerID  [IDLen]byte
}

func NewHandShakeMsg(infoSHA [SHALEN]byte, peerID [IDLen]byte) *HandShakeMsg {
	return &HandShakeMsg{
		PreStr:  "BitTorrent Protocol",
		InfoSHA: infoSHA,
		PeerID:  peerID,
	}
}

func WriteHandShake(w io.Writer, msg *HandShakeMsg) (int, error) {
	buf := make([]byte, 1+len(msg.PreStr)+HsMsgLen)
	buf[0] = byte(len(msg.PreStr))
	wLen := 1
	wLen += copy(buf[wLen:], msg.PreStr)
	wLen += copy(buf[wLen:], make([]byte, Reserved))
	wLen += copy(buf[wLen:], msg.InfoSHA[:])
	wLen += copy(buf[wLen:], msg.PeerID[:])
	return w.Write(buf)
}

func ReadHandShake(r io.Reader) (*HandShakeMsg, error) {
	// 读协议长度
	lenBuf := make([]byte, 1)
	if _, err := io.ReadFull(r, lenBuf); err != nil {
		return nil, err
	}
	preLen := int(lenBuf[0])
	// 读消息体
	msgBuf := make([]byte, preLen+HsMsgLen)
	if _, err := io.ReadFull(r, msgBuf); err != nil {
		return nil, err
	}
	var infoSHA [SHALEN]byte
	var peerID [IDLen]byte

	// 拷贝哈希值
	copy(infoSHA[:], msgBuf[preLen+Reserved:preLen+Reserved+SHALEN])
	// 拷贝peer id
	copy(peerID[:], msgBuf[preLen+Reserved+SHALEN:])

	// 封装消息返回
	return &HandShakeMsg{
		PreStr:  string(msgBuf[0:preLen]),
		InfoSHA: infoSHA,
		PeerID:  peerID,
	}, nil
}
