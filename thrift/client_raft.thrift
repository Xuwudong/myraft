namespace go raft
include "common.thrift"

struct DoCommandReq {
    1: string id
	2: common.Command command
}

struct DoCommandResp {
    1: bool succuess
    2: i64 value
}

service ClientRaftServer{
    DoCommandResp DoCommand(1:DoCommandReq req);
}
