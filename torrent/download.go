package torrent

import (
	"bytes"
	"crypto/sha1"
	"log"
	"os"
	"time"
)

const (
	BLOCKSIZE  = 1024 * 16 // block = sub-piece
	MAXBACKLOG = 5         // 一个peer协程最多同时发送5个请求
)

// TorrentTask 下载任务的抽象
type TorrentTask struct {
	peerID   [IDLen]byte    // 客户端id
	PeerList []PeerInfo     // 获取到的peer列表
	InfoSHA  [SHALEN]byte   // 文件哈希值
	FileName string         // 文件名
	FileLen  int            // 文件长度
	PieceLen int            // 分片长度
	PieceSHA [][SHALEN]byte // 所有分片哈希值
}

func (t *TorrentTask) getPieceBounds(index int) (int, int) {
	begin := t.PieceLen * index
	end := begin + t.PieceLen
	if end > t.FileLen {
		end = t.FileLen
	}
	return begin, end
}

func (t *TorrentTask) peerRoutine(peer PeerInfo, taskQueue chan *pieceTask, resultQueue chan *pieceResult) {
	// 建立peer的连接
	conn, err := NewPeerConn(peer, t.InfoSHA, t.peerID)
	if err != nil {
		log.Println("connect to peer error = ", err)
		return
	}
	defer conn.Close()
	log.Printf("complete handshake with peer, ip = [%s], port = [%d]", conn.peer.IP.String(), conn.peer.Port)
	// 给peer发送interested消息表示想要下载
	if _, err = conn.WriteMsg(&PeerMsg{MsgInterested, nil}); err != nil {
		log.Println("write msg to conn error = ", err)
		return
	}
	// 从任务队列获取分片进行下载
	for task := range taskQueue {
		// 如果这个peer没有我们想要的分片，就把任务重新放回队列中
		if !conn.Field.HasPiece(task.index) {
			taskQueue <- task
			continue
		}
		log.Printf("get task, index = [%d], peer = [%s]\n", task.index, peer.IP.String())
		res, err := downloadPiece(conn, task)
		if err != nil {
			// 下载失败，把任务重新放到队列中，让其他peer下载这些任务
			taskQueue <- task
			log.Printf("download piece error = [%v]\n", err)
			return
		}
		if !checkPiece(task, res) {
			taskQueue <- task
			continue
		}
		// 校验哈希值通过，把分片下载结果发送到通道中
		resultQueue <- res
	}
}

// 分片下载任务的抽象
type pieceTask struct {
	index  int          // 分片序号
	sha    [SHALEN]byte // 分片哈希值
	length int          // 分片长度，最后一片可能不均等
}

// 下载中间态的抽象
type taskState struct {
	index      int       // 分片序号
	conn       *PeerConn // peer之间的连接
	requested  int       // 还剩多少字节没有请求
	downloaded int       // 已经下载的字节数
	backlog    int       // 并发度
	data       []byte
}

func (ts *taskState) handleMsg() error {
	msg, err := ts.conn.ReadMsg()
	if err != nil {
		return err
	}
	// 空消息是探活消息
	if msg == nil {
		return nil
	}

	switch msg.ID {
	case MsgChoke: // 对方拒绝上传数据，默认状态
		ts.conn.Choked = true
	case MsgUnchoke: // 对方上传数据
		ts.conn.Choked = false
	case MsgHave: // 通知拥有某个分片
		index, err := GetIndex(msg)
		if err != nil {
			return err
		}
		ts.conn.Field.SetPiece(index)
	case MsgPiece: // 数据消息
		n, err := CopyPieceData(ts.index, ts.data, msg)
		if err != nil {
			return err
		}
		ts.downloaded += n
		ts.backlog--
	}
	return nil
}

// 分片下载结果
type pieceResult struct {
	index int    // 分片序号
	data  []byte // 下载的内容
}

func Download(task *TorrentTask) error {
	log.Println("start downloading ", task.FileName)
	// 初始化通道：分片任务通道，下载结果通道
	taskQueue := make(chan *pieceTask, len(task.PieceSHA))
	resultQueue := make(chan *pieceResult)

	for index, sha := range task.PieceSHA {
		begin, end := task.getPieceBounds(index)
		taskQueue <- &pieceTask{
			index:  index,
			sha:    sha,
			length: end - begin,
		}
	}
	// 每个peer开一个协程处理
	for _, peer := range task.PeerList {
		go task.peerRoutine(peer, taskQueue, resultQueue)
	}
	buf := make([]byte, task.FileLen)
	cnt := 0
	for cnt < len(task.PieceSHA) {
		res := <-resultQueue
		begin, end := task.getPieceBounds(res.index)
		copy(buf[begin:end], res.data)
		cnt++
		// 打印进度条日志
		ratio := float64(cnt) / float64(len(task.PieceSHA)) * 100
		log.Printf("downloading, progress = (%0.2f%%)\n", ratio)
	}
	// 关闭通道
	close(taskQueue)
	close(resultQueue)
	// 把数据从内存写入文件
	file, err := os.Create(task.FileName)
	if err != nil {
		log.Println("create file error = ", err)
		return err
	}
	if _, err = file.Write(buf); err != nil {
		log.Println("write to file error = ", err)
		return err
	}
	return nil
}

// 下载单个分片
func downloadPiece(conn *PeerConn, task *pieceTask) (*pieceResult, error) {
	state := &taskState{
		index: task.index,
		conn:  conn,
		data:  make([]byte, task.length),
	}
	conn.SetDeadline(time.Now().Add(15 * time.Second))
	defer conn.SetDeadline(time.Time{})
	// 等全部下载完成再退出
	for state.downloaded < task.length {
		// 轮询判断如果对面发送了unchoked消息，就能继续下载
		if !conn.Choked {
			// 最多同时发起5次请求
			for state.backlog < MAXBACKLOG && state.requested < task.length {
				// 默认一个sub piece是16k，如果最后一个子分片不足16k，就手动计算一下最后一片的大小
				length := BLOCKSIZE
				if task.length-state.requested < BLOCKSIZE {
					length = task.length - state.requested
				}
				// 封装请求体并发送
				msg := NewRequestMsg(state.index, state.requested, length)
				if _, err := conn.WriteMsg(msg); err != nil {
					return nil, err
				}
				state.requested += length
				state.backlog++
			}
		}
		// 分类处理各类消息，包括探活，拒绝上传，通知拥有分片，下载分片等
		if err := state.handleMsg(); err != nil {
			return nil, err
		}
	}
	return &pieceResult{state.index, state.data}, nil
}

func checkPiece(task *pieceTask, res *pieceResult) bool {
	sha := sha1.Sum(res.data)
	if !bytes.Equal(task.sha[:], sha[:]) {
		log.Printf("check sha1 sum error, index = [%d]\n", task.index)
		return false
	}
	return true
}
