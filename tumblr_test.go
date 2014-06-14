package main

import (
	"testing"
)

func TestExtractPostId(t *testing.T) {
	testUrls := []string{
		"http://lacecard.tumblr.com/post/76803575816/emacs-in-tron",
		"http://lacecard.tumblr.com/post/76803575816",
		"lacecard.tumblr.com/post/76803575816/emacs-in-tron",
	}
	for _, u := range testUrls {
		bh, pid, err := extractPostId(u)
		if bh != "lacecard.tumblr.com" {
			t.Errorf("Did not parse lacecard.tumblr.com (got %s)", bh)
			return
		}
		if pid != 76803575816 {
			t.Errorf("Did not parse correct PID (got %d)", pid)
			return
		}
		if err != nil {
			t.Error(err)
			return
		}
	}
}
