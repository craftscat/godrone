package main

import (
	"github.com/taqboz/gotello/app/models"
	"time"
)

func main()  {
	droneManager := models.NewDroneManager()
	droneManager.TakeOff()
	time.Sleep(10*time.Second)
	droneManager.Land()
}
