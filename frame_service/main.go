package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"./models"
	"./readconfig"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

const hashCost = 8

// AnotherHandlerLatest is the newest version of AnotherHandler
func AnotherHandlerLatest(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello from AnotherHandlerLatest.")
}

// ExampleHandlerLatest is the newest version of ExampleHandler
func ExampleHandlerLatest(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello from ExampleHandlerLatest.")
}

// ExampleHandlerV1 is a v1-compatible version of ExampleHandler
func ExampleHandlerV1(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello from ExampleHandlerv1.")
}

// ExampleHandlerV1 is a v2-compatible version of ExampleHandler
func ExampleHandlerV2(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello from ExampleHandlerv2.")
}

// AddV1Routes takes a router or subrouter and adds all the v1
// routes to it
func AddV1Routes(r *mux.Router) {
	r.HandleFunc("/example", ExampleHandlerV1)
	AddRoutes(r)
}

// AddV2Routes takes a router or subrouter and adds all the v2
// routes to it, note that these should probably match
// AddRoutes(r *muxRouter) alternatively you can do
// var AddV2Routes = AddRoutes
func AddV2Routes(r *mux.Router) {
	r.HandleFunc("/example2", ExampleHandlerV2)
	AddRoutes(r)
}

// AddRoutes takes a router or subrouter and adds all the latest
// routes to it
func AddRoutes(r *mux.Router) {
	r.HandleFunc("/signin", models.Signin).Methods("POST")
	r.HandleFunc("/signup", models.Signup).Methods("POST")
	r.HandleFunc("/getcommand/", models.GetCommands).Methods("GET")
	r.HandleFunc("/getcams", models.Cam_adr_get).Methods("GET")
	r.HandleFunc("/upload/", models.UploadHandler).Methods("POST")
	r.HandleFunc("/diskstate", models.DiskStateHandler).Methods("POST")
	//r.HandleFunc("/ip", models.GetIP).Methods("GET")
	//r.HandleFunc("/get-token", models.GetTokenHandler).Methods("GET")
}

func main() {

	//logging
	file, err := os.OpenFile("./log/frame_service.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	log.SetOutput(file)
	log.Println("Logging to a file in Go!")

	// read config
	readcfg := readconfig.Config_reader("./readconfig/frame_conf.conf")
	models.InitDB(readcfg)

	router := mux.NewRouter()
	// latest
	AddRoutes(router)
	// v1
	AddV1Routes(router.PathPrefix("/v1").Subrouter())
	// v2
	AddV2Routes(router.PathPrefix("/v2").Subrouter())

	log.Fatal(http.ListenAndServe(":8080", router))

}
