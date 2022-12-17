package bencode

import "errors"

var (
	TypeError  = errors.New("wrong type")
	NumError   = errors.New("expect num")
	ColonError = errors.New("expect colon")
	CharIError = errors.New("expect char i")
	CharEError = errors.New("expect char e")
)
