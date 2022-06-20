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
	"log"
	"testing"

	"github.com/salesforce/UniTAO/lib/Schema"
	"github.com/salesforce/UniTAO/lib/Util"
)

func TestSchemaValidate(t *testing.T) {
	log.Print("test Run")
	filePath := "data/infrastructure.json"
	testData, err := Util.LoadJSONMap(filePath)
	if err != nil {
		t.Fatalf("failed loading data from [path]=[%s], Err:%s", filePath, err)
	}
	schemaRecord, ok := testData[Schema.Schema].(map[string]interface{})
	if !ok {
		t.Fatalf("missing field [%s] from test data", Schema.Schema)
	}
	data, ok := testData[Schema.RecordData].(map[string]interface{})
	if !ok {
		t.Fatalf("missing field [%s] from test data", Schema.RecordData)
	}
	schema, err := Schema.LoadSchemaOps(schemaRecord)
	if err != nil {
		t.Fatalf("failed to load schema record, Error:\n%s", err)
	}
	err = schema.Validate(data)
	if err != nil {
		t.Fatalf("schema validation failed. Error:\n%s", err)
	}
	negativeData, ok := testData["negativeData"].([]interface{})
	if !ok {
		return
	}
	for idx, data := range negativeData {
		err = schema.Validate(data.(map[string]interface{}))
		if err == nil {
			t.Fatalf("failed to alert schema err on negative data [idx]=[%d]", idx)
		}
		log.Printf("spot err at idx=[%d], Err:\n%s", idx, err)
	}

}
