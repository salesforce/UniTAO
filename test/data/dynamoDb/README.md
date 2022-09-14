# Run DynamoDB-LOCAL to provide Test Environment

## Install docker and docker-compose
MacOS:
```
brew install --cask docker
brew install docker-compose
```

## Docker Compose commands:
 ```
 docker-compose up -d dynamodb
 docker logs -f my-dynamodb
 ```
## Connect to the Database:
#### DynamoDB Admin
**URL:** https://medium.com/swlh/a-gui-for-local-dynamodb-dynamodb-admin-b16998323f8e
```
npm install -g dynamodb-admin
```
MacOS/Linux
```
dynamodb-admin
```
Windows
```
export DYNAMO_ENDPOINT=http://localhost:8000 dynamodb-admin -p 8002
```

#### AWS CLI:
 ```
 export AWS_ACCESS_KEY_ID=123
 export AWS_SECRET_ACCESS_KEY=456
 export AWS_DEFAULT_REGION=us-west-2

 aws dynamodb <command> --endpoint-url http://localhost:8000
 e.g.
 aws dynamodb list-tables --endpoint-url http://localhost:8000
 ```
#### AWS-SDK Way:
 ```
 2.2.1 — Node.js
 import { DynamoDB, DocumentClient } from "aws-sdk"
 // via DynamoDB
 const dbClient = new DynamoDB({
   endpoint: "http://localhost:8000",
   .....
 })
 // via DocumentClient
 const docClient = new DocumentClient({
   endpoint: "http://localhost:8000",
   .....
 })
 2.2.2 — Golang
 import (
     "github.com/aws/aws-sdk-go/aws"
     "github.com/aws/aws-sdk-go/aws/session"
     "github.com/aws/aws-sdk-go/service/dynamodb"
 )
 sess, err := session.NewSession(&aws.Config{     
     Endpoint: aws.String("http://localhost:8000")}, 
 )
 if err != nil {     
     // Handle Session creation error 
 }
 // Create DynamoDB client 
 svc := dynamodb.New(sess)
 2.2.3 - Java
 import software.amazon.awssdk.regions.Region;
 import software.amazon.awssdk.services.dynamodb.DynamoDbClient;
 DynamoDbClient client = DynamoDbClient.builder()
    .region(Region.US_WEST_2) 
    .endpointOverride(
         URI.create("http://localhost:8000")
    ).build();
 2.2.4 - Python (Boto3)
 import boto3
 # Get the service client
 db = boto3.client('dynamodb',endpoint_url='http://localhost:8000')
 # Get the service resource
 db = boto3.resource('dynamodb',endpoint_url='http://localhost:8000')
 ```
