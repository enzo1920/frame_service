package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Set_cmds struct {
	Cmd_id     int `json:"cmd_id"`
	Cmd_status int `json:"cmd_status"`
}
type CmdsToExec struct {
	Cmd_id   int    `json:"cmd_id"`
	Cmd_name string `json:"cmd_name"`
}

func GetCommands(serv_url string, dev_name string) ([]CmdsToExec, error) {

	req, err := http.NewRequest("GET", serv_url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Host", dev_name)

	client := &http.Client{}

	r, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	body, _ := ioutil.ReadAll(r.Body)
	fmt.Println("Body", string(body))

	var ce []CmdsToExec
	json.Unmarshal(body, &ce)
	fmt.Printf("cmds : %+v\n", ce)
	/*for _, v := range ce {
		fmt.Println(v.Cmd_id, v.Cmd_name)
	}*/

	return ce, nil
	//fmt.Println("response Body cmds:", string(cmds))
}

func SetCommandStatus(serv_url string, cmd_id int, cmd_status int) error {

	set_command := &Set_cmds{}
	set_command.Cmd_id = cmd_id
	set_command.Cmd_status = cmd_status
	jsonDisk, err := json.Marshal(set_command)
	if err != nil {
		//fmt.Println(err)
		return err
	}

	fmt.Println("cmd_set json:", string(jsonDisk))

	req, err := http.NewRequest("POST", serv_url, bytes.NewBuffer(jsonDisk))
	req.Header.Set("X-Custom-Header", "cmd_set")
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
