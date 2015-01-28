package main

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"encoding/json"
	"compress/gzip"
	"bytes"
	"io/ioutil"
	"strings"
	"os"
)

type FileStat struct {
	Path     string
	Hash     string
	Size     int64
	Hostname string
}

func cater(path string) {
	fmt.Printf("search path: %s\n", path)
	conn, _ := redis.Dial("tcp", ":6379")
	defer conn.Close()
	x, _ := redis.Values(conn.Do("KEYS", path))
	var fs FileStat
	for _, z := range x {
		z = fmt.Sprintf("%s", z)
		reply, _ := redis.String(conn.Do("GET", z))
		y := []byte(reply)

		json.Unmarshal(y, &fs)
		file_contents_gzipped, _ := redis.String(conn.Do("GET", fs.Hash))

		var b bytes.Buffer
		b.Write([]byte(file_contents_gzipped))

		gr, _ := gzip.NewReader(&b)
		defer gr.Close()
		plaintext, _ := ioutil.ReadAll(gr)

		for _, a := range strings.Split(string(plaintext), "\n") {
			fmt.Printf("%s:%s\n", fs.Hostname, a)
		}
	}
	
}

func main() {
	cater(string(os.Args[1]))
}
