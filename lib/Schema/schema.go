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
	"path"
	"strings"

	"github.com/salesforce/UniTAO/lib/Schema/JsonKey"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

const (
	Data       = "Data"
	DataType   = "__type"
	DataId     = "__id"
	DataVer    = "__ver"
	Schema     = "schema"
	RecordData = "data"
	Inventory  = "inventory"
)

type Record struct {
	Id      string                 `json:"__id"`
	Type    string                 `json:"__type"`
	Version string                 `json:"__ver"`
	Data    map[string]interface{} `json:"data"`
	Raw     string
}

type SchemaOps struct {
	Record *Record
	Schema *SchemaDoc
	Meta   *jsonschema.Schema
}

func LoadRecord(data map[string]interface{}) (*Record, error) {
	recordBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal doc to string, Err:%s", err)
	}
	record := Record{}
	err = json.Unmarshal(recordBytes, &record)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal data to Record. Error:%s", err)
	}
	record.Raw = string(recordBytes)
	return &record, nil
}

func LoadSchemaOps(data map[string]interface{}) (*SchemaOps, error) {
	record, err := LoadRecord(data)
	if err != nil {
		return nil, err
	}
	recWithSchema := SchemaOps{
		Record: record,
	}
	recWithSchema.init()
	return &recWithSchema, nil
}

func (schema *SchemaOps) init() error {
	if schema.Record.Type != Schema {
		return fmt.Errorf("schema record has wrong [%s], [%s]!=[%s]", DataType, schema.Record.Id, Schema)
	}
	doc, err := NewSchemaDoc(schema.Record.Data, schema.Record.Id, nil)
	if err != nil {
		return fmt.Errorf("failed to create Schema Doc, err: %s", err)
	}
	schema.Schema = doc
	schemaBytes, err := json.MarshalIndent(schema.Record.Data, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to MarshalIndent value [field]=[data], Err:%s", err)
	}
	meta, err := jsonschema.CompileString(schema.Record.Id, string(schemaBytes))
	if err != nil {
		return fmt.Errorf("failed to compile schema, [%s]=[%s] Err:%s", DataId, schema.Record.Id, err)
	}
	schema.Meta = meta
	return nil
}

func (schema *SchemaOps) Validate(payload map[string]interface{}) error {
	record, err := LoadRecord(payload)
	if err != nil {
		return fmt.Errorf("failed to load payload as Record. Error:%s", err)
	}
	if schema.Record.Id != record.Type {
		return fmt.Errorf("schema id and payload data type does not match, [%s]!=[%s]", schema.Record.Id, record.Type)
	}
	err = schema.Meta.Validate(record.Data)
	if err != nil {
		return fmt.Errorf("schema validation failed. Error:\n%s", err)
	}
	return nil
}

type SchemaDoc struct {
	Id          string
	Parent      *SchemaDoc
	Data        map[string]interface{}
	Definitions map[string]*SchemaDoc
	Refs        []*SchemaDocRef
	SubDocs     map[string]*SchemaDoc
}

type SchemaDocRef struct {
	Doc         *SchemaDoc
	Name        string
	ContentType string
}

func NewSchemaDoc(data map[string]interface{}, id string, parent *SchemaDoc) (*SchemaDoc, error) {
	parentPath := ""
	if parent != nil {
		parentPath = parent.Path()
	}
	_, ok := data[JsonKey.Properties].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to create doc path=[%s/%s]", parentPath, id)
	}
	doc := SchemaDoc{
		Id:      id,
		Parent:  parent,
		Data:    data,
		Refs:    []*SchemaDocRef{},
		SubDocs: map[string]*SchemaDoc{},
	}
	docDefs, ok := data[JsonKey.Definitions]
	if ok {
		defMap, ok := docDefs.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("failed to parse field=[%s], path=[%s/%s]", JsonKey.Definitions, parentPath, id)
		}
		doc.Definitions = make(map[string]*SchemaDoc, len(defMap))
		for key, def := range defMap {
			defObj, ok := def.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("failed to parse definition to object. path=[%s/%s/%s/%s]", parentPath, id, JsonKey.Definitions, key)
			}
			defDoc, err := NewSchemaDoc(defObj, key, &doc)
			if err != nil {
				return nil, err
			}
			doc.Definitions[key] = defDoc
		}
	}
	err := doc.preprocess()
	if err != nil {
		return nil, fmt.Errorf("failed @preprocess, Err:\n%s", err)
	}
	return &doc, nil
}

func (d *SchemaDoc) Path() string {
	if d.Parent == nil {
		return d.Id
	}
	return path.Join(d.Parent.Path(), d.Id)
}

