// Example program that uses blakjack/webcam library
// for working with V4L2 devices.
// The application reads frames from device and writes them to stdout
// If your device supports motion formats (e.g. H264 or MJPEG) you can
// use it's output as a video stream.
// Example usage: go run stdout_streamer.go | vlc -
package main

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	"os"
	"sort"

	"github.com/blackjack/webcam"
)

type FrameSizes []webcam.FrameSize

func (slice FrameSizes) Len() int {
	return len(slice)
}

// For sorting purposes
func (slice FrameSizes) Less(i, j int) bool {
	ls := slice[i].MaxWidth * slice[i].MaxHeight
	rs := slice[j].MaxWidth * slice[j].MaxHeight
	return ls < rs
}

// For sorting purposes
func (slice FrameSizes) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

func compressAscii(img image.Image, w int, h int) string {
	var a, b int

	a = img.Bounds().Max.X
	b = img.Bounds().Max.Y
	ratiow := a / w
	ratioh := b / h
	result := ""
	list := " .:-=+*#%@"

	for j := 0; j < h; j++ {
		for i := 0; i < w; i++ {
			cpt := 0
			for l := 0; l < ratioh; l++ {
				for k := 0; k < ratiow; k++ {
					r, g, blue, _ := img.At(ratiow*i+k, ratioh*j+l).RGBA()
					cpt += int((r >> 8) + (g >> 8) + (blue >> 8))
				}
			}
			cpt /= (ratiow * ratioh * 3)
			result += string(list[cpt*len(list)/256])
		}
		result += "\n"
	}
	return result
}

func disp(value string, w int, h int) {
	fmt.Print(value)
	for i := 0; i < h; i++ {
		fmt.Printf("\033[1A")
	}
	fmt.Printf("\r")
}

func quitter() {
	var a string
	for true {
		fmt.Scanln(&a)
		if a == "q" {
			os.Exit(0)
		}
	}
}

func main() {
	const fps = 13
	const width = 320
	const height = 90

	cam, err := webcam.Open("/dev/video0")
	if err != nil {
		panic(err.Error())
	}
	defer cam.Close()

	format_desc := cam.GetSupportedFormats()
	var formats []webcam.PixelFormat
	for f := range format_desc {
		formats = append(formats, f)
	}

	frames := FrameSizes(cam.GetSupportedFrameSizes(formats[0]))
	sort.Sort(frames)

	f, m, h, err := cam.SetImageFormat(formats[0], uint32(frames[0].MaxWidth), uint32(frames[0].MaxHeight))
	if err != nil {
		panic(err.Error())
	}

	if err != nil {
		panic(err.Error())
	}

	n, _ := cam.GetFramerate()

	fmt.Printf("%s : %dx%d | %vfps\n", format_desc[f], m, h, n)

	err = cam.StartStreaming()
	if err != nil {
		panic(err.Error())
	}

	timeout := uint32(5) //5 seconds

	go quitter()

	for {
		var frame []byte
		var err error
		for i := 0; i < int(float32(n)/float32(fps)); i++ {
			err = cam.WaitForFrame(timeout)
			switch err.(type) {
			case nil:
			case *webcam.Timeout:
				fmt.Fprint(os.Stderr, err.Error())
				continue
			default:
				panic(err.Error())
			}

			frame, err = cam.ReadFrame()
		}

		if err != nil {
			panic(err.Error())
		} else if len(frame) != 0 {

			img, _, error := image.Decode(bytes.NewReader(frame))
			if error != nil {
				panic(error.Error())
			}

			fin := compressAscii(img, width, height)
			go disp(fin, width, height)
		}
	}
}
