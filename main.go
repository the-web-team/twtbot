package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"twtbot/karma"
	"twtbot/points"
	"twtbot/rearrange"
)

const Prefix = "!b"

var AuthToken string

var messageHandlers = []interface{}{
	points.HandleMessageCreate,
	rearrange.HandleMessageCreate,
	karma.HandleMessageCreate,
}

var services = []func(){
	points.StartService,
}

func init() {
	AuthToken = os.Getenv("DISCORD_AUTH_TOKEN")

	if AuthToken == "" {
		flag.StringVar(&AuthToken, "t", "", "Bot Token")
		flag.Parse()
	}

	if AuthToken == "" {
		log.Fatal(errors.New("no discord auth token supplied"))
	}
}

func main() {
	client, clientError := NewDiscordClient(AuthToken)
	if clientError != nil {
		log.Fatal(clientError)
	}

	fmt.Println(messageHandlers)

	client.AttachHandlers(messageHandlers)
	client.AttachServices(services)

	if runError := client.Run(); runError != nil {
		log.Fatal(runError)
	}
}
