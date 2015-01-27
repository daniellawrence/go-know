package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"github.com/rakyll/magicmime"
	"crypto/md5"
	"bytes"
	"compress/gzip"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func GrabWalk(dirPath string) bytes.Buffer {

	fullPath, err := filepath.Abs(dirPath)
	var whole_buffer bytes.Buffer;

	if err != nil {
		return whole_buffer
	}

	
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
	b, _ := ioutil.ReadFile(path)

	md5sum := md5.Sum(b)
	buffer.WriteString(fmt.Sprintf(">>>>> FILE_MD5: %s <<<<<<\n", path))
	buffer.WriteString(fmt.Sprintf("%x\n", md5sum))

	buffer.WriteString(fmt.Sprintf(">>>>> FILE_CONTENT: %s <<<<<<\n", path))
	buffer.WriteString(string(b))
	
	buffer.WriteString(fmt.Sprintf(">>>>> FILE_STAT: %s <<<<<<\n", path))
	buffer.WriteString(fmt.Sprintf("size: %s\n", fi.Size()))

	return buffer

}

func main() {
	b := GrabWalk("/etc/")

	hostname, _ := os.Hostname()
	output_gz := fmt.Sprintf("%s-etc.gz", hostname)

	var c bytes.Buffer
	w := gzip.NewWriter(&c)
	fmt.Fprintf(w, b.String())
	w.Close()
	ioutil.WriteFile(output_gz, c.Bytes(), 0666)

	fmt.Printf("Wrote stats from /etc into %s\n", output_gz)
}
