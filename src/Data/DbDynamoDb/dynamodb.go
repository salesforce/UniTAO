/*
************************************************************************************************************
Copyright (c) 2022 Salesforce, Inc.
All rights reserved.

UniTAO was originally created in 2022 by Shai Herzog & Yi Huo as an
Universal No-Coding Heterogeneous Infrastructure Maintenance & Inventory system that is holistically driven by open/community-developed semantic models/schemas.

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published
by the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>

This copyright notice and license applies to all files in this directory or sub-directories, except when stated otherwise explicitly.
************************************************************************************************************
*/

package DbDynamoDb

import (
	"encoding/json"
	"fmt"
	"log"

	"Data/DbConfig"
	"Data/DbIface"

	"github.com/salesforce/UniTAO/lib/Schema/Record"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
)

const (
	Name      = "dynamodb"
	UpdateOps = "updateops"
	TableName = "TableName"
)

type Database struct {
	config   DbConfig.DatabaseConfig
	sess     *session.Session
	database *dynamodb.DynamoDB
}

func ParseQueryOutput(output *dynamodb.QueryOutput) ([]map[string]interface{}, error) {
	result := make([]map[string]interface{}, len(output.Items))
	for idx, value := range output.Items {
		item := make(map[string]interface{})
		err := dynamodbattribute.UnmarshalMap(value, &item)
		if err != nil {
			err = fmt.Errorf("failed to unmarshal item: %d, Error:%s", idx, err)
			return nil, err
		}
		result[idx] = item
	}
	return result, nil
}

func (db *Database) Query(table string, index string, queryArgs map[string]interface{}) ([]map[string]interface{}, error) {
	init := false
	var cond expression.KeyConditionBuilder
	for key, value := range queryArgs {
		if !init {
			cond = expression.Key(key).Equal(expression.Value(value))
			init = true
			continue
		}
		cond = cond.And(expression.Key(key).Equal(expression.Value(value)))
	}
	expr, err := expression.NewBuilder().WithKeyCondition(cond).Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build index query. Error:%s", err)
	}
	queryInput := &dynamodb.QueryInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String(table),
	}
	if index != "" {
		queryInput.IndexName = aws.String(index)
	}
	output, err := db.database.Query(queryInput)
	if err != nil {
		return nil, fmt.Errorf("failed to query version for [table]=[%s]. Error: %s", table, err)
	}
	return ParseQueryOutput(output)
}

func (db *Database) Get(queryArgs map[string]interface{}) ([]map[string]interface{}, error) {
	tableName, ok := queryArgs[DbIface.Table].(string)
	if !ok {
		return nil, fmt.Errorf("missing parameter [%s] from queryArgs", DbIface.Table)
	}
	dataType, ok := queryArgs[Record.DataType].(string)
	if !ok {
		return nil, fmt.Errorf("missing parameter [%s] from queryArgs", Record.DataType)
	}
	args := make(map[string]interface{})
	args[Record.DataType] = dataType
	dataId, ok := queryArgs[Record.DataId].(string)
	if ok {
		args[Record.DataId] = dataId
	}
	return db.Query(tableName, "", args)
}

func (db *Database) ListTable() ([]*string, error) {
	tableList := []*string{}
	input := &dynamodb.ListTablesInput{}
	for {
		// Get the list of tables
		result, err := db.database.ListTables(input)
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				switch aerr.Code() {
				case dynamodb.ErrCodeInternalServerError:
					fmt.Println(dynamodb.ErrCodeInternalServerError, aerr.Error())
				default:
					fmt.Println(aerr.Error())
				}
			} else {
				// Print the error, cast err to awserr.Error to get the Code and
				// Message from an error.
				fmt.Println(err.Error())
			}
			return nil, err
		}

		tableList = append(tableList, result.TableNames...)

		// assign the last read tablename as the start for our next call to the ListTables function
		// the maximum number of table names returned in a call is 100 (default), which requires us to make
		// multiple calls to the ListTables function to retrieve all table names
		input.ExclusiveStartTableName = result.LastEvaluatedTableName

		if result.LastEvaluatedTableName == nil {
			break
		}
	}
	return tableList, nil
}

