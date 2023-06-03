package main

import (
	"dominocount"
	"os"
)

func main() {
	dominocount.RunServer(os.Stdout)
}
