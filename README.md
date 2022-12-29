### Background
#### Technical Lingo
* CS：传统client-server架构，客户端请求服务器获取资源，缺点是服务器如果故障，所有客户端都无法下载；服务器带宽有限，客户端多了速度就会下降。
* P2P：peer-to-peer架构，是一种在对等者（Peer）之间分配任务和工作负载的分布式应用架构，是对等计算模型在应用层形成的一种网络形式。在这种架构下，每个机器既是客户端也是服务器，从别人那里获取资源的同时，也提供资源给别人。
* BT：BitTorrent，也称比特洪流，是一种基于P2P的通信协议，是一个应用层协议
* 种子(seed)：指的是下载下来的.torrent文件，这个二进制文件采用了bencode编码方式进行编码

#### Goal

### Usage
```
cd ./cmd
go run main.go ../testfile/debian-iso.torrent
```