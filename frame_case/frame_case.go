package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"syscall"

	"./readconfig"
)

type User struct {
	Login    string `json:"username"`
	Password string `json:"password"`
}

type Cmd struct {
	Login string `json:"username"`
}

type DiskStatus struct {
	Device    string `json:"device"`
	Disk_part string `json:"disk_part"`
	All       uint64 `json:"all"`
	Used      uint64 `json:"used"`
	Free      uint64 `json:"free"`
}

const (
	B  = 1
	KB = 1024 * B
	MB = 1024 * KB
	GB = 1024 * MB
)

/*
func login() {
	user := &User{Login: "****", Password: "****"}
	jsonStr, err := json.Marshal(user)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(jsonStr))

	url := "http://host:8080/signin"
	fmt.Println("URL:>", url)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("X-Custom-Header", "myvalue")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	fmt.Println("req:>", req)

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
}
*/
//
func GetCommands(serv_url string, dev_name string) {
	user := &User{Login: dev_name}
	jsonStr, err := json.Marshal(user)
	if err != nil {
		fmt.Println(err)
		return
	}
	//fmt.Println(string(jsonStr))

	url := serv_url + "/v1/commands/?device=" + dev_name

	req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("X-Custom-Header", "myvalue")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)

	body, _ := ioutil.ReadAll(resp.Body)
	cmds := string(body)
	fmt.Println("response Body cmds:", string(cmds))
}

func UploadImage(serv_url string, dev_name string, filename string) {

	file, err := os.Open("./upload/" + filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	client := &http.Client{}
	req, err := http.NewRequest("POST", serv_url+"/upload", file)
	if err != nil {
		panic(err)
	}
	req.Header.Add("X-Custom-Header", dev_name)
	req.Header.Add("Content-Disposition", "attachment; filename="+filename)
	resp, err := client.Do(req)
	defer resp.Body.Close()
	fmt.Println(resp.Status)

}

func Getfilesdir() []string {

	files_to_upload := make([]string, 0)
	dirname := "./upload" + string(filepath.Separator)
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

func DiskUsage(serv_url string, device_name string, path string) {
	fs := syscall.Statfs_t{}
	err := syscall.Statfs(path, &fs)
	if err != nil {
		return
	}
	disk := &DiskStatus{}
	disk.Device = device_name
	disk.Disk_part = path
	disk.All = fs.Blocks * uint64(fs.Bsize)
	disk.Free = fs.Bfree * uint64(fs.Bsize)
	disk.Used = disk.All - disk.Free

	jsonDisk, err := json.Marshal(disk)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("fs_state_ json:", string(jsonDisk))

	url := serv_url + "/v1/diskstate"

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonDisk))
	req.Header.Set("X-Custom-Header", "myvalue")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)

	//return
}

func main() {
	//login()
	readcfg := readconfig.Config_reader("./readconfig/frame_case.conf")
	server_url := "http://" + readcfg.Host + ":" + strconv.Itoa(readcfg.Port)
	device := readcfg.Devicename
	GetCommands(server_url, device)

	files := Getfilesdir()
	for _, filename := range files {
		fmt.Println("files in dir is:", filename)
		UploadImage(server_url, device, filename)
	}
	DiskUsage(server_url, device, "/")
}
