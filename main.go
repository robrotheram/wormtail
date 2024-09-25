// // The tshello server demonstrates how to use Tailscale as a library.
package main

import (
	"log"
	"warptail/pkg/api"
	"warptail/pkg/router"
	"warptail/pkg/utils"
)

func main() {
	config := utils.LoadConfig()
	r, err := router.NewRouter(config)
	if err != nil {
		log.Fatalf("unable to start %v", err)
	}
	defer r.Close()
	r.StartAll()
	server := api.NewApi(r, config.Dasboard)
	server.Start(":8081")
}
