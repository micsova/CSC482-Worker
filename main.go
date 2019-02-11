package main

import (
	"encoding/json"
	"fmt"
	"github.com/jamesPEarly/loggly"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type Artist struct {
	Name string
}

type Album struct {
	Name string
	Artists []Artist
	ID string
	HREF string
}

type Track struct {
	Name string
	Artists []Artist
	Album Album
	TrackNumber int
	ID string
	HREF string
}

type Playlist struct {
	Name string
	Description string
	Tracks []Track
	TotalTracks int
	Followers int
	ID string
	HREF string
}

func main() {
	// Check for token
	fmt.Println("JPE--LOGGLY_TOKEN:", os.Getenv("LOGGLY_TOKEN"))
	//for true {
		tag := "Spotify"
		logglyClient := loggly.New(tag)
		client := &http.Client{}

		//Get access token using refresh token
		dat := url.Values{}
		dat.Add("grant_type", "refresh_token")
		dat.Add("refresh_token", "AQAYxcApQyHMgY5M1q9sBCKRYCXwEF6ez5kvNPlQD1Oyd6_6H0TAxwJLJnJZ5cQWWar47gQMr_06YqFTW0sCIYztQTt1XePiaPURAqZiuhU9eYYfAIBC7sXweODnJ1vIWEeTpA")
		req, _ := http.NewRequest(http.MethodPost, "https://accounts.spotify.com/api/token", strings.NewReader(dat.Encode()))
		req.Header.Add("Authorization", "Basic ZjNmOGQ3MWNiZDQ2NDYwNGExZTA0MTJlZGMxM2IzOGU6M2MwMGFlNmJjNjVkNGM4ZWE2YTY4YTFhYTE4NDVkY2Y=")
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		res, _ := client.Do(req)
		body, _ := ioutil.ReadAll(res.Body)
		byt := []byte(string(body))
		var data map[string]interface{}
		if err := json.Unmarshal(byt, &data); err != nil {
			fmt.Println(err)
		}
		token := data["access_token"].(string)

		//Use access token to make a request
		endpoint := "https://api.spotify.com/v1/playlists/37i9dQZF1DX4JAvHpjipBk"
		req, _ = http.NewRequest(http.MethodGet, endpoint, nil)
		req.Header.Add("Authorization", "Bearer "+token)
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		res, _ = client.Do(req)
		body, _ = ioutil.ReadAll(res.Body)

		//Parse the resulting JSON file
		byt = []byte(string(body))
		var file interface{}
		if err := json.Unmarshal(byt, &file); err != nil {
			fmt.Println(err)
		}
		//fmt.Println(file)    // string interface matches values to keys   //Gets keys/values inside overall "albums" object
		f := file.(map[string]interface{})
		var playlist Playlist
		playlistName := f["name"].(string)
		playlistDescription := f["description"].(string)
		playlistFollowers := int(f["followers"].(map[string]interface{})["total"].(float64))
		playlistID := f["id"].(string)
		playlistHREF := f["href"].(string)
		var playlistTracks []Track
		for key, value := range file.(map[string]interface{})["tracks"].(map[string]interface{}) { //Iterate through tracks data
			if key == "items" {
				for _, item := range value.([]interface{}) { //Iterate through tracks
					for k, v := range item.(map[string]interface{}) {
						if k == "track" {
							track := v.(map[string]interface{})
							trackName := track["name"].(string)
							trackID := track["id"].(string)
							trackHREF := track["href"].(string)
							trackNumber := int(track["track_number"].(float64))
							var trackArtists []Artist
							for _, art := range track["artists"].([]interface{}) {
								artist := art.(map[string]interface{})
								trackArtists = append(trackArtists, Artist{
									Name: artist["name"].(string),
								})
							}
							album := v.(map[string]interface{}) //Get album of track
							albumID := album["id"].(string)
							albumHREF := album["href"].(string)
							albumName := album["name"].(string)
							var albumArtists []Artist
							for _, art := range album["artists"].([]interface{}) { //Iterate through artists
								artist := art.(map[string]interface{})
								albumArtists = append(albumArtists, Artist{ //Create list of artists for the album
									Name: artist["name"].(string),
								})
							}
							trackAlbum := Album{
								ID:      albumID,
								HREF:    albumHREF,
								Name:    albumName,
								Artists: albumArtists,
							}
							playlistTracks = append(playlistTracks, Track{
								Name:        trackName,
								Artists:     trackArtists,
								Album:       trackAlbum,
								TrackNumber: trackNumber,
								ID:          trackID,
								HREF:        trackHREF,
							})
						}
					}
				}
			}
		}
		playlist = Playlist {
			Name:playlistName,
			Description:playlistDescription,
			Tracks:playlistTracks,
			TotalTracks:len(playlistTracks),
			Followers:playlistFollowers,
			ID:playlistID,
			HREF:playlistHREF,
		}
		var sb strings.Builder
		sb.WriteString("Name: " + playlist.Name + "\n")
		sb.WriteString("Description: " + playlist.Description + "\n")
		sb.WriteString("Followers: " + strconv.Itoa(playlist.Followers) + "\n")
		sb.WriteString("Tracks: (" + strconv.Itoa(playlist.TotalTracks) + ")\n")
		for _, track := range playlist.Tracks {
			sb.WriteString("    " + track.Name + " by ")
			for _, artist := range track.Artists {
				sb.WriteString(artist.Name + ", ")
			}
			sb.WriteString("\n")
		}
		fmt.Print(sb.String())
		//_ = logglyClient.Send("info", sb.String())
		logglyClient = loggly.New("Data")
		_ = logglyClient.Send("info", "Name=" + playlist.Name + ",Followers=" + strconv.Itoa(playlist.Followers))
	//}
}