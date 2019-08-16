package models

import (
	"context"
	"image"
	"image/color"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os/exec"
	"strconv"
	"time"

	"github.com/hybridgroup/mjpeg"
	"gobot.io/x/gobot"
	"gobot.io/x/gobot/platforms/dji/tello"
	"gocv.io/x/gocv"
	"golang.org/x/sync/semaphore"
)

const (
	DefaultSpeed      = 10
	WaitDroneStartSec = 5
	frameX            = 960 / 3
	frameY            = 720 / 3
	frameCenterX      = frameX / 2
	frameCenterY      = frameY / 2
	frameArea         = frameX * frameY
	frameSize         = frameArea * 3
	faceDetectXMLFile = "./app/models/haarcascade_frontalface_default.xml"
	snapshotsFolder   = "./static/img/snapshots/"
)

type DroneManager struct {
	*tello.Driver
	Speed                int
	patrolSem            *semaphore.Weighted
	patrolQuit           chan bool
	isPatrolling         bool
	ffmpegIn             io.WriteCloser
	ffmpegOut            io.ReadCloser
	Stream               *mjpeg.Stream
	faceDetectTrackingOn bool
	isSnapShot           bool
}

func NewDroneManager() *DroneManager {
	// 8889でdroneに接続
	drone := tello.NewDriver("8889")

	ffmpeg := exec.Command("ffmpeg", "-hwaccel", "auto", "-hwaccel_device", "opencl", "-i", "pipe:0", "-pix_fmt", "bgr24",
		"-s", strconv.Itoa(frameX)+"x"+strconv.Itoa(frameY), "-f", "rawvideo", "pipe:1")
	ffmpegIn, _ := ffmpeg.StdinPipe()
	ffmpegOut, _ := ffmpeg.StdoutPipe()

	droneManager := &DroneManager{
		// tello.Driverに要素を追加する
		Driver:               drone,
		Speed:                DefaultSpeed,
		patrolSem:            semaphore.NewWeighted(1),
		patrolQuit:           make(chan bool),
		isPatrolling:         false,
		ffmpegIn:             ffmpegIn,
		ffmpegOut:            ffmpegOut,
		Stream:               mjpeg.NewStream(),
		faceDetectTrackingOn: false,
		isSnapShot:           false,
	}
	work := func() {
		if err := ffmpeg.Start(); err != nil {
			log.Println(err)
			return
		}

		drone.On(tello.ConnectedEvent, func(data interface{}) {
			log.Println("Connected")
			drone.StartVideo()
			drone.SetVideoEncoderRate(tello.VideoBitRateAuto)
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
			case <-t.C:
				// droneを静止
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
			case <-d.patrolQuit:
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
	if d.isPatrolling {
		d.Patrol()
	}
}

func (d *DroneManager) StreamVideo() {
	go func(d *DroneManager) {
		classifier := gocv.NewCascadeClassifier()
		defer classifier.Close()
		if !classifier.Load(faceDetectXMLFile) {
			log.Printf("Error reading cascade file: %v\n", faceDetectXMLFile)
			return
		}
		blue := color.RGBA{0, 0, 255, 0}

		for {
			buf := make([]byte, frameSize)
			if _, err := io.ReadFull(d.ffmpegOut, buf); err != nil {
				log.Println(err)
			}
			img, _ := gocv.NewMatFromBytes(frameY, frameX, gocv.MatTypeCV8UC3, buf)

			if img.Empty() {
				continue
			}

			if d.faceDetectTrackingOn {
				d.StopPatrol()
				rects := classifier.DetectMultiScale(img)
				log.Printf("found %d faces\n", len(rects))

				if len(rects) == 0 {
					d.Hover()
				}
				for _, r := range rects {
					gocv.Rectangle(&img, r, blue, 3)
					pt := image.Pt(r.Max.X, r.Min.Y-5)
					gocv.PutText(&img, "Human", pt, gocv.FontHersheyPlain, 1.2, blue, 2)

					faceWidth := r.Max.X - r.Min.X
					faceHeight := r.Max.Y - r.Min.Y
					faceCenterX := r.Min.X + (faceWidth / 2)
					faceCenterY := r.Min.Y + (faceHeight / 2)
					faceArea := faceWidth * faceHeight
					diffX := frameCenterX - faceCenterX
					diffY := frameCenterY - faceCenterY
					percentF := math.Round(float64(faceArea) / float64(frameArea) * 100)

					move := false
					if diffX < -20 {
						d.Right(15)
						move = true
					}
					if diffX > 20 {
						d.Left(15)
						move = true
					}
					if diffY < -30 {
						d.Down(25)
						move = true
					}
					if diffY > 30 {
						d.Up(25)
						move = true
					}
					if percentF > 7.0 {
						d.Backward(10)
						move = true
					}
					if percentF < 0.9 {
						d.Forward(10)
						move = true
					}
					if !move {
						d.Hover()
					}

					break
				}
			}

			jpegBuf, _ := gocv.IMEncode(".jpg", img)

			if d.isSnapShot {
				backupFileName := snapshotsFolder + time.Now().Format(time.RFC3339) + ".jpg"
				ioutil.WriteFile(backupFileName, jpegBuf, 0644)
				snapshotFileName := snapshotsFolder + "snapshot.jpg"
				ioutil.WriteFile(snapshotFileName, jpegBuf, 0644)
				d.isSnapShot = false
			}

			d.Stream.UpdateJPEG(jpegBuf)
		}
	}(d)
}

func (d *DroneManager) TakeSnapshot() {
	d.isSnapShot = true
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	for {
		if !d.isSnapShot || ctx.Err() != nil {
			break
		}
	}
	d.isSnapShot = false
}

func (d *DroneManager) EnableFaceDetectTracking() {
	d.faceDetectTrackingOn = true
}

func (d *DroneManager) DisableFaceDetectTracking() {
	d.faceDetectTrackingOn = false
	d.Hover()
}
