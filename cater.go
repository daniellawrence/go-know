package main

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"encoding/json"
	"compress/gzip"
	"bytes"
	"io/ioutil"
	"strings"
	//"os"
)

/*

type FileStat struct {
	Path     string
	Hash     string
	Size     int64
	Hostname string
}
*/

// Make the Hostnames an array,
// Store every hostname that matches the hash
// Fake the file for each to the stdout buffer
type Hash struct {
	Hostnames  []string
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
			if len(a) == 0 {
				continue
			}

			for _, b := range hash.Hostnames {
				line := fmt.Sprintf("%s:%s\n", b, a)
				response.Write([]byte(line))
			}
		}
	}
	return response

}


func grab(path string) map[string]Hash {
	var file_contents_gzipped []byte
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
			file_contents_gzipped_string, _ := redis.String(conn.Do("GET", fs.Hash))
			file_contents_gzipped = []byte(file_contents_gzipped_string)
			
			fmt.Println("NEW HASH")
			knownHashes[fs.Hash] = Hash{
				Hostnames: []string{fs.Hostname},
				Contents: []byte(file_contents_gzipped),
			}
		} else {
			file_contents_gzipped = knownHashes[fs.Hash].Contents
			old_hostnames := knownHashes[fs.Hash].Hostnames
			new_hostnames := append(old_hostnames, fs.Hostname)
			knownHashes[fs.Hash] = Hash{
				Hostnames: new_hostnames,
				Contents: file_contents_gzipped,
			}
		}
	}
	return  knownHashes
}

/*
func main() {
	if len(os.Args) != 2 {
		fmt.Printf("usage: %s filepattern\n", os.Args[0])
		return
	}
	cater_response := cater(os.Args[1])
	fmt.Printf(cater_response.String())
}
*/
