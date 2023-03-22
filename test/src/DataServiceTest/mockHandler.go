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

package DataServiceTest

import (
	"Data/DbConfig"
	"Data/DbIface"
	"DataService/Config"
	"DataService/DataHandler"
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"

	"github.com/salesforce/UniTAO/lib/Schema/Record"
	"github.com/salesforce/UniTAO/lib/Util"
	"github.com/salesforce/UniTAO/lib/Util/Json"
)

func GetSchemaOfSchema() (string, error) {
	rootDir, err := Util.RootDir()
	if err != nil {
		return "", fmt.Errorf("failed to get running dir")
	}
	schemaFile, err := filepath.Abs(filepath.Join(rootDir, "lib/Schema/data/schema.json"))
	if err != nil {
		return "", fmt.Errorf("failed to get ABS path of schema.json")
	}
	schemaData, err := Json.LoadJsonFile(schemaFile)
	if err != nil {
		return "", err
	}
	DataCache := map[string]interface{}{}
	schemaList := schemaData.(map[string]interface{})["data"].([]interface{})
	for idx, recObj := range schemaList {
		record, err := Record.LoadMap(recObj.(map[string]interface{}))
		if err != nil {
			return "", fmt.Errorf("failed to load schema record @[%d]", idx)
		}
		if _, ok := DataCache[record.Type]; !ok {
			DataCache[record.Type] = map[string]interface{}{}
		}
		DataCache[record.Type].(map[string]interface{})[record.Id] = record.Map()
	}
	dataStr, err := json.MarshalIndent(DataCache, "", "    ")
	if err != nil {
		return "", err
	}
	return string(dataStr), nil
}

func MockHandler() (*DataHandler.Handler, error) {
	configStr := `
	{
		"database": {
			"type": "dynamodb",
			"dynamodb": {
				"region": "us-west-2",
				"endpoint": "http://localhost:8000"
			}
		},
		"table": {
			"data": "DataService01"
		},
		"http": {
			"type": "http",
			"dns": "localhost",
			"port": "8002",
			"id": "DataService01"
		},
		"inventory": {
			"url": "http://localhost:8004"
		}
	}
	`
	config := Config.Confuguration{}
	err := json.Unmarshal([]byte(configStr), &config)
	if err != nil {
		return nil, fmt.Errorf("faild to load config str. invalid format. Error:%s", err)
	}
	log.Print("config loaded")
	dbStr, err := GetSchemaOfSchema()
	if err != nil {
		return nil, err
	}
	connectDb := func(config DbConfig.DatabaseConfig, logger *log.Logger) (DbIface.Database, error) {
		mockDb, err := NewMockDb(config, dbStr, logger)
		if err != nil {
			return nil, err
		}
		return mockDb, nil
	}
	handler, ex := DataHandler.New(config, nil, connectDb)
	if ex != nil {
		return nil, fmt.Errorf("failed to create handler")
	}
	return handler, nil
}
