package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/derzombiiie/yt-archive"
	//	"google.golang.org/api/googleapi"
	"google.golang.org/api/youtube/v3"
)

var service *youtube.Service

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: get_vids <in: videos.txt> <out: meta.txt> [startid no=\"\"]")
		os.Exit(-1)
	}

	startID := os.Args[3]

	var err error
	service, err = archive.NewService()
	if err != nil {
		log.Fatalf("Error creating Service: %s\n", err)
	}

	f, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatalf("Error opening %s: %s\n", os.Args[1], err)
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	s.Scan()

	of, err := os.OpenFile(os.Args[2], os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		log.Fatalf("Error optining '%s' RDWR:%s\n", os.Args[2], err)
	}

	defer of.Close()

	e := json.NewEncoder(of)

	idCh := make(chan string, 1)

	go dlThread(e, idCh)
	var line int64

	log.Printf("Waiting for ID '%s'\n", startID)
	scan := startID == ""
	for s.Scan() {
		line++

		if !scan {
			scan = startID == s.Text()
		} else {
			fmt.Printf("%5d ", line)
			idCh <- s.Text()
			continue
		}

		if scan {
			fmt.Printf("%5d ", line)
			idCh <- s.Text()
		}
	}
}

func dlThread(enc *json.Encoder, idCh chan string) {
	for {
		id := <-idCh

		if err := getMeta(id, enc); err != nil {
			log.Fatalf("Error downloading video '%s': %s\n", id, err)
		}

	}
}

func getMeta(id string, enc *json.Encoder) error {
	log.Printf("Getting ID '%s'\n", id)

	vcall := service.Videos.List([]string{"snippet", "contentDetails"})
	vcall.Id(id)

	vres, err := vcall.Do()
	if err != nil {
		return err
	}

	if len(vres.Items) == 0 {
		return fmt.Errorf("vres.Items==0!")
	}

	video := vres.Items[0]

	enc.Encode(video)

	return nil
}
