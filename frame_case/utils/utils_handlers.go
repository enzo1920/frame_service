package utils

import (
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"time"
)

func TokenGenerator() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

// remove dirs older than xx days

func isOlderThanXXDay(t time.Time, multiplier int) bool {
	//multiplier -24 one day
	multiplier = 24 * multiplier
	fmt.Println("duration in hours", time.Duration(multiplier)*time.Hour)
	return time.Now().Sub(t) > time.Duration(multiplier)*time.Hour
}

func findDirsOlderThanXXDay(startDir string, multip int) (fidnDirs []os.FileInfo, err error) {
	tmpdirs, err := ioutil.ReadDir(startDir)
	if err != nil {
		return
	}
	for _, dir := range tmpdirs {
		if dir.IsDir() {
			if isOlderThanXXDay(dir.ModTime(), multip) {
				fidnDirs = append(fidnDirs, dir)
			}
		}

	}
	return
}

func RemoveOldThanXX(multip int) (err error) {
	startDir := "/mnt/flash/img"
	fdirs, _ := findDirsOlderThanXXDay(startDir, multip)
	for _, dir := range fdirs {
		fmt.Println(dir.Name())
		os.RemoveAll(path.Join(startDir, dir.Name()))

	}
	return
}

func Getfilesdir(startDir string) []string {

	files_to_upload := make([]string, 0)
	dirname := path.Join(startDir, string(filepath.Separator))
	d, err := os.Open(dirname)
	if err != nil {
		panic(err)
	}
	defer d.Close()

	files, err := d.Readdir(-1)
	if err != nil {
		panic(err)
	}

	fmt.Println("Reading " + dirname)

	for _, file := range files {
		if file.Mode().IsRegular() {
			if filepath.Ext(file.Name()) == ".jpeg" {
				files_to_upload = append(files_to_upload, file.Name())
			}
		}
	}
	return files_to_upload
}
