package models

import (
	"time"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/platforms/dji/tello"
)

const (
	DefaultSpeed = 10
	WaitDroneStartSec = 5
)

// 要素を追加するための構造体
type DroneManager struct {
	*tello.Driver
	Speed int
}

func NewDroneManager() *DroneManager  {
	// 8889でdroneに接続
	drone := tello.NewDriver("8889")
	// tello.Driverに要素を追加する
	droneManager := &DroneManager{
		Driver: drone,
		Speed:  DefaultSpeed,
	}
	// 挙動の定義
	work := func() {
		//TODO
	}
	// robotの定義
	robot := gobot.NewRobot("tello", []gobot.Connection{}, []gobot.Device{drone}, work)
	// robotの挙動の開始
	go robot.Start()
	time.Sleep(WaitDroneStartSec * time.Second)
	// WEBサーバーでの操作用
	return droneManager
}
