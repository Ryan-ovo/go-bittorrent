### Background
#### Technical Lingo
* CS架构：传统client-server架构，客户端请求服务器获取资源，缺点是服务器如果故障，所有客户端都无法下载；服务器带宽有限，客户端多了速度就会下降。
* P2P架构：peer-to-peer架构，是一种在对等者（Peer）之间分配任务和工作负载的分布式应用架构，是对等计算模型在应用层形成的一种网络形式。在这种架构下，每个机器既是客户端也是服务器，从别人那里获取资源的同时，也提供资源给别人。
* BT：BitTorrent，也称比特洪流，是一种基于P2P的通信协议，是一个应用层协议
* Seed：指的是下载下来的.torrent文件，这个二进制文件采用了bencode编码方式进行编码
* Seeder：制作并发布种子的用户
* Peer：对等用户，每个p2p通信的用户都是peer
* Leecher：下载种子的用户
* Tracker：Tracker服务器，作为peer之间沟通的桥梁，Tracker本身不保存任何资源的信息，但可以追踪文件副本在其他peer上的位置，帮助peer之间进行连接和文件重组
* Piece：文件被分片，以离散的形式保存在每个peer中，并且拥有每个片的哈希值可以进行校验
* Bit Field：位域，用来标记每个peer中对于某个文件分片的拥有情况

#### Goal
* 本项目基于Go实现了一个bt下载器客户端
* 使用bencode编码实现torrent文件的序列化和反序列化
* 可以与Tracker进行交互，获取拥有文件分片的对等客户端
* 作为peer在对等网络模型中实现种子文件的并发下载

### Usage
```
cd ./cmd
go run main.go ../testfile/debian-iso.torrent
```