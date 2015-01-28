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


// Make the Hostnames an array,
// Store every hostname that matches the hash
// Fake the file for each to the stdout buffer
type Hash struct {
	Hostnames  string
	Contents   []byte
}


func StringInArray(a string, list map[string]Hash) bool {
	for q, _ := range list {
		if q == a {
			return true
		}
	}
	return false
}


func cater(path string) bytes.Buffer {
	var response bytes.Buffer
	knownHashes := grab(path)

	for _, hash := range knownHashes {

		var b bytes.Buffer
		b.Write([]byte(hash.Contents))

		gr, _ := gzip.NewReader(&b)
		defer gr.Close()
		plaintext, _ := ioutil.ReadAll(gr)

		for _, a := range strings.Split(string(plaintext), "\n") {
			line := fmt.Sprintf("%s:%s\n", hash.Hostnames, a)
			response.Write([]byte(line))
		}
	}
	return response

}


func grab(path string) map[string]Hash {
	var file_contents_gzipped string
	knownHashes := make(map[string]Hash)

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

		// Check if we already have this hash, before wasting redis time
		if ! StringInArray(fs.Hash, knownHashes) {
			file_contents_gzipped, _ = redis.String(conn.Do("GET", fs.Hash))
			knownHashes[fs.Hash] = Hash{
				// Need to switch this to an array of hostnames
				// to fake the output and save lookups, ect
				Hostnames: fs.Hostname,
				Contents: []byte(file_contents_gzipped),
			}
		}
	}
	return  knownHashes
}

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("usage: %s filepattern\n", os.Args[0])
		return
	}
	cater_response := cater(os.Args[1])
	fmt.Printf(cater_response.String())
}

