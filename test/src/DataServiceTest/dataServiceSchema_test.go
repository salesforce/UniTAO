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
	"DataService/DataHandler"
	"net/http"
	"testing"

	"github.com/salesforce/UniTAO/lib/Schema/JsonKey"
	"github.com/salesforce/UniTAO/lib/Schema/Record"
	"github.com/salesforce/UniTAO/lib/Schema/SchemaDoc"
	"github.com/salesforce/UniTAO/lib/Util/Http"
)

func AddData(handler *DataHandler.Handler, data string) *Http.HttpError {
	record, ex := Record.LoadStr(data)
	if ex != nil {
		return Http.WrapError(ex, "failed to load data as record", http.StatusBadRequest)
	}
	err := handler.Add(record)
	if err != nil {
		return err
	}
	return nil
}

func TestAddSchema(t *testing.T) {
	handler, ex := MockHandler()
	if ex != nil {
		t.Fatalf(ex.Error())
	}
	schemaList, err := handler.List(JsonKey.Schema)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if len(schemaList) == 0 {
		t.Fatalf("failed to init db, schema should be empty")
	}
	baseSchema := `{
		"__id": "test",
		"__type": "schema",
		"__ver": "0.0.1",
		"data": {
			"name": "test",
			"version": "0.0.1",
			"properties": {
				"testAttr1": {
					"type": "string"
				}
			}
		}
	}`
	err = AddData(handler, baseSchema)
	if err != nil {
		t.Fatalf("failed to add init schema. Error: %s", err)
	}
	data, err := handler.Get(JsonKey.Schema, "test")
	if err != nil {
		t.Fatalf(err.Error())
	}
	record, ex := Record.LoadMap(data)
	if ex != nil {
		t.Fatalf(ex.Error())
	}
	if record.Data[JsonKey.Version] != "0.0.1" {
		t.Fatal("failed to create base schema")
	}
	newSchema := `{
		"__id": "test",
		"__type": "schema",
		"__ver": "0.0.1",
		"data": {
			"name": "test",
			"version": "0.0.2",
			"properties": {
				"testAttr1": {
					"type": "string"
				}
			}
		}
	}`
	err = AddData(handler, newSchema)
	if err != nil {
		t.Fatalf("failed to add new schema. Error: %s", err)
	}
	data, err = handler.Get(JsonKey.Schema, "test")
	if err != nil {
		t.Fatalf(err.Error())
	}
	record, ex = Record.LoadMap(data)
	if ex != nil {
		t.Fatalf(ex.Error())
	}
	if record.Data[JsonKey.Version] != "0.0.2" {
		t.Fatal("failed to create new schema")
	}
	base, err := handler.LocalSchema("test", "0.0.1")
	if err != nil {
		t.Fatalf("failed to load archived schema, Error:%s", err)
	}
	if base.Record.Id != SchemaDoc.ArchivedSchemaId("test", "0.0.1") {
		t.Fatalf("failed to load archived schema.")
	}
	newS, err := handler.LocalSchema("test", "")
	if err != nil {
		t.Fatalf("failed to load new schema.Error:%s", err)
	}
	if newS.Record.Id != "test" {
		t.Fatalf("failed to add new schema")
	}
	if newS.Schema.Version != "0.0.2" {
		t.Fatal("failed to load new schema version 00.2")
	}
}

func TestAddData(t *testing.T) {
	handler, ex := MockHandler()
	if ex != nil {
		t.Fatalf(ex.Error())
	}
	baseSchema := `{
		"__id": "test",
		"__type": "schema",
		"__ver": "0.0.1",
		"data": {
			"name": "test",
			"version": "0.0.1",
			"properties": {
				"testAttr1": {
					"type": "string"
				}
			}
		}
	}`
	err := AddData(handler, baseSchema)
	if err != nil {
		t.Fatalf("failed to add init schema. Error: %s", err)
	}
	baseData := `{
		"__id": "base01",
		"__type": "test",
		"__ver": "0.0.1",
		"data": {
			"testAttr1": "test"
		}
	}`
	err = AddData(handler, baseData)
	if err != nil {
		t.Fatalf("failed to add data on base schema. Error:%s", err)
	}
	newData := `{
		"__id": "new01",
		"__type": "test",
		"__ver": "0.0.2",
		"data": {
			"testAttr2": "test"
		}
	}`
	err = AddData(handler, newData)
	if err == nil {
		t.Fatalf("failed to catch schema error on new schema data")
	}
	newSchema := `{
		"__id": "test",
		"__type": "schema",
		"__ver": "0.0.1",
		"data": {
			"name": "test",
			"version": "0.0.2",
			"properties": {
				"testAttr2": {
					"type": "string"
				}				
			}
		}
	}`
	err = AddData(handler, newSchema)
	if err != nil {
		t.Fatalf("failed to upgrade to new schema")
	}
	err = AddData(handler, newData)
	if err != nil {
		t.Fatalf("failed to data for new schema")
	}
	baseData02 := `{
		"__id": "base02",
		"__type": "test",
		"__ver": "0.0.1",
		"data": {
			"testAttr1": "test02"
		}
	}`
	err = AddData(handler, baseData02)
	if err != nil {
		t.Fatalf("failed to add data on archived schema")
	}
}
