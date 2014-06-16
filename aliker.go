package main

import (
	"fmt"
	//	"github.com/bradfitz/iter"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/hut8/tumblr-go"
	"github.com/kr/pretty"
	"html/template"
	"net/http"
	"os"
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

type beginNotification struct {
	BaseHostname string `json:"base-hostname"`
	PostID       int64  `json:"pid"`
	MsgType      string `json:"msg-type"`
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
	return c.WriteJSON(&struct {
		MsgType string `json:"msg-type"`
		Message string `json:"message"`
	}{
		"error",
		err.Error(),
	})
}

func sendBlogsLikingPostData(c *websocket.Conn, blogs []string) error {
	return c.WriteJSON(&struct {
		MsgType string   `json:"msg-type"`
		Blogs   []string `json:"blogs"`
	}{
		"blogs-liking-post",
		blogs,
	})
}

func sendBlogLikesData(c *websocket.Conn, blog string, likes []int64) error {
	return c.WriteJSON(&struct {
		MsgType string  `json:"msg-type"`
		Blog    string  `json:"blog"`
		Likes   []int64 `json:"likes"`
	}{
		"blog-likes",
		blog,
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

	// Find every blog that likes the input post
	likingBlogs, err := blogsLikingPost(bh, pid)
	ensureNil(err)
	err = sendBlogsLikingPostData(conn, likingBlogs)
	ensureNil(err)

	// Find every liked post from every blog that likes the input post
	// postId -> []blogUrl
	blogLikeMap := make(map[int64][]string)
	// postId -> Post
	postMap := make(map[int64]tumblr.Post)
	for _, blogName := range likingBlogs {
		b := tumblrClient.NewBlog(blogName)
		fmt.Printf("Requesting likes for: %s\n", b.BaseHostname)
		likeCollection, err := b.Likes(tumblr.LimitOffset{})
		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		// How many pages are we going to loop through
		totalPages := likeCollection.TotalCount / 20
		if (likeCollection.TotalCount % 20) != 0 {
			totalPages += 1
		}

		// Loop over pages.
		// Note that we already retrieved the first page, so fetch at end of loop
		for currentPage := int64(0); currentPage < totalPages; currentPage++ {
			pagePostIDs := []int64{}
			for _, likedPost := range likeCollection.Likes.Posts {
				pagePostIDs = append(pagePostIDs, likedPost.PostId())
				postMap[likedPost.PostId()] = likedPost
				// Initialize if key if needbe
				_, ok := blogLikeMap[likedPost.PostId()]
				if !ok {
					blogLikeMap[likedPost.PostId()] = []string{}
				}
				blogLikeMap[likedPost.PostId()] = append(
					blogLikeMap[likedPost.PostId()], b.BaseHostname)
			}
			sendBlogLikesData(conn, b.BaseHostname, pagePostIDs)

			// Fetch next page
			likeCollection, err = b.Likes(tumblr.LimitOffset{})
			if err != nil {
				fmt.Println(err.Error())
				break
			}
		}
		//sendBlogsLikingPostData(conn, blogLikeMap[likedPost.PostId()])
	}
	fmt.Printf("PostID->Blogs who like it%# v\n", pretty.Formatter(blogLikeMap))
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
