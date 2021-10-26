package main

import (
	"fmt"
	"strings"
)

type myLogger struct{}

func (l myLogger) Log(msg string) {
	if strings.Index(msg, "INFO:") == -1 {
		fmt.Printf("%v\n", msg)
	}
}
