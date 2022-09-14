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

package SchemaTest

import (
	"encoding/json"
	"testing"

	"github.com/salesforce/UniTAO/lib/Schema/Record"
	"github.com/salesforce/UniTAO/lib/Schema/SchemaDoc"
)

func TestOjectKeyAttrRequired(t *testing.T) {
	schemaStr := `{
		"name": "test",
		"description": "test schema",
		"key": "{type}_{name}_{version}",
		"properties": {
			"name": {
				"type": "string"
			},
			"type": {
				"type": "string"
			},
			"version": {
				"type": "string",
				"required": false
			}
		}
	}`
	data := map[string]interface{}{}
	err := json.Unmarshal([]byte(schemaStr), &data)
	if err != nil {
		t.Errorf("failed to load schemaStr. Error:%s", err)
	}
	_, err = SchemaDoc.New(data, "test", nil)
	if err == nil {
		t.Errorf("failed to validate key [attr]=[version]] as required")
	}
}

func TestObjectKey(t *testing.T) {
	schemaStr := `{
		"name": "test",
		"description": "test schema",
		"key": "{type}_{name}_{version}",
		"properties": {
			"name": {
				"type": "string"
			},
			"type": {
				"type": "string"
			},
			"version": {
				"type": "string"
			}
		}
	}`
	recordStr := `{
		"__id": "test_key_01",
		"__type": "test",
		"__ver": "0.0.1",
		"data": {
			"name": "key",
			"type": "test",
			"version": "01"
		}
	}`
	data := map[string]interface{}{}
	err := json.Unmarshal([]byte(schemaStr), &data)
	if err != nil {
		t.Errorf("failed to load schemaStr. Error:%s", err)
	}
	schema, err := SchemaDoc.New(data, "test", nil)
	if err != nil {
		t.Errorf("failed to load schemaStr as SchemaDoc")
	}
	record, err := Record.LoadStr(recordStr)
	if err != nil {
		t.Errorf("failed to load recordStr as Record")
	}
	recordKey, err := schema.BuildKey(record.Data)
	if err != nil {
		t.Errorf("failed to build record key. Error:%s", err)
	}
	if recordKey != record.Id {
		t.Errorf("build the wrong key. [%s]!=[%s]", recordKey, record.Id)
	}
}
