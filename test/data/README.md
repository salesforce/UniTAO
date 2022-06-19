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
  db-ds-02: ./dynamoDbLocal/db-ds-02/startDb.sh
  ```
- to stop and remove DynamoDb instance.
   - cd to database instance folder. [db-ds-01, db-ds-02]
   - shutdown command: docker-compose down


