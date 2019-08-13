package main

import (
	"github.com/taqboz/gotello/app/models"
	"github.com/taqboz/gotello/config"
	"github.com/taqboz/gotello/utils"
	"time"
)

func main()  {
	utils.LoggingSetting(config.Config.LogFile)
	droneManager := models.NewDroneManager()
	droneManager.TakeOff()
	time.Sleep(10*time.Second)
	droneManager.Patrol()
	time.Sleep(30*time.Second)
	droneManager.Patrol()
	time.Sleep(10*time.Second)
	droneManager.Land()
}
