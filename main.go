package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

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
	if (tableName) == "" {
		fmt.Println("Error: Table_Name environment variable is not set")
		panic(("TABLE_NAME environment variable required"))
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
func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	fmt.Printf("HandleRequest() invoked via request of type '%s' and path '%s' \n", request.HTTPMethod, request.Path)
	switch request.HTTPMethod {
	case "POST":
		return addRoutineHandler(request)
	case "GET":
		fmt.Println("Incoming request matched the HTTP verb GET")
		return getRoutinesHandler(request)
	case "OPTIONS":
		return handleOptionsRequest()
	default:
		fmt.Println("Couldn't match any HTTP Verbs and so, returning default response!")
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusMethodNotAllowed,
			Headers:    getCommonHeaders(),
			Body:       `{"error":"Method not allowed"}`,
		}, nil
	}
}

func addRoutineHandler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var routine Routine
	err := json.Unmarshal([]byte(request.Body), &routine)
	if err != nil {
		fmt.Println("Error unmarshalling request:", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Headers:    getCommonHeaders(),
			Body:       `{"error":"Invalid request"}`,
		}, nil
	}

	av, err := dynamodbattribute.MarshalMap(routine)
	if err != nil {
		fmt.Println("Error marshalling routine:", err)
		return events.APIGatewayProxyResponse{
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
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Headers:    getCommonHeaders(),
			Body:       `{"error":"failed to add routine"}`,
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusCreated,
		Headers:    getCommonHeaders(),
		Body:       `{"message":"Routine added successfully"}`,
	}, nil
}

func getRoutinesHandler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	fmt.Println("getRoutinesHandler() invoked..")
	date, exists := request.QueryStringParameters["date"]
	if !exists || date == "" {
		return events.APIGatewayProxyResponse{
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
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       `{"error":"Failed to fetch routines"}`,
		}, nil
	}
	routines := []Routine{}
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &routines)
	if err != nil {
		fmt.Printf("Error Unmarshalling DB items for date %s: %v\n", date, err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       `{"error":"Failed to process routines"}`,
		}, nil
	}
	responseBody, err := json.Marshal(routines)
	if err != nil {
		fmt.Println("Error marshalling response", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       `{"error":"Failed to prepare response"}`,
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers:    getCommonHeaders(),
		Body:       string(responseBody),
	}, nil

}

func handleOptionsRequest() (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers:    getCommonHeaders(),
		Body:       "",
	}, nil
}
func main() {
	lambda.Start(HandleRequest)
}
