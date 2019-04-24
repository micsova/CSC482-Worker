package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/gorilla/mux"
	"net/http"
	"os"
	"strconv"
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

type Playlist struct {
	TableID     string `json:"unique_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	TrackList   struct {
		Items []struct {
			Song Track `json:"track"`
		} `json:"items"`
	} `json:"tracks"`
	//TrackList	Tracks		`json:"tracks"`
	PlaylistID string `json:"id"`
	HREF       string `json:"href"`
	Followers  struct {
		Total int `json:"total"`
	} `json:"followers"`
}

type Item struct {
	Playlists []Playlist `json:"items"`
}

var svc *dynamodb.DynamoDB

func main() {
	//Router setup
	//Initialize
	router := mux.NewRouter()
	//Route handlers (endpoints)
	router.HandleFunc("/msova/all", getAll).Methods("GET")
	router.HandleFunc("/msova/status", getStatus).Methods("GET")
	//Create a DynamoDB session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")},
	)
	if err != nil {
		fmt.Println("Error creating session:")
		fmt.Println(err.Error())
		os.Exit(1)
	}
	// Create DynamoDB client
	svc = dynamodb.New(sess)
	//Run server
	if err := http.ListenAndServe(":8080", router); err != nil {
		fmt.Println(err)
	}
}

//Gets JSON representation of the entire table
func getAll(w http.ResponseWriter, r *http.Request) {
	//Define what you want to get
	proj := expression.NamesList(expression.Name("name"),
		expression.Name("description"), expression.Name("followers"), expression.Name("tracks"),
		expression.Name("id"), expression.Name("href"), expression.Name("unique_id"))

	expr, err := expression.NewBuilder().WithProjection(proj).Build()
	if err != nil {
		fmt.Println("Got error building expression:")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	params := &dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String("SpotifyGrab"),
	}

	result, err := svc.Scan(params)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	var playlists []Playlist
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &playlists)

	ba, err := json.MarshalIndent(playlists, "", "   ")
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(ba)
}

//Gets status of the table
func getStatus(w http.ResponseWriter, r *http.Request) {
	result, err := svc.Scan(&dynamodb.ScanInput{
		TableName: aws.String("SpotifyGrab"),
	})

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	var playlists []Playlist
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &playlists)

	//Create JSON-formatted string with table name and item count
	info := []byte(`{"table":"SpotifyGrab","recordCount": `)
	info = append(info, []byte(strconv.Itoa(len(playlists)))...)
	info = append(info, []byte(`}`)...)
	var prettyJson bytes.Buffer
	_ = json.Indent(&prettyJson, info, "", "   ")

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(prettyJson.Bytes())
}
