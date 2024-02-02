// Updater is console application to change Strava activities.
//
// Usage example:
//
//	updater -accessToken <access_token> -from 2021-01-01 -to 2021-12-31
//
// See strava-updater.http on how to get access token.
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

var (
	accessToken = flag.String("accessToken", "", "Strava access_token with read and write activities permissions.")
	from        = flag.String("from", "", "Start date in format 'YYYY-MM-DD'.")
	to          = flag.String("to", "", "Finish date in format 'YYYY-MM-DD'.")
	debug       = flag.Bool("debug", false, "Print debug information.")
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: updater [options]\n")
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime)
	log.SetPrefix("updater: ")

	flag.Usage = usage
	flag.Parse()

	if *accessToken == "" {
		log.Printf("AccessToken is required\n\n")
		usage()
	}

	if *from == "" || *to == "" {
		log.Printf("Both from and to dates are required\n\n")
		usage()
	}

	fromDate, err := time.Parse(time.DateOnly, *from)
	if err != nil {
		log.Printf("Failed to parse from date: %v\n", err)
		usage()
	}
	toDate, err := time.Parse(time.DateOnly, *to)
	if err != nil {
		log.Printf("Failed to parse to date: %v\n\n", err)
		usage()
	}

	client, err := strava.New(*accessToken, http.DefaultClient, *debug)
	if err != nil {
		log.Fatalf("Failed to create Strava client: %v", err)
	}

	athlete, err := client.Athlete(context.Background())
	if err != nil {
		log.Fatalf("Failed to get athlete: %v", err)
	}
	log.Printf("Current logger in athlete: %#v\n", athlete)

	var activitiesFromTo []strava.SummaryActivity
	for page := 1; ; page++ {
		activities, hasNext, err := client.Activities(context.Background(), fromDate, toDate, page)
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

	log.Printf("Number of activities from %v to %v: %d\n", fromDate, toDate, len(activitiesFromTo))

	translator := translate.New()
	var translated int
	for _, activity := range activitiesFromTo {
		log.Printf("Activity: %#v\n", activity)

		trActivityName := translator.ActivityName(activity.Name)
		if trActivityName == activity.Name {
			continue
		}

		log.Printf("Translating name: %q -> %q\n", activity.Name, trActivityName)
		err = client.UpdateActivity(context.Background(), activity.ID, trActivityName)
		if err != nil {
			log.Fatalf("Failed to update activity=%d: %v", activity.ID, err)
		}
		translated++
	}

	log.Printf("Number of translated activities: %d\n", translated)
}
