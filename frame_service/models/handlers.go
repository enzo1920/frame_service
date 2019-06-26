package models

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
)

const (
	B  = 1
	KB = 1024 * B
	MB = 1024 * KB
	GB = 1024 * MB
)

// Create a struct that models the structure of a user, both in the request body, and in the DB
type Credentials struct {
	Password string `json:"password", db:"password"`
	Username string `json:"username", db:"username"`
}

type CamState struct {
	CamIp    string `json:"cam_ip"`
	CamState int    `json:"cam_state"`
}

type DiskStatus struct {
	Device    string `json:"device"`
	Disk_part string `json:"disk_part"`
	All       uint64 `json:"all"`
	Used      uint64 `json:"used"`
	Free      uint64 `json:"free"`
}

type CmdsToExec struct {
	Cmd_id   int    `json:"cmd_id"`
	Cmd_name string `json:"cmd_name"`
}

type DevTemp struct {
	Dev_name   string    `json:"dev_name"`
	Dev_temp string `json:"dev_temp"`
	Temp_date string `json:"temp_date"`
}
type DevDescr struct {
	Dev_name   string    `json:"dev_name"`
	Dev_descr string `json:"dev_descr"`
}

type Set_cmds struct {
	Cmd_id     int `json:"cmd_id"`
	Cmd_status int `json:"cmd_status"`
}

func FloatToString(input_num float64) string {
	// to convert a float number to a string
	return strconv.FormatFloat(input_num, 'f', 6, 64)
}

