package service

import "io"

type Writer interface {
	JsonWrite(io.Writer) error
}
