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
	"UniTao/Schema/JsonKey"
	"encoding/json"
	"fmt"
	"path"

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

func NewRecordData(dataType string, dataId string, data interface{}) map[string]interface{} {
	record := make(map[string]interface{})
	record[DataId] = dataId
	record[DataType] = dataType
	record[RecordData] = data
	return record
}

type Record struct {
	Id     string
	Raw    string
	Data   map[string]interface{}
	Schema *Doc
	Meta   *jsonschema.Schema
}

func LoadRecord(data map[string]interface{}) (*Record, error) {
	recordBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal doc to string, Err:%s", err)
	}
	record := Record{
		Raw:  string(recordBytes),
		Data: data,
	}
	record.init()
	return &record, nil
}

func (rec *Record) init() error {
	schemaId, ok := rec.Data[DataId].(string)
	if !ok {
		return fmt.Errorf("invalid format for schema, missing field [%s]", DataId)
	}
	rec.Id = schemaId
	schemaDataType, ok := rec.Data[DataType].(string)
	if !ok {
		return fmt.Errorf("invalid format for schema, missing field [%s]", DataType)
	}
	if schemaDataType != Schema {
		return fmt.Errorf("schema record has wrong [%s], [%s]!=[%s]", DataType, schemaDataType, Schema)
	}
	schemaData, ok := rec.Data[RecordData].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid format for schema, missing field [%s]", RecordData)
	}
	doc, err := NewDoc(schemaData, schemaId, rec, nil)
	if err != nil {
		return fmt.Errorf("failed to create Schema Doc, err: %s", err)
	}
	rec.Schema = doc
	schemaBytes, err := json.MarshalIndent(doc.Data, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to MarshalIndent value [field]=[%s], Err:%s", RecordData, err)
	}
	meta, err := jsonschema.CompileString(rec.Id, string(schemaBytes))
	if err != nil {
		return fmt.Errorf("failed to compile schema, [%s]=[%s] Err:%s", DataId, schemaId, err)
	}
	rec.Meta = meta
	return nil
}

func (rec *Record) Validate(payload map[string]interface{}) error {
	payloadType, ok := payload[DataType].(string)
	if !ok {
		return fmt.Errorf("missing field [%s] from payload", DataType)
	}
	if rec.Id != payloadType {
		return fmt.Errorf("schema id and payload data type does not match, [%s]!=[%s]", rec.Id, payloadType)
	}
	payloadData, ok := payload[RecordData]
	if !ok {
		return fmt.Errorf("missing field [%s] from payload", RecordData)
	}

	err := rec.Meta.Validate(payloadData)
	if err != nil {
		return fmt.Errorf("schema validation failed. Error:\n%s", err)
	}
	return nil
}

type Doc struct {
	Id          string
	Record      *Record
	Parent      *Doc
	Raw         string
	Data        map[string]interface{}
	Properties  map[string]interface{}
	Definitions map[string]*Doc
	Refs        []*DocRef
	SubDoc      map[string]*Doc
}

type DocRef struct {
	Doc         *Doc
	Name        string
	ContentType string
}

func NewDoc(data map[string]interface{}, id string, record *Record, parent *Doc) (*Doc, error) {
	rawBytes, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return nil, fmt.Errorf("invalid format, field=[%s] failed Marshal, Error:%s", RecordData, err)
	}
	doc := Doc{
		Id:     id,
		Raw:    string(rawBytes),
		Record: record,
		Parent: parent,
		Data:   data,
		Refs:   []*DocRef{},
		SubDoc: make(map[string]*Doc),
	}
	err = doc.init()
	if err != nil {
		return nil, fmt.Errorf("failed @init, Err:\n%s", err)
	}
	err = doc.preprocess()
	if err != nil {
		return nil, fmt.Errorf("failed @preprocess, Err:\n%s", err)
	}
	return &doc, nil
}

