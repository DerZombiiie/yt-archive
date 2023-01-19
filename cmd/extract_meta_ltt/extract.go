package main

import (
	"image"
	_ "image/png"
	"os"
	"os/exec"
	"path"

	"github.com/eaciit/gocr"
)

func extract(p string) (m *Meta, err error) {
	tmp := path.Join(os.TempDir(), "image.png")

	// get last frame of video:
	//> ffmpeg -sseof -3 -i input -update 1 -q:v 1 last.jpg
	cmd := exec.Command("ffmpeg",
		"-sseof", "-2", // seek to last 2 seconds
		"-i", p, // input path
		"-update", "1", // overwrite file
		"-q:v", "1", // map video (i think)
		tmp, // output file
	)

	if err = cmd.Run(); err != nil {
		return
	}

	// readout the file
	f, err := os.Open(tmp)
	if err != nil {
		return
	}

	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return
	}

	return
}
