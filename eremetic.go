package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"

	log "github.com/dmuth/google-go-log4go"
	"github.com/kardianos/osext"
	"github.com/spf13/viper"

	"github.com/alde/eremetic/database"
	"github.com/alde/eremetic/routes"
	"github.com/alde/eremetic/scheduler"
)

func readConfig() {
	path, _ := osext.ExecutableFolder()
	viper.AddConfigPath("/etc/eremetic")
	viper.AddConfigPath(path)
	viper.AutomaticEnv()
	viper.SetConfigName("eremetic")
	viper.SetDefault("loglevel", "debug")
	viper.SetDefault("database", "db/eremetic.db")
	viper.ReadInConfig()
}

func setupLogging() {
	log.SetLevelString(viper.GetString("loglevel"))
	log.SetDisplayTime(true)
}

func main() {
	if len(os.Args) == 2 && os.Args[1] == "--version" {
		fmt.Println(Version)
		os.Exit(0)
	}
	readConfig()
	setupLogging()
	defer database.Close()

	bind := fmt.Sprintf("%s:%d", viper.GetString("address"), viper.GetInt("port"))

	// Catch interrupt
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, os.Kill)
		s := <-c
		if s != os.Interrupt && s != os.Kill {
			return
		}

		log.Info("Eremetic is shutting down")
		os.Exit(0)
	}()

	router := routes.Create()
	log.Infof("listening to %s", bind)
	go scheduler.Run()
	err := http.ListenAndServe(bind, router)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}
