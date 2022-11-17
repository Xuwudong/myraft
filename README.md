## MyRaft
>基于[raft](https://github.com/maemual/raft-zh_cn)算法实现的分布式线性一致性kv系统
### 功能列表
1. 强一致性，支持简单k(string), v(int) 读写
2. 高可用，自动master选举（集群启动、leader宕机场景）
3. 容错，集群中只要有半数实例存活即可以支持写,follower宕机恢复后自动同步日志

### 如何跑起来
1. 安装thrift，执行以下命令生成rpc相关代码
```shell
    thrift -r --gen go thrift/raft.thrift
    thrift -r --gen go thrift/client_raft.thrift
```
2. 系统配置文件在conf/raft.conf,默认三个实例，可以自行增加或修改接口，注意需要手动创建系统数据目录：/var/myraft
3. 假如没有修改配置，则使用以下shell命令启动服务
```shell
    # id参数指的是实例编号，对应raft.conf配置中的server.#{id}
    go run main.go 
    go run main.go -id 2 
    go run main.go -id 3
```

### TODO
1. 支持自动扩容
2. 提供 go client sdk
3. 实现实例之间通信连接池
4. 日志压缩
