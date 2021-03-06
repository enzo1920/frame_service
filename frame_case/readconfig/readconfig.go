package readconfig

import (
	"encoding/json"
	"flag"
	"log"
	"os"
)

/*type Configuration struct {
		Host  string
		Devicename  string
		Password string
		Port  int

	log_file_name string
}*/

type Configuration struct {
	Connection struct {
		Host       string
		Devicename string
		Password   string
		Port       int
	}
	Cameras_block struct {
		Cameras_address []struct {
			Type  int
			Value string `json:"ip_address"`
		}
		Cameras_type []struct {
			Type  int
			Value string
		}
		Cameras_stream []struct {
			Cameras_type int
			Stream       string
		}
	}
	Api_Urls []struct {
		Api_command string
		Url         string
	}
	Api_token string `json:"api_token"`
}

type Ip_stream struct {
	Ip     []string
	Type   string
	Stream string
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
