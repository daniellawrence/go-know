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


func GrabWalk(dirPath string) bytes.Buffer {

	fullPath, err := filepath.Abs(dirPath)
	var whole_buffer bytes.Buffer;

	check(err)
	
	callback := func(path string, fi os.FileInfo, err error) error {
		if strings.Contains(path, ".git") {
			return nil
		}
		single_file_buffer := getFileInfo(path, fi)
		whole_buffer.WriteString(single_file_buffer.String())
		
		return nil
	}
	filepath.Walk(fullPath, callback)

	return whole_buffer
}

func getFileInfo(path string, fi os.FileInfo) bytes.Buffer {
	var buffer bytes.Buffer;

	if fi.IsDir() {
		return buffer
	}
	
	mm, _ := magicmime.New(magicmime.MAGIC_MIME_TYPE | magicmime.MAGIC_SYMLINK | magicmime.MAGIC_ERROR)
	mimetype, _ := mm.TypeByFile(path)

	if strings.Contains(mimetype, "application"){
		return buffer
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

	hostname_key := fmt.Sprintf("%s:%s", global_hostname, path)
	WriteToRedis(hostname_key, fileStatJson)

	WriteToRedisCompressed(file_md5sum, file_contents)

	return buffer

}

func main() {
	global_hostname, _ = os.Hostname()
	GrabWalk("/etc/")
	fmt.Printf("\nDone\n")
}
