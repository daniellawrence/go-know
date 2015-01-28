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



func cater(path string) bytes.Buffer {
	var response bytes.Buffer;

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
			line := fmt.Sprintf("%s:%s\n", fs.Hostname, a)
			response.Write([]byte(line))
		}
	}
	return response
	
}

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("usage: %s filepattern\n", os.Args[0])
		return
	}
	cater_response := cater(os.Args[1])
	fmt.Printf(cater_response.String())
}

