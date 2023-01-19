package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"

	//	"google.golang.org/api/googleapi"
	"google.golang.org/api/youtube/v3"
)

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: download_image <in: meta.txt> <out: outimgs/> <defualt|maxres|medium|standard> [skipto]")
		os.Exit(-1)
	}

	println(len(os.Args))

	var skipto int64
	if len(os.Args) >= 5 {
		var err error
		skipto, err = strconv.ParseInt(os.Args[4], 10, 64)
		if err != nil {
			log.Fatal(err)
		}
	}

	f, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatalf("Error openining '%s': %s\n", os.Args[1], err)
	}

	defer f.Close()

	out := os.Args[2]
	res := os.Args[3]

	dec := json.NewDecoder(f)

	var line int64 = -1

	for {
		line++

		var vid youtube.Video

		if err := dec.Decode(&vid); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			log.Fatalf("Decode failed: %s\n", err)
		}

		if line < skipto {
			fmt.Printf("%d skip!\n", line)
			continue
		}

		var url string

		switch res { // defualt|maxres|medium|standard
		case "maxres":
			if vid.Snippet.Thumbnails.Maxres != nil {
				url = vid.Snippet.Thumbnails.Maxres.Url
				break
			}

			fallthrough

		case "medium":
			if vid.Snippet.Thumbnails.Medium != nil {
				url = vid.Snippet.Thumbnails.Medium.Url
				break
			}

			fallthrough

		case "standard":
			if vid.Snippet.Thumbnails.Standard != nil {
				url = vid.Snippet.Thumbnails.Standard.Url
				break
			}

			fallthrough

		case "default":
			url = vid.Snippet.Thumbnails.Default.Url

		default:
			log.Fatal("Invalid resolution")
		}

		log.Printf("%6d, GET: %s\n", line, url)

		resp, err := http.Get(url)
		if err != nil {
			log.Fatalf("Error GET url: '%s': %s\n", url, err)
		}

		defer resp.Body.Close()

		p := path.Join(out, vid.Id+".jpg")
		imgfile, err := os.OpenFile(p, os.O_CREATE|os.O_RDWR, 0777)
		if err != nil {
			log.Fatalf("Error Opening/Creating '%s': %s\n", p, err)
		}

		defer imgfile.Close()

		imgfile.Seek(0, 0)

		io.Copy(imgfile, resp.Body)
	}
}
