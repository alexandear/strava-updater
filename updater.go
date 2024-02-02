// Updater is console application to rewrite Strava activities
//
// Usage:
//
//	updater [options]
//
// The options are:
//
//	-accessToken string
//		Strava access_token with read and write activities permissions.
//
//	-debug
//		Whether to print debug information.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/alexandear/strava-rewriter/strava"
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: updater [options]\n")
	flag.PrintDefaults()
}

var (
	accessToken = flag.String("accessToken", "", "Strava access_token with read and write activities permissions.")
	debug       = flag.Bool("debug", false, "Print debug information.")
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("updater: ")

	flag.Usage = usage
	flag.Parse()

	client, err := strava.New(*accessToken, http.DefaultClient, *debug)
	if err != nil {
		log.Fatalf("Failed to create Strava client: %v", err)
	}
	athlete, err := client.Athlete(context.Background())
	if err != nil {
		log.Fatalf("Failed to get athlete: %v", err)
	}
	log.Printf("Athlete: %#v", athlete)
}
