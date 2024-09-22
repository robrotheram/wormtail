// // The tshello server demonstrates how to use Tailscale as a library.
package main

import (
	"log"
	"wormtail/pkg/api"
	"wormtail/pkg/router"
	"wormtail/pkg/utils"
)

var Router *router.Router

func main() {
	config := utils.LoadConfig()
	r, err := router.NewRouter(config)
	if err != nil {
		log.Fatalf("unable to start %v", err)
	}

	Router = r
	defer Router.Close()
	Router.StartAll()

	server := api.NewApi(r, config.Dasboard)
	server.Start(":8081")
}
