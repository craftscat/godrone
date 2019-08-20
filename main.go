package main

import (
	"log"

	"github.com/taqboz/gotello/app/controllers"
	"github.com/taqboz/gotello/config"
	"github.com/taqboz/gotello/utils"
)

func main()  {
	utils.LoggingSetting(config.Config.LogFile)
	log.Println(controllers.StartWebServer())
}
