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

import (
	"testing"

	"github.com/salesforce/UniTAO/lib/Schema/JsonKey"
	"github.com/salesforce/UniTAO/lib/Schema/Record"
	"github.com/salesforce/UniTAO/lib/Schema/SchemaDoc"
	"github.com/salesforce/UniTAO/lib/Util"
)

// test schema record operation

func TestProcessArchivedSchema(t *testing.T) {
	env := prepEnv(t)
	schemaV1 := `{
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
				"attr2": {
					"type": "string"
				}
			}
		}
	}`
	addSchema(env, schemaV1)
	schemaV2 := `{
		"__id": "test",
		"__type": "schema",
		"__ver": "0.0.1",
		"data": {
			"name": "test",
			"version": "0.0.2",
			"properties": {
				"attr1": {
					"type": "string"
				},
				"attr2": {
					"type": "string"
				},
				"attr3": {
					"type": "string",
					"required": false
				}
			}
		}
	}`
	addSchema(env, schemaV2)
	typeList, err := env.Handler.List(JsonKey.Schema)
	if err != nil {
		t.Fatalf("failed to get type list. Error:%s", err)
	}
	typeMap := Util.IdxList(typeList)
	v1Id := SchemaDoc.ArchivedSchemaId("test", "0.0.1")
	if _, ok := typeMap[v1Id]; !ok {
		t.Fatalf("v1 schema still need to be alive")
	}
	if _, ok := typeMap["test"]; !ok {
		t.Fatalf("missing test v2 schema")
	}
}

func TestProcessArchivedSchemaWithData(t *testing.T) {
	env := prepEnv(t)
	schemaV1 := `{
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
				"attr2": {
					"type": "string"
				}
			}
		}
	}`
	addSchema(env, schemaV1)
	test01 := `{
		"__id": "test01",
		"__type": "test",
		"__ver": "0.0.1",
		"data": {
			"attr1": "test",
			"attr2": "test"
		}
	}`
	addData(env, test01)
	schemaV2 := `{
		"__id": "test",
		"__type": "schema",
		"__ver": "0.0.1",
		"data": {
			"name": "test",
			"version": "0.0.2",
			"properties": {
				"attr1": {
					"type": "string"
				},
				"attr2": {
					"type": "string"
				},
				"attr3": {
					"type": "string",
					"required": false
				}
			}
		}
	}`
	addSchema(env, schemaV2)
	typeList, err := env.Handler.List(JsonKey.Schema)
	if err != nil {
		t.Fatalf("failed to get type list. Error:%s", err)
	}
	typeMap := Util.IdxList(typeList)
	v1Id := SchemaDoc.ArchivedSchemaId("test", "0.0.1")
	if _, ok := typeMap[v1Id]; !ok {
		t.Fatalf("v1 schema should not be removed")
	}
	if _, ok := typeMap["test"]; !ok {
		t.Fatalf("missing test v2 schema")
	}
	test02Str := `{
		"__id": "test02",
		"__type": "test",
		"__ver": "0.0.1",
		"data": {
			"attr1": "test",
			"attr2": "test"
		}
	}`
	test02Rec, ex := Record.LoadStr(test02Str)
	if ex != nil {
		t.Fatalf("failed to load test01 as record")
	}
	err = env.Handler.Add(test02Rec)
	if err == nil {
		t.Fatalf("failed to catch submit data to archived schema")
	}
	t.Logf("now upgrade test01 to 0.0.2")
	_, err = env.Handler.Patch("test", "test01/__ver", nil, "0.0.2")
	if err != nil {
		t.Fatalf("failed to upgrade test01 to version 0.0.2")
	}
	processJournal(env, "test", "test01")

}
