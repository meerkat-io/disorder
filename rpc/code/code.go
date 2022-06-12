package code

type Code byte

// Custom code can be start from 16
const (
	OK                  Code = 0
	Unknown             Code = 1
	InvalidRequest      Code = 2
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
)
