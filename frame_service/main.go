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
//func AnotherHandlerLatest(w http.ResponseWriter, r *http.Request) {
//	fmt.Fprintf(w, "hello from AnotherHandlerLatest.")
//}

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
	r.HandleFunc("/example1", ExampleHandlerV1)
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

	r.HandleFunc("/get/command/", models.GetCommands).Methods("GET")
	r.HandleFunc("/get/cams/", models.GetCameraState).Methods("GET")
	r.HandleFunc("/get/relay/reset/", models.GetResetRelay).Methods("GET")
	r.HandleFunc("/set/cams/", models.SetCamState).Methods("POST")
	r.HandleFunc("/upload/image/", models.UploadHandler).Methods("POST")
	r.HandleFunc("/upload/temp/", models.UploadTempHandler).Methods("POST")
	r.HandleFunc("/upload/voltage/", models.UploadVoltageHandler).Methods("POST")
	r.HandleFunc("/upload/volumeinfo/", models.DiskStateHandler).Methods("POST")
	r.HandleFunc("/set/command/", models.SetCommandHandler).Methods("POST")
	r.HandleFunc("/get/img", models.GetImg).Methods("GET")
	//r.HandleFunc("/ip", models.GetIP).Methods("GET")
	//r.HandleFunc("/get-token", models.GetTokenHandler).Methods("GET")
}

func main() {

	//logging
	log_dir := "./log"
	if _, err := os.Stat(log_dir); os.IsNotExist(err) {
		os.Mkdir(log_dir, 0644)
	}

	file, err := os.OpenFile("./log/frame_service.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	log.SetOutput(file)
	log.Println("Logging to a file frameservice!")

	// read config
	readcfg := readconfig.Config_reader("./readconfig/frame_conf.conf")
	models.InitDB(readcfg)

	router := mux.NewRouter()
	//s := http.StripPrefix("/index/", http.FileServer(http.Dir("front/html")))

	// latest
	AddRoutes(router)
	// v1
	AddV1Routes(router.PathPrefix("/v1").Subrouter())
	// v2
	AddV2Routes(router.PathPrefix("/v2").Subrouter())
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("front/html")))
	//router.PathPrefix("/index/").Handler(http.StripPrefix("/index/", http.FileServer(http.Dir("front/html"))))
	//http.Handle("/", router)

	log.Fatal(http.ListenAndServe(":8080", router))

}
