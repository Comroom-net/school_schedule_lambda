package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/aws"
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
			return handler(req)
		default:
			return events.APIGatewayProxyResponse{}, errors.New("StatusMethodNotAllowed")
	}
}


type Schedule struct {
	School_id int
	Hours     string
}

func post_handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// TODO: put item to dynamodb
	return events.APIGatewayProxyResponse{
		Body:       fmt.Sprintf("Hello, %v", request.Body),
		StatusCode: 200,
	}, nil
}

func get_handler(request events.APIGatewayProxyRequest, svc *dynamodb.DynamoDB) (events.APIGatewayProxyResponse, error) {
	school_id := request.PathParameters["sid"]
	tableName := "comroom_schedule"

	result, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"school_id": {
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
		Body:       fmt.Sprintf("Hello, %v", request.Body),
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
