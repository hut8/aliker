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

type beginNotification struct {
	BaseHostname string
	PostID       int64
	MsgType      string
}

func sendBeginNotification(c *websocket.Conn, bh string, pid int64) error {
	msg := &beginNotification{
		BaseHostname: bh,
		PostID:       pid,
		MsgType:      "begin-notification",
	}
	return c.WriteJSON(msg)
}

func sendErrorNotification(c *websocket.Conn, err error) error {
	return c.WriteJSON(&struct{
		MsgType string
		Message string
	}{
		"error",
		err.Error(),
	})
}

func SimilarHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	ensureNil(err)

	spr := &SimilarPostRequest{}
	err = conn.ReadJSON(spr)
	ensureNil(err)

	bh, _, err := extractPostId(spr.PostUri)
	if err != nil {
		msg := fmt.Sprintf("invalid post uri: %s", spr.PostUri)
		fmt.Println(msg)
		conn.WriteJSON(map[string]string{
			"error": msg,
		})
		conn.Close()
	}

	for {
		err = conn.WriteJSON(map[string]string{
			"bh": bh,
		})
		if err != nil {
			break
		}
		<-time.After(5 * time.Second)
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
	router.HandleFunc("/post", SimilarHandler)

	n := negroni.New()
	n.UseHandler(router)
	n.Run(fmt.Sprintf(":%s", os.Getenv("PORT")))
}
