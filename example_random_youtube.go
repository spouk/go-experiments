package main

import (
	"log"

	. "github.com/kkdai/youtube"
	"fmt"
	"strings"
)

func main() {
	//currentFile, _ := filepath.Abs(os.Args[0])
	currentFile := "/tmp/example1.mp4"
	log.Println("download to file=", currentFile)

	// NewYoutube(debug) if debug parameter will set true we can log of messages
	y := NewYoutube(true)
	err := y.DecodeURL("https://www.youtube.com/watch?v=yoDF5SjsiVg")
	if err != nil {
		fmt.Printf("ERROR : %v\n", err)
	}
	fmt.Printf("VideoID:\n `%v`\n:`%v`\n:`%v`\n", y.VideoID, y.DebugMode, y.StreamList)
	for i, z := range y.StreamList {
		fmt.Printf("[%d] %v\n", i, z)
		if len(z) > 0 {
			for k, v := range z {
				fmt.Printf(" %v = `%v`\n", k, v)

			}
		}
	}
	go showPercent(y.DownloadPercent)
	y.StartDownload(currentFile)

	var r = makeFileName(y)
	fmt.Printf("Result: %v\n%v\n%v\n", r.Type, r.Author, r.Filename)
}
func showPercent(info chan int64) {
	log.Println("starting show percent...")
	defer log.Println("end show percent")
	for {
		select {
		case pr := <-info:
			log.Printf("%d", pr)
		default:
			continue
		}
	}
}

type Element struct {
	Type     string
	Author   string
	Filename string
}

func makeFileName(y *Youtube) *Element {
	fmt.Printf(">>>FIRSTELEMENT: %v\n", y.StreamList[0])
	var tt = strings.Split(y.StreamList[0]["type"], ";")[0]
	var typeVideo = strings.Split(tt, "/")[1]
	var fname  = strings.Join([]string{y.StreamList[0]["author"], typeVideo}, ".")
	return &Element{Type: typeVideo, Author: y.StreamList[0]["author"], Filename:fname}

}
