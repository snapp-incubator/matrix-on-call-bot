package command

import "errors"

var (
	ErrInvalidCommand = errors.New("invalid command")
	ErrUnknownCommand = errors.New("unknown command")
	ErrInvalidBody    = errors.New("invalid body")
	ErrInvalidType    = errors.New("invalid type")
)
