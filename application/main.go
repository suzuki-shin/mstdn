package main

import (
	"context"
	"fmt"
	"log"

	"github.com/mattn/go-mastodon"
)

func main() {
	app, err := mastodon.RegisterApp(context.Background(), &mastodon.AppConfig{
		Server:     "https://friends.nico",
		ClientName: "mstdn",
		Scopes:     "read write follow",
		Website:    "",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("client-id    : %s\n", app.ClientID)
	fmt.Printf("client-secret: %s\n", app.ClientSecret)
}
