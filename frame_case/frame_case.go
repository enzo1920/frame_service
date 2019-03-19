package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"./readconfig"
	"./utils"
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

type Ip_stream struct {
	Ip     []string
	Type   string
	Stream string
}

type Api_Url struct {
	Command string
	Url     string
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

	file, err := os.Open(path.Join("./upload/", filename))
	if err != nil {
		panic(err)
	}
	fmt.Println("open file to send:", filename)
	defer file.Close()

	client := &http.Client{}
	req, err := http.NewRequest("POST", serv_url+"/upload", file)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Content-Disposition", "form-data; name="+dev_name+"; filename="+filename)
	resp, err := client.Do(req)
	defer resp.Body.Close()
	fmt.Println(resp.Status)

}
func GetUrlFromConfig(cfg readconfig.Configuration) {
	//api_token := cfg.Api_token
	//api_url := cfg.Api_Urls
	//dict := map[int]*Api_Url{}
	for _, cmd_url := range cfg.Api_Urls {
		fmt.Println("cmd_url", cmd_url)
	}

}

// func for formating rtsp-stream from config
func FormatCommands(cfg readconfig.Configuration) []string {

	t := time.Now().Format("2006-01-02_15-04-05")
	commands_capture := make([]string, 0)

	dictionary := map[int]*Ip_stream{}

	for _, i := range cfg.Cameras_block.Cameras_stream {
		ipStream := Ip_stream{
			Stream: i.Stream}
		dictionary[i.Cameras_type] = &ipStream
	}
	for _, i := range cfg.Cameras_block.Cameras_type {
		obj := dictionary[i.Type]
		obj.Type = i.Value
	}

	for _, i := range cfg.Cameras_block.Cameras_address {
		obj := dictionary[i.Type]
		obj.Ip = append(obj.Ip, i.Value)
	}

	for _, x := range dictionary {
		//fmt.Println(x)
		for _, ip := range x.Ip {
			sa := strings.Split(ip, ".")
			//get addr last byte of ip
			camera_name := strings.Title(x.Type) + sa[3]
			//fmt.Println(camera_name)
			stream := strings.Replace(x.Stream, "{ip}", ip, 1)
			stream = strings.Replace(stream, "{port}", "554", 1)
			stream = stream + " ./upload/" + camera_name + "_" + t + ".jpeg"
			//fmt.Println(stream)
			commands_capture = append(commands_capture, stream)
		}

	}

	for _, cmd := range commands_capture {
		fmt.Println("cmd:", cmd)
	}
	return commands_capture
}

func Overlay_fonter(file string) {
	file_path := path.Join("./upload", file)
	compose := "composite?-gravity center ?/home/src_img/fonter.png ?" + file_path + " ?" + file_path
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go exe_cmd(compose, wg)
	wg.Wait()
}

func Overlay_info(file string) {
	file_path := path.Join("./upload", file)
	camera_name := "test_camera"
	device_name := "test_device"
	temp := "36,6*C"

	//convert_commands := []string{"convert " + file_path + "  -pointsize 20 -draw 'gravity southeast fill yellow  text 20,8 " + camera_name + "' " + file_path}
	convert_commands := []string{"composite?-gravity?center?/home/src_img/fonter.png?" + file_path + "?" + file_path,
		"convert?" + file_path + "?-pointsize?20?-gravity?Southeast?-fill?yellow?-draw?text 20,8 '" + camera_name + "'?" + file_path,
		"convert?" + file_path + "?-pointsize?20?-gravity?Southwest?-fill?yellow?-draw?text 20,8 '" + device_name + "'?" + file_path,
		"convert?" + file_path + "?-pointsize?20?-gravity?South?-fill?yellow?-draw?text 20,8 '" + temp + "'?" + file_path}
	//wg := new(sync.WaitGroup)
	//wg.Add(len(convert_commands))
	for _, str := range convert_commands {

		//fmt.Println("convert cmd-----------------", str)
		exe_cmd_one(str)
	}
	//wg.Wait()
}

func Getfilesdir() []string {

	files_to_upload := make([]string, 0)
	dirname := path.Join("./upload", string(filepath.Separator))
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

func exe_cmd(cmd string, wg *sync.WaitGroup) {

	// splitting head => g++ parts => rest of the command
	parts := strings.Split(cmd, "?")
	head := parts[0]
	args := parts[1:len(parts)]
	fmt.Println("parts is ", parts)
	cmd_exec := exec.Command(head, args...)
	//	Sanity check -- capture stdout and stderr:
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd_exec.Stdout = &out
	cmd_exec.Stderr = &stderr

	//	Run the command
	cmd_exec.Run()

	//	Output our results
	//fmt.Printf("Result: %v / %v", out.String(), stderr.String())
	wg.Done() // Need to signal to waitgroup that this goroutine is done
}

func exe_cmd_one(cmd string) {

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
	err := cmd_exec.Run()
	if err != nil {
		fmt.Println(err)
	}

	//	Output our results
	//fmt.Printf("Result: %v / %v", out.String(), stderr.String())

}

func main() {
	//login()

	fmt.Println("max parallelism is:", utils.MaxParallelism())
	readcfg := readconfig.Config_reader("./readconfig/frame_case.conf")
	server_url := "http://" + readcfg.Connection.Host + ":" + strconv.Itoa(readcfg.Connection.Port)
	device := readcfg.Connection.Devicename
	token := readcfg.Api_token
	api_urls := readcfg.Api_Urls
	fmt.Println("api urls:", api_urls)
	fmt.Println("api token:", token)
	GetUrlFromConfig(readcfg)
	GetCommands(server_url, device)
	/*
		//capture images and send
		   	formated_cmds := FormatCommands(readcfg)

		   	wg := new(sync.WaitGroup)
		       for _, str := range formated_cmds {
		           wg.Add(1)
		           go exe_cmd(str, wg)
		       }
		       wg.Wait()
	*/

	files := Getfilesdir()
	for _, filename := range files {
		fmt.Println("files in dir is:", filename)
		//Overlay_fonter(filename)
		Overlay_info(filename)
		UploadImage(server_url, device, filename)
	}

	DiskUsage(server_url, device, "/")

}
