//go:build windows

package main

import (
	"log"
	"os"
	"strconv"

	"github.com/ghp3000/go-wimgapi/wimgapi"
)

func main() {
	if len(os.Args) != 4 {
		log.Fatalf("usage: %s <path-to-wim> <index> <target-dir>", os.Args[0])
	}

	index, err := strconv.Atoi(os.Args[2])
	if err != nil {
		log.Fatal(err)
	}

	f, err := wimgapi.Open(os.Args[1], wimgapi.OpenOptions{})
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	img, err := f.LoadImage(index)
	if err != nil {
		log.Fatal(err)
	}
	defer img.Close()

	if err := img.Apply(os.Args[3], wimgapi.ApplyOptions{}); err != nil {
		log.Fatal(err)
	}
}
