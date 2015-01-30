package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"crypto/md5"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"log"
	"github.com/rakyll/magicmime"
	"github.com/garyburd/redigo/redis"

)

var global_hostname string

type FileStat struct {
	Path     string
	Hash     string
	Size     int64
	Hostname string
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func WriteToRedis(key string, value []byte) {
	// fmt.Printf("wrote key='%s' to tcp:6379\n", key)
	fmt.Printf(".")
	conn, err := redis.Dial("tcp", ":6379")
	check(err)
	defer conn.Close()
	_, err = conn.Do("SET", key, value)
	check(err)
}

func WriteToRedisCompressed(key string, value []byte) {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	w.Write(value)
	w.Close()
	WriteToRedis(key, b.Bytes())
}


func GrabWalk(dirPath string) {

	fullPath, err := filepath.Abs(dirPath)

	check(err)
	
	callback := func(path string, fi os.FileInfo, err error) error {
		if strings.Contains(path, ".git") {
			return nil
		}

		file_md5sum, fileStatJson, file_contents := getFileInfo(path, fi)

		hostname_key := fmt.Sprintf("%s:%s", global_hostname, path)
		WriteToRedis(hostname_key, fileStatJson)
		WriteToRedisCompressed(file_md5sum, file_contents)
		
		return nil
	}
	filepath.Walk(fullPath, callback)

	return 
}

func getFileInfo(path string, fi os.FileInfo) (string, []byte, []byte) {
	var empty_bytes []byte
	var empty_string string;

	if fi.IsDir() {
		return empty_string, empty_bytes, empty_bytes
	}
	
	mm, _ := magicmime.New(magicmime.MAGIC_MIME_TYPE | magicmime.MAGIC_SYMLINK | magicmime.MAGIC_ERROR)
	mimetype, _ := mm.TypeByFile(path)

	if strings.Contains(mimetype, "application"){
		return empty_string, empty_bytes, empty_bytes
	}


	file, _ := os.Open(path)
	defer file.Close()
	file_contents, _ := ioutil.ReadFile(path)

	file_md5sum := fmt.Sprintf("%x", md5.Sum(file_contents))
	file_size := fi.Size()

	fs := FileStat{
		Path: path,
		Hash: file_md5sum,
		Size: file_size,
		Hostname: global_hostname,
	}
	fileStatJson, _ := json.Marshal(fs)
	

	return file_md5sum, fileStatJson, file_contents

}

func main() {
	global_hostname, _ = os.Hostname()
	GrabWalk("/etc/")
	fmt.Printf("\nDone\n")
}
