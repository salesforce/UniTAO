# UniTao

## Description

UniTAO was originally created in 2022 by Shai Herzog & Yi Huo as an 
Universal No-Coding Heterogeneous Infrastructure Maintenance & Inventory system that is holistically driven by open/community-developed semantic models/schemas.

It is designed to be self-contained and self-documenting to serve as foundation for human-free automation of heterogeneous computing infrastructure.

The model/schema uses the standard JSON schema format as basis, with additional enhancements. Those can be found in the documentation. 

**schema of schema is defined in path:**
```
lib/Schema/data/schema.json
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


