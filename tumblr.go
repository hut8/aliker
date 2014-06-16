package main

import (
	"fmt"
	"github.com/hut8/tumblr-go"
	"os"
	"regexp"
	"strconv"
)

var (
	tumblrDataRegexp *regexp.Regexp = regexp.MustCompile("http://([^/]+)/post/(\\d+)(?:/.+)?")
	tumblrClient *tumblr.Tumblr = &tumblr.Tumblr{
		Credentials: getCredentials(),
	}
)

// http://lacecard.tumblr.com/post/76803575816/emacs-in-tron -> (lacecard.tumblr.com, 76803575816)
func extractPostId(u string) (string, int64, error) {
	m := tumblrDataRegexp.FindStringSubmatch(u)
	if m == nil {
		return "", 0, fmt.Errorf("Could not extract base hostname and post ID")
	}
	bh := m[1]
	pid, err := strconv.ParseInt(m[2], 10, 64)
	ensureNil(err) // For 2^64+
	return bh, pid, nil
}

// Return a slice of blog names who like the given post
// FIXME: The stupid API only returns 20 of these friggin things
func blogsLikingPost(baseHostname string, postId int64) ([]string, error) {
	// Get the post from the API (specifying ID guarantees only one)
	blog := tumblrClient.NewBlog(baseHostname)
	params := tumblr.PostRequestParams{
		Id:        postId,
		NotesInfo: true,
	}
	postCollection, err := blog.Posts(params)
	ensureNil(err)

	// Extract the single post we're looking for
	posts := postCollection.Posts
	if len(posts) == 0 {
		return nil, fmt.Errorf("No such post was found")
	}
	if len(posts) > 1 {
		panic("Multiple posts return from API with same ID")
	}
	p := posts[0]

	// Make set of blog names
	blogNames := make(map[string]struct{})
	for _, note := range p.PostNotes() {
		blogNames[note.BlogURL] = struct{}{}
	}

	// Uniquify blog names
	uniqueBlogNames := []string{}
	for name, _ := range blogNames {
		uniqueBlogNames = append(uniqueBlogNames, name)
	}

	return uniqueBlogNames, nil
}

func getCredentials() tumblr.APICredentials {
	c := tumblr.APICredentials{
		Key:    os.Getenv("ALIKER_KEY"),
		Secret: os.Getenv("ALIKER_SECRET"),
	}
	if c.Key == "" || c.Secret == "" {
		msg := "ALIKER_KEY and ALIKER_SECRET variables unset"
		panic(msg)
	}
	return c
}
