package tracer

const (
	traceSysUnknown = iota
	traceSysSocket
	traceSysConnect
	traceSysSendto
	traceSysSendmsg
	traceSysFcntl
	traceSysClose
)
