package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/philips/go-ghost"
	"github.com/writeas/go-mobiledoc"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Fprintf(os.Stderr, "usage: sample GHOST_URL API_KEY POST_ID_TO_EDIT\n")
		os.Exit(1)
	}

	url := os.Args[1]
	apiKey := os.Args[2]
	id := os.Args[3]

	c := ghost.NewClient(url, apiKey)

	// 1. Get the UpdatedAt field which is required for a PUT
	resp, err := c.Request(http.MethodGet, c.EndpointForID("admin", "posts", id), nil)
	if err != nil {
		log.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()

	pr := &ghost.PostRequest{}
	err = json.NewDecoder(resp.Body).Decode(pr)
	if err != nil {
		log.Fatalf("decode: %v", err)
	}
	fmt.Printf("title: %v Updated At: %v\n", *pr.Posts[0].Title, pr.Posts[0].UpdatedAt.String())

	// 2. Generate new Post content
	md := `This is a **test post** made with the [go-ghost](https://github.com/writeas/go-ghost) library.`

	mobdoc, err := json.Marshal(mobiledoc.FromMarkdown(md))
	if err != nil {
		log.Fatalf("decode: %v", err)
	}

	presp := ghost.PostRequest{
		Posts: []ghost.Post{
			ghost.Post{
				Title:     ghost.String("Asset Transparency Log Metrics"),
				UpdatedAt: pr.Posts[0].UpdatedAt,
				Mobiledoc: ghost.String(string(mobdoc)),
			}},
	}

	// 3. PUT the new Post contents to the ID
	resp, err = c.Request(http.MethodPut, c.EndpointForID("admin", "posts", id), presp)
	if err != nil {
		fmt.Printf("UpdatePost: %v\n", err)
	}
	defer resp.Body.Close()

	rbody, err := ioutil.ReadAll(resp.Body)
	fmt.Printf("response: %s\n", rbody)
}
