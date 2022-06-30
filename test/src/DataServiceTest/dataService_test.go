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
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"testing"

	"DataService/Config"
	"DataService/DataHandler"
	"DataService/DataServer"

	"github.com/salesforce/UniTAO/lib/Schema/Record"
	"github.com/salesforce/UniTAO/lib/Util"
)

// Make sure both data service and inventory Service are running before running the test
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
	log.Print("get positive data for test")
	record, err := Record.LoadMap(testData["data"].(map[string]interface{}))
	if err != nil {
		t.Fatalf("failed to load data as record. Error:%s", err)
	}
	log.Printf("[type]=[%s],[id]=[%s]", record.Type, record.Id)
	_, err = handler.Validate(record)
	if err != nil {
		t.Fatalf("failed to validate positive data. Error:%s", err)
	}
	log.Printf("positive data validate passed")
	log.Print("get negative Data for test")
	record, err = Record.LoadMap(testData["negativeData"].(map[string]interface{}))
	if err != nil {
		t.Fatalf("failed to load negativeData as record. Error:%s", err)
	}
	log.Printf("[type]=[%s],[id]=[%s]", record.Type, record.Id)
	_, err = handler.Validate(record)
	if err == nil {
		t.Fatalf("failed to validate negative data.")
	}
	log.Printf("negative data validate passed")
}

func TestParseRecord(t *testing.T) {
	server := DataServer.Server{}
	payload := make(map[string]interface{})
	dataType := "test"
	typeVer := "00_00_00"
	dataId := "test_01"
	record := Record.NewRecord(dataType, typeVer, dataId, payload)
	_, code, _ := server.ParseRecord([]string{}, record.Map(), dataType, dataId)
	if code != http.StatusAccepted {
		t.Fatalf("failed to parse record. type=[%s], id=[%s]", dataType, dataId)
	}
	pRecord, _, err := server.ParseRecord([]string{"true"}, record.Data, dataType, dataId)
	if err != nil {
		t.Fatalf("failed to parse record with no Reacod header. Error:%s", err)
	}
	if pRecord.Version != "0_00_00" {
		t.Fatalf("failed to create record with correct version")
	}
}

func loadConfig() (*Config.Confuguration, error) {
	config := Config.Confuguration{}
	configPathStr := "../../data/DataService01/config.json"
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
