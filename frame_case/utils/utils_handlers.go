package utils

import (
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"os"
	"path"
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
