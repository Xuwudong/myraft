package conf

type MemberConfState int8

const COld MemberConfState = 0
const COldNew MemberConfState = 1

//const CNew MemberConfState = 2

type MemberConf struct {
	ServerAddrMap    map[int64]string
	ClientAddrMap    map[int64]string
	NewServerAddrMap map[int64]string

	State MemberConfState
}
