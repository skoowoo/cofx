package main

import "github.com/sirupsen/logrus"

func main() {
	logrus.SetLevel(logrus.ErrorLevel)
	Execute()
}
