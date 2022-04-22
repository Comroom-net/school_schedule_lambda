package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)
	
var (
	// DefaultHTTPGetAddress Default Address
	DefaultHTTPGetAddress = "https://checkip.amazonaws.com"

	// ErrNoIP No IP found in response
	ErrNoIP = errors.New("No IP in HTTP response")

	// ErrNon200Response non 200 status code in response
	ErrNon200Response = errors.New("Non 200 Response found")
)

func router(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	
	svc := dynamodb.New(sess)
	switch req.HTTPMethod {
		case "GET":
			return get_handler(req, svc)
		case "POST":
			return post_handler(req, svc)
		default:
			return events.APIGatewayProxyResponse{}, errors.New("StatusMethodNotAllowed")
	}
}


type Schedule struct {
	School_id int
	Hours     string
}

func post_handler(request events.APIGatewayProxyRequest, svc *dynamodb.DynamoDB) (events.APIGatewayProxyResponse, error) {
	// TODO: put item to dynamodb
	var schedule Schedule
	json.Unmarshal([]byte(request.Body), &schedule)
	schedule.School_id, _ = strconv.Atoi(request.PathParameters["sid"])
	tableName := "comroom_schedule"

	log.Printf("schedule: %v", schedule)

	item, err := dynamodbattribute.MarshalMap(schedule)
    if err != nil {
        log.Fatalf("Got error marshalling map: %s", err)
    }
	log.Printf("item: %v", item)

    // Create item in table Movies
    input := &dynamodb.PutItemInput{
        Item: item,
        TableName: aws.String(tableName),
    }

    _, err = svc.PutItem(input)
    if err != nil {
        log.Fatalf("Got error calling PutItem: %s", err)
    }

    school_id := strconv.Itoa(schedule.School_id)

    fmt.Println("Successfully added '" + schedule.Hours + "' (" + school_id + ") to table " + tableName)

	return events.APIGatewayProxyResponse{
		Body:       fmt.Sprintf("Post Good, %v", request.Body),
		StatusCode: 200,
	}, nil
}

func get_handler(request events.APIGatewayProxyRequest, svc *dynamodb.DynamoDB) (events.APIGatewayProxyResponse, error) {
	school_id := request.PathParameters["sid"]
	tableName := "comroom_schedule"

	result, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"School_id": {
				N: aws.String(school_id),
			},
		},
	})
	if err != nil {
		log.Fatalf("Got error calling GetItem: %s", err)
	}

	if result.Item == nil {
		msg := "Could not find "
		return events.APIGatewayProxyResponse{}, errors.New(msg)
	}
		
	item := Schedule{}
	
	err = dynamodbattribute.UnmarshalMap(result.Item, &item)
	if err != nil {
		panic(fmt.Sprintf("Failed to unmarshal Record, %v", err))
	}
	
	fmt.Println("Found item:")
	fmt.Println("school_id:  ", item.School_id)
	fmt.Println("hours: ", item.Hours)
	


	return events.APIGatewayProxyResponse{
		Body:       fmt.Sprintf("Get Good, %v", request.Body),
		StatusCode: 200,
	}, nil
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	resp, err := http.Get(DefaultHTTPGetAddress)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	if resp.StatusCode != 200 {
		return events.APIGatewayProxyResponse{}, ErrNon200Response
	}

	ip, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	if len(ip) == 0 {
		return events.APIGatewayProxyResponse{}, ErrNoIP
	}

	return events.APIGatewayProxyResponse{
		Body:       fmt.Sprintf("Hello, %v", string(ip)),
		StatusCode: 200,
	}, nil
}

func main() {
	lambda.Start(router)
}
