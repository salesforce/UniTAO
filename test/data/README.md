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
  db-ds-01: ./dynamoDbLocal/db-ds-01/startDb.sh
  Admin-01: ./dynamoDbLocal/db-ds-01/adminGUI.sh
  db-ds-02: ./dynamoDbLocal/db-ds-02/startDb.sh
  Admin-02: ./dynamoDbLocal/db-ds-02/adminGUI.sh
  ```
- to Init table of database
  ```
  db-ds-01: ./dynamoDbLocal/db-ds-01/initTable.sh
  db-ds-02: ./dynamoDbLocal/db-ds-02/initTable.sh
  ```
- to run DataService
  ```
  db-ds-01: ./dynamoDbLocal/db-ds-01/runServer.sh
  db-ds-02: ./dynamoDbLocal/db-ds-02/runServer.sh
  ```
- to start Inventory Service
  ```
  ./SysDirFile/runServer.sh
  ```
- to importTestData from Test Data Service 01 and 02
  - make sure both DataService 01 and 02 is running at 8004 and 8005
  ```
  ./SysDirFile/importDataService.sh
  ```
- to import test data
  ```
  db-ds-01: ./dynamoDbLocal/db-ds-01/importData.sh
  db-ds-02: ./dynamoDbLocal/db-ds-02/importData.sh
  ```
- to stop and remove DynamoDb instance.
  ```
  db-ds-01: ./dynamoDbLocal/db-ds-01/stopDb.sh
  db-ds-02: ./dynamoDbLocal/db-ds-02/stopDb.sh
  ```
#### Test Environment Port Arrangement:
 - 8000:    DB-DS-01
 - 8001:    DB-DS-02
 - 8002:    DB-DS-01-GUI-Admin
 - 8003:    DB-DS-02-GUI-Admin
 - 8004:    DataService01
 - 8005:    DataService02
 - 8006:    InventoryService