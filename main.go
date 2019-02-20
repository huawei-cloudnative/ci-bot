package main

import (
	"github.com/Huawei-PaaS/ci-bot/handlers"

	"github.com/spf13/pflag"
)

func main() {
	s := handlers.NewWebHookServer()
	handlers.AddFlags(pflag.CommandLine, s)
	handlers.Run(s)
}
