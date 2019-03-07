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
func GetCommands(cfg readconfig.Configuration) {
	user := &User{Login: cfg.Devicename}
	jsonStr, err := json.Marshal(user)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(jsonStr))

	url := "http://" + cfg.Host + ":" + strconv.Itoa(cfg.Port) + "/v1/commands/?device=" + cfg.Devicename
	fmt.Println("URL:>", url)

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

// it's simple upload func let's do it!
func upload(cfg readconfig.Configuration) {
	file, err := os.Open("./upload/Cotier17_2017-07-17_13-22-02.jpeg")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	res, err := http.Post("http://"+cfg.Host+":"+strconv.Itoa(cfg.Port)+"/upload", "binary/octet-stream", file)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	message, _ := ioutil.ReadAll(res.Body)
	fmt.Printf(string(message))
}

func main() {
	//login()
	readcfg := readconfig.Config_reader("./readconfig/frame_case.conf")
	GetCommands(readcfg)
	upload(readcfg)
}