func (d *Doc) Path() string {
	if d.Parent == nil {
		return d.Id
	}
	return path.Join(d.Parent.Path(), d.Id)
}

func (d *Doc) init() error {
	propDef, ok := d.Data[JsonKey.Properties]
	if !ok {
		return nil
	}
	propMap, ok := propDef.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid data value, [path]=[%s/%s]", d.Path(), JsonKey.Properties)
	}
	d.Properties = propMap
	definitions, ok := d.Data[JsonKey.Definitions]
	if !ok {
		return nil
	}
	defMap, ok := definitions.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid data type, [path]=[%s/%s]", d.Path(), JsonKey.Definitions)
	}
	d.Definitions = make(map[string]*Doc)
	for defId, defItem := range defMap {
		defData, ok := defItem.(map[string]interface{})
		if !ok {
			return fmt.Errorf("invalid data type, [path]=[%s/%s/%s]", d.Path(), JsonKey.Definitions, defId)
		}
		defDoc, err := NewDoc(defData, defId, d.Record, d)
		if err != nil {
			return fmt.Errorf("failed to parse definition, [path]=[%s/%s/%s]", d.Path(), JsonKey.Definitions, defId)
		}
		d.Definitions[defId] = defDoc
	}
	for pname, prop := range d.Properties {
		err := d.getSubDoc(pname, prop)
		if err != nil {
			return fmt.Errorf("failed to find get subdoc,[path]=[%s]", path.Join(d.Path(), JsonKey.Properties, pname))
		}
	}
	return nil
}

func (d *Doc) getDefinition(dataType string) (*Doc, error) {
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

func (d *Doc) getSubDoc(pname string, prop interface{}) error {
	dataType := prop.(map[string]interface{})[JsonKey.Type].(string)
	doc, err := d.getDefinition(dataType)
	if err != nil {
		return err
	}
	if doc != nil {
		d.SubDoc[pname] = doc
		return nil
	}
	items, ok := prop.(map[string]interface{})[JsonKey.Items]
	if !ok {
		return nil
	}
	itemName := path.Join(pname, "key")
	return d.getSubDoc(itemName, items)
}

func (d *Doc) preprocess() error {
	err := d.processRequired()
	if err != nil {
		err := fmt.Errorf("preprocess failed @processRequired, [path]=[%s], Error:%s", d.Path(), err)
		return err
	}
	err = d.processInvRefs()
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

func (d *Doc) processRequired() error {
	propPath := path.Join(d.Path(), JsonKey.Properties)
	requiredList := make([]string, 0, len(d.Properties))
	for pname, prop := range d.Properties {
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

func (d *Doc) processInvRefs() error {
	for pname, prop := range d.Properties {
		ref, err := d.getCmtRef(pname, prop)
		if err != nil {
			return err
		}
		if ref != nil {
			d.Refs = append(d.Refs, ref)
			continue
		}
		itemDef, ok := prop.(map[string]interface{})[JsonKey.Items]
		if ok {
			ref, err = d.getCmtRef(pname, itemDef)
			if err != nil {
				return err
			}
			if ref != nil {
				d.Refs = append(d.Refs, ref)
			}
		}
	}
	return nil
}

func (d *Doc) getCmtRef(pname string, prop interface{}) (*DocRef, error) {
	propDef, ok := prop.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid property definition. [path]=[%s]", path.Join(d.Path(), JsonKey.Properties, pname))
	}
	cmt, ok := propDef[JsonKey.ContentMediaType]
	if !ok {
		return nil, nil
	}
	if propDef[JsonKey.Type] != JsonKey.String {
		return nil, fmt.Errorf("expect [%s]=[%s], found [%s], [path]=[%s]", JsonKey.Type, JsonKey.String, propDef[JsonKey.Type], path.Join(d.Path(), JsonKey.Properties, pname))
	}
	ref := DocRef{
		Doc:         d,
		Name:        pname,
		ContentType: cmt.(string),
	}
	return &ref, nil
}
