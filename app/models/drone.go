package models

import (
	"time"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/platforms/dji/tello"
	"golang.org/x/sync/semaphore"
)

const (
	DefaultSpeed = 10
	WaitDroneStartSec = 5
)

// 要素を追加するための構造体
type DroneManager struct {
	*tello.Driver
	Speed int
	patrolSem *semaphore.Weighted
	patrolQuit chan bool
	isPatrolling bool
}

func NewDroneManager() *DroneManager  {
	// 8889でdroneに接続
	drone := tello.NewDriver("8889")
	// tello.Driverに要素を追加する
	droneManager := &DroneManager{
		Driver: drone,
		Speed:  DefaultSpeed,
		patrolSem: semaphore.NewWeighted(1),
		patrolQuit: make(chan bool),
		isPatrolling: false,
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

func (d *DroneManager) Patrol() {
	go func() {
		// ロック
		isAcquire := d.patrolSem.TryAcquire(1)
		if !isAcquire {
			d.patrolQuit <- true
			d.isPatrolling = false
			return
		}

		defer d.patrolSem.Release(1)
		// パトロールの開始
		d.isPatrolling = true
		status := 0
		// 3秒ごとに時刻を刻む
		t := time.NewTicker(3 * time.Second)

		for {
			select {
			// tickerが動いてるときに
			case <- t.C:
				// droneをホバー
				d.Hover()
				// status毎に挙動を変更
				switch status {
				case 1:
					d.Forward(d.Speed)
				case 2:
					d.Right(d.Speed)
				case 3:
					d.Backward(d.Speed)
				case 4:
					d.Left(d.Speed)
				case 5:
					status = 0
				}
				status++
			// patrolQuitがtrueで入ってきた場合、静止->着陸
			case <- d.patrolQuit:
				t.Stop()
				d.Hover()
				return
			}
		}
	}()
}
