package main

import (
	"flag"
	"fmt"

	"github.com/kataras/pkg/config"
)

type (
	// the Host is the only one set-ed by yaml file
	// so it shouldn't ask for that, but the rest will be asked for their values.
	databaseCredentials struct {
		Username string `yaml:"Username"`
		// that tag's value will replace writing chars to *****, you can use "secret" as well
		Password string `yaml:"Password" config:"password"`
		Host     string `yaml:"Host"`
		DBName   string `yaml:"DBName"`
	}

	configuration struct {
		// a required string which is not defaulted.
		Addr string `yaml:"Addr"`
		// a non required string which is defaulted
		ServerName string `yaml:"ServerName"`
		// a required string which is loaded from the file.
		Author string `yaml:"Author"`
		// a boolean.
		Debug bool `yaml:"Debug"`
		// a required boolean, should be asked from survey and have a default value of 'false'.
		Required bool `yaml:"Required"`

		// a required integer, should be set-ed by the command line as flag's value, if not then it will be asked from survey.
		Year int `yaml:"Year"`
		//

		// unexported fields are ignored.
		unexportedIgnored string `yaml:"unexportedIgnored"`
		// this is ignored by both yaml (and survey).
		ExportedIgnored string `yaml:"-" config:"-"`

		// a struct with fields which should be parsed correctly.
		DBCredentials databaseCredentials `yaml:"DBCredentials"`
	}
)

// $ go run main.go -year 2017
func main() {
	flag.Int("year", 0, "set the year")

	c := configuration{ServerName: "iris: https://iris-go.com"}
	if err := config.Load("./config_example.yml", &c, config.WithFlags(flag.CommandLine)); err != nil {
		panic(err)
	}

	fmt.Println("Configuration loaded:")
	fmt.Printf("%#+v\n", c)
}
