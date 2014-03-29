package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/cellofellow/gopiano"
	"github.com/cellofellow/gopiano/responses"
	"github.com/GeertJohan/go.linenoise"
	"github.com/howeyc/gopass"
	"github.com/turbowookie/gompd/mpd"
)

var pandora *gopiano.Client

func main() {
	// Create client and login partner.
	pandora, err := gopiano.NewClient(gopiano.AndroidClient)
	if err != nil {
		fmt.Println("Unexpected error: %v", err)
		os.Exit(1)
	}

	_, err = pandora.AuthPartnerLogin()
	if err != nil {
		fmt.Println("Unexpected error: %v", err)
		os.Exit(1)
	}

	fmt.Println("Welcome to pianopub! Press ? for a list of commands.")
	email, err := linenoise.Line("[?] Email: ")
	if err == linenoise.KillSignalError {
		os.Exit(0)
	}
	if err != nil {
		fmt.Println("Unexpected error: %v", err)
		os.Exit(1)
	}
	fmt.Printf("[?] Password: ")
	password := string(gopass.GetPasswd())

	fmt.Printf("(i) Login... ")
	_, err = pandora.AuthUserLogin(email, password)
	if err != nil {
		fmt.Println("Unexpected error: %v", err)
		os.Exit(1)
	}
	fmt.Printf("Ok.\n")

	fmt.Printf("(i) Get stations... ")
	stations, err := pandora.UserGetStationList(true)
	if err != nil {
		fmt.Println("Unexpected error: %v", err)
		os.Exit(1)
	}
	fmt.Printf("Ok\n")

	sort.Sort(stations.Result.Stations)

	var stationChoices map[int]responses.Station = make(map[int]responses.Station, 20)
	var quickmixStations map[string]bool = make(map[string]bool, 20)
	for i, s := range stations.Result.Stations {
		stationChoices[i] = s
		if s.IsQuickMix {
			for _, id := range s.QuickMixStationIDs {
				quickmixStations[id] = true
			}
		}
	}

	fstring := "\t%2d) %s%s %s\n"
	for i := 0; i < len(stationChoices); i++ {
		s := stationChoices[i]
		var inQuickMix, isQuickMix string
		if _, yes := quickmixStations[s.StationID]; yes {
			inQuickMix = "q"
		} else {
			inQuickMix = " "
		}
		if s.IsQuickMix {
			isQuickMix = "Q"
		} else {
			isQuickMix = " "
		}
		fmt.Printf(fstring, i, inQuickMix, isQuickMix, s.StationName)
	}

	str, err := linenoise.Line("[?] Select station: ")
	if err == linenoise.KillSignalError {
		os.Exit(0)
	}
	if err != nil {
		fmt.Println("Unexpected error: %v", err)
		os.Exit(1)
	}

	fields := strings.Fields(str)
	choice, err := strconv.ParseInt(fields[0], 10, 0)
	if err != nil {
		fmt.Println("You must enter an integer.")
		os.Exit(1)
	}
	var station responses.Station
	if s, ok := stationChoices[int(choice)]; ok {
		station = s
	} else {
		fmt.Println("Station %d not found.", choice)
		os.Exit(1)
	}

	playlistResponse, err := pandora.StationGetPlaylist(station.StationToken)
	if err != nil {
		fmt.Println("Unexpected error: %v", choice)
		os.Exit(1)
	}

	mpdclient, err := mpd.Dial("tcp", "localhost:6600")
	if err != nil {
		fmt.Println("Unexpected error: %v", err)
		os.Exit(1)
	}
	defer mpdclient.Close()
	for _, item := range playlistResponse.Result.Items {
		fmt.Printf("%s by %s from %s\n", item.SongName, item.ArtistName, item.AlbumName)
		err = mpdclient.Add(item.AudioURLMap["mediumQuality"].AudioURL)
		if err != nil {
			fmt.Println("Unexpected error: %v", err)
			os.Exit(1)
		}

	}
	err = mpdclient.Play(0)
	if err != nil {
		fmt.Println("Unexpected error: %v", err)
		os.Exit(1)
	}
}

