package utils

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"../readconfig"
)

type CamState struct {
	CamIp    string `json:"cam_ip"`
	CamState int    `json:"cam_state"`
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

func PingCmr(serv_url string, cfg readconfig.Configuration) error {
	cmd := "ping"

	//Common Channel for the goroutines
	//tasks := make(chan *exec.Cmd, 64)
	//tasks := make(chan *map[string]int)
	//Spawning  goroutines
	//var wg sync.WaitGroup
	//for _, i := range cfg.Cameras_block.Cameras_address {
	//	wg.Add(1)
	//	go RunPing(i.Value, tasks, &wg)
	//}

	fmt.Println("cameras count is ", len(cfg.Cameras_block.Cameras_address))
	//Generate Tasks
	cmr_status := []*CamState{}
	for _, i := range cfg.Cameras_block.Cameras_address {
		cstate := &CamState{}
		out, err := exec.Command(cmd, "-c3", i.Value).Output()
		cstate.CamIp = i.Value
		if err != nil {
			log.Printf("can't get stdout:%v", err)
		}
		if strings.Contains(string(out), "100% packet loss") {

			cstate.CamState = 1 //1-down
		} else {
			cstate.CamState = 2 //2-alive
		}
		cmr_status = append(cmr_status, cstate)

	}
	jsonCmrStat, err := json.Marshal(cmr_status)
	if err != nil {
		fmt.Println(err)
		//return err
	}
	fmt.Println("cameras_state_json:", string(jsonCmrStat))

	req, err := http.NewRequest("POST", serv_url, bytes.NewBuffer(jsonCmrStat))
	req.Header.Set("X-Custom-Header", "cam_state")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)

	fmt.Printf("\n=========ping done=====>>>>\n")
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

func RunPing(camera_ip string, tasks chan *exec.Cmd, w *sync.WaitGroup) {
	defer w.Done()
	var (
		out []byte
		err error
	)
	cmr_status := map[string]int{}
	for cmd := range tasks { // this will exit the loop when the channel closes

		//cmrs := CameraStatus{} //struct for export to server

		out, err = cmd.Output()
		if err != nil {
			log.Printf("can't get stdout:%v", err)
		}
		if strings.Contains(string(out), "100% packet loss") {
			cmr_status[camera_ip] = 1 //1-down
		} else {
			cmr_status[camera_ip] = 2 //2-alive
		}

	}
	jsonCmrStat, err := json.Marshal(cmr_status)
	if err != nil {
		fmt.Println(err)
		//return err
	}
	fmt.Println("cameras_state_json:", string(jsonCmrStat))

}
