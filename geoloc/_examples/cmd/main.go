package main

import (
	"flag"
	"fmt"

	"github.com/kataras/pkg/geoloc"
)

// $ go run main.go -ip 138.197.168.210
// > Fetching, pelase wait...
// > loc.Info{...}
// OR
// $ go run main.go
// > Please write a remote IP: 138.197.168.210
// > Fetching, pelase wait...
// > loc.Info{...}
func main() {
	var ip string
	flag.StringVar(&ip, "ip", "", "the remote IP address. Usage: -ip 138.197.168.210")
	flag.Parse()

	if ip == "" {
		fmt.Printf("Please write a remote IP: ")
		fmt.Scan(&ip)
	}

	fmt.Println("Fetching, please wait...")
	if info, ok := geoloc.Fetch(ip); !ok {
		fmt.Printf("Failed...")
	} else {
		fmt.Printf("%#+v", info)
	}
}
