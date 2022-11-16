package errno

type NewClientErr struct {
	msg string
}

func (e *NewClientErr) Error() string {
	return "new client err:" + e.msg
}

func NewNewClientErr(msg string) *NewClientErr {
	return &NewClientErr{
		msg: msg,
	}
}
