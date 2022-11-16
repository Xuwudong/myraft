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

struct LogEntry {
    1: i64  term
    2: Command command
}

struct Command {
    1: Entity entity
    3: Opt opt
}

struct Entity {
    1: string key
    2: i64 value
}