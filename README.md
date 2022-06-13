# UniTao

## Description

UniTAO was originally created in 2022 by Shai Herzog & Yi Huo as an 
Universal No-Coding Heterogeneous Infrastructure Maintenance & Inventory system that is holistically driven by open/community-developed semantic models/schemas.

It is designed to be self-contained and self-documenting to serve as foundation for human-free automation of heterogeneous computing infrastructure.

The model/schema uses the standard JSON schema format as basis, with additional enhancements. Those can be found in the documentation. 

## Components
 
### Data Layer
 - Supported Databases: DynamoDb, MongoDb
 - plugin-able data layer to support multiple types of database
 - Language: GoLang
 - sub folder: ./data


#### REST API Service
 - Language: GOLANG
 - Workspace: all project folders are included in go.work
 - install golang on MacOS
 ```
 brew install go
 ```
 - follow src/README.md to init and run the Inventory Service