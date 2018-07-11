package main

import (
	"log"

	. "github.com/kkdai/youtube"
	"fmt"
	"strings"
	"os"

	"sync"
	"net/http"
	"github.com/iris-contrib/errors"
	"io"
)

var pool = sync.WaitGroup{}

func main() {
	////currentFile, _ := filepath.Abs(os.Args[0])
	//currentFile := "/tmp/example1.mp4"
	//log.Println("download to file=", currentFile)
	//
	//// NewYoutube(debug) if debug parameter will set true we can log of messages
	//y := NewYoutube(true)
	//err := y.DecodeURL("https://www.youtube.com/watch?v=rFejpH_tAHM")
	//if err != nil {
	//	fmt.Printf("ERROR : %v\n", err)
	//}
	//fmt.Printf("VideoID:\n `%v`\n:`%v`\n:`%v`\n", y.VideoID, y.DebugMode, y.StreamList)
	//for i, z := range y.StreamList {
	//	fmt.Printf("[%d] %v\n", i, z)
	//	if len(z) > 0 {
	//		for k, v := range z {
	//			fmt.Printf(" %v = `%v`\n", k, v)
	//
	//		}
	//	}
	//}
	////go showPercent(y.DownloadPercent)
	//y.StartDownload(currentFile)

	//var r = makeFileName(y)
	//fmt.Printf("Result: %v\n%v\n%v\n", r.Type, r.Author, r.Filename)
	var down = new(Downloader)
	var link = "https://www.youtube.com/watch?v=rFejpH_tAHM"
	var linkImage = "https://user-images.githubusercontent.com/6550035/32120326-5f420d42-bb15-11e7-89d4-c502864e78eb.gif"

	down.DownloadImage("/tmp/testingimage.gif", linkImage, false)
	os.Exit(1)

	for x := 0; x < 3; x++ {
		pool.Add(1)
		go down.DownloaderVideo(fmt.Sprintf("[GOR#%d]", x), "/home/spouk/tmp/TestingDownloadVideoDir", link, true)
	}
	pool.Wait()

	fmt.Printf("All ok")
}
func showPercent(namegoroutine, filename string, info chan int64) {
	log.Println("starting show percent...")
	defer log.Println("end show percent")
	for {
		select {
		case pr := <-info:
			log.Printf("[%s][%s] %d/100%%", namegoroutine, filename, pr)
		default:
			continue
		}
	}
}

type Element struct {
	Type     string
	Author   string
	Filename string
	Path     string
}
func makeFileName(y *Youtube) *Element {
	var tt = strings.Split(y.StreamList[0]["type"], ";")[0]
	var typeVideo = strings.Split(tt, "/")[1]
	var fname = strings.Join([]string{y.StreamList[0]["author"], typeVideo}, ".")
	return &Element{Type: typeVideo, Author: y.StreamList[0]["author"], Filename: fname}
}


type Downloader struct {
	sync.WaitGroup
}
func (d *Downloader) DownloaderVideo(nameGoroutine, pathSaveDir, linkYoutube string, pool bool) (error) {
	if pool {
		defer func() {
			d.Done()
		}()
	}

	//check exists save path
	_, err := os.Stat(pathSaveDir)
	if err != nil {
		//create new dir
		err := os.Mkdir(pathSaveDir, os.ModeDir|os.ModePerm)
		if err != nil {
			return err
		}
	}
	//download instance
	y := NewYoutube(true)
	err = y.DecodeURL(linkYoutube)
	if err != nil {
		return err
	}
	//extract filename + ext + make outputfilename
	var tt = strings.Split(y.StreamList[0]["type"], ";")[0]
	var typeVideo = strings.Split(tt, "/")[1]                                    //ex: mp4
	var name = strings.Join([]string{y.StreamList[0]["author"], typeVideo}, ".") //fname
	//var fname = strings.Join([]string{name, typeVideo}, ".")

	//download video
	ps := strings.Join([]string{pathSaveDir, name}, "/")
	fmt.Printf("PATH SAVE: %v\n", ps)
	go showPercent(nameGoroutine, ps, y.DownloadPercent)
	//y.StartDownload("/tmp/tester.mp4")
	y.StartDownload(ps)

	return nil
}

func (d *Downloader) DownloadImage(pathToSaveWithFilename, link string, pool bool) (error) {
	if pool {
		defer func() {
			d.Done()
		}()
	}
	resp, err := http.Get(link)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return errors.New(fmt.Sprintf("http error: %v\n", http.StatusForbidden))
	}
	file, err := os.Create(pathToSaveWithFilename)
	if err != nil {
		return err
	}

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}
	file.Close()
	log.Printf("%s success download\n", pathToSaveWithFilename)
	return nil
}
