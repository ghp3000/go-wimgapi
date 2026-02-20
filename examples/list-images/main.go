//go:build windows

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ghp3000/go-wimgapi/wimgapi"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("usage: %s <path-to-wim>", os.Args[0])
	}

	f, err := wimgapi.Open(os.Args[1], wimgapi.OpenOptions{})
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	images, err := f.Images()
	if err != nil {
		log.Fatal(err)
	}

	for _, img := range images {
		fmt.Printf("%d %s\n", img.Index, img.Name)
	}
}
