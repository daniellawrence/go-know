package main

import (
	"testing"
	"os"
	"io/ioutil"
	"fmt"
)

func TestGetFileInfo(t *testing.T) {

	test_file_path := "/tmp/test_file"
	test_content := []byte("Hello World\n")

	ioutil.WriteFile(test_file_path, test_content, 0644)

	test_file_info, _ := os.Stat(test_file_path)

	file_md5sum, fileStatJson, file_contents := getFileInfo(test_file_path, test_file_info)

	if file_md5sum != "e59ff97941044f85df5297e1c302d260" {
		t.Error(
			"For", "/tmp/test_file",
			"expected", "md5_sum: xxxx",
			"got", file_md5sum,
		)
	}
	if string(file_contents) != "Hello World\n" {
		t.Error(
			"For", "/tmp/test_file",
			"expected", "Hello World\n",
			"got", string(file_contents),
		)
	}
}
