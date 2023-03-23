# UniTao

## Main Repo Moved To
https://github.com/TuringCompute/UniTAO

## NOTE from Author (Yi Huo):

Apparently, The project lead Shai Herzog was laid off by Salesforce. and this opensource project is terminated by salesforce.
all further development will be fully based on my own requirement.

**Thus, I decided to fork this project to my own Organization (TuringCompute) for further development**

https://github.com/TuringCompute

**The new repo will be a fork to this repo.**

https://github.com/TuringCompute/UniTAO

**But, all future development will happens to that repo first. I will sync all changes back from the fork as long as I have the permission.**

## Description

UniTAO was originally created in 2022 by Shai Herzog & Yi Huo as an 
Universal No-Coding Heterogeneous Infrastructure Maintenance & Inventory system that is holistically driven by open/community-developed semantic models/schemas.

It is designed to be self-contained and self-documenting to serve as foundation for human-free automation of heterogeneous computing infrastructure.

The model/schema uses the standard JSON schema format as basis, with additional enhancements. Those can be found in the documentation. 

**schema of schema is defined in path:**
```
lib/Schema/data/schema.json
```
### Demo

#### **Docker**
We can build DataService and Inventory Service into docker container images. so we can run these in a docker environment for demo.

to build all images, run one of the following command based on your target environment.

```
# Powershell script for Windows environment
./docker/buildAll.ps1

# Bash script for Mac or Linux
./docker/buildAll.sh
```

#### **Environment**
after all docker images build successfully. we can bring a set of docker instances up as demo environment.
 - Data Service 01
 - Data Service 01 Admin
 - Data Service 02
 - Data Service 02 Admin
 - Inventory Service
 - Inventory Service Admin
 - Database (DynamoDB / MongoDB)
 - Database Admin (UI interface)

pick a folder and run the command 
**docker compose up -d**
```
# demo environment with DynamoDB
cd ./docker-compose/2data1inv

# demo environment with MongoDB
cd ./docker-compose/2data1invMongo

```

#### **Run**
to run demo:
 - bring up demo environment. wait all instance running stable after multiple restart for data setup
 - login to demo folder
 ```
 ./demo
 ```
 - choose a demo folder and run demo script depending on demo environment.
 ```
 # Windows environment:
 demo.ps1

 # Mac/Linux environment
 demo.sh
 ```


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

## JSON Schema Extensions
the JSON schema of schema to define data format is the key feature of this Data Service.
it give enough flexibility to define JSON data in order to support any coding requirement.

in order to achieve certain feature of Data Service, we have extended JSON schema as following:

### **contentMediaType**
the original meaning of contentMediaType is define how to parse value of the attached attribute.

**Example:**

the following code means value of testAttr is a json string that we can use json library to parse it
```
{
    "name": "testAttr",
    "type": "string",
    'contentMediaType': 'json'
}

```

in our schema, we introduced

**contentMediaType = [inventory/{dataType}]**

The following attribute definition means, value of testId is the id of type **test** from **inventory** service
```
{
    "name": "testId",
    "type": "string",
    'contentMediaType': 'inventory/test'
}
```

#### **Reason：**
in JSON schema, there is already a key **$ref** that can reference remote schema.

but after close look, we found **$ref** is just a simple include method. it does not include the meaning of parsing and validation. 

so we decide **conentMediaType** are closer to what we really means here. 

that is:
```
we only specify how to find the data that referenced by the value, but we don't really tell you how to use it.
this will seperate the logic of parsing and using of the schema
```

### **indexTemplate**
this attribute is for DataService automation in order to auto fill in reference value in registry attribute of other record.

**Example**

what we want to achieve here is to automatically fill m01 into machines attribute from listA of type machineList when add record machine **id=m01**
```
{
    "machine": [
        {
            "id": "m01"
            "listId": "listA"
        }
    ],
    "machineList": [
        {
            "id": "listA",
            "machines": [
                "m01"
            ]
        }
    ]
}
```


