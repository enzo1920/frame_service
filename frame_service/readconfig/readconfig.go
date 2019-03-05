package readconfig

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"../frame"
)


//func config_reader(cfg_file string)([]string){
func Config_reader(cfg_file string) frame.Configuration {

	c := flag.String("c", cfg_file, "Specify the configuration file.")
	flag.Parse()
	file, err := os.Open(*c)
	if err != nil {
		log.Println("can't open config file: ", err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	Config := frame.Configuration{}
	err = decoder.Decode(&Config)
	if err != nil {
		log.Println("can't decode config JSON: ", err)
	}

	return Config
}