func (d *SchemaDoc) preprocess() error {
	err := d.processRequired()
	if err != nil {
		err := fmt.Errorf("preprocess failed @processRequired, [path]=[%s], Error:%s", d.Path(), err)
		return err
	}
	err = d.processRefs()
	if err != nil {
		err := fmt.Errorf("preprocess failed @processInvRefs, [path]=[%s], Error:%s", d.Path(), err)
		return err
	}
	if d.Definitions != nil {
		for _, defDoc := range d.Definitions {
			err = defDoc.preprocess()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *SchemaDoc) processRequired() error {
	propPath := path.Join(d.Path(), JsonKey.Properties)
	propMap := d.Data[JsonKey.Properties].(map[string]interface{})
	requiredList := make([]string, 0, len(propMap))
	for pname, prop := range propMap {
		propDef, ok := prop.(map[string]interface{})
		if !ok {
			return fmt.Errorf("invalid property value !=[map[string]interface{}], path:[%s/%s]", propPath, pname)
		}
		required, ok := propDef[JsonKey.Required]
		if !ok {
			requiredList = append(requiredList, pname)
			continue
		}
		requiredBool, ok := required.(bool)
		if !ok {
			return fmt.Errorf("invalid data type, [type] != [bool], [path]=[%s/%s/%s]", propPath, pname, JsonKey.Required)
		}
		if requiredBool {
			requiredList = append(requiredList, pname)
		}
		delete(propDef, JsonKey.Required)
	}
	if len(requiredList) > 0 {
		d.Data[JsonKey.Required] = requiredList
	}
	return nil
}

func (d *SchemaDoc) processRefs() error {
	for pname, prop := range d.Data[JsonKey.Properties].(map[string]interface{}) {
		propDef, ok := prop.(map[string]interface{})
		if !ok {
			return fmt.Errorf("failed to parse property [%s], path=[%s]", pname, d.Path())
		}
		success, err := d.processRef(pname, propDef)
		if err != nil {
			return err
		}
		if success {
			continue
		}
		item, ok := propDef[JsonKey.Items]
		if ok {
			itemDef, ok := item.(map[string]interface{})
			if !ok {
				return fmt.Errorf("failed to parse field=[%s], path=[%s/%s]", JsonKey.Items, d.Path(), pname)
			}
			_, err := d.processRef(pname, itemDef)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *SchemaDoc) processRef(pname string, prop map[string]interface{}) (bool, error) {
	ref, err := d.getCmtRef(pname, prop)
	if err != nil {
		return false, err
	}
	if ref != nil {
		d.Refs = append(d.Refs, ref)
		return true, nil
	}
	doc, err := d.getSubDoc(pname, prop)
	if err != nil {
		return false, err
	}
	if doc != nil {
		d.SubDocs[pname] = doc
		return true, nil
	}
	return false, nil
}

func (d *SchemaDoc) getCmtRef(pname string, prop map[string]interface{}) (*SchemaDocRef, error) {
	cmt, ok := prop[JsonKey.ContentMediaType]
	if !ok {
		return nil, nil
	}
	if prop[JsonKey.Type] != JsonKey.String {
		return nil, fmt.Errorf("expect [%s]=[%s], found [%s], [path]=[%s]", JsonKey.Type, JsonKey.String, prop[JsonKey.Type], path.Join(d.Path(), JsonKey.Properties, pname))
	}
	ref := SchemaDocRef{
		Doc:         d,
		Name:        pname,
		ContentType: cmt.(string),
	}
	return &ref, nil
}

func (d *SchemaDoc) getSubDoc(pname string, prop map[string]interface{}) (*SchemaDoc, error) {
	dataType := prop[JsonKey.Type].(string)
	if dataType != JsonKey.Object {
		return nil, nil
	}
	ref, ok := prop[JsonKey.Ref]
	if !ok {
		return nil, fmt.Errorf("object type no schema definition. path=[%s/%s]", d.Path(), pname)
	}
	refVal, ok := ref.(string)
	if !ok {
		return nil, fmt.Errorf("failed to parse ref value. path=[%s/%s/%s]", d.Path(), pname, JsonKey.Ref)
	}
	if !strings.HasPrefix(refVal, JsonKey.DefinitionPrefix) {
		return nil, fmt.Errorf("unknown ref value=[%s], path=[%s/%s/%s]", refVal, d.Path(), pname, JsonKey.Ref)
	}
	docType := refVal[len(JsonKey.DefinitionPrefix):]
	doc, err := d.getDefinition(docType)
	if err != nil {
		return nil, fmt.Errorf("failed to get Definition=[%s], path=[%s/%s/%s], Error:%s", docType, d.Path(), pname, JsonKey.Ref, err)
	}
	if doc == nil {
		return nil, fmt.Errorf("cannot find definition=[%s], path=[%s/%s/%s], no error", docType, d.Path(), pname, JsonKey.Ref)
	}
	return doc, nil
}

func (d *SchemaDoc) getDefinition(dataType string) (*SchemaDoc, error) {
	if d.Definitions != nil {
		doc, ok := d.Definitions[dataType]
		if ok {
			return doc, nil
		}
	}
	if d.Parent != nil {
		doc, err := d.Parent.getDefinition(dataType)
		if err != nil {
			return nil, err
		}
		return doc, nil
	}
	return nil, nil
}
