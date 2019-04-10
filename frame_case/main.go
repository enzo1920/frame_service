package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"./models"
	"./readconfig"
	"./utils"
	"github.com/enzo1920/frame_service/frame_case/version"
)

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

type CameraStatus struct {
	cmrIp     string `json:"camera_ip"`
	cmrStatus int    `json:"camera_status"`
}

const (
	B  = 1
	KB = 1024 * B
	MB = 1024 * KB
	GB = 1024 * MB
)

func check(e error) {
	if e != nil {
		log.Println("error ", e)
	}
}

func UploadImage(serv_url string, dev_name string, filename string) error {

	file, err := os.Open(path.Join("./upload/", filename))
	if err != nil {
		return err
	}
	//fmt.Println("open file to send:", filename)
	defer file.Close()

	client := &http.Client{}
	fmt.Println("post", serv_url)
	req, err := http.NewRequest("POST", serv_url, file)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Disposition", "form-data; name="+dev_name+"; filename="+filename)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	fmt.Println(resp.Status)
	return nil

}
func GetUrlFromConfig(cfg readconfig.Configuration, cmd_ret_url string) string {
	api_token := cfg.Api_token
	//api_url := cfg.Api_Urls
	//dict := map[int]*Api_Url{}
	default_cmd := "/"
	dictionary := map[int]*Api_Url{}

	for i, cmd_url := range cfg.Api_Urls {
		dictionary[i] = &Api_Url{cmd_url.Api_command, strings.Replace(cmd_url.Url, "{api_token}", api_token, 1)}

		//fmt.Println("cmd_url", cmd_url.Api_command, strings.Replace(cmd_url.Url, "{api_token}", api_token, 1))
	}
	for _, ds := range dictionary {
		if cmd_ret_url == ds.Command {
			fmt.Println("dict struct:", ds.Command, ds.Url)
			return ds.Url
		}

	}
	return default_cmd

}

