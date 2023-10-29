package rpc

import "github.com/meerkat-io/disorder/rpc/code"

type Error struct {
	Code  code.Code
	Error error
}
