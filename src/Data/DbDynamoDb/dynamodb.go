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

type dynamoDB struct {
	logger   *log.Logger
	config   DbConfig.DatabaseConfig
	sess     *session.Session
	database *dynamodb.DynamoDB
}

func (db *dynamoDB) Name() string {
	return Name
}

func ParseQueryOutput(output *dynamodb.QueryOutput) ([]map[string]interface{}, error) {
	result := make([]map[string]interface{}, 0, len(output.Items))
	for idx, value := range output.Items {
		item := make(map[string]interface{})
		err := dynamodbattribute.UnmarshalMap(value, &item)
		if err != nil {
			err = fmt.Errorf("failed to unmarshal item: %d, Error:%s", idx, err)
			return nil, err
		}
		result = append(result, item)
	}
	return result, nil
}

func (db *dynamoDB) Query(table string, index string, queryArgs map[string]interface{}) ([]map[string]interface{}, error) {
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

func (db *dynamoDB) Get(queryArgs map[string]interface{}) ([]map[string]interface{}, error) {
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

func (db *dynamoDB) ListTable() ([]interface{}, error) {
	tableList := []interface{}{}
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
		for _, name := range result.TableNames {
			tableList = append(tableList, *name)
		}

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

func (db *dynamoDB) CreateTable(name string, meta map[string]interface{}) error {
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

func (db *dynamoDB) DeleteTable(name string) error {
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

func MarshalMapWithCustomEncoder(data interface{}) (map[string]*dynamodb.AttributeValue, error) {
	encoder := dynamodbattribute.NewEncoder(func(e *dynamodbattribute.Encoder) {
		e.EnableEmptyCollections = true
	})
	av, err := encoder.Encode(data)
	if err != nil || av == nil || av.M == nil {
		return map[string]*dynamodb.AttributeValue{}, err
	}
	return av.M, nil
}

func (db *dynamoDB) Create(table string, data interface{}) error {
	return db.createRecord(table, data)
}

func (db *dynamoDB) createRecord(table string, data interface{}) error {
	av, err := MarshalMapWithCustomEncoder(data)
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

func (db *dynamoDB) Replace(table string, keys map[string]interface{}, data interface{}) error {
	queryArgs := map[string]interface{}{
		DbIface.Table: table,
	}
	for key, value := range keys {
		queryArgs[key] = value
	}
	currentList, err := db.Get(queryArgs)
	if err != nil {
		return fmt.Errorf("failed to query record exists.Error:%s", err)
	}
	if len(currentList) > 0 {
		err = db.deleteRecord(table, keys)
		if err != nil {
			return err
		}
	}
	return db.createRecord(table, data)
}

func (db *dynamoDB) Delete(table string, keys map[string]interface{}) error {
	return db.deleteRecord(table, keys)
}

func (db *dynamoDB) deleteRecord(table string, keys map[string]interface{}) error {
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

func (db *dynamoDB) Update(table string, keys map[string]interface{}, data interface{}) (map[string]interface{}, error) {
	dataType, ok := keys[Record.DataType].(string)
	if !ok {
		return nil, fmt.Errorf("missing key=[%s]", Record.DataType)
	}
	dataId, ok := keys[Record.DataId].(string)
	if !ok {
		return nil, fmt.Errorf("missing key=[%s]", Record.DataId)
	}
	if dataType == "" || dataId == "" {
		return nil, fmt.Errorf("missing dataType/dataId, expect format:[{dataType}/{dataId}/{dataPath}]")
	}
	queryPath, ok := keys[DbIface.PatchPath].(string)
	if !ok {
		return nil, fmt.Errorf("missing patch key=[%s]", DbIface.PatchPath)
	}
	patchData, err := db.Get(map[string]interface{}{
		DbIface.Table:   table,
		Record.DataType: dataType,
		Record.DataId:   dataId,
	})
	if err != nil {
		return nil, err
	}
	if len(patchData) == 0 {
		return nil, fmt.Errorf("data [%s/%s] does not exists", dataType, dataId)
	}
	subData, attrPath, err := DbIface.GetDataOnPath(patchData[0], queryPath, fmt.Sprintf("%s/%s/%s", dataType, dataId, queryPath))
	if err != nil {
		return nil, err
	}
	err = DbIface.SetPatchData(subData, attrPath, data)
	if err != nil {
		return nil, err
	}
	err = db.Create(table, patchData[0])
	if err != nil {
		return nil, err
	}
	return patchData[0], nil
}

func Connect(config DbConfig.DatabaseConfig, logger *log.Logger) (DbIface.Database, error) {
	if logger == nil {
		logger = log.Default()
	}
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
	database := dynamoDB{
		logger:   logger,
		config:   config,
		sess:     dbSession,
		database: dbSvc,
	}
	return &database, nil
}
