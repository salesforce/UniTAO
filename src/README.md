# UniTao Inventory Service

Provide Inventory Service and Data Service for multi-nodes cross reference inventory query system

#### Language
GO-LANG

#### Data Service
provide CRUD for all local Inventory

#### Inventory Service
- Able to list all inventory registry records for all related inventory services
- [TODO] Able to perform cross reference query to get data across Inventory Services

#### prepare to run the service
- Install docker and docker-compose and run dynamodb-local and listen at port http://localhost:8000
- Install dynamodb-admin according to the test/data/DynamoDB/README.md
- Run init server command to initializ the database.
```
go run src/Server/server.go init -config test/data/DynamoDB/config.json -db-data dbSchemas/DynamoDB/Data.json -force true
```
- Run dynamodb-admin and check in the Web UI to see if table and data are being populated

#### Run Inventory Service
```
go run src/Server/server.go run -config test/data/DynamoDB/config.json
```
if no error happens in console, the service should be listen at http://localhost:8002
use post man or web browser to check following url paths to see if REST API works
```
http://localhost:8002
http://localhost:8002/data
http://localhost:8002/data/schema
http://localhost:8002/data/schema/schema
http://localhost:8002/data/schema/inventory
http://localhost:8002/inventory
http://localhost:8002/inventory/self
```


