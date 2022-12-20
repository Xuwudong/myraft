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

enum EntryType {
    KV = 1
    MemberChange = 2
    MemberChangeNew = 3
}

struct LogEntry {
    1: i64  term
    2: Entry entry
}

struct Command {
    1: Entry entry
    3: Opt opt
}

struct Entry {
    1: string key
    2: i64 value
    3: EntryType entry_type
    4: list<Member> addMembers
    5: list<Member> subMembers
}

struct Member {
    1: i64 member_id
    2: string server_addr // 内部通信地址
    3: string client_addr // 外部通信地址
}