### 相关术语
* CS：传统client-server架构，客户端请求服务器获取资源，缺点是服务器如果故障，所有客户端都无法下载；服务器带宽有限，客户端多了速度就会下降。
* P2P：peer-to-peer架构，是一种在对等者（Peer）之间分配任务和工作负载的分布式应用架构，是对等计算模型在应用层形成的一种网络形式。在这种架构下，每个机器既是客户端也是服务器，从别人那里获取资源的同时，也提供资源给别人。
* BT：BitTorrent，也称比特洪流，是一种基于P2P的通信协议，是一个应用层协议
* 种子：指的是下载下来的.torrent文件，这个二进制文件采用了bencode编码方式进行编码
* Tracker：
### Bencode编码
#### String
将一个字符串的前面加上长度标识和符号（冒号），这就是 Bencode 编码后的字符串了，如：
`"Hello"`编码后为`5:Hello`
`How are you`编码后为：`11:How are you`
#### Int
一个整数起始以 i 作为标识，结尾以 e 来作为标识，把数字写在中间即可，如：
`123`编码后为`i123e`
`666`编码后为`i666e`
`0`编码后为`i0e`
#### List
列表可以类比为 Python 中的列表，是一种容器性质的数据结构，每个元素可以是四种数据结构中的任意一组，没有长度限制。语法是，列表的开头和结尾分别用 `l` 和 `e` 作为标识符，中间的值就是任意的数据结构。
`[123,666,0]`编码后为`li123ei666ei0ee`
`[123,'hello',456]`编码后为`li123e5:helloi456ee`
#### Dict
字典的开头和结尾以 `d` 和 `e` 作为标识符，bencode中的字典，key 要求必须是字符串格式的，value 的格式可以随便。另外，编码过程，key 要**根据字符串的字典序进行升序排序**。比如：
`{'a':1,'cd':[3,4],'b':2}`编码后为`d1:ai1e1:bi2e2:cdli3ei4eee`
### .Torrent种子格式

- announce：Tracker主服务器的url
- announce-list：备用tracker的url，可选
- comment：备注信息
- created by：创建的工具签名
- creation date：种子创建时间
- info_hash：整个文件的哈希值，采用SHA-1算法
- info：
    - length：文件长度，单位为字节
    - name：文件名
    - piece length：每块的哈希值长度
    - pieces：分成每个块的哈希值大小
    - files：包含的文件信息
```json
{
    "announce": "https://torrent.ubuntu.com/announce",
    "announce-list": [
        [
            "https://torrent.ubuntu.com/announce"
        ],
        [
            "https://ipv6.torrent.ubuntu.com/announce"
        ]
    ],
    "comment": "Ubuntu CD releases.ubuntu.com",
    "created by": "mktorrent 1.1",
    "creation date": "2021-02-12 03:02:32",
    "info_hash": "4ba4fbf7231a3a660e86892707d25c135533a16a",
    "info": {
        "length": 2877227008,
        "name": "ubuntu-20.04.2.0-desktop-amd64.iso",
        "piece length": 262144,
        "pieces": [
            "d89b853053ac28e09d6d322658636d9663aa80fe",
            "287528aae8bda9ef962918ba8db2ceb0638454e4",
            "149987b3a98147d9b5cc1e249b2fea7dc3401eb1",
            "539f5c519a5fcb058d5978b415188340f57039df",
            "c5ac6a46748abef691e96f7913c60c22990d5123",
            "e87e684ca1c31cc029560514058c75c306a6b41c",
            "c19e41f1c980b91ff735af99a2c4ab4d90946344",
            "4707444be592ae107ddd614a3ef79fbc21e090a3",
            "3acce815ec86a6d5bc0677874ab98dba424ddf35",
            "d4e0d04c15514509c14fa97b1eb09f3bdbaff144",
            "f03a8f9c698568221b4582995716b1123b7e7390",
            "3efe825e140ab8137525f2ecaa0b32d46ec62851",
            "数量太多，这里截断，一共10976行 ......."
        ]
    }
}
```
### BT下载流程

### Tracker

