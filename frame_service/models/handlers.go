package models

import (
	
	"encoding/json"
	"fmt"
	"io"
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

type DiskStatus struct {
	Device    string `json:"device"`
	Disk_part string `json:"disk_part"`
	All       uint64 `json:"all"`
	Used      uint64 `json:"used"`
	Free      uint64 `json:"free"`
}


func FloatToString(input_num float64) string {
	// to convert a float number to a string
	return strconv.FormatFloat(input_num, 'f', 6, 64)
}

func Cam_adr_get(w http.ResponseWriter, r *http.Request) {

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
	rows, err := db.Query("SELECT cmd_name FROM commands as c inner join commands_ex as ce on c.id=ce.cmd_id"+
		" INNER join devices as d on d.id= ce.device_id"+
		" INNER join command_status cs on cs.id=ce.status_id WHERE d.dev_token =$1", dev_token)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	cmds := make([]string, 0)

	for rows.Next() {
		var cmd string
		if err := rows.Scan(&cmd); err != nil {
			// Check for a scan error.
			// Query rows will be closed with defer.
			log.Fatal(err)
		}
		cmds = append(cmds, cmd)
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
	fmt.Fprintf(w, "%s", strings.Join(cmds, ","))

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
	if _, err := db.Query("insert into  dev_upload_log values (DEFAULT,$1,$2,$3,$4, CURRENT_TIMESTAMP)", device_id, string(ip), port, filename); err != nil {
		// If there is any issue with inserting into the database, return a 500 error
		log.Println(" insert dev_upload_log err", err)
		//w.WriteHeader(http.StatusInternalServerError)
		return
	}

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

	if _, err = db.Query("insert into  device_disk_state values (DEFAULT,$1,$2,$3,$4,$5, CURRENT_TIMESTAMP)", device_id, diskstate.Disk_part,
		FloatToString(float64(diskstate.All)/float64(GB)),
		FloatToString(float64(diskstate.Used)/float64(GB)),
		FloatToString(float64(diskstate.Free)/float64(GB))); err != nil {
		// If there is any issue with inserting into the database, return a 500 error
		log.Println(" insert err", err)
		//w.WriteHeader(http.StatusInternalServerError)
		return
	}
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
