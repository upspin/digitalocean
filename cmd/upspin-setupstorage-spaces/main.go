package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"upspin.io/subcmd"
)

const help = `
setupstorage-spaces is the second step in establishing an upspinserver.
It sets up DigitalOcean's Spaces storage for your Upspin installation. You may skip this step
if you wish to store Upspin data on your server's local disk.
The first step is 'setupdomain' and the final step is 'setupserver'.
setupstorage-spaces updates the server configuration files in $where/$domain/ to use
the specified spaces and region.
Before running this command, you should ensure you have an Digital Ocean account 
with proper Spaces's API key and secret.
`

type state struct {
	*subcmd.State
}

func main() {
	const name = "setupstorage-spaces"

	log.SetFlags(0)
	log.SetPrefix("upspin setupstorage-spaces: ")

	s := &state{
		State: subcmd.NewState(name),
	}

	var (
		where  = flag.String("where", filepath.Join(os.Getenv("HOME"), "upspin", "deploy"), "`directory` to store private configuration files")
		domain = flag.String("domain", "", "domain `name` for this Upspin installation")
		region = flag.String("region", "nyc3", "region for the Spaces' name")
		root   = flag.String("root", "", "root for the Spaces' path")
	)

	s.ParseFlags(flag.CommandLine, os.Args[1:], help,
		"setupstorage-spaces -domain=<name> [-region=<region>] <bucket_name>")
	if flag.NArg() != 1 {
		s.Exitf("a single bucket name must be provided")
	}
	if len(*domain) == 0 {
		s.Exitf("the -domain flag must be provided")
	}

	spaceName := flag.Arg(0)

	cfgPath := filepath.Join(*where, *domain)
	cfg := s.ReadServerConfig(cfgPath)

	cfg.StoreConfig = []string{
		"backend=Spaces",
		"spacesName=" + spaceName,
		"spacesRegion=" + *region,
		"spacesRoot=" + *root,
	}
	s.WriteServerConfig(cfgPath, cfg)
}
