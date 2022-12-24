## MyRaft

> 基于[raft](https://github.com/maemual/raft-zh_cn)算法实现的分布式线性一致性kv系统

### 功能列表

1. 线性一致性，支持简单k(string), v(int) 读写
2. 高可用，自动master选举（集群启动、leader宕机场景）
3. 容错，集群中只要有半数实例存活即可以支持写,follower宕机恢复后自动同步日志
4. 支持成员变更

### 如何跑起来

1. 需要手动创建系统日志目录：/var/myraft
2. 使用以下shell命令启动服务

```shell
    # 启动leader，由于没有服务发现，leader默认为1 8080 9090
    sh run.sh 1 8080 9090 
    sleep 2
    # 启动控制面工具
    cd control_kit && sh run.sh && cd ..
    sleep 3
    # 启动followers
    sh run.sh 2 8082 9092
    sh run.sh 3 8083 9093
    sh run.sh 4 8084 9094
    sh run.sh 5 8085 9095
    sh run.sh 6 8086 9096
    ...
```

### TODO

1. 引入服务发现
2. 提供 go client sdk
3. 日志压缩