func SetCamState(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["token"]

	if !ok || len(keys[0]) < 1 {
		log.Println("Url Param 'token' is missing")
		return
	}
	token := string(keys[0])
	device_name := "TARS"
	device_id := 0
	err1 := db.QueryRow("select id, device_name from devices where dev_token=$1", token).Scan(&device_id, &device_name)
	if err1 != nil {
		// If there is an issue with the database, return a 500 error
		fmt.Println("--->", err1)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	body, _ := ioutil.ReadAll(r.Body)
	var camstate []*CamState
	/*err := json.NewDecoder(r.Body).Decode(camstate)
	if err != nil {
		fmt.Println("err to decode ", err)
		// If there is something wrong with the request body, return a 400 status
		w.WriteHeader(http.StatusBadRequest)
		return
	}*/

	json.Unmarshal(body, &camstate)
	for _, v := range camstate {
		fmt.Println(v.CamIp, v.CamState)

		rows, err := db.Query("update cameras_addr set cmr_status=$1 where ip_addr=$2", v.CamState, v.CamIp)
		if err != nil {
			// If there is any issue with inserting into the database, return a 500 error
			log.Println("update cameras_addr", err)
			//w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer rows.Close()

	}
	w.WriteHeader(http.StatusOK)
}

func GetCameraState(w http.ResponseWriter, r *http.Request) {

	cam_type := 0
	rows, err := db.Query("SELECT ip_addr FROM cameras_addr WHERE cmr_type=$1", cam_type)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	ip_addrs := make([]string, 0)

	for rows.Next() {
		var ip_addr string
		if err := rows.Scan(&ip_addr); err != nil {
			// Check for a scan error.
			// Query rows will be closed with defer.
			log.Fatal(err)
		}
		ip_addrs = append(ip_addrs, ip_addr)
	}
	// If the database is being written to ensure to check for Close
	// errors that may be returned from the driver. The query may
	// encounter an auto-commit error and be forced to rollback changes.
	rerr := rows.Close()
	if rerr != nil {
		log.Fatal(err)
	}

	// Rows.Err will report the last error encountered by Rows.Scan.
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(w, "%s are %d ", strings.Join(ip_addrs, ", "), cam_type)
}

func GetCommands(w http.ResponseWriter, r *http.Request) {

	keys, ok := r.URL.Query()["token"]

	if !ok || len(keys[0]) < 1 {
		log.Println("Url Param 'device' is missing")
		return
	}

	// Query()["key"] will return an array of items,
	// we only want the single item.
	dev_token := string(keys[0])
	rows, err := db.Query("SELECT ce.id, cmd_name FROM commands as c inner join commands_ex as ce on c.id=ce.cmd_id"+
		" INNER join devices as d on d.id= ce.device_id"+
		" INNER join command_status cs on cs.id=ce.status_id WHERE d.dev_token =$1 and  ce.status_id < 2", dev_token)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	//собираем команды в массив структур

	//cmds := make([]string, 0)
	var commands []CmdsToExec
	for rows.Next() {
		var ce CmdsToExec
		if err := rows.Scan(&ce.Cmd_id, &ce.Cmd_name); err != nil {
			// Check for a scan error.
			// Query rows will be closed with defer.
			log.Fatal(err)
		}
		//cmds = append(cmds, cmd)
		commands = append(commands, ce)
	}

	sendcommans, err := json.Marshal(commands)
	if err != nil {
		//fmt.Println(err)
		fmt.Println(err)
	}
	fmt.Println("sendcommands json:", string(sendcommans))
	// If the database is being written to ensure to check for Close
	// errors that may be returned from the driver. The query may
	// encounter an auto-commit error and be forced to rollback changes.
	rerr := rows.Close()
	if rerr != nil {
		log.Fatal(err)
	}

	// Rows.Err will report the last error encountered by Rows.Scan.
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(sendcommans)

}

//reset relay handler
func GetResetRelay(w http.ResponseWriter, r *http.Request) {

	keys, ok := r.URL.Query()["token"]

	if !ok || len(keys[0]) < 1 {
		log.Println("Url Param 'device' is missing")
		return
	}

	// Query()["key"] will return an array of items,
	// we only want the single item.
	dev_token := string(keys[0])

	relayEnable := 0
	err1 := db.QueryRow("select relay_enbl from dev_relay dr inner join devices d on dr.dev_id=d.id where d.dev_token=$1", dev_token).Scan(&relayEnable)
	if err1 != nil {
		// If there is an issue with the database, return a 500 error
		log.Println("GetResetRelay err", err1)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	//w.Write(relayEnable)
	w.Write([]byte(fmt.Sprintf(string(relayEnable))))

}

func GetImg(w http.ResponseWriter, r *http.Request) {
	//from request get datime and device name
	//return urls of files to web interface
	//fmt.Println("get img--->")
	//fmt.Println("GET params were:", r.URL.Query())
	//devname := r.URL.Query().Get("dev_name")
	//fmt.Println("GET devname were:", devname)

	file, err := ioutil.ReadFile("uploaded/Cotier20.jpeg")
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-type", "image/jpeg")
	w.Write(file)

}
//get devices with description only rasp
func GetDevices(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("select device_name, dev_descr from devices d "+
		" INNER join  devices_descr dcr on d.id=dcr.dev_id"+
		" INNER join  dev_types dt on dt.id= d.dev_type  where device_status=true and dev_type=1")
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()

    var devicesdescr []DevDescr
	for rows.Next() {
		var dd DevDescr
		if err := rows.Scan(&dd.Dev_name, &dd.Dev_descr); err != nil {

			log.Println(err)
		}
		devicesdescr = append(devicesdescr, dd)
	}

	ddjson, err := json.Marshal(devicesdescr)
	if err != nil {
		log.Println(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(ddjson)

}


//get dev folders
func GetDevFolders(w http.ResponseWriter, r *http.Request) {
	devname := strings.ToUpper(r.URL.Query().Get("device"))
	cntdev := 0
	err := db.QueryRow("select count(*) from devices where dev_name=$1",devname).Scan(&cntdev)
	if err != nil {
		log.Println(err)
	}

    if cntdev == 0 {
		log.Println("GET: no devices found in DB")
	}

    deviceDir, err := ioutil.ReadDir("uploaded/"+devname+"/")
	if err != nil {
		return
	}
	var fidnDirs []string
	for _, dir := range deviceDir {
		if dir.IsDir() {
			fidnDirs = append(fidnDirs, dir.Name())
		}

	}
	foldersjson, err := json.Marshal(fidnDirs)
	if err != nil {
		log.Println(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(foldersjson)
}


// получаем json список картинок в папке
func GetImgFromFolders(w http.ResponseWriter, r *http.Request) {
	devname := strings.ToUpper(r.URL.Query().Get("device"))
	folder := strings.ToUpper(r.URL.Query().Get("folder"))
	cntdev := 0
	err := db.QueryRow("select count(*) from devices where dev_name=$1",devname).Scan(&cntdev)
	if err != nil {
		log.Println(err)
	}

    if cntdev == 0 {
		log.Println("GET: no devices found in DB")
	}

	var filesindir []string
	d, err := os.Open("uploaded/"+devname+"/"+folder+"/")
	if err != nil {
		log.Println("error open dir in GetImgFromFolders",err)
	}
	defer d.Close()

	files, err := d.Readdir(-1)
	if err != nil {
		log.Println("error read dir in GetImgFromFolders",err)
	}

	fmt.Println("Reading " +"uploaded/"+devname+"/"+folder+"/")

	for _, file := range files {
		if file.Mode().IsRegular() {
			if filepath.Ext(file.Name()) == ".jpeg" {
				filesindir = append(filesindir, file.Name())
			}
		}
	}

	
	filesjson, err := json.Marshal(filesindir)
	if err != nil {
		log.Println(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(filesjson)
}


func GetImgByName(w http.ResponseWriter, r *http.Request) {
	devname := r.URL.Query().Get("device")
	folder := r.URL.Query().Get("folder")
	picture := r.URL.Query().Get("picture")
	//fmt.Println("need to open:","uploaded/"+devname+"/"+folder+"/"+picture)

	file, err := ioutil.ReadFile("uploaded/"+devname+"/"+folder+"/"+picture)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-type", "image/jpeg")
	w.Write(file)

}



//get current temperature from db
func GetCurrentTemp(w http.ResponseWriter, r *http.Request) {
	var dt DevTemp
	err := db.QueryRow("SELECT d.device_name, ts.temp, to_char(ts.date_temp, 'yyyy-mm-dd HH24:MI:SS')  FROM temp_stat ts INNER JOIN devices d ON d.id=ts.dev_id ORDER BY ts.id desc  limit 1").Scan(&dt.Dev_name, &dt.Dev_temp,&dt.Temp_date)
	if err != nil {
		log.Fatal(err)
	}
	
	senddevtemps, err := json.Marshal(dt)
	if err != nil {
		//fmt.Println(err)
		log.Println(err)
	}
	//fmt.Println("senddevtemps json:", string(senddevtemps))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(senddevtemps)
}

func UploadVoltageHandler(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["token"]

	if !ok || len(keys[0]) < 1 {
		log.Println("Url Param 'token' is missing")
		return
	}
	token := string(keys[0])

	device_id := 0
	err1 := db.QueryRow("select id from devices where dev_token=$1", token).Scan(&device_id)
	if err1 != nil {
		// If there is an issue with the database, return a 500 error
		fmt.Println("--->", err1)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Println("dev_id upload temp:", device_id)
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)

	}
	volatge, _ := strconv.ParseFloat(string(body), 64)
	log.Printf("volt is:%v, token is %s\n", volatge, token)
	rows, err := db.Query("insert into  volt_stat values (DEFAULT,$1,$2, CURRENT_TIMESTAMP)", &device_id, &volatge)
	if err != nil {
		// If there is any issue with inserting into the database, return a 500 error
		log.Println(" insert volt_stat err", err)
		//w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer rows.Close()
}

func UploadTempHandler(w http.ResponseWriter, r *http.Request) {

	//fmt.Println("upload temp query:",r.URL.Query())
	keys, ok := r.URL.Query()["token"]

	if !ok || len(keys[0]) < 1 {
		log.Println("Url Param 'token' is missing")
		return
	}
	token := string(keys[0])

	device_id := 0
	err1 := db.QueryRow("select id from devices where dev_token=$1", token).Scan(&device_id)
	if err1 != nil {
		// If there is an issue with the database, return a 500 error
		fmt.Println("--->", err1)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Println("dev_id upload temp:", device_id)
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)

	}
	bodyString := string(body)
	//fmt.Printf("temp is:%s, token is %s\n", bodyString, token)
	log.Printf("temp is:%s, token is %s\n", bodyString, token)
	rows, err := db.Query("insert into  temp_stat values (DEFAULT,$1,$2, CURRENT_TIMESTAMP)", &device_id, &bodyString)
	if err != nil {
		// If there is any issue with inserting into the database, return a 500 error
		log.Println(" insert temp_stat err", err)
		//w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer rows.Close()
}

func SetCommandHandler(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["token"]

	if !ok || len(keys[0]) < 1 {
		log.Println("Url Param 'token' is missing")
		return
	}
	token := string(keys[0])

	device_id := 0
	err1 := db.QueryRow("select id from devices where dev_token=$1", token).Scan(&device_id)
	if err1 != nil {
		// If there is an issue with the database, return a 500 error
		fmt.Println("--->", err1)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Println("dev_id set_cmd_status:", device_id)

	cmd_status := &Set_cmds{}
	err := json.NewDecoder(r.Body).Decode(cmd_status)
	if err != nil {
		// If there is something wrong with the request body, return a 400 status
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Printf("cmd is:%s, status is %s\n", cmd_status.Cmd_id, cmd_status.Cmd_status)
	rows, err := db.Query("update commands_ex set status_id=$1 where id=$2 and device_id=$3", &cmd_status.Cmd_status, &cmd_status.Cmd_id, &device_id)
	if err != nil {
		// If there is any issue with inserting into the database, return a 500 error
		log.Println("update commands_exerr", err)
		//w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer rows.Close()
}

//upload image handler
func UploadHandler(w http.ResponseWriter, r *http.Request) {

	keys, ok := r.URL.Query()["token"]

	if !ok || len(keys[0]) < 1 {
		log.Println("Url Param 'token' is missing")
		return
	}

	contentDisposition := r.Header.Get("Content-Disposition")
	_, params, err := mime.ParseMediaType(contentDisposition)
	filename := params["filename"] // get filename from header
	token := string(keys[0])
	//get dev_id from db
	device_name := "TARS"
	device_id := 0
	err1 := db.QueryRow("select id, device_name from devices where dev_token=$1", token).Scan(&device_id, &device_name)
	if err1 != nil {
		// If there is an issue with the database, return a 500 error
		fmt.Println("--->", err1)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Println("dev_name:", device_name)
	fmt.Println("select dev_id is ok!")
	//work with remote addr
	ip, port, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		//return nil, fmt.Errorf("userip: %q is not IP:port", req.RemoteAddr)

		log.Println("userip: %q is not IP:port", r.RemoteAddr)
	}
	userIP := net.ParseIP(ip)
	if userIP == nil {
		//return nil, fmt.Errorf("userip: %q is not IP:port", req.RemoteAddr)
		log.Println("userip: %q is not IP:port \n", r.RemoteAddr)
		return
	}
	rows, err := db.Query("insert into  dev_upload_log values (DEFAULT,$1,$2,$3,$4, CURRENT_TIMESTAMP)", device_id, string(ip), port, filename)
	if err != nil {
		// If there is any issue with inserting into the database, return a 500 error
		log.Println(" insert dev_upload_log err", err)
		//w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer rows.Close()
	//devicename := r.Header.Get("X-Custom-Header")// get devname from custom header
	//fmt.Printf("Custom header is: %s\n", devicename)

	//devicename := params["name"]
	//fmt.Printf("header is: %s\n", devicename)
	fmt.Printf("response file %s\n", filename)
	//get date from filename
	re := regexp.MustCompile(`\d{4}-\d{2}-\d{2}`)
	submatchall := re.FindAllString(filename, -1)
	for _, element := range submatchall {
		newpath := filepath.Join(".", "uploaded", device_name, element)
		os.MkdirAll(newpath, os.ModePerm)
		file, err := os.Create(filepath.Join(newpath, filename))
		if err != nil {
			log.Fatal(err)
		}
		n, err := io.Copy(file, r.Body)
		if err != nil {
			log.Fatal(err)
		}
		w.Write([]byte(fmt.Sprintf("%d bytes are recieved.\n", n)))
		fmt.Println(element)
	}

}

func DiskStateHandler(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["token"]

	if !ok || len(keys[0]) < 1 {
		log.Println("Url Param 'token' is missing")
		return
	}
	token := string(keys[0])
	device_name := "TARS"
	device_id := 0
	err1 := db.QueryRow("select id, device_name from devices where dev_token=$1", token).Scan(&device_id, &device_name)
	if err1 != nil {
		// If there is an issue with the database, return a 500 error
		fmt.Println("--->", err1)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	diskstate := &DiskStatus{}
	err := json.NewDecoder(r.Body).Decode(diskstate)
	if err != nil {
		// If there is something wrong with the request body, return a 400 status
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Printf("Device %s disk_part %s\n", diskstate.Device, diskstate.Disk_part)
	fmt.Printf("All: %.2f GB\n", float64(diskstate.All)/float64(GB))
	fmt.Printf("Used: %.2f GB\n", float64(diskstate.Used)/float64(GB))
	fmt.Printf("Free: %.2f GB\n", float64(diskstate.Free)/float64(GB))

	rows, err := db.Query("insert into  device_disk_state values (DEFAULT,$1,$2,$3,$4,$5, CURRENT_TIMESTAMP)", device_id, diskstate.Disk_part,
		FloatToString(float64(diskstate.All)/float64(GB)),
		FloatToString(float64(diskstate.Used)/float64(GB)),
		FloatToString(float64(diskstate.Free)/float64(GB)))
	if err != nil {
		// If there is any issue with inserting into the database, return a 500 error
		log.Println(" insert err", err)
		//w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	w.WriteHeader(http.StatusOK)

}

/*
//get client ip
func GetIP(w http.ResponseWriter, req *http.Request){
    fmt.Fprintf(w, "<h1>static file server</h1><p><a href='./uploaded'>folder</p></a>")

    ip, port, err := net.SplitHostPort(req.RemoteAddr)
    if err != nil {
        //return nil, fmt.Errorf("userip: %q is not IP:port", req.RemoteAddr)

        fmt.Fprintf(w, "userip: %q is not IP:port", req.RemoteAddr)
    }

    userIP := net.ParseIP(ip)
    if userIP == nil {
        //return nil, fmt.Errorf("userip: %q is not IP:port", req.RemoteAddr)
        fmt.Fprintf(w, "userip: %q is not IP:port", req.RemoteAddr)
        return
    }

    // This will only be defined when site is accessed via non-anonymous proxy
    // and takes precedence over RemoteAddr
    // Header.Get is case-insensitive
    forward := req.Header.Get("X-Forwarded-For")
    fmt.Fprintf(w, "<p>IP: %s</p>", ip)
    fmt.Fprintf(w, "<p>Port: %s</p>", port)
    fmt.Fprintf(w, "<p>Forwarded for: %s</p>", forward)
}
*/
