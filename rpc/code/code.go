package code

type Code int32

// Custom code start from 100
const (
	Unknown             Code = 0
	InvalidRequest      Code = 1
	DataCorrupt         Code = 2
	NotFound            Code = 3
	AlreadyExists       Code = 4
	PermissionDenied    Code = 5
	ResourceExhausted   Code = 6
	Aborted             Code = 7
	OutOfRange          Code = 8
	Unimplemented       Code = 9
	Internal            Code = 10
	Unavailable         Code = 11
	Unauthenticated     Code = 12
	NetworkDisconnected Code = 13
	InvalidHost         Code = 14
)
