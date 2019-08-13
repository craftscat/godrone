package controllers

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/taqboz/gotello/config"
)

func viewIndexhandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("app/view/index.html")
	err := t.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func StartWebServer() error {
	http.HandleFunc("/", viewIndexhandler)
	return http.ListenAndServe(fmt.Sprintf("%s:%d", config.Config.Address, config.Config.Port), nil)
}

