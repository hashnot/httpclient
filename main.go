package main

import (
	"bitbucket.org/hashnot/httpclient/httpclient"
	"flag"
	"github.com/hashnot/function"
	"log"
)

func failOnErr(err error, msg string) {
	if err != nil {
		log.Fatalf("%s %s", msg, err)
	}
}

func main() {
	client := new(httpclient.HttpClient)
	err := function.UnmarshalFile("httpclient.yaml", client)
	failOnErr(err, "")

	verbose := flag.Bool("verbose", false, "Verbose")
	flag.Parse()

	err = client.Setup(*verbose)
	failOnErr(err, "Failed to start")

	err = function.StartWithConfig(client, client.Function)
	failOnErr(err, "Failed to start")
}
