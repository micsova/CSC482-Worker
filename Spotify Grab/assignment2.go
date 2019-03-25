package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/jamesPEarly/loggly"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Artist struct {
	Name string `json:"name"`
	HREF string `json:"href"`
	ID   string `json:"id"`
}

type Album struct {
	Name    string   `json:"name"`
	Artists []Artist `json:"artists"`
	ID      string   `json:"id"`
	HREF    string   `json:"href"`
}

type Track struct {
	Name     string   `json:"name"`
	Artists  []Artist `json:"artists"`
	Album    Album    `json:"album"`
	TrackNum int      `json:"track_number"`
	ID       string   `json:"id"`
	HREF     string   `json:"href"`
}

type Follower struct {
	Total int `json:"total"`
}

//type Item struct {
//	TrackList Track `json:"track"`
//}
//
//type Tracks struct {
//	Items []Item `json:"items"`
//}

type Playlist struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	TrackList   struct {
		Items []struct {
			Song Track `json:"track"`
		} `json:"items"`
	} `json:"tracks"`
	//TrackList	Tracks		`json:"tracks"`
	ID        string   `json:"id"`
	HREF      string   `json:"href"`
	Followers Follower `json:"followers"`
}

const endpoint = "https://api.spotify.com/v1/playlists/37i9dQZF1DX4JAvHpjipBk"

func main() {
	// Check for token
	fmt.Println("JPE--LOGGLY_TOKEN:", os.Getenv("LOGGLY_TOKEN"))
	for {
		tag := "Spotify"
		logglyClient := loggly.New(tag)
		client := &http.Client{}

		//Get access token using refresh token
		token := getToken(client)

		//Use access token to make a request
		playlist := getPlaylist(token, client)

		/*
			//Print data locally
			var sb strings.Builder
			sb.WriteString("Name: " + playlist.Name + "\n")
			sb.WriteString("Description: " + playlist.Description + "\n")
			sb.WriteString("Followers: " + strconv.Itoa(playlist.Followers.Total) + "\n")
			sb.WriteString("Tracks: (" + strconv.Itoa(len(playlist.SongList.Items.Tracks)) + ")\n")
			for _, song := range playlist.SongList.Items.Tracks {
				sb.WriteString("    " + song.Name + " by ")
				for _, artist := range song.Artists {
					sb.WriteString(artist.Name + ", ")
				}
				sb.WriteString("\n")
			}
			fmt.Print(sb.String())
		*/

		//Loggly Reporting
		logglyClient = loggly.New("Data")
		_ = logglyClient.Send("info", "{\n\"name\": "+playlist.Name+"\",\n"+
			"\"followers\": "+strconv.Itoa(playlist.Followers.Total)+"\n}")

		//Create a DynamoDB session
		sess, err := session.NewSession(&aws.Config{
			Region: aws.String("us-east-1")},
		)

		if err != nil {
			fmt.Println("Error creating session:")
			fmt.Println(err.Error())
			os.Exit(1)
		}
		svc := dynamodb.New(sess)

		//Put the playlist info into the DynamoDB table
		table(playlist, svc)

		//Wait 15 minutes before polling again
		time.Sleep(15 * time.Minute)
	}
}

func getToken(client *http.Client) string {
	dat := url.Values{}
	dat.Add("grant_type", "refresh_token")
	dat.Add("refresh_token", "AQAYxcApQyHMgY5M1q9sBCKRYCXwEF6ez5kvNPlQD1Oyd6_6H0TAxwJLJnJZ5cQWWar47gQMr_06YqFTW0sCIYztQTt1XePiaPURAqZiuhU9eYYfAIBC7sXweODnJ1vIWEeTpA")
	req, _ := http.NewRequest(http.MethodPost, "https://accounts.spotify.com/api/token", strings.NewReader(dat.Encode()))
	req.Header.Add("Authorization", "Basic ZjNmOGQ3MWNiZDQ2NDYwNGExZTA0MTJlZGMxM2IzOGU6M2MwMGFlNmJjNjVkNGM4ZWE2YTY4YTFhYTE4NDVkY2Y=")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	res, _ := client.Do(req)
	body, _ := ioutil.ReadAll(res.Body)
	stringFile := []byte(string(body))
	var data map[string]interface{}
	if err := json.Unmarshal(stringFile, &data); err != nil {
		fmt.Println(err)
	}
	token := data["access_token"].(string)
	return token
}

func getPlaylist(token string, client *http.Client) Playlist {
	req, _ := http.NewRequest(http.MethodGet, endpoint, nil)
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	res, _ := client.Do(req)
	body, _ := ioutil.ReadAll(res.Body)

	//Parse the resulting JSON file
	stringFile := []byte(string(body))
	//fmt.Println(string(stringFile))
	var playlist Playlist
	if err := json.Unmarshal(stringFile, &playlist); err != nil {
		fmt.Println(err)
	}
	fmt.Println("Got playlist")
	return playlist
}

func table(playlist Playlist, svc *dynamodb.DynamoDB) {
	// Add each item to Movies table:
	av, err := dynamodbattribute.MarshalMap(playlist)

	if err != nil {
		fmt.Println("Got error marshalling map:")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	// Create item in table Movies
	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String("Spotify-New-Music-Playlist"),
	}

	_, err = svc.PutItem(input)

	if err != nil {
		fmt.Println("Got error calling PutItem:")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println("Successfully added '", playlist.Name, "' (", playlist.Followers, ") to Spotify-New-Music-Playlist table")
}
