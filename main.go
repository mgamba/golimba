package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/hajimehoshi/oto"
	"github.com/hajimehoshi/go-mp3"
)

func run() error {
	f, err := os.Open("kit.mp3")
	if err != nil {
		return err
	}
	defer f.Close()

	d, err := mp3.NewDecoder(f)
	if err != nil {
		return err
	}

	p, err := oto.NewPlayer(d.SampleRate(), 2, 2, 8192)
	if err != nil {
		return err
	}
	defer p.Close()

	fmt.Printf("Length: %d[bytes]\n", d.Length())

	if _, err := io.CopyN(p, d, 88000); err != nil {
		return err
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
