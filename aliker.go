package main

import (
	"fmt"
	//	"github.com/bradfitz/iter"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/hut8/tumblr-go"
	//"github.com/kr/pretty"
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

func sendProcessNotification(c *websocket.Conn, bh string, pid int64) error {
	return c.WriteJSON(&struct {
		MsgType      string `json:"msg-type"`
		BaseHostname string `json:"base-hostname"`
		PostID       int64  `json:"pid"`
	}{
		"process-post",
		bh,
		pid,
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

func sendBlogLikesData(c *websocket.Conn, bhn string, likes []int64,
	page int, totalPages int) error {
	return c.WriteJSON(&struct {
		MsgType string  `json:"msg-type"`
		Blog    string  `json:"blog"`
		Likes   []int64 `json:"likes"`
	}{
		"blog-likes",
		bhn,
		likes,
	})
}

func sendPostSimilarities(c *websocket.Conn, s map[int64]int64) error {
	return c.WriteJSON(&struct {
		MsgType       string          `json:"msg-type"`
		SimilarityMap map[int64]int64 `json:"similarities"`
	}{
		"post-similarities",
		s,
	})
}

func sendPostsData(c *websocket.Conn, pd map[int64]tumblr.Post) error {
	return c.WriteJSON(&struct {
		MsgType       string          `json:"msg-type"`
		PostData map[int64]tumblr.Post `json:"posts-data"`
	}{
		"posts-data",
		pd,
	})
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

func SimilarHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	ensureNil(err)
	defer conn.Close()

	// Figure out what post they want and extract the details we need
	spr := &SimilarPostRequest{}
	err = conn.ReadJSON(spr)
	ensureNil(err)
	bh, pid, err := extractPostId(spr.PostUri)
	// bh, pid will be zero values if invalid. the JS knows this.
	mustSend(sendProcessNotification(conn, bh, pid))
	if err != nil {
		return
	}

	// Find every blog that likes the input post
	likingBlogs, err := blogsLikingPost(bh, pid)
	ensureNil(err)

	mustSend(sendBlogsLikingPostData(conn, likingBlogs))

	// Find every liked post from every blog that likes the input post
	// postId -> []blogUrl
	blogLikeMap := make(map[int64][]string)
	// postId -> Post
	postMap := make(map[int64]tumblr.Post)
	// TODO Break out each blog into a separate goroutine!
	for _, blogName := range likingBlogs {
		b := tumblrClient.NewBlog(blogName)
		fmt.Printf("Requesting likes for: %s\n", b.BaseHostname)
		likeCollection, err := b.Likes(tumblr.LimitOffset{})
		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		// How many pages are we going to loop through
		totalPages := int(likeCollection.TotalCount) / 20
		if (likeCollection.TotalCount % 20) != 0 {
			totalPages += 1
		}
		if totalPages > 20 {
			totalPages = 20
		}

		// Loop over pages.
		// Note that we already retrieved the first page, so fetch at end of loop
		for currentPage := 1; currentPage < totalPages; currentPage++ {
			pagePostIDs := []int64{}
			for _, likedPost := range likeCollection.Likes.Posts {
				pagePostIDs = append(pagePostIDs, likedPost.PostId())
				postMap[likedPost.PostId()] = likedPost
				// Initialize if key if needbe. No autovivification here!
				_, ok := blogLikeMap[likedPost.PostId()]
				if !ok {
					blogLikeMap[likedPost.PostId()] = []string{}
				}
				blogLikeMap[likedPost.PostId()] = append(
					blogLikeMap[likedPost.PostId()], b.BaseHostname)
			}

			mustSend(sendBlogLikesData(conn, b.BaseHostname, pagePostIDs,
				currentPage, totalPages))

			// Fetch next page
			likeCollection, err = b.Likes(tumblr.LimitOffset{
				Offset: currentPage * 20,
			})
			if err != nil {
				sendErrorNotification(conn, err)
				fmt.Println(err.Error())
				break
			}
		}
	}

	// Send a map of postId -> tumblr.Post
	mustSend(sendPostsData(conn, postMap))

	// Finally, calculate overall similarities now that we have all the data
	// postId -> how many users like it
	postPopularityMap := make(map[int64]int64)
	for postId, blogLikes := range blogLikeMap {
		postPopularityMap[postId] = int64(len(blogLikes))
	}
	mustSend(sendPostSimilarities(conn, postPopularityMap))
	//fmt.Printf("PostID->Blogs who like it%# v\n", pretty.Formatter(blogLikeMap))
}

func mustSend(x interface{}) {
	ensureNil(x)
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
