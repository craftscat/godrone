package models

import (
	"github.com/hybridgroup/mjpeg"
	"gocv.io/x/gocv"
	"io"
	"log"
	"os/exec"
	"strconv"
	"time"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/platforms/dji/tello"
	"golang.org/x/sync/semaphore"
)

const (
	DefaultSpeed = 10
	WaitDroneStartSec = 5
	frameX = 960 / 3
	frameY = 720 / 3
	frameCenterX = frameX / 2
	frameCenterY = frameY / 2
	frameArea = frameX * frameY
	frameSize = frameArea * 3
)

// 要素を追加するための構造体
type DroneManager struct {
	*tello.Driver
	Speed int
	patrolSem *semaphore.Weighted
	patrolQuit chan bool
	isPatrolling bool
	ffmpegIn io.WriteCloser
	ffmpegOut io.ReadCloser
	Stream *mjpeg.Stream
}

func NewDroneManager() *DroneManager  {
	// 8889でdroneに接続
	drone := tello.NewDriver("8889")

	ffmpeg := exec.Command("ffmpeg", "-hwaccel", "auto", "-hwaccel_device", "opencl", "-i", "pipe:0", "-pix_fmt", "bgr24",
		"-s", strconv.Itoa(frameX)+"x"+strconv.Itoa(frameY), "-f", "rawvideo", "pipe:1")
	ffmpegIn, _ := ffmpeg.StdinPipe()
	ffmpegQut, _ := ffmpeg.StdoutPipe()

	// tello.Driverに要素を追加する
	droneManager := &DroneManager{
		Driver: drone,
		Speed:  DefaultSpeed,
		patrolSem: semaphore.NewWeighted(1),
		patrolQuit: make(chan bool),
		isPatrolling: false,
		ffmpegIn: ffmpegIn,
		ffmpegOut: ffmpegQut,


	}
	// ビデオ
	work := func() {
		drone.On(tello.ConnectedEvent, func(data interface{}) {
			if err := ffmpeg.Start(); err != nil {
				log.Println(err)
				return
			}
			log.Println("Connected")
			drone.StartVideo()
			drone.SetVideoEncoderRate(tello.VideoBitRateAuto)
			// 露光
			drone.SetExposure(0)

			gobot.Every(100*time.Millisecond, func() {
				drone.StartVideo()
			})

			droneManager.StreamVideo()
		})

		drone.On(tello.VideoFrameEvent, func(data interface{}) {
			pkt := data.([]byte)
			if _, err := ffmpegIn.Write(pkt); err != nil {
				log.Println(err)
			}
		})
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
				d.isPatrolling = false
				return
			}
		}
	}()
}

func (d *DroneManager) StartPatrol() {
	if !d.isPatrolling {
		d.Patrol()
	}
}

func (d *DroneManager) StopPatrol() {
	if !d.isPatrolling {
		d.Patrol()
	}
}

func (d *DroneManager) StreamVideo() {
	go func(d *DroneManager) {
		for {
			buf := make([]byte, frameSize)
			if _, err := io.ReadFull(d.ffmpegOut, buf); err != nil {
				log.Println(err)
			}
			img, _ := gocv.NewMatFromBytes(frameY, frameX, gocv.MatTypeCV8UC3, buf)

			if img.Empty(){
				continue
			}

			jpegBuf, _ := gocv.IMEncode(".jpg", img)
			d.Stream.UpdateJPEG(jpegBuf)
		}
	}(d)
}
