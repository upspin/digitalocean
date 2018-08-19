package main

import (
	"upspin.io/cloud/https"
	"upspin.io/serverutil/upspinserver"

	// Storage on Spaces.
	_ "digitalocean.upspin.io/cloud/storage/spaces"
)

func main() {
	ready := upspinserver.Main()
	https.ListenAndServe(ready, https.OptionsFromFlags())
}
