package main

import (
	"github.com/taqboz/gotello/app/controllers"
	"github.com/taqboz/gotello/config"
	"github.com/taqboz/gotello/utils"
	"log"
)

func main()  {
	utils.LoggingSetting(config.Config.LogFile)
	log.Println(controllers.StartWebServer())
	//droneManager := models.NewDroneManager()
	//droneManager.TakeOff()
	//time.Sleep(10*time.Second)
	//droneManager.Patrol()
	//time.Sleep(30*time.Second)
	//droneManager.Patrol()
	//time.Sleep(10*time.Second)
	//droneManager.Land()
}
