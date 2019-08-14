package controllers

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/taqboz/gotello/config"
)

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

func StartWebServer() error {
	http.HandleFunc("/", viewIndexHandler)
	// "static"をファイルサーバーとして使用する
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	return http.ListenAndServe(fmt.Sprintf("%s:%d", config.Config.Address, config.Config.Port), nil)
}

