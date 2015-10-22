package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/alde/eremetic/handler"
	"github.com/alde/eremetic/routes"
	"github.com/kardianos/osext"
	"github.com/spf13/viper"
)

func readConfig() {
	path, _ := osext.ExecutableFolder()
	viper.AddConfigPath("/etc/eremetic")
	viper.AddConfigPath(path)
	viper.AutomaticEnv()
	viper.SetConfigName("eremetic")
	viper.ReadInConfig()
}

func main() {
	readConfig()
	bind := fmt.Sprintf("%s:%d", viper.GetString("address"), viper.GetInt("port"))

	router := routes.Create()
	log.Printf("listening to %s", bind)
	go handler.Run()
	log.Fatal(http.ListenAndServe(bind, router))
}
