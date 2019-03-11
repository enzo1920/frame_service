package models

import (
	"database/sql"
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
	"golang.org/x/crypto/bcrypt"
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

/*
func HomeRouterHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm() //анализ аргументов,
	fmt.Println(r.Method)
	fmt.Println(r.Form)  // ввод информации о форме на стороне сервера
	fmt.Println("path", r.URL.Path)
	fmt.Println("scheme", r.URL.Scheme)
	fmt.Println(r.Form["url_long"])
	for k, v := range r.Form {
			fmt.Println("key:", k)
			fmt.Println("val:", strings.Join(v, ""))
	}
	fmt.Fprintf(w, "Hello serg!") // отправляем данные на клиентскую сторону
}
*/
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

func Signup(w http.ResponseWriter, r *http.Request) {
	// Parse and decode the request body into a new `Credentials` instance
	fmt.Println("Signup is ok!")
	creds := &Credentials{}
	err := json.NewDecoder(r.Body).Decode(creds)
	if err != nil {
		// If there is something wrong with the request body, return a 400 status
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Println(creds.Username)
	// Salt and hash the password using the bcrypt algorithm
	// The second argument is the cost of hashing, which we arbitrarily set as 8 (this value can be more or less, depending on the computing power you wish to utilize)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), 8)
	// Next, test connect to DB
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	fmt.Println("DB is ok!")

	// Next, insert the username, along with the hashed password into the database
	if _, err = db.Query("insert into users values  ($1, $2)  ON CONFLICT (username) DO NOTHING", creds.Username, string(hashedPassword)); err != nil {
		// If there is any issue with inserting into the database, return a 500 error
		log.Println(" insert err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Println("insert  is ok!")
	// We reach this point if the credentials we correctly stored in the database, and the default status of 200 is sent back
}

func Signin(w http.ResponseWriter, r *http.Request) {

	// Parse and decode the request body into a new `Credentials` instance
	fmt.Println("Signin is ok!")
	creds := &Credentials{}
	err := json.NewDecoder(r.Body).Decode(creds)
	if err != nil {
		// If there is something wrong with the request body, return a 400 status
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Println(creds.Username, creds.Password)
	// Get the existing entry present in the database for the given username
	result := db.QueryRow("select password from users where username=$1", creds.Username)
	if err != nil {
		// If there is an issue with the database, return a 500 error
		fmt.Println("--->", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Println("Signin  select is ok!")

	// We create another instance of `Credentials` to store the credentials we get from the database
	storedCreds := &Credentials{}
	// Store the obtained password in `storedCreds`
	err = result.Scan(&storedCreds.Password)
	if err != nil {
		// If an entry with the username does not exist, send an "Unauthorized"(401) status
		if err == sql.ErrNoRows {
			log.Println(" sql error in scan !")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		// If the error is of any other type, send a 500 status
		log.Println(" error is of any other type  !")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Compare the stored hashed password, with the hashed version of the password that was received
	if err = bcrypt.CompareHashAndPassword([]byte(storedCreds.Password), []byte(creds.Password)); err != nil {
		// If the two passwords don't match, return a 401 status
		w.WriteHeader(http.StatusUnauthorized)

	}
	fmt.Println("authorized --->")
	// If we reach this point, that means the users password was correct, and that they are authorized
	// The default 200 status is sent
}

func GetCommands(w http.ResponseWriter, r *http.Request) {

	keys, ok := r.URL.Query()["device"]

	if !ok || len(keys[0]) < 1 {
		log.Println("Url Param 'device' is missing")
		return
	}

	// Query()["key"] will return an array of items,
	// we only want the single item.
	device := string(keys[0])
	rows, err := db.Query("SELECT cmd_name FROM commands as c inner join commands_ex as ce on c.id=ce.cmd_id"+
		" INNER join devices as d on d.id= ce.device_id"+
		" INNER join command_status cs on cs.id=ce.status_id WHERE d.device_name =$1", device)
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
	fmt.Fprintf(w, "%s", strings.Join(cmds, ", "))

}

//upload image handler
func UploadHandler(w http.ResponseWriter, r *http.Request) {

	contentDisposition := r.Header.Get("Content-Disposition")
    devicename := r.Header.Get("X-Custom-Header")// get devname from custom header
	fmt.Printf("Custom header is: %s\n", devicename)
	_, params, err := mime.ParseMediaType(contentDisposition)
	filename := params["filename"] // get filename from header
	fmt.Println("response file", filename)
	//get date from filename
	re := regexp.MustCompile(`\d{4}-\d{2}-\d{2}`)
	submatchall := re.FindAllString(filename, -1)
	for _, element := range submatchall {
		newpath := filepath.Join(".", "uploaded", element)
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
		//fmt.Println(element)
	}

	//get dev_id from db
	device_id := 0
	err = db.QueryRow("select id from devices where device_name=$1", devicename).Scan(&device_id)
	if err != nil {
		// If there is an issue with the database, return a 500 error
		fmt.Println("--->", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Println("id is:", device_id)
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
    if _, err = db.Query("insert into  dev_upload_log values (DEFAULT,$1,$2,$3,$4, CURRENT_TIMESTAMP)", device_id, string(ip), port,filename ); err != nil {
		// If there is any issue with inserting into the database, return a 500 error
		log.Println(" insert dev_upload_log err", err)
		//w.WriteHeader(http.StatusInternalServerError)
		return
	}

}

func DiskStateHandler(w http.ResponseWriter, r *http.Request) {
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

	if _, err = db.Query("insert into  device_disk_state values (DEFAULT,$1,$2,$3,$4,$5, CURRENT_TIMESTAMP)", 11, diskstate.Disk_part,
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
