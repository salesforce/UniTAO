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
	"DataService/Config"
	"DataService/DataJournal"
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/salesforce/UniTAO/lib/Schema/Record"
	"github.com/salesforce/UniTAO/lib/Util"
)

func getSchemaOfSchema() (string, error) {
	rootDir, err := Util.RootDir()
	if err != nil {
		return "", fmt.Errorf("failed to get running dir")
	}
	schemaFile, err := filepath.Abs(filepath.Join(rootDir, "lib/Schema/data/schema.json"))
	if err != nil {
		return "", fmt.Errorf("failed to get ABS path of schema.json")
	}
	schemaData, err := Util.LoadJsonFile(schemaFile)
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

func mockDbConfig() (*Config.Confuguration, error) {
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
			"id": "DataService_01"
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
	return &config, nil
}

func NewDb(config *Config.Confuguration) (*MockDatabase, error) {
	dataStr, err := getSchemaOfSchema()
	if err != nil {
		return nil, err
	}
	mockDb, err := NewMockDb(config.Database, dataStr)
	if err != nil {
		return nil, err
	}
	return mockDb, nil
}

func TestParseJournalId(t *testing.T) {
	journalId := "dataType:test_dataId:test01_page:1"
	dataType, dataId, idx, err := DataJournal.ParseJournalId(journalId)
	if err != nil {
		t.Fatalf("failed to parse id. Error:%s", err)
	}
	if dataType != "test" || dataId != "test01" || idx != 1 {
		t.Fatalf("failed to parse journalId=[%s]", journalId)
	}
	journalId = "dataType:test_dataId:test01_1"
	dataType, dataId, idx, err = DataJournal.ParseJournalId(journalId)
	if err != nil {
		t.Fatalf("failed to parse id. Error:%s", err)
	}
	if dataType != "test" || dataId != "test01_1" || idx != 0 {
		t.Fatalf("failed to parse journalId=[%s]", journalId)
	}
	journalId = "dataType:test_dataId:test01_page:1a"
	_, _, _, err = DataJournal.ParseJournalId(journalId)
	if err == nil {
		t.Fatalf("fail to catch id format error")
	}
}

func TestAddJournal(t *testing.T) {
	config, err := mockDbConfig()
	if err != nil {
		t.Fatalf("failed to create MockDbConfig. Error:%s", err)
	}
	mockDb, err := NewDb(config)
	if err != nil {
		t.Fatalf("failed to create MockDb. Error:%s", err)
	}
	journal := DataJournal.NewJournalLib(mockDb, config.DataTable.Data)
	page, err := journal.AddJournal("test", "testid_123", nil, map[string]interface{}{"attr": "test"})
	if len(mockDb.Data[DataJournal.KeyJournal].(map[string]interface{})) != 1 {
		t.Fatalf("invalid add Journal Entry.")
	}
	if page != 1 {
		t.Fatalf("failed to create first journal page")
	}
	record, err := Record.LoadMap(mockDb.Data[DataJournal.KeyJournal].(map[string]interface{})["dataType:test_dataId:testid_123_page:1"].(map[string]interface{}))
	if err != nil {
		t.Fatalf("new entry record failed to load as record")
	}
	if len(record.Data["active"].([]interface{})) != 1 {
		t.Fatal("failed add the first entry")
	}
	journal.AddJournal("test", "testid_123", nil, map[string]interface{}{"attr": "test"})
	if len(mockDb.Data[DataJournal.KeyJournal].(map[string]interface{})) != 1 {
		t.Fatalf("invalid add Journal Entry.")
	}
	record, err = Record.LoadMap(mockDb.Data[DataJournal.KeyJournal].(map[string]interface{})["dataType:test_dataId:testid_123_page:1"].(map[string]interface{}))
	if err != nil {
		t.Fatalf("new entry record failed to load as record")
	}
	if len(record.Data["active"].([]interface{})) != 2 {
		t.Fatal("failed add the first entry")
	}
	for i := 0; i < 8; i++ {
		journal.AddJournal("test", "testid_123", nil, map[string]interface{}{"attr": fmt.Sprintf("test_%d", i)})
		if len(mockDb.Data[DataJournal.KeyJournal].(map[string]interface{})) != 1 {
			t.Fatalf("invalid add Journal Entry.")
		}
		record, err = Record.LoadMap(mockDb.Data[DataJournal.KeyJournal].(map[string]interface{})["dataType:test_dataId:testid_123_page:1"].(map[string]interface{}))
		if err != nil {
			t.Fatalf("new entry record failed to load as record")
		}
		if len(record.Data["active"].([]interface{})) != i+3 {
			t.Fatal("failed add the first entry")
		}
	}
	journal.AddJournal("test", "testid_123", nil, map[string]interface{}{"attr": fmt.Sprintf("test_%d", 0)})
	if len(mockDb.Data[DataJournal.KeyJournal].(map[string]interface{})) != 2 {
		t.Fatalf("invalid add Journal Entry.")
	}
	record, err = Record.LoadMap(mockDb.Data[DataJournal.KeyJournal].(map[string]interface{})["dataType:test_dataId:testid_123_page:1"].(map[string]interface{}))
	if err != nil {
		t.Fatalf("new entry record failed to load as record")
	}
	if len(record.Data["active"].([]interface{})) != 10 {
		t.Fatal("failed add the first entry")
	}
	record, err = Record.LoadMap(mockDb.Data[DataJournal.KeyJournal].(map[string]interface{})["dataType:test_dataId:testid_123_page:2"].(map[string]interface{}))
	if err != nil {
		t.Fatalf("new entry record failed to load as record")
	}
	if len(record.Data["active"].([]interface{})) != 1 {
		t.Fatal("failed add the first entry")
	}
}

func TestArchiveJournal(t *testing.T) {
	config, err := mockDbConfig()
	if err != nil {
		t.Fatalf("failed to create MockDbConfig. Error:%s", err)
	}
	mockDb, err := NewDb(config)
	if err != nil {
		t.Fatalf("failed to create MockDb. Error:%s", err)
	}
	journal := DataJournal.NewJournalLib(mockDb, config.DataTable.Data)
	for i := 0; i < 16; i++ {
		page, e := journal.AddJournal("test", "testid_123", nil, map[string]interface{}{"attr": fmt.Sprintf("test_%d", i)})
		if e != nil {
			t.Fatalf("failed to add hournal. Error: %s", e)
		}
		if i < 10 {
			if len(mockDb.Data[DataJournal.KeyJournal].(map[string]interface{})) != 1 {
				t.Fatalf("invalid add Journal Entry.")
			}
			if page != 1 {
				t.Fatalf("invalid add Journal Entry to page %d.", page)
			}
		} else {
			if len(mockDb.Data[DataJournal.KeyJournal].(map[string]interface{})) != 2 {
				t.Fatalf("invalid add Journal Entry.")
			}
			if page != 2 {
				t.Fatalf("invalid add Journal Entry. to page %d", page)
			}
		}
		pageId := DataJournal.PageId("test", "testid_123", page)
		record, err := Record.LoadMap(mockDb.Data[DataJournal.KeyJournal].(map[string]interface{})[pageId].(map[string]interface{}))
		if err != nil {
			t.Fatalf("new entry record failed to load as record")
		}
		if i < 10 {
			if len(record.Data[DataJournal.KeyActive].([]interface{})) != i+1 {
				t.Fatalf("failed add the entry to page[%d]", page)
			}
		} else {
			if len(record.Data[DataJournal.KeyActive].([]interface{})) != i-9 {
				t.Fatalf("failed add the entry to page[%d]", page)
			}
		}

	}
	entry, e := journal.NextJournalEntry("test", "testid_123")
	if e != nil {
		t.Fatal(e)
	}
	pageIdx := int(entry[DataJournal.KeyPage].(float64))
	if pageIdx != 1 {
		t.Fatal("failed to get the entry from first page")
	}
	entryIdx := int(entry[DataJournal.KeyIdx].(float64))
	if entryIdx != 1 {
		t.Fatal("failed to get first entry")
	}
	if entry[DataJournal.KeyBefore] != nil {
		t.Fatalf("entry.%s should be nil", DataJournal.KeyBefore)
	}
	if entry[DataJournal.KeyAfter] == nil {
		t.Fatalf("entry.%s should not be nil", DataJournal.KeyAfter)
	}
	if entry[DataJournal.KeyAfter].(map[string]interface{})["attr"].(string) != "test_0" {
		t.Fatalf("got the wrong entry, attr[%s]!=[test_0]", entry[DataJournal.KeyAfter].(map[string]interface{})["attr"].(string))
	}
	e = journal.ArchiveJournalEntry("test", "testid_123", entry)
	if e != nil {
		t.Fatal(e)
	}
	record, err := Record.LoadMap(mockDb.Data[DataJournal.KeyJournal].(map[string]interface{})["dataType:test_dataId:testid_123_page:1"].(map[string]interface{}))
	if err != nil {
		t.Fatalf("new entry record failed to load as record")
	}
	if len(record.Data[DataJournal.KeyActive].([]interface{})) != 9 {
		t.Fatal("failed archive entry")
	}
	if len(record.Data[DataJournal.KeyArchived].([]interface{})) != 1 {
		t.Fatal("failed archive entry")
	}
	for i := 0; i < 15; i++ {
		entry, e := journal.NextJournalEntry("test", "testid_123")
		if e != nil {
			t.Fatal(e)
		}
		pageIdx := int(entry[DataJournal.KeyPage].(float64))
		entryIdx := int(entry[DataJournal.KeyIdx].(float64))
		if i < 9 {
			if pageIdx != 1 {
				t.Fatal("failed to get the entry from first page")
			}
			if entryIdx != i+2 {
				t.Fatalf("failed to get entry [%d]!=[%d]", entryIdx, i+2)
			}
		} else {
			if pageIdx != 2 {
				t.Fatal("failed to get the entry from first page")
			}
			if entryIdx != i-8 {
				t.Fatalf("failed to get entry [%d]!=[%d]", entryIdx, i+2)
			}
		}
		journal.ArchiveJournalEntry("test", "testid_123", entry)
	}
}
