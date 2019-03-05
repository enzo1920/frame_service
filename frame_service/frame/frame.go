package frame

type Configuration struct {
	Database struct {
		Host  string 
		DBname  string
		DBuser string
		Password string
		Port  int
	}
	log_file_name string
}