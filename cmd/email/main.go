package main

import (
	"github.com/chenji/email/internal/interface/cli"
)

// Version 构建时通过 ldflags 注入
var Version = "dev"

func main() {
	cli.Execute(Version)
}