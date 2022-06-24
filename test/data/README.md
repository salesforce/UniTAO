# Data Layer for uniTao

## Currently Support

#### DynamoDB
 - Test Method: Docker-Compose DynamoDB-local
 - Test Location: ./test/DynamoDB/
 - Ref URL: https://medium.com/platform-engineer/running-aws-dynamodb-local-with-docker-compose-6f75850aba1e
 - Instruction to run DynamoDB locally:
   - ./dynamoDbLocal/README.md
   - Commands:
  ```
  dynamoDb: ./dynamoDb/startDb.sh
  dynamoDB-Admin: ./dynamoDb/adminGUI.sh
  ```
- to Init table of database
  ```
  db-ds-01: ./DataServce01/initTable.sh    # init table: DataServce01
  db-ds-02: ./DataServce02/initTable.sh    # init table: DataServce02
  ```
- to run DataService
  ```
  db-ds-01: ./DataServce01/runServer.sh
  db-ds-02: ./DataServce02/runServer.sh
  ```
- to start Inventory Service
  ```
  ./InventoryService/runServer.sh
  ```
- to importTestData from Test Data Service 01 and 02
  - make sure both DataService 01 and 02 is running at 8004 and 8005
  ```
  ./InventoryService/importDataService.sh
  ```
- to import test data
  ```
  db-ds-01: ./DataServce01/importData.sh
  db-ds-02: ./DataServce02/importData.sh
  ```
- to stop and remove DynamoDb instance.
  ```
  db-ds-01: ./DataServce01/stopDb.sh
  db-ds-02: ./DataServce02/stopDb.sh
  ```
#### Test Environment Port Arrangement:
 - 8000:    DynamoDB Database
 - 8001:    DynamoDB Admin UI
 - 8002:    DataService01
 - 8003:    DataService02
 - 8004:    InventoryService