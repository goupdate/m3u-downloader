package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/valyala/fasthttp"
)

func main() {
	m3u := flag.String("m3u", "", "m3u link")
	domain := flag.String("d", "", "domain?")
	cutFrom := flag.Int("cut", 0, "position to cut name from")
	cnt := flag.Int("cnt", 0, "count of files to download 0=all")

	flag.Parse()
	if *m3u == "" || *cutFrom == 0 {
		flag.Usage()
		return
	}

	var dst []byte
	_, body, _ := fasthttp.GetTimeout(dst, *m3u, time.Minute)

	list := strings.Split(string(body), "\n")
	fmt.Printf("m3u contains: %d files\n", len(list))

	ioutil.WriteFile("current.m3u", body, 0644)

	os.MkdirAll("./files", 0644)

	c := make(chan string)
	var w sync.WaitGroup

	num := int32(1)

	downloader := func() {
		defer w.Done()
		for file := range c {
			if len(file)<(*cutFrom+2) {
				continue
			}
			var dst []byte
			name := strings.ReplaceAll(file[*cutFrom:], "/", "-")
			name = strings.ReplaceAll(name, "&", "_")
			name = "./files/" + name

			file = strings.ReplaceAll(file,"/domain/","/"+*domain+"/")

			info, _ := os.Stat(name)
			//file not exists
			if info == nil || info.Size() == 0 {
				fmt.Printf("%d  : %s\n", atomic.AddInt32(&num,1), file)					
				_, body, _ := fasthttp.GetTimeout(dst, file, time.Minute*2)
				ioutil.WriteFile(name, body, 0644)
			}
		}
	}

	for i := 0; i < 10; i++ {
		w.Add(1)
		go downloader()
	}

	for i, v := range list {
		c <- v
		if i > *cnt && *cnt > 0 {
			break
		}
	}

	close(c)
	w.Wait()
}
