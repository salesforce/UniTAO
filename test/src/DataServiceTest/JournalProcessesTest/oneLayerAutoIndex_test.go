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

package JournalProcessTest

// test 1 layer of parent above leaf and direct registry path at root level of parent record

import (
	"DataService/DataHandler"
	"DataService/DataJournal"
	"DataService/DataJournal/ProcessIface"
	"UniTao/Test/DataServiceTest"
	"encoding/json"
	"log"
	"net/http"
	"testing"

	"github.com/salesforce/UniTAO/lib/Schema/CmtIndex"
	"github.com/salesforce/UniTAO/lib/Schema/JsonKey"
	"github.com/salesforce/UniTAO/lib/Schema/Record"
	"github.com/salesforce/UniTAO/lib/Schema/SchemaDoc"
)

var handler *DataHandler.Handler
var journal *DataJournal.JournalLib
var processList []ProcessIface.JournalProcess

func getHandlerAndJournal(t *testing.T) {
	hdl, err := DataServiceTest.MockHandler()
	if err != nil {
		t.Fatalf("failed to load MockHandler")
	}
	handler = hdl
	jnl, ex := DataJournal.NewJournalLib(handler.DB, handler.Config.DataTable.Data)
	if ex != nil {
		t.Fatalf("failed to create Journal Library. Error: %s", err)
	}
	journal = jnl
	handler.AddJournal = journal.AddJournal
}

func prepEnv(t *testing.T) {
	if handler != nil {
		log.Printf("environment is already created")
		return
	}
	getHandlerAndJournal(t)
	processes, err := DataJournal.LoadProcesses(handler, log.Default())
	if err != nil {
		t.Fatalf("failed to create schema process. Error: %s", err)
	}
	processList = processes
}

func addSchema(t *testing.T, schemaStr string) int {
	schemaRec, err := Record.LoadStr(schemaStr)
	if err != nil {
		t.Fatal(err)
	}
	if schemaRec.Type != JsonKey.Schema {
		t.Fatalf("[%s] is not [%s] to add", schemaRec.Type, JsonKey.Schema)
	}
	schemaDoc, err := SchemaDoc.New(schemaRec.Data)
	if err != nil {
		t.Fatalf("invalid schema data. Error:%s", err)
	}
	_, ex := handler.LocalSchema(schemaRec.Id, schemaDoc.Version)
	if ex == nil {
		log.Printf("schema [%s] already created", schemaRec.Id)
		return 0
	}
	if ex.Status != http.StatusNotFound {
		t.Fatal(ex)
		return 0
	}
	ex = handler.Add(schemaRec)
	if ex != nil {
		t.Fatalf("failed to add first schema record. Error: %s", ex)
	}
	return processJournal(t, JsonKey.Schema, schemaRec.Id)
}

func addData(t *testing.T, recordStr string) int {
	test01, err := Record.LoadStr(recordStr)
	if err != nil {
		t.Fatalf("failed to load test01 as record")
	}
	ex := handler.Add(test01)
	if ex != nil {
		t.Fatalf("failed to add record of test 0.0.1")
	}
	return processJournal(t, test01.Type, test01.Id)
}

func setData(t *testing.T, recordStr string) int {
	test01, err := Record.LoadStr(recordStr)
	if err != nil {
		t.Fatalf("failed to load test01 as record")
	}
	ex := handler.Set(test01)
	if ex != nil {
		t.Fatalf("failed to add record of test 0.0.1")
	}
	return processJournal(t, test01.Type, test01.Id)
}

func processJournal(t *testing.T, dataType string, dataId string) int {
	count := 0
	for {
		err := DataJournal.ProcessNextEntry(journal, processList, dataType, dataId, log.Default())
		if err != nil {
			if err.Status == http.StatusNotFound {
				log.Printf("run out of Journal Entry on %s/%s", dataType, dataId)
				break
			}
			t.
				Fatal(err)
		}
		count += 1
	}
	log.Printf("processed %d Entries", count)
	return count
}

func TestCmdIdxMap(t *testing.T) {
	idx := CmtIndex.CmtIndex{
		DataType: "test",
		Subscriber: map[string]CmtIndex.CmtSubscriber{
			"test": {
				DataType: "source",
				VersionIndex: map[string]CmtIndex.VersionIndex{
					"0.0.2": {
						Version: "0.0.2",
						IndexTemplate: []interface{}{
							"test",
						},
					},
				},
			},
		},
	}
	dataBytes, err := json.Marshal(idx)
	if err != nil {
		t.Fatalf(err.Error())
	}
	dataStr := string(dataBytes)
	log.Print(dataStr)
	target := map[string]interface{}{}
	err = json.Unmarshal([]byte(dataStr), &target)
	if err != nil {
		t.Fatalf(err.Error())
	}
	targetBytes, err := json.Marshal(target)
	if err != nil {
		t.Fatal(err)
	}
	targetStr := string(targetBytes)
	log.Print(targetStr)
	data := idx.Map()
	if data == nil {
		t.Fatalf("failed to convert idx to map")
	}

}

