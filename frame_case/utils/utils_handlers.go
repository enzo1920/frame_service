package utils

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"../readconfig"
)

type CameraStatus struct {
	cmrIp     string `json:"camera_ip"`
	cmrStatus int    `json:"camera_status"`
}

type Worker struct {
	Command string
	Output  chan string
	Outerr  chan string
}

func (cmd *Worker) Run() {

	parts := strings.Split(cmd.Command, "?")
	head := parts[0]
	args := parts[1:len(parts)]
	fmt.Printf("programm is: %v params is: %v \n", head, args)

	cmdExec, err := exec.Command(head, args...).Output()
	if err != nil {
		log.Fatal(err)
	}
	/*
		var out bytes.Buffer
		var stderr bytes.Buffer
		cmdExec.Stdout = &out
		cmdExec.Stderr = &stderr
		fmt.Printf("\nResult: %v / %v", out.String(), stderr.String())*/

	cmd.Output <- string(cmdExec)
	cmd.Outerr <- string("111")
}

func Collect(c chan string, e chan string) {
	for {
		msg := <-c
		err := <-e
		fmt.Printf("The command result is %v error is %v \n", msg, err)
	}
}

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

func PingCmr(cfg readconfig.Configuration) error {
	cmd := "ping"

	//Common Channel for the goroutines
	tasks := make(chan *exec.Cmd, 64)

	//Spawning  goroutines
	var wg sync.WaitGroup
	for _, i := range cfg.Cameras_block.Cameras_address {
		wg.Add(1)
		go Run(i.Value, tasks, &wg)
	}

	//cmr_status := map[int]*CameraStatus{}

	fmt.Println("cameras count is ", len(cfg.Cameras_block.Cameras_address))
	//Generate Tasks
	for _, i := range cfg.Cameras_block.Cameras_address {
		tasks <- exec.Command(cmd, "-c3", i.Value)
	}
	close(tasks)

	// wait for the workers to finish
	wg.Wait()

	fmt.Printf("\n=========ping done=====>>>>\n")

	/*

		for k, i := range cfg.Cameras_block.Cameras_address {
			cmrs := CameraStatus{}
			ping_cmd := "ping?-c6?" + i.Value
			fmt.Println("Camera ip is", i.Value)

			if strings.Contains(out, "100% packet loss") {
				fmt.Println("Camera DOWN")
				cmrs.cmrIp = i.Value
				cmrs.cmrStatus = 1
			} else {
				cmrs.cmrIp = i.Value
				cmrs.cmrStatus = 2
				fmt.Println("Camera is  ALIVEEE")
			}
			cmr_status[k] = &cmrs
		}

		//test print
		for _, v := range cmr_status {
			fmt.Printf("\nCamera %s  state is %d", v.cmrIp, v.cmrStatus)
		}

		jsonCmrStat, err := json.Marshal(cmr_status)
		if err != nil {
			//fmt.Println(err)
			return err
		}
		fmt.Println("cameras_state_ json:", string(jsonCmrStat))*/
	return nil
}

//exec external command

func ExecCmd(cmd string, wg *sync.WaitGroup) {
	// splitting head => g++ parts => rest of the command
	parts := strings.Split(cmd, "?")
	head := parts[0]
	args := parts[1:len(parts)]
	//fmt.Println("parts is ", parts)
	cmd_exec := exec.Command(head, args...)
	//	Sanity check -- capture stdout and stderr:
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd_exec.Stdout = &out
	cmd_exec.Stderr = &stderr

	//	Run the command
	cmd_exec.Run()
	wg.Done() // Need to signal to waitgroup that this goroutine is done
}

func ExecCmdone(cmd string) (string, string) {

	// splitting head => g++ parts => rest of the command
	parts := strings.Split(cmd, "?")
	head := parts[0]
	args := parts[1:len(parts)]
	//fmt.Printf("programm is: %v params is: %v ", head, args)
	cmd_exec := exec.Command(head, args...)
	//	Sanity check -- capture stdout and stderr:
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd_exec.Stdout = &out
	cmd_exec.Stderr = &stderr

	//	Run the command
	err := cmd_exec.Run()
	if err != nil {
		log.Println(err)
	}

	//	Output our results
	//fmt.Printf("\nResult: %v / %v", out.String(), stderr.String())
	return out.String(), stderr.String()

}

func Run(camera string, tasks chan *exec.Cmd, w *sync.WaitGroup) {
	defer w.Done()
	var (
		out []byte
		err error
	)
	for cmd := range tasks { // this will exit the loop when the channel closes
		out, err = cmd.Output()
		if err != nil {
			log.Printf("can't get stdout:%v", err)
		}
		//fmt.Printf("goroutine %d command output:%s", num, string(out))
		if strings.Contains(string(out), "100% packet loss") {
			fmt.Printf("\nHost DOWN  %s", camera)

		} else {

			fmt.Printf("\nHost is  ALIVEEE  %s", camera)
		}

	}
}
