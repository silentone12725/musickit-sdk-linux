// Command go-recommendations shows how to use the SDK libraries directly,
// without running the HTTP engine: it fetches the listener's Listen Now feed
// through sdk/ampapi.
//
//	MUSICKIT_USER_TOKEN=<Music-User-Token> go run ./go-recommendations
//
// The developer token is fetched automatically when MUSICKIT_DEV_TOKEN is unset.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/silentone12725/musickit-sdk-linux/sdk/ampapi"
)

func main() {
	userToken := os.Getenv("MUSICKIT_USER_TOKEN")
	if userToken == "" {
		log.Fatal("set MUSICKIT_USER_TOKEN to the listener's Music-User-Token")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	body, err := ampapi.GetRecommendations(ctx, ampapi.KindRecommendations,
		os.Getenv("MUSICKIT_DEV_TOKEN"), userToken,
		ampapi.RecommendationsOptions{Language: "en-US", Limit: 10})
	if err != nil {
		log.Fatal(err)
	}

	var feed struct {
		Data []struct {
			ID         string `json:"id"`
			Type       string `json:"type"`
			Attributes struct {
				Title struct {
					StringForDisplay string `json:"stringForDisplay"`
				} `json:"title"`
			} `json:"attributes"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &feed); err != nil {
		log.Fatal(err)
	}
	for _, item := range feed.Data {
		fmt.Printf("%-24s %s\n", item.Type, item.Attributes.Title.StringForDisplay)
	}
}
