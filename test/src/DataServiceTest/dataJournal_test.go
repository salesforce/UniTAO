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
	"DataService/Common"
	"DataService/Config"
	"DataService/DataJournal"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/salesforce/UniTAO/lib/Schema/Record"
)

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
	dataStr, err := GetSchemaOfSchema()
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
	journal, e := DataJournal.NewJournalLib(mockDb, config.DataTable.Data)
	if e != nil {
		t.Fatalf("failed to create Journal Library. Error: %s", err)
	}
	e = journal.AddJournal("test", "testid_123", nil, map[string]interface{}{"attr": "test"})
	if e != nil {
		t.Fatalf(e.Error())
	}
	if len(mockDb.Data[Common.KeyJournal].(map[string]interface{})) != 1 {
		t.Fatalf("invalid add Journal Entry.")
	}
	if len(journal.Cache["test"]["testid_123"].Sort) != 1 {
		t.Fatalf("failed to create first journal page")
	}
	record, err := Record.LoadMap(mockDb.Data[Common.KeyJournal].(map[string]interface{})["dataType:test_dataId:testid_123_page:0"].(map[string]interface{}))
	if err != nil {
		t.Fatalf("new entry record failed to load as record")
	}
	if len(record.Data["active"].([]interface{})) != 1 {
		t.Fatal("failed add the first entry")
	}
	journal.AddJournal("test", "testid_123", nil, map[string]interface{}{"attr": "test"})
	if len(mockDb.Data[Common.KeyJournal].(map[string]interface{})) != 1 {
		t.Fatalf("invalid add Journal Entry.")
	}
	record, err = Record.LoadMap(mockDb.Data[Common.KeyJournal].(map[string]interface{})["dataType:test_dataId:testid_123_page:0"].(map[string]interface{}))
	if err != nil {
		t.Fatalf("new entry record failed to load as record")
	}
	if len(record.Data["active"].([]interface{})) != 2 {
		t.Fatal("failed add the first entry")
	}
	for i := 0; i < 8; i++ {
		journal.AddJournal("test", "testid_123", nil, map[string]interface{}{"attr": fmt.Sprintf("test_%d", i)})
		if len(mockDb.Data[Common.KeyJournal].(map[string]interface{})) != 1 {
			t.Fatalf("invalid add Journal Entry.")
		}
		record, err = Record.LoadMap(mockDb.Data[Common.KeyJournal].(map[string]interface{})["dataType:test_dataId:testid_123_page:0"].(map[string]interface{}))
		if err != nil {
			t.Fatalf("new entry record failed to load as record")
		}
		if len(record.Data["active"].([]interface{})) != i+3 {
			t.Fatal("failed add the first entry")
		}
	}
	journal.AddJournal("test", "testid_123", nil, map[string]interface{}{"attr": fmt.Sprintf("test_%d", 0)})
	if len(mockDb.Data[Common.KeyJournal].(map[string]interface{})) != 2 {
		t.Fatalf("invalid add Journal Entry.")
	}
	record, err = Record.LoadMap(mockDb.Data[Common.KeyJournal].(map[string]interface{})["dataType:test_dataId:testid_123_page:0"].(map[string]interface{}))
	if err != nil {
		t.Fatalf("new entry record failed to load as record")
	}
	if len(record.Data["active"].([]interface{})) != 10 {
		t.Fatal("failed add the first entry")
	}
	record, err = Record.LoadMap(mockDb.Data[Common.KeyJournal].(map[string]interface{})["dataType:test_dataId:testid_123_page:1"].(map[string]interface{}))
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
	journal, e := DataJournal.NewJournalLib(mockDb, config.DataTable.Data)
	if e != nil {
		t.Fatalf("failed to create Journal Library. Error: %s", err)
	}
	for i := 0; i < 16; i++ {
		e = journal.AddJournal("test", "testid_123", nil, map[string]interface{}{"attr": fmt.Sprintf("test_%d", i)})
		if e != nil {
			t.Fatalf("failed to add hournal. Error: %s", e)
		}
		if i < 10 {
			if len(mockDb.Data[Common.KeyJournal].(map[string]interface{})) != 1 {
				t.Fatalf("invalid add Journal Entry.")
			}
		} else {
			if len(mockDb.Data[Common.KeyJournal].(map[string]interface{})) != 2 {
				t.Fatalf("invalid add Journal Entry.")
			}
		}
		pageCache := journal.Cache["test"]["testid_123"]
		page := pageCache.Sort[len(pageCache.Sort)-1]
		pageId := DataJournal.PageId("test", "testid_123", page)
		record, err := Record.LoadMap(mockDb.Data[Common.KeyJournal].(map[string]interface{})[pageId].(map[string]interface{}))
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
	entry := journal.NextJournalEntry("test", "testid_123")
	if entry.Page != 0 {
		t.Fatal("failed to get the entry from first page")
	}
	if entry.Idx != 1 {
		t.Fatal("failed to get first entry")
	}
	if entry.Before != nil {
		t.Fatalf("entry.%s should be nil", DataJournal.KeyBefore)
	}
	if entry.After == nil {
		t.Fatalf("entry.%s should not be nil", DataJournal.KeyAfter)
	}
	if entry.After["attr"].(string) != "test_0" {
		t.Fatalf("got the wrong entry, attr[%s]!=[test_0]", entry.After["attr"].(string))
	}
	e = journal.ArchiveJournalEntry("test", "testid_123", entry)
	if e != nil {
		t.Fatal(e)
	}
	record, err := Record.LoadMap(mockDb.Data[Common.KeyJournal].(map[string]interface{})["dataType:test_dataId:testid_123_page:0"].(map[string]interface{}))
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
		entry := journal.NextJournalEntry("test", "testid_123")
		if i < 9 {
			if entry.Page != 0 {
				t.Fatal("failed to get the entry from first page")
			}
			if entry.Idx != i+2 {
				t.Fatalf("failed to get entry [%d]!=[%d]", entry.Idx, i+2)
			}
		} else {
			if entry.Page != 1 {
				t.Fatal("failed to get the entry from first page")
			}
			if entry.Idx != i-8 {
				t.Fatalf("failed to get entry [%d]!=[%d]", entry.Idx, i+2)
			}
		}
		journal.ArchiveJournalEntry("test", "testid_123", entry)
	}
}
