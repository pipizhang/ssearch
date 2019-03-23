package main

import (
	_md5 "crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
)

func md5(text string) string {
	hashMd5 := _md5.New()
	io.WriteString(hashMd5, text)
	return fmt.Sprintf("%x", hashMd5.Sum(nil))
}

func readFile(file string) []byte {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		log.Println("[error] Read error")
	}
	return content
}

func fileExist(file string) bool {
	if s, err := os.Stat(file); (err == nil || os.IsExist(err)) && !s.IsDir() {
		return true
	}
	return false
}

func dirExist(dir string) bool {
	if s, err := os.Stat(dir); err == nil && s.IsDir() {
		return true
	}
	return false
}