func (db *Database) CreateTable(name string, meta map[string]interface{}) error {
	log.Printf("create table %s in dynamodb", name)
	meta[TableName] = name
	rawJson, _ := json.Marshal(meta)
	input := &dynamodb.CreateTableInput{}
	json.Unmarshal([]byte(rawJson), &input)
	_, err := db.database.CreateTable(input)
	if err != nil {
		log.Printf("Got error calling CreateTable: %s", err)
		return err
	}
	return nil
}

func (db *Database) DeleteTable(name string) error {
	params := &dynamodb.DeleteTableInput{
		TableName: aws.String(name),
	}
	_, err := db.database.DeleteTable(params)
	if err != nil {
		log.Printf("Got error calling DeleteTable: %s", err)
		return err
	}
	return nil
}

func (db *Database) Create(table string, data interface{}) error {
	av, err := dynamodbattribute.MarshalMap(data)
	if err != nil {
		log.Printf("Got error marshalling map: %s", err)
		return err
	}
	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(table),
	}
	_, err = db.database.PutItem(input)
	if err != nil {
		log.Printf("Got error calling PutItem: %s", err)
		return err
	}
	return nil
}

func (db *Database) Update(table string, keys map[string]interface{}, cmd DbIface.UpdateCommand) (map[string]interface{}, error) {
	return nil, nil
}

func (db *Database) Replace(table string, keys map[string]interface{}, data interface{}) error {
	return db.Create(table, data)
}

func (db *Database) Delete(table string, keys map[string]interface{}) error {
	av, err := dynamodbattribute.MarshalMap(keys)
	if err != nil {
		log.Printf("Got error marshalling map: %s", err)
		return err
	}
	input := &dynamodb.DeleteItemInput{
		Key:       av,
		TableName: aws.String(table),
	}
	_, err = db.database.DeleteItem(input)
	if err != nil {
		log.Printf("Got error calling DeleteItem: %s", err)
		return err
	}
	return nil
}

func Connect(config DbConfig.DatabaseConfig) (DbIface.Database, error) {
	if config.Dynamodb.Region == "" {
		err := fmt.Errorf("missing configuration region")
		return nil, err
	}
	if config.Dynamodb.EndPoint == "" {
		err := fmt.Errorf("missing configuration endpoint")
		return nil, err
	}
	if config.Dynamodb.AccessKey == "" {
		config.Dynamodb.AccessKey = "dummyAccessKey"
	}
	if config.Dynamodb.SecretKey == "" {
		config.Dynamodb.SecretKey = "dummySecret"
	}
	if config.Dynamodb.AccessToken == "" {
		config.Dynamodb.AccessToken = "dummyToken"
	}

	dbSession, err := session.NewSession(&aws.Config{
		Region:      aws.String(config.Dynamodb.Region),
		Endpoint:    aws.String(config.Dynamodb.EndPoint),
		Credentials: credentials.NewStaticCredentials(config.Dynamodb.AccessKey, config.Dynamodb.SecretKey, config.Dynamodb.AccessToken),
	})
	if err != nil {
		newErr := fmt.Errorf("failed to create AWS session, region:%s, endpoint:%s, error:%s", config.Dynamodb.Region, config.Dynamodb.EndPoint, err.Error())
		return nil, newErr
	}
	dbSvc := dynamodb.New(dbSession)
	input := &dynamodb.ListTablesInput{}
	_, dbErr := dbSvc.ListTables(input)
	if dbErr != nil {
		panic("failed to list tables")
	}
	database := Database{
		config:   config,
		sess:     dbSession,
		database: dbSvc,
	}
	return &database, nil
}
