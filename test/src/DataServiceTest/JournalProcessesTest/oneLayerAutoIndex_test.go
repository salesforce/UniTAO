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

type TestEnv struct {
	T         *testing.T
	Handler   *DataHandler.Handler
	Journal   *DataJournal.JournalLib
	Processes []ProcessIface.JournalProcess
}

func getHandlerAndJournal(t *testing.T) (*DataHandler.Handler, *DataJournal.JournalLib) {
	handler, err := DataServiceTest.MockHandler()
	if err != nil {
		t.Fatalf("failed to load MockHandler")
	}
	journal, ex := DataJournal.NewJournalLib(handler.DB, handler.Config.DataTable.Data, nil)
	if ex != nil {
		t.Fatalf("failed to create Journal Library. Error: %s", err)
	}
	handler.AddJournal = journal.AddJournal
	return handler, journal
}

func prepEnv(t *testing.T) *TestEnv {
	handler, journal := getHandlerAndJournal(t)
	processes, err := DataJournal.LoadProcesses(handler, log.Default())
	if err != nil {
		t.Fatalf("failed to create schema process. Error: %s", err)
	}
	env := TestEnv{
		T:         t,
		Handler:   handler,
		Journal:   journal,
		Processes: processes,
	}
	return &env
}

func addSchema(t *TestEnv, schemaStr string) int {
	schemaRec, err := Record.LoadStr(schemaStr)
	if err != nil {
		t.T.Fatal(err)
	}
	if schemaRec.Type != JsonKey.Schema {
		t.T.Fatalf("[%s] is not [%s] to add", schemaRec.Type, JsonKey.Schema)
	}
	schemaDoc, err := SchemaDoc.New(schemaRec.Data)
	if err != nil {
		t.T.Fatalf("invalid schema data. Error:%s", err)
	}
	_, ex := t.Handler.LocalSchema(schemaRec.Id, schemaDoc.Version)
	if ex == nil {
		log.Printf("schema [%s] already created", schemaRec.Id)
		return 0
	}
	if ex.Status != http.StatusNotFound {
		t.T.Fatal(ex)
		return 0
	}
	ex = t.Handler.Add(schemaRec)
	if ex != nil {
		t.T.Fatalf("failed to add first schema record. Error: %s", ex)
	}
	return processJournal(t, JsonKey.Schema, schemaRec.Id)
}

func addData(t *TestEnv, recordStr string) int {
	test01, err := Record.LoadStr(recordStr)
	if err != nil {
		t.T.Fatalf("failed to load test01 as record")
	}
	ex := t.Handler.Add(test01)
	if ex != nil {
		t.T.Fatalf("failed to add record of test 0.0.1")
	}
	return processJournal(t, test01.Type, test01.Id)
}

func setData(t *TestEnv, recordStr string) int {
	test01, err := Record.LoadStr(recordStr)
	if err != nil {
		t.T.Fatalf("failed to load test01 as record")
	}
	ex := t.Handler.Set("", "", test01)
	if ex != nil {
		t.T.Fatalf("failed to add record of test 0.0.1")
	}
	return processJournal(t, test01.Type, test01.Id)
}

func processJournal(t *TestEnv, dataType string, dataId string) int {
	worker := DataJournal.NewJournalWorker(t.Journal, dataType, dataId, nil, t.Processes)
	count := 0
	for {
		err := worker.ProcessNextEntry()
		if err != nil {
			if err.Status == http.StatusNotFound {
				log.Printf("run out of Journal Entry on %s/%s", dataType, dataId)
				break
			}
			t.T.Fatal(err)
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

func addLeafSchemaAndData(t *TestEnv) {
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

func addLayer1SchemaDataAndUpgradSchema(t *TestEnv) {
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
	_, err := t.Handler.LocalData(CmtIndex.KeyCmtIdx, "test")
	if err == nil {
		t.T.Fatal("there should be no CmtIndex of Type=[test]")
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
						"indexTemplate": "categoryLayer1/{categoryLayer1}/testAry"
					}
				}
			}
		}
	}`
	addSchema(t, catLayer1SchemaV2)
	cmtIdxData, err := t.Handler.LocalData(CmtIndex.KeyCmtIdx, "test")
	if err != nil {
		t.T.Fatal("failed to add CmtIndex of Type=[test]")
	}
	idxRecord, ex := Record.LoadMap(cmtIdxData)
	if ex != nil {
		t.T.Fatal(ex)
	}
	cmtIdx, ex := CmtIndex.LoadMap(idxRecord.Data)
	if ex != nil {
		t.T.Fatal(ex)
	}
	if cmtIdx.DataType != "test" {
		t.T.Fatalf("generated wrong cmtIndex, type [%s]!=[test]", cmtIdx.DataType)
	}

	catSub, ok := cmtIdx.Subscriber["categoryLayer1"]
	if !ok {
		t.T.Fatal("missing cmtIdx for type=[categoryLayer1]")
	}
	if catSub.DataType != "categoryLayer1" {
		t.T.Fatalf("invalid sub type [%s]!=[categoryLayer1]", catSub.DataType)
	}
	verPath, ok := catSub.VersionIndex["0.0.2"]
	if !ok {
		t.T.Fatal("missing cmtIdx for version [0.0.2]")
	}
	if verPath.Version != "0.0.2" {
		t.T.Fatalf("cmdIdx invalid version [%s]!=[0.0.2]", verPath.Version)
	}
	if len(verPath.IndexTemplate) != 1 {
		t.T.Fatalf("invalid index template number. [%d]!=[1]", len(verPath.IndexTemplate))
	}
	if verPath.IndexTemplate[0].(string) != "categoryLayer1/{categoryLayer1}/testAry" {
		t.T.Fatalf("cmtIdx path[%s]!=[categoryLayer1/{categoryLayer1}/testAry]", verPath.IndexTemplate[0].(string))
	}
}

func addLayer1DataAndLeafDataAndDelete(t *TestEnv) {
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
		t.T.Fatal(ex)
	}
	err := t.Handler.Set("", "", catLayer1)
	if err != nil {
		t.T.Fatal(err)
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
	err = t.Handler.Delete("test", "test02")
	if err != nil {
		t.T.Fatal(err)
	}
	processJournal(t, "test", "test02")
}

func TestAddSchemaWithoutCmtIndex(t *testing.T) {
	env := prepEnv(t)
	addLeafSchemaAndData(env)
}

func TestAddCmtSchema(t *testing.T) {
	env := prepEnv(t)
	addLeafSchemaAndData(env)
	addLayer1SchemaDataAndUpgradSchema(env)
	addLayer1DataAndLeafDataAndDelete(env)
	processJournal(env, "categoryLayer1", "Layer1Cat01")
}
