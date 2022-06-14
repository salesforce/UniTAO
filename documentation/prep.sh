#! /bin/sh
#
# This is to prep the tools after reboot
#
# Start tje dynamoDB admin UI (http://localhost:8001)
# dynamodb-admin &
#
# run server and initialize the DB 
go run src/Server/server.go init -config test/data/DynamoDB/config.json -db-data dbSchemas/DynamoDB/Data.json -force true
#
go run src/Server/server.go run -config test/data/DynamoDB/config.json

