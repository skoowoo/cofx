package service

import "io"

type Printer interface {
	PrettyPrint(io.Writer) error
	JsonPrint(io.Writer) error
}
