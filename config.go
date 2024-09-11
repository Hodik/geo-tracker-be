package main

func GetConfig() *Config {
	var conf Config
	result := db.FirstOrCreate(&conf, Config{})

	if result.Error != nil {
		panic(result.Error)
	}
	return &conf
}