func addLeafSchemaAndData(t *testing.T) {
	firstSchemaStr := `{
		"__id": "test",
		"__type": "schema",
		"__ver": "0.0.1",
		"data": {
			"name": "test",
			"version": "0.0.1",
			"properties": {
				"attr1": {
					"type": "string"
				},
				"categoryLayer1": {
					"type": "string"
				}
			}
		}
	}`
	addSchema(t, firstSchemaStr)
	test01Str := `{
		"__id": "test01",
		"__type": "test",
		"__ver": "0.0.1",
		"data": {
			"attr1": "test",
			"categoryLayer1": "Layer1Cat01"
		}
	}`
	addData(t, test01Str)
}

func addLayer1SchemaDataAndUpgradSchema(t *testing.T) {
	catLayer1Schema := `{
		"__id": "categoryLayer1",
		"__type": "schema",
		"__ver": "0.0.1",
		"data": {
			"name": "categoryLayer1",
			"version": "0.0.1",
			"properties": {
				"testAry": {
					"type": "array",
					"items": {
						"type": "string",
						"contentMediaType": "inventory/test"
					}
				}
			}
		}
	}`
	addSchema(t, catLayer1Schema)
	catLayer1Data1 := `{
		"__id": "Layer1Cat01",
		"__type": "categoryLayer1",
		"__ver": "0.0.1",
		"data": {
			"testAry": []
		}
	}`
	addData(t, catLayer1Data1)
	_, err := handler.Get(CmtIndex.KeyCmtIdx, "test")
	if err == nil {
		t.Fatal("there should be no CmtIndex of Type=[test]")
	}
	catLayer1SchemaV2 := `{
		"__id": "categoryLayer1",
		"__type": "schema",
		"__ver": "0.0.1",
		"data": {
			"name": "categoryLayer1",
			"version": "0.0.2",
			"properties": {
				"test1": {
					"type": "string",
					"contentMediaType": "inventory/test"
				},
				"testAry": {
					"type": "array",
					"items": {
						"type": "string",
						"contentMediaType": "inventory/test",
						"indexTemplate": "{categoryLayer1}/testAry"
					}
				}
			}
		}
	}`
	addSchema(t, catLayer1SchemaV2)
	cmtIdxData, err := handler.Get(CmtIndex.KeyCmtIdx, "test")
	if err != nil {
		t.Fatal("failed to add CmtIndex of Type=[test]")
	}
	idxRecord, ex := Record.LoadMap(cmtIdxData)
	if ex != nil {
		t.Fatal(ex)
	}
	cmtIdx, ex := CmtIndex.LoadMap(idxRecord.Data)
	if ex != nil {
		t.Fatal(ex)
	}
	if cmtIdx.DataType != "test" {
		t.Fatalf("generated wrong cmtIndex, type [%s]!=[test]", cmtIdx.DataType)
	}

	catSub, ok := cmtIdx.Subscriber["categoryLayer1"]
	if !ok {
		t.Fatal("missing cmtIdx for type=[categoryLayer1]")
	}
	if catSub.DataType != "categoryLayer1" {
		t.Fatalf("invalid sub type [%s]!=[categoryLayer1]", catSub.DataType)
	}
	verPath, ok := catSub.VersionIndex["0.0.2"]
	if !ok {
		t.Fatal("missing cmtIdx for version [0.0.2]")
	}
	if verPath.Version != "0.0.2" {
		t.Fatalf("cmdIdx invalid version [%s]!=[0.0.2]", verPath.Version)
	}
	if len(verPath.IndexTemplate) != 1 {
		t.Fatalf("invalid index template number. [%d]!=[1]", len(verPath.IndexTemplate))
	}
	if verPath.IndexTemplate[0].(string) != "{categoryLayer1}/testAry" {
		t.Fatalf("cmtIdx path[%s]!=[{categoryLayer1}/testAry]", verPath.IndexTemplate[0].(string))
	}
}

func addLayer1DataAndLeafDataAndDelete(t *testing.T) {
	catLayer1Data2 := `{
		"__id": "Layer1Cat01",
		"__type": "categoryLayer1",
		"__ver": "0.0.2",
		"data": {
			"test1": "test01",
			"testAry": []
		}
	}`
	catLayer1, ex := Record.LoadStr(catLayer1Data2)
	if ex != nil {
		t.Fatal(ex)
	}
	err := handler.Set(catLayer1)
	if err != nil {
		t.Fatal(err)
	}
	processJournal(t, catLayer1.Type, catLayer1.Id)
	test02Str := `{
		"__id": "test02",
		"__type": "test",
		"__ver": "0.0.1",
		"data": {
			"attr1": "test",
			"categoryLayer1": "Layer1Cat01"
		}
	}`
	addData(t, test02Str)
	err = handler.Delete("test", "test02")
	if err != nil {
		t.Fatal(err)
	}
	processJournal(t, "test", "test02")
}

func TestAddSchemaWithoutCmtIndex(t *testing.T) {
	prepEnv(t)
	addLeafSchemaAndData(t)
}

func TestAddCmtSchema(t *testing.T) {
	prepEnv(t)
	addLeafSchemaAndData(t)
	addLayer1SchemaDataAndUpgradSchema(t)
	addLayer1DataAndLeafDataAndDelete(t)
	processJournal(t, "categoryLayer1", "Layer1Cat01")
}
