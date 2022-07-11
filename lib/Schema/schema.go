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

package Schema

import (
	"encoding/json"
	"fmt"

	"github.com/salesforce/UniTAO/lib/Schema/JsonKey"
	"github.com/salesforce/UniTAO/lib/Schema/Record"
	"github.com/salesforce/UniTAO/lib/Schema/SchemaDoc"
	"github.com/salesforce/UniTAO/lib/Util"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

const (
	Inventory = "inventory"
)

type SchemaOps struct {
	Record     *Record.Record
	SchemaData map[string]interface{}
	Schema     *SchemaDoc.SchemaDoc
	Meta       *jsonschema.Schema
}

func LoadSchemaOpsRecord(record *Record.Record) (*SchemaOps, error) {
	recWithSchema := SchemaOps{
		Record: record,
	}
	err := recWithSchema.init()
	if err != nil {
		return nil, fmt.Errorf("failed while init SchemaOps. Error:%s", err)
	}
	return &recWithSchema, nil
}

func LoadSchemaOpsData(dataType string, typeVer string, data map[string]interface{}) (*SchemaOps, error) {
	dataId, ok := data[JsonKey.Name]
	if !ok {
		return nil, fmt.Errorf("missing required field=[%s] from data", JsonKey.Name)
	}
	record := Record.NewRecord(dataType, typeVer, dataId.(string), data)
	return LoadSchemaOpsRecord(record)
}

func (schema *SchemaOps) init() error {
	if schema.Record.Type != JsonKey.Schema {
		return fmt.Errorf("schema record has wrong [%s], [%s]!=[%s]", Record.DataType, schema.Record.Id, JsonKey.Schema)
	}
	schemaData, err := Util.JsonCopy(schema.Record.Data)
	if err != nil {
		return fmt.Errorf("copy schema.Record.Data failed. Error: %s", err)
	}
	doc, err := SchemaDoc.New(schemaData.(map[string]interface{}), schema.Record.Id, nil)
	if err != nil {
		return fmt.Errorf("failed to create Schema Doc, err: %s", err)
	}
	schema.Schema = doc
	schemaBytes, err := json.MarshalIndent(doc.Data, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to MarshalIndent value [field]=[data], Err:%s", err)
	}
	meta, err := jsonschema.CompileString(schema.Record.Id, string(schemaBytes))
	if err != nil {
		return fmt.Errorf("failed to compile schema, [%s]=[%s] Err:%s", Record.DataId, schema.Record.Id, err)
	}
	schema.Meta = meta
	return nil
}

func (schema *SchemaOps) ValidateRecord(record *Record.Record) error {
	if schema.Record.Id != record.Type {
		return fmt.Errorf("schema id and payload data type does not match, [%s]!=[%s]", schema.Record.Id, record.Type)
	}
	return schema.ValidateData(record.Data)
}

func (schema *SchemaOps) ValidateData(data map[string]interface{}) error {
	err := schema.Meta.Validate(data)
	if err != nil {
		return fmt.Errorf("schema validation failed. Error:\n%s", err)
	}
	return nil
}
