package main

import (
	"fmt"
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


func usersLikingPost(baseHostname string, postId int64) {

}
