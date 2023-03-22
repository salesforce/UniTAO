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

package MongoDb

import (
	"Data/DbConfig"
	"Data/DbIface"
	"context"
	"fmt"
	"log"

	"github.com/salesforce/UniTAO/lib/Schema/Record"
	"github.com/salesforce/UniTAO/lib/Util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	Name        = "mongodb"
	UpdateOps   = "updateops"
	TableName   = "TableName"
	Indexes     = "Indexes"
	IndexFields = "IndexFields"
	IndexSort   = "IndexSort"
	UniqueIndex = "UniqueIndex"

	IdKey   = "_id_"
	IdxName = "name"
)

type mongoDb struct {
	logger *log.Logger
	config DbConfig.DatabaseConfig
	client *mongo.Client
}

func (db *mongoDb) Name() string {
	return Name
}

func (db *mongoDb) Database() (*mongo.Database, error) {
	dbList, err := db.client.ListDatabaseNames(context.TODO(), bson.D{})
	if err != nil {
		return nil, fmt.Errorf("failed to list database. Error: %s", err)
	}
	for _, dbName := range dbList {
		if dbName == db.config.Mongodb.Database {
			database := db.client.Database(db.config.Mongodb.Database)
			return database, nil
		}
	}
	return nil, nil
}

func (db *mongoDb) ListTable() ([]interface{}, error) {
	tableList := []interface{}{}
	database, err := db.Database()
	if err != nil {
		return nil, err
	}
	if database == nil {
		return tableList, nil
	}
	colList, err := database.ListCollectionNames(context.TODO(), bson.D{})
	if err != nil {
		return nil, fmt.Errorf("failed to list table. Error:%s", err)
	}
	for _, colName := range colList {
		tableList = append(tableList, colName)
	}
	return tableList, nil
}

func (db *mongoDb) CreateTable(tableName string, data map[string]interface{}) error {
	database := db.client.Database(db.config.Mongodb.Database)
	tableList, err := db.ListTable()
	if err != nil {
		return fmt.Errorf("failed to list table. Error:%s", err)
	}
	tableMap := Util.IdxList(tableList)
	if _, ok := tableMap[tableName]; !ok {
		err := database.CreateCollection(context.TODO(), tableName)
		if err != nil {
			return err
		}
	}
	table := database.Collection(tableName)
	idxCur, err := table.Indexes().List(context.TODO())
	if err != nil {
		return fmt.Errorf("failed to get indexes from table[%s], Error:%s", tableName, err)
	}
	currentIdxMap := map[string]int{}
	for idxCur.Next(context.TODO()) {
		var result map[string]interface{}
		err = idxCur.Decode(&result)
		if err != nil {
			return fmt.Errorf("failed while query table[%s] indexes, Errors: %s", tableName, err)
		}
		if result[IdxName] != IdKey {
			currentIdxMap[IdxName] = result["v"].(int)
		}
	}
	idxMap := data[Indexes].(map[string]interface{})
	for idxName, idxData := range idxMap {
		if _, ok := currentIdxMap[idxName]; ok {
			continue
		}
		index := idxData.(map[string]interface{})
		if _, ok := index[IndexFields]; !ok {
			return fmt.Errorf("invalid index[%s] for table[%s], missing [%s]", idxName, tableName, IndexFields)
		}
		idxFields := index[IndexFields].(map[string]interface{})
		if len(idxFields) == 0 {
			return fmt.Errorf("invalid index[%s] for table[%s], no field in [%s]", idxName, tableName, IndexFields)
		}
		var keyList bson.D
		for k, v := range idxFields {
			keyList = append(keyList, bson.E{Key: k, Value: v})
		}
		idxModel := mongo.IndexModel{
			Keys: keyList,
		}
		idxModel.Options = options.Index().SetName(idxName)
		if isUnique, ok := index[UniqueIndex]; ok {
			idxModel.Options = idxModel.Options.SetUnique(isUnique.(bool))
		}
		table.Indexes().CreateOne(context.TODO(), idxModel)
	}
	return nil
}

func (db *mongoDb) DeleteTable(name string) error {
	database := db.client.Database(db.config.Mongodb.Database)
	tableList, err := db.ListTable()
	if err != nil {
		return err
	}
	for _, table := range tableList {
		if table == name {
			collecton := database.Collection(table.(string))
			collecton.Drop(context.TODO())
			break
		}
	}
	return nil
}

func (db *mongoDb) Get(queryArgs map[string]interface{}) ([]map[string]interface{}, error) {
	database := db.client.Database(db.config.Mongodb.Database)
	tableName, ok := queryArgs[DbIface.Table].(string)
	if !ok {
		return nil, fmt.Errorf("missing parameter [%s] from queryArgs", DbIface.Table)
	}
	table := database.Collection(tableName)
	if table == nil {
		return nil, fmt.Errorf("table [%s] does not exists", tableName)
	}
	dataType, ok := queryArgs[Record.DataType].(string)
	if !ok {
		return nil, fmt.Errorf("missing parameter [%s] from queryArgs", Record.DataType)
	}

	filter := bson.M{
		Record.DataType: bson.M{
			"$eq": dataType,
		},
	}
	dataId, ok := queryArgs[Record.DataId].(string)
	if ok {
		filter = bson.M{
			"$and": bson.A{
				filter,
				bson.M{
					Record.DataId: dataId,
				},
			},
		}
	}
	cursor, err := table.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}
	result := []map[string]interface{}{}
	err = cursor.All(context.TODO(), &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (db *mongoDb) Create(tableName string, data interface{}) error {
	database := db.client.Database(db.config.Mongodb.Database)
	table := database.Collection(tableName)
	if table == nil {
		return fmt.Errorf("table [%s] does not exists", tableName)
	}
	_, err := table.InsertOne(context.TODO(), data)
	if err != nil {
		return err
	}
	return nil
}
func (db *mongoDb) Update(table string, keys map[string]interface{}, data interface{}) (map[string]interface{}, error) {
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
	return nil, nil
}

func (db *mongoDb) Replace(tableName string, keys map[string]interface{}, data interface{}) error {
	database := db.client.Database(db.config.Mongodb.Database)
	table := database.Collection(tableName)
	if table == nil {
		return fmt.Errorf("table [%s] does not exists", tableName)
	}
	opts := options.Replace().SetUpsert(true)
	table.ReplaceOne(context.TODO(), keys, data, opts)
	return nil
}

func (db *mongoDb) Delete(tableName string, keys map[string]interface{}) error {
	database := db.client.Database(db.config.Mongodb.Database)
	table := database.Collection(tableName)
	if table == nil {
		return fmt.Errorf("table [%s] does not exists", tableName)
	}
	table.DeleteOne(context.TODO(), keys)
	return nil
}

func Connect(config DbConfig.DatabaseConfig, logger *log.Logger) (DbIface.Database, error) {
	if logger == nil {
		logger = log.Default()
	}
	credential := options.Credential{
		AuthMechanism: "SCRAM-SHA-1",
		Username:      config.Mongodb.UserName,
		Password:      config.Mongodb.Password,
	}
	clientOpts := options.Client().ApplyURI(config.Mongodb.EndPoint).SetAuth(credential)
	client, err := mongo.Connect(context.TODO(), clientOpts)
	if err != nil {
		return nil, err
	}
	database := mongoDb{
		logger: logger,
		config: config,
		client: client,
	}
	return &database, nil
}
