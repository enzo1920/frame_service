package readconfig

import (
	"encoding/json"
	"flag"
	"log"
	"os"
)

type Configuration struct {
	Host  string 
	Devicename  string
	Password string
	Port  int
}


//func config_reader(cfg_file string)([]string){
func Config_reader(cfg_file string) Configuration {

	c := flag.String("c", cfg_file, "Specify the configuration file.")
	flag.Parse()
	file, err := os.Open(*c)
	if err != nil {
		log.Println("can't open config file: ", err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	Config := Configuration{}
	err = decoder.Decode(&Config)
	if err != nil {
		log.Println("can't decode config JSON: ", err)
	}

	return Config
}
