package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"strings"
	"syscall"
)

var kproc *os.Process

func main() {
	if len(os.Args) < 3 {
		log.Fatal("Usage: download <videos.txt> <skipto> [yt-dlp args]...")
	}

	out := "./outfiles/"

	f, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatalf("OPEN: %s\n", err)
	}

	have := make(map[string]string)

	// read ids allready downloaded:
	e, err := os.ReadDir(out)
	if err != nil {
		log.Fatalf("ERROR: %s", err)
	}

	for _, v := range e {
		name := strings.SplitN(v.Name(), ".", 2)[0]

		have[name] = out + v.Name()
	}

	s := bufio.NewScanner(f)
	s.Scan()

	var count int
	var session int

	interruptCh := MKiChan()

	exit := func() {
		log.Printf("Stopped at count: %d, downloads this session: %d", count, session)
		os.Exit(0)
	}

	for s.Scan() {
		count++

		id := s.Text()
		// check if allready have:
		if path, ok := have[id]; ok {
			fmt.Printf("Already have %s: %s\n", id, path)
			continue
		}

		session++

		args := append(os.Args[3:], "-o", path.Join(out, id), "--", id)
		fmt.Printf(">> yt-dlp %s -> %d\n", strings.Join(args, " "), count)

		cmd := exec.Command("yt-dlp", args...)
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setpgid: true,
			Pgid:    0,
		}

		err := cmd.Start()
		if err != nil {
			log.Fatalf("Error starting yt-dlp: %s\n", err)
		}

		kproc = cmd.Process

		err = cmd.Wait()
		if err != nil {
			log.Fatalf("Error waiting for yt-dlp: %s\n", err)
		}

		select {
		case i := <-interruptCh:
			fmt.Printf("Got signal: %s\n", i)
			if i == os.Interrupt {
				exit()
			}

		default:
			continue
		}
	}

	fmt.Println("Done!")
	exit()
}

func MKiChan() <-chan os.Signal {
	ch := make(chan os.Signal, 1)
	ch2 := make(chan os.Signal, 2)

	signal.Notify(ch, os.Interrupt)

	go func() {
		last := false

		for {
			s := <-ch
			if last {
				log.Printf("[meta] ABORT!")
				kproc.Kill()
				os.Exit(-1)
			}
			log.Printf("[meta] Got signal: %s, stopping after this download Ctrl+c again to abort\n", s)
			last = true
			ch2 <- s
		}
	}()

	return ch2
}
