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
	"github.com/codegangsta/cli"
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

func greper(path string, pattern string, countonly bool) {
	var response bytes.Buffer
	var count int
	count = 0
	knownHashes := grab(path)
	fmt.Printf("searching for '%s' in %s\n", pattern, path)

	for _, hash := range knownHashes {

		var b bytes.Buffer
		b.Write([]byte(hash.Contents))

		gr, _ := gzip.NewReader(&b)
		defer gr.Close()
		plaintext, _ := ioutil.ReadAll(gr)

		if countonly {
			if strings.Contains(string(plaintext), pattern) {
				count += len(hash.Hostnames)
				continue
			}
		}


		for _, a := range strings.Split(string(plaintext), "\n") {
			if len(a) == 0 {
				continue
			}

			if ! strings.Contains(a, pattern) {
				continue
			}

			for _, b := range hash.Hostnames {

				line := fmt.Sprintf("%s:%s\n", b, a)
				response.Write([]byte(line))
			}
		}
	}
	if countonly {
		fmt.Printf("Count: %d\n", count)
	}
	fmt.Printf(response.String())

}



func cater(path string) {
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
	// return response
	fmt.Printf(response.String())

}


func grab(path string) map[string]Hash {
	var file_contents_gzipped []byte
	knownHashes := make(map[string]Hash)

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

func main() {

	app := cli.NewApp()
	app.Name = "grab"
	app.Usage = "Grab file data from redis server"

	app.Commands = []cli.Command{
		{
			Name:      "cat",
			Usage:     "add a task to the list",
			Action:    func(c *cli.Context) {
				if len(c.Args()) != 1 {
					return
				}
				path := c.Args().First()
				cater(path)
			},
		},
		{
			Name:      "grep",
			Usage:     "grep for pattern",
			Action:    func(c *cli.Context) {
				if len(c.Args()) != 2 {
					fmt.Printf("missing file and pattern\n")
					return
				}
				path := c.Args().First()
				pattern := c.Args()[1]
				greper(path, pattern, false)
			},
		},
		{
			Name:      "grepcount",
			Usage:     "count pattern",
			Action:    func(c *cli.Context) {
				if len(c.Args()) != 2 {
					fmt.Printf("missing file and pattern\n")
					return
				}
				path := c.Args().First()
				pattern := c.Args()[1]
				greper(path, pattern, true)
			},
		},
	}
	
	app.Run(os.Args)
}
