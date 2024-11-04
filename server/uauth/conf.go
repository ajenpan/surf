package uauth

type Config struct {
	DBAddr         string
	HttpListenAddr string
}

var DefaultConf = &Config{
	DBAddr:         "sa1:sa1@tcp(test41:3306)/surf?charset=utf8mb4&parseTime=True&loc=Local",
	HttpListenAddr: ":9999",
}
