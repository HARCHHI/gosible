package main

import (
	"github.com/HARCHHI/gosible/cmd"
	"log"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
