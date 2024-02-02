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
	"time"

	"github.com/alexandear/strava-rewriter/strava"
	"github.com/alexandear/strava-rewriter/translate"
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
	log.SetFlags(log.Ldate | log.Ltime)
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
	log.Printf("Current logger in athlete: %#v\n", athlete)

	from := time.Date(2017, time.August, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2018, time.January, 1, 0, 0, 0, 0, time.UTC)

	var activitiesFromTo []strava.SummaryActivity
	for page := 1; ; page++ {
		activities, hasNext, err := client.Activities(context.Background(), from, to, page)
		if err != nil {
			log.Printf("Failed to get activities for page=%d: %v\n", page, err)
			break
		}
		if len(activities) == 0 {
			break
		}
		activitiesFromTo = append(activitiesFromTo, activities...)
		if !hasNext {
			break
		}
	}
	translator := translate.New()
	for _, activity := range activitiesFromTo {
		log.Printf("Activity: %#v\n", activity)
		log.Printf("Translated Activity Name: %#v\n", translator.ActivityName(activity.Name))
	}
}
