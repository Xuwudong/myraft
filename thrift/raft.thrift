namespace go raft

enum Role{
    Leader  = 1
    Follower = 2
    Candidater = 3
}

enum Opt{
    Read = 1
    Write = 2
}

struct Command {
    1: Entry entry
    3: Opt opt
}

struct Entry {
    1: string key
    2: i64 value
}


struct AppendEntriesReq {
	1: i64 term
	2: i64 leaderId
    3: i64 preLogIndex
    4: i64 preLogTerm
    5: list<Command> commands
    6: i64 leaderCommit
}

struct AppendEntriesResp {
    1: i64 term
    2: bool succuess
}

struct RequestVoteReq{
    1: i64 term
    2: i64 candidateId
    3: i64 lastLogIndex
    4: i64 lastLogTerm
}

struct RequestVoteResp{
    1: i64 term
    2: bool voteGranted
}

service RaftServer{
    AppendEntriesResp AppendEntries(1: AppendEntriesReq req);
    RequestVoteResp RequestVote(1: RequestVoteReq req);
}
