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
	"UniTao/DataService/lib/Config"
	"UniTao/DataService/lib/DataHandler"
	"fmt"
	"github.com/salesforce/UniTAO/lib/Schema"
	"github.com/salesforce/UniTAO/lib/Util"
	"log"
	"path/filepath"
	"testing"
)

func TestDataHander(t *testing.T) {
	log.Print("test start")
	config, err := loadConfig()
	if err != nil {
		t.Fatalf("failed to read config file. Error:%s", err)
	}
	log.Print("config loaded")
	handler, err := DataHandler.New(*config)
	if err != nil {
		t.Fatalf("failed to create DataHandler from config, Error:%s", err)
	}
	filePath := "data/region.json"
	testData, err := Util.LoadJSONMap(filePath)
	if err != nil {
		t.Fatalf("failed loading data from [path]=[%s], Err:%s", filePath, err)
	}
	log.Print("data loaded")
	data, ok := testData[Schema.RecordData].(map[string]interface{})
	if !ok {
		t.Fatalf("missing field [%s] from test data", Schema.RecordData)
	}
	log.Print("get data for test")
	dataType := data[Schema.DataType].(string)
	dataId := data[Schema.DataId].(string)
	log.Printf("[type]=[%s],[id]=[%s]", dataType, dataId)
	_, err = handler.Validate(dataType, dataId, data)
	if err != nil {
		t.Fatalf("failed to validate positive data.")
	}
	log.Printf("positive data validate passed")
	negativeData, ok := testData["negativeData"].(map[string]interface{})
	if !ok {
		t.Fatalf("missing field [%s] from test data", Schema.RecordData)
	}
	log.Print("get data for test")
	dataType = negativeData[Schema.DataType].(string)
	dataId = negativeData[Schema.DataId].(string)
	log.Printf("[type]=[%s],[id]=[%s]", dataType, dataId)
	_, err = handler.Validate(dataType, dataId, negativeData)
	if err == nil {
		t.Fatalf("failed to validate negative data.")
	}
	log.Printf("negative data validate passed")

}

func loadConfig() (*Config.Confuguration, error) {
	config := Config.Confuguration{}
	configPathStr := "../../data/dynamoDbLocal/db-ds-01/config.json"
	configPath, err := filepath.Abs(configPathStr)
	if err != nil {
		return nil, fmt.Errorf("failed to get ABS Path [%s], Error:%s", configPathStr, err)
	}
	err = Config.Read(configPath, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to read Data Service config file, [%s], Error:%s", configPath, err)
	}
	return &config, nil
}