// func for formating rtsp-stream from config
func FormatCommands(cfg readconfig.Configuration) ([]string, error) {

	t := time.Now().Format("2006-01-02_15-04-05")
	commands_capture := make([]string, 0)

	dictionary := map[int]*Ip_stream{}

	for _, i := range cfg.Cameras_block.Cameras_stream {
		ipStream := Ip_stream{Stream: i.Stream}
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
	return commands_capture, nil
}

func PingCmr(cfg readconfig.Configuration) error {

	cmr_status := map[int]*CameraStatus{}
	for k, i := range cfg.Cameras_block.Cameras_address {
		cmrs := CameraStatus{}
		ping_cmd := "ping?-c6?" + i.Value
		fmt.Println("camer ip is", i.Value)
		out, _ := exe_cmd_one(ping_cmd)
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
	jsonCmrStat, err := json.Marshal(cmr_status)
	if err != nil {
		//fmt.Println(err)
		return err
	}
	fmt.Println("cameras_state_ json:", jsonCmrStat)
	return nil
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

/*
func Montage_img(cmrName string, cmrDate string, cmrHour string) (err error) {
	startDir := "/mnt/flash/img/" + cmrDate
	pattern := cmrDate + "_" + cmrHour
	findedfiles := make([]string, 0)
	files := Getfilesdir(startDir)
	for _, filename := range files {
		if strings.Contains(filename, pattern) {
			findedfiles = append(findedfiles, startDir+"/"+filename)
		}
	}
	fmt.Println(strings.Join(findedfiles[:], ","))

	montageFile := cmrName + "_mont_" + cmrDate + "_" + cmrHour + ".jpeg"
	cmdToMontage := "montage?`find ./upload -type f -name *" + cmrDate + "_" + cmrHour + "*.jpeg -not -name mini_*`?-geometry?640x360+2+2?-background?yellow?" + montageFile
	fmt.Println("montage:", cmdToMontage)
	exe_cmd_one(cmdToMontage)
	return nil
}*/

func DiskUsage(serv_url string, device_name string, path string) error {
	fs := syscall.Statfs_t{}
	err := syscall.Statfs(path, &fs)
	if err != nil {
		return err
	}
	disk := &DiskStatus{}
	disk.Device = device_name
	disk.Disk_part = path
	disk.All = fs.Blocks * uint64(fs.Bsize)
	disk.Free = fs.Bfree * uint64(fs.Bsize)
	disk.Used = disk.All - disk.Free

	jsonDisk, err := json.Marshal(disk)
	if err != nil {
		//fmt.Println(err)
		return err
	}
	fmt.Println("fs_state_ json:", string(jsonDisk))

	req, err := http.NewRequest("POST", serv_url, bytes.NewBuffer(jsonDisk))
	req.Header.Set("X-Custom-Header", "myvalue")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)

	return nil
}

func exe_cmd(cmd string, wg *sync.WaitGroup) {

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

func exe_cmd_one(cmd string) (string, string) {

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

func main() {

	//logging
	log_dir := "./log"
	if _, err := os.Stat(log_dir); os.IsNotExist(err) {
		os.Mkdir(log_dir, 0644)
	}
	file, err := os.OpenFile("./log/framecase.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	log.SetOutput(file)
	log.Println("Logging to a file in Go!")

	log.Printf(
		"Starting the service...\ncommit: %s, build time: %s, release: %s",
		version.Commit, version.BuildTime, version.Release,
	)

	log.Println("max parallelism is:", utils.MaxParallelism())
	readcfg := readconfig.Config_reader("./readconfig/frame_case.conf")
	server_url := "http://" + readcfg.Connection.Host + ":" + strconv.Itoa(readcfg.Connection.Port)
	device := readcfg.Connection.Devicename
	token := readcfg.Api_token
	api_urls := readcfg.Api_Urls
	url_upload := GetUrlFromConfig(readcfg, "upload")
	url_command := GetUrlFromConfig(readcfg, "getcommand")
	url_voluminfo := GetUrlFromConfig(readcfg, "volumeinfo")
	url_setcmd := GetUrlFromConfig(readcfg, "setcmd")
	fmt.Println("api urls:", api_urls)
	fmt.Println("api token:", token)
	fmt.Println("api ret url:", url_command)
	PingCmr(readcfg)
	// while true loop
	for {
		log.Println("Starting connect to server")
		//получаем и сразу выставляем статус
		gcmds, err := models.GetCommands(server_url+url_command, device)
		if err != nil {
			log.Println("error GetCommands", err)
		}
		for _, v := range gcmds {
			fmt.Println(v.Cmd_id, v.Cmd_name)
			err_set := models.SetCommandStatus(server_url+url_setcmd, v.Cmd_id, 2)
			if err_set != nil {
				log.Println("error Set command", err_set)
			}
		}

		/*
			//capture images and send
				formated_cmds, err_cmds := FormatCommands(readcfg)
				check(err_cmds)

				wg := new(sync.WaitGroup)
				for _, str := range formated_cmds {
					wg.Add(1)
					go exe_cmd(str, wg)
				}
				wg.Wait()
		*/
		files := utils.Getfilesdir("./upload")
		for _, filename := range files {
			//fmt.Println("files in dir is:", filename)
			//Overlay_fonter(filename)
			Overlay_info(filename)
			err := UploadImage(server_url+url_upload, device, filename)
			if err != nil {
				fmt.Println("error UploadImage", err)
			}
		}

		err1 := DiskUsage(server_url+url_voluminfo, device, "/")
		if err1 != nil {
			fmt.Println("error DiskUsage", err1)
		}

		err_rem := utils.RemoveOldThanXX(1)
		if err_rem != nil {
			fmt.Println("error RemoveOldThanXX", err_rem)
		}

		/*err_mnt := Montage_img("Besder24", "2019-03-17", "22")
		check(err_mnt)*/

		time.Sleep(10 * time.Second)
	}

}
