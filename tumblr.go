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
func blogsLikingPost(baseHostname string, postId int64) []string {
	client := tumblr.Tumblr{
		Credentials: getCredentials(),
	}
	blog := client.NewBlog(baseHostname)
	params := tumblr.PostRequestParams{
		Id:        postId,
		NotesInfo: true,
	}
	posts, err := blog.Posts(params)
	ensureNil(err)
	// posts must be of length 1 because we specified an ID
	p := posts[0].(map[string]interface{})
	notes := p["notes"].([]interface{})

	blogNames := make(map[string]struct{})
	for _, rawNote := range notes {
		note := rawNote.(map[string]interface{})
		blogName := note["blog_name"].(string)
		blogNames[blogName] = struct{}{}
	}

	uniqueBlogNames := []string{}
	for name, _ := range blogNames {
		uniqueBlogNames = append(uniqueBlogNames, name)
	}

	// Filter out only "likes" and "reblogs"
	return uniqueBlogNames
}

func getCredentials() tumblr.APICredentials {
	return tumblr.APICredentials{
		Key:    os.Getenv("ALIKER_KEY"),
		Secret: os.Getenv("ALIKER_SECRET"),
	}
}
