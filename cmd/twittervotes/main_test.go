package main

import "testing"

func TestGenerateTweets(t *testing.T) {
	people := make([]string, 0)
	for p := range generateTweets() {
		people = append(people, p)
	}

	if len(people) != 41 {
		t.Errorf("Expected length 41, got %d", len(people))
	}
}
