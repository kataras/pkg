package main

import (
	"github.com/kataras/pkg/geoloc"
)

func main() {
	// http://localhost:8080/138.197.168.210
	//
	// http://localhost:8080/me
	// note that the above '/me' path will not work if you try to visit it from your localhost.
	geoloc.Listen(":8080")
}
