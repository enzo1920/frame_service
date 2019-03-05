package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

// Create a struct that models the structure of a user, both in the request body, and in the DB
type Credentials struct {
	Password string `json:"password", db:"password"`
	Username string `json:"username", db:"username"`
}

func bad_method(w http.ResponseWriter, method string) {
	if method != "POST" {
		http.Error(w, http.StatusText(405), 405)
		return
	}

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
func Cam_adr_get(w http.ResponseWriter, r *http.Request) {
	bad_method(w, r.Method)
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
	bad_method(w, r.Method)
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

	bad_method(w, r.Method)

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

	fmt.Println("Url Param 'device' is: " + device)
	rows, err := db.Query("SELECT cmd_name FROM commands as c inner join commands_ex as ce on c.id=ce.cmd_id INNER join devices as d on d.id= ce.device_id INNER join command_status cs on cs.id=ce.status_id WHERE d.device_name =$1", device)
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

/*// Глобальный секретный ключ
 var mySigningKey = []byte("framesecretkey")

func GetTokenHandler(w http.ResponseWriter, r *http.Request){
	 // Создаем новый токен
	 token := jwt.New(jwt.SigningMethodHS256)

	 // Устанавливаем набор параметров для токена
	 token.Claims["admin"] = true
	 token.Claims["name"] = "CASE1"
	 token.Claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	 // Подписываем токен нашим секретным ключем
	 tokenString, _ := token.SignedString(mySigningKey)

	 // Отдаем токен клиенту
	 w.Write([]byte(tokenString))
 }*/
