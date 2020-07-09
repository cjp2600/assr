package main

import (
	"github.com/cjp2600/assr/cmd"
	"github.com/cjp2600/assr/log"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Error(r.(error))
		}
	}()
	cmd.Execute()
}
