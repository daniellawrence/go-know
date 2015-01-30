package main

import (
	"testing"
	"os"
	"io/ioutil"
	"fmt"
)

func TestgetFileInfo(t *testing.T) {

	test_file_path := "/tmp/test_file"
	test_content := []byte("Hello World\n")

	ioutil.WriteFile(test_file_path, test_content, 0644)

	test_file_info, _ := os.Stat(test_file_path)

	file_md5sum, fileStatJson, file_contents := getFileInfo(test_file_path, test_file_info)
	fmt.Println(file_md5sum)
	fmt.Println(fileStatJson)
	fmt.Println(file_contents)

	t.Error(
		"For", "/tmp/test_file",
		"expected", "md5_sum: xxxx",
		"got", file_md5sum,
	)
}
