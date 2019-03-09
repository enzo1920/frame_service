package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"./readconfig"
	"path/filepath"
)

type User struct {
	Login    string `json:"username"`
	Password string `json:"password"`
}

type Cmd struct {
	Login string `json:"username"`
}

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


func upload(serv_url string, filename string) {

	file, err := os.Open("./upload/"+filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	client := &http.Client{}
	req, err := http.NewRequest("POST", serv_url+"/upload", file)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Content-Disposition", "attachment; filename="+filename)
	resp, err := client.Do(req)
	defer resp.Body.Close()
	fmt.Println(resp.Status)

}


func Getfilesdir()([]string){

	files_to_upload := make([]string, 0)
	dirname := "./upload" + string(filepath.Separator)
	d, err := os.Open(dirname)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer d.Close()

	files, err := d.Readdir(-1)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("Reading "+ dirname)

	for _, file := range files {
		if file.Mode().IsRegular() {
			if filepath.Ext(file.Name()) == ".jpeg" {
				files_to_upload = append(files_to_upload, file.Name())
			}
		}
	}
return files_to_upload
}

func main() {
	//login()
	readcfg := readconfig.Config_reader("./readconfig/frame_case.conf")
	server_url :="http://"+readcfg.Host+":"+strconv.Itoa(readcfg.Port)
	device :=readcfg.Devicename
	GetCommands(server_url, device)

	files := Getfilesdir()
	for _, filename := range files{
		fmt.Println("files in dir is:", filename)
		upload(server_url, filename)
	}
}
