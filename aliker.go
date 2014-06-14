package main

import (
	"fmt"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"html/template"
	"net/http"
	"os"
//	"time"
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

func sendBlogsLikingPostData(c *websocket.Conn, blogs []string) error {
	return c.WriteJSON(&struct{
		MsgType string
		Blogs []string
	}{
		"blogs-liking-post",
		blogs,
	})
}

func sendBlogLikesData(c *websocket.Conn, likes []string) error {
	return c.WriteJSON(&struct{
		MsgType string
		Likes []string
	}{
		"blog-likes",
		likes,
	})
}

func SimilarHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	ensureNil(err)
	defer conn.Close()

	// Figure out what post they want and extract the details we need
	spr := &SimilarPostRequest{}
	err = conn.ReadJSON(spr)
	ensureNil(err)
	bh, pid, err := extractPostId(spr.PostUri)
	if err != nil {
		msg := fmt.Errorf("invalid post uri: %s", spr.PostUri)
		sendErrorNotification(conn, msg)
		return
	}

	err = sendBeginNotification(conn, bh, pid)
	ensureNil(err)

	likingBlogs := blogsLikingPost(bh, pid)

	err = sendBlogsLikingPostData(conn, likingBlogs)
	ensureNil(err)

	for _, blogName := range likingBlogs {
		fmt.Println(blogName)
		// TODO Get page one of that blogs' likes
		// TODO Request the other pages in parallel
		// TODO Send all of the post data at once for this blog
		// message type: blog-likes
		// { "blog-name": "some dude", "likes": [...] }
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
