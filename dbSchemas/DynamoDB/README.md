## Data Schema for DynamoDB

#### load table data
Each **.json** file stored under DynamoDB, represent a table in Dynamo db
Run following command to load schema into dynamodb instance
```
# example, load table Model in local dynamoDB
# this command should be run at data folder
aws dynamodb create-table --endpoint-url http://localhost:8000 --cli-input-json file://dbSchemas/DynamoDB/Model.json
```
