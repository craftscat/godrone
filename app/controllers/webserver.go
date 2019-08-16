package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/taqboz/gotello/app/models"
	"github.com/taqboz/gotello/config"
	"html/template"
	"log"
	"net/http"
	"regexp"
	"strconv"
)

var appContext struct {
	DroneManager *models.DroneManager
}

func init() {
	appContext.DroneManager = models.NewDroneManager()
}

// テンプレートの読み込み
func getTemplate(temp string) (*template.Template, error) {
	return template.ParseFiles("app/view/layout.html", temp)
}

func viewIndexHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := getTemplate("app/view/index.html")
	err := t.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func viewControllerHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := getTemplate("app/view/controller.html")
	err := t.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// 以下APIの処理
type APIResult struct {
	Result interface{} `json:"result"`
	Code   int         `json:"code"`
}

func APIResponse(w http.ResponseWriter, result interface{}, code int)  {
	res := APIResult{Result: result, Code: code}
	js, err := json.Marshal(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(js)
}

var apiValidPath = regexp.MustCompile("^/api/(command|shake|video)")

func apiMakeHandler(fn func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := apiValidPath.FindStringSubmatch(r.URL.Path)
		if len(m) == 0 {
			// TODO
			return
		}
		fn(w, r)
	}
}

func getSpeed(r *http.Request) int {
	strSpeed := r.FormValue("speed")
	if strSpeed == "" {
		return models.DefaultSpeed
	}
	speed, err := strconv.Atoi(strSpeed)
	if err != nil {
		return models.DefaultSpeed
	}
	return speed
}

func apiCommandHandler(w http.ResponseWriter, r *http.Request)  {
	command := r.FormValue("command")
	log.Printf("action=apiCommandHandler command=%s", command)
	drone := appContext.DroneManager
	switch command {
	case "ceaseRotation":
		drone.CeaseRotation()
	case "takeOff":
		drone.TakeOff()
	case "land":
		drone.Land()
	case "hover":
		drone.Hover()
	case "up":
		drone.Up(drone.Speed)
	case "clockwise":
		drone.Clockwise(drone.Speed)
	case "counterClockWise":
		drone.CounterClockwise(drone.Speed)
	case "down":
		drone.Down(drone.Speed)
	case "forward":
		drone.Forward(drone.Speed)
	case "left":
		drone.Left(drone.Speed)
	case "right":
		drone.Right(drone.Speed)
	case "backward":
		drone.Backward(drone.Speed)
	case "speed":
		drone.Speed = getSpeed(r)
	case "frontFlip":
		drone.FrontFlip()
	case "leftFlip":
		drone.LeftFlip()
	case "rightFlip":
		drone.RightFlip()
	case "backFlip":
		drone.BackFlip()
	case "throwTakeOff()":
		drone.ThrowTakeOff()
	case "bounce":
		drone.Bounce()

	// 関数内同じだからまとめるか実装を変えたほうがいいかも?
	case "patrol":
		drone.StartPatrol()
	case "stopPatrol":
		drone.StopPatrol()
	default:
		APIResponse(w, "Not found", http.StatusNotFound)
		return
	}
	APIResponse(w, "OK", http.StatusOK)
}

func StartWebServer() error {
	http.HandleFunc("/", viewIndexHandler)
	http.HandleFunc("/controller/", viewControllerHandler)
	http.HandleFunc("/api/command/", apiMakeHandler(apiCommandHandler))
	http.Handle("/video/streaming", appContext.DroneManager.Stream)
	// "static"をファイルサーバーとして使用する
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	return http.ListenAndServe(fmt.Sprintf("%s:%d", config.Config.Address, config.Config.Port), nil)
}
