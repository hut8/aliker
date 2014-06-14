package main

import (
	"fmt"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"html/template"
	"net/http"
	"os"
	"time"
)

var upgrader websocket.Upgrader

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/index.go.html")
	ensureNil(err)
	err = t.Execute(w, nil)
	ensureNil(err)
}

type SimilarPostRequest struct {
	PostUri string
}

type TumblrCredentials struct {
	Key    string
	Secret string
}

func getCredentials() *TumblrCredentials {
	return &TumblrCredentials{
		Key:    os.Getenv("ALIKER_KEY"),
		Secret: os.Getenv("ALIKER_SECRET"),
	}
}

func SimilarHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	ensureNil(err)

	for {
		err = conn.WriteJSON(map[string]string{
			"hello": "world",
		})
		if err != nil {
			break
		}
		<- time.After(5 * time.Second)
	}
}

func ensureNil(x interface{}) {
	if x != nil {
		panic(x)
	}
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/", HomeHandler)
	router.HandleFunc("/post/{postPath:\\S+}", SimilarHandler)
	n := negroni.New()
	n.UseHandler(router)
	n.Run(fmt.Sprintf(":%s", os.Getenv("PORT")))
}
