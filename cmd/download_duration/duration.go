package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: download_duration <videos/> [videos.txt]")
		os.Exit(-1)
	}

	var overlay map[string]struct{}
	if len(os.Args) >= 3 {
		overlay = make(map[string]struct{})
		f, err := os.Open(os.Args[2])
		if err != nil {
			log.Fatalf("Error opening %s: %s\n", os.Args[2], err)
		}

		defer f.Close()

		s := bufio.NewScanner(f)

		for s.Scan() {
			overlay[s.Text()] = struct{}{}
		}
	}

	videos := os.Args[1]

	log.Print("Videosdir: ", videos)
	e, err := os.ReadDir(videos)
	if err != nil {
		log.Fatal(err)
	}

	var files []string
	for _, f := range e {
		files = append(files, path.Join(videos, f.Name()))
		log.Printf("name %s\n", f.Name())
	}

	var total time.Duration

loop:
	for _, pa := range files {
		if overlay != nil {
			_, file := path.Split(pa)
			ext := path.Ext(file)

			if ext == ".part" {
				continue
			}

			id := strings.SplitN(file, ".", 2)[0]
			if _, ok := overlay[id]; !ok {
				fmt.Printf("skipping: %s <- %s\n", pa, id)
				//				fmt.Print(".")

				continue loop
			} else {
				fmt.Print("!")
			}
		}

		t := probe(pa)
		fmt.Printf("File: %s -> %10s\n", pa, t)

		total += t
	}

	log.Printf("\nTotalDuration: %s\n", total)
	log.Printf("TotalDuration: %s\n", total)
}

func probe(file string) time.Duration {
	//ffprobe -v 0 -of compact=p=0:nk=1 -show_entries format=duration
	cmd := exec.Command("ffprobe",
		"-v", "0",
		"-of", "compact=p=0:nk=1",
		"-show_entries", "format=duration", // only print duration
		"-i", file,
	)

	cmd.Stderr = os.Stderr

	buf := &bytes.Buffer{}

	cmd.Stdout = buf

	if err := cmd.Run(); err != nil {
		log.Printf("Error executing ffprobe: %s; (%s)\n", err, file)
		return 0
	}

	secs, err := strconv.ParseFloat(buf.String()[:len(buf.String())-1], 64)
	if err != nil {
		log.Fatalf("Error converting %s to float: %s (%s)\n", buf.String(), err, file)
	}

	return time.Millisecond * (time.Duration(secs) * 1000)
}
