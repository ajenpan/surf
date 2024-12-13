package uauth

type Config struct {
	DBAddr string
}

var DefaultConf = &Config{
	DBAddr: "root:123456@tcp(localhost:3306)/surf?charset=utf8mb4&parseTime=True&loc=Local",
}
