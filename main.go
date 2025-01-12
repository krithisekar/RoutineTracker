package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type FunctionURLRequest struct {
	Version               string            `json:"version"`
	RawPath               string            `json:"rawPath"`
	Headers               map[string]string `json:"headers"`
	QueryStringParameters map[string]string `json:"queryStringParameters"`
	RequestContext        struct {
		HTTP struct {
			Method string `json:"method"`
		} `json:"http"`
	} `json:"requestContext"`
	Body string `json:"body"`
}

type FunctionURLResponse struct {
	StatusCode int               `json:"statusCode"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
}

type Routine struct {
	Date        string `json:"date" dynamodbav:"date"`
	Time        string `json:"time" dynamodbav:"time"`
	RoutineText string `json:"text" dynamodbav:"text"`
}

var db *dynamodb.DynamoDB
var tableName string

func init() {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("ap-south-1"),
	}))
	db = dynamodb.New(sess)

	tableName = os.Getenv("TABLE_NAME")
	if tableName == "" {
		fmt.Println("Error: Table_Name environment variable is not set")
		panic("TABLE_NAME environment variable required")
	}
}

func getCommonHeaders() map[string]string {
	return map[string]string{
		"Access-Control-Allow-Origin":  "*", // Replace with your S3 website URL in production
		"Access-Control-Allow-Methods": "OPTIONS,POST,GET",
		"Access-Control-Allow-Headers": "Content-Type",
		"Content-Type":                 "application/json",
	}
}

func HandleRequest(ctx context.Context, request FunctionURLRequest) (FunctionURLResponse, error) {
	fmt.Printf("HandleRequest() invoked via method '%s' and path '%s' \n",
		request.RequestContext.HTTP.Method, request.RawPath)

	switch request.RequestContext.HTTP.Method {
	case "POST":
		return addRoutineHandler(request)
	case "GET":
		return getRoutinesHandler(request)
	case "OPTIONS":
		return handleOptionsRequest()
	default:
		return FunctionURLResponse{
			StatusCode: http.StatusMethodNotAllowed,
			Headers:    getCommonHeaders(),
			Body:       `{"error":"Method not allowed"}`,
		}, nil
	}
}

func addRoutineHandler(request FunctionURLRequest) (FunctionURLResponse, error) {
	var routine Routine
	err := json.Unmarshal([]byte(request.Body), &routine)
	if err != nil {
		return FunctionURLResponse{
			StatusCode: http.StatusBadRequest,
			Headers:    getCommonHeaders(),
			Body:       `{"error":"Invalid request"}`,
		}, nil
	}

	av, err := dynamodbattribute.MarshalMap(routine)
	if err != nil {
		fmt.Println("Error marshalling routine:", err)
		return FunctionURLResponse{
			StatusCode: http.StatusInternalServerError,
			Headers:    getCommonHeaders(),
			Body:       `{"error":"Failed to process routine"}`,
		}, nil
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(tableName),
	}

	_, err = db.PutItem(input)
	if err != nil {
		fmt.Println("Error putting into DynamoDB", err)
		return FunctionURLResponse{
			StatusCode: http.StatusInternalServerError,
			Headers:    getCommonHeaders(),
			Body:       `{"error":"failed to add routine"}`,
		}, nil
	}

	return FunctionURLResponse{
		StatusCode: http.StatusCreated,
		Headers:    getCommonHeaders(),
		Body:       `{"message":"Routine added successfully"}`,
	}, nil
}

func getRoutinesHandler(request FunctionURLRequest) (FunctionURLResponse, error) {
	date, exists := request.QueryStringParameters["date"]
	if !exists || date == "" {
		return FunctionURLResponse{
			StatusCode: http.StatusBadRequest,
			Headers:    getCommonHeaders(),
			Body:       `{"error":"Missing date query parameter"}`,
		}, nil
	}
	input := &dynamodb.QueryInput{
		TableName:              aws.String(tableName),
		KeyConditionExpression: aws.String("#d = :date"),
		ExpressionAttributeNames: map[string]*string{
			"#d": aws.String("date"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":date": {
				S: aws.String(date),
			},
		},
	}

	result, err := db.Query(input)
	if err != nil {
		fmt.Printf("Error querying DB for date %s: %v\n", date, err)
		return FunctionURLResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       `{"error":"Failed to fetch routines"}`,
		}, nil
	}
	routines := []Routine{}
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &routines)
	if err != nil {
		fmt.Printf("Error Unmarshalling DB items for date %s: %v\n", date, err)
		return FunctionURLResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       `{"error":"Failed to process routines"}`,
		}, nil
	}
	responseBody, err := json.Marshal(routines)
	if err != nil {
		fmt.Println("Error marshalling response", err)
		return FunctionURLResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       `{"error":"Failed to prepare response"}`,
		}, nil
	}

	return FunctionURLResponse{
		StatusCode: http.StatusOK,
		Headers:    getCommonHeaders(),
		Body:       string(responseBody),
	}, nil

}

func handleOptionsRequest() (FunctionURLResponse, error) {
	return FunctionURLResponse{
		StatusCode: http.StatusOK,
		Headers:    getCommonHeaders(),
		Body:       "",
	}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
