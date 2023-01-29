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

package CmtIndex

import (
	"fmt"

	"github.com/salesforce/UniTAO/lib/Schema/JsonKey"
	"github.com/salesforce/UniTAO/lib/Schema/Record"
	"github.com/salesforce/UniTAO/lib/Schema/SchemaDoc"
	"github.com/salesforce/UniTAO/lib/Util/Json"
)

const (
	CurrentVersion   = "0.0.1"
	KeyCmtIdx        = "cmtIdx"
	KeyCmtSubscriber = "cmtSubscriber"
	KeyVersionIndex  = "versionIndex"
)

type CmtIndex struct {
	DataType   string                   `json:"dataType"`
	Subscriber map[string]CmtSubscriber `json:"cmtSubscriber"`
}

type CmtSubscriber struct {
	DataType     string                  `json:"dataType"`
	VersionIndex map[string]VersionIndex `json:"versionIndex"`
	// mapping version with indexTemplate
}

type VersionIndex struct {
	Version       string        `json:"version"`
	IndexTemplate []interface{} `json:"indexTemplate"`
}

func LoadMap(data interface{}) (*CmtIndex, error) {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to convert data to map")
	}
	record, err := Record.LoadMap(dataMap)
	if err == nil {
		dataMap = record.Data
	}
	idx := CmtIndex{}
	err = Json.CopyTo(dataMap, &idx)
	if err != nil {
		return nil, fmt.Errorf("failed to load data as Index")
	}
	return &idx, nil
}

func (c *CmtIndex) Map() map[string]interface{} {
	data := map[string]interface{}{}
	Json.CopyTo(c, &data)
	return data
}

func (c *CmtIndex) Record() *Record.Record {
	data := c.Map()
	rec := Record.NewRecord(KeyCmtIdx, CurrentVersion, c.DataType, data)
	return rec
}

func FindAutoIndex(schema *SchemaDoc.SchemaDoc, path string) []*AutoIndex {
	// load schema data into schema doc
	// loop through all path from doc root and find the string attribute with indexTemplate in the attribute definition
	linkList := []*AutoIndex{}
	for attr, def := range schema.Data[JsonKey.Properties].(map[string]interface{}) {
		attrDef := def.(map[string]interface{})
		attrList := getAttrAutoIndex(schema, attr, attrDef, fmt.Sprintf("%s/%s", path, attr))
		linkList = append(linkList, attrList...)
	}
	return linkList
}

func getStrIndex(schema *SchemaDoc.SchemaDoc, attr string, attrDef map[string]interface{}, path string) *AutoIndex {
	cmtRef, ok := schema.CmtRefs[attr]
	if !ok {
		return nil
	}
	idxTemp, ok := attrDef[JsonKey.IndexTemplate]
	if !ok {
		return nil
	}
	idx := AutoIndex{
		AttrPath:      path,
		ContentType:   cmtRef.ContentType,
		IndexTemplate: idxTemp.(string),
	}
	return &idx
}

func getAttrAutoIndex(schema *SchemaDoc.SchemaDoc, attr string, attrDef map[string]interface{}, path string) []*AutoIndex {
	linkList := []*AutoIndex{}
	attrType := attrDef[JsonKey.Type].(string)
	switch attrType {
	case JsonKey.String:
		idx := getStrIndex(schema, attr, attrDef, path)
		if idx != nil {
			linkList = append(linkList, idx)
		}
	case JsonKey.Object:
		if SchemaDoc.IsMap(attrDef) {
			itemDef := attrDef[JsonKey.AdditionalProperties].(map[string]interface{})
			nextList := getItemAutoIndex(schema, attr, itemDef, path)
			linkList = append(linkList, nextList...)

		} else {
			nextSchema := schema.SubDocs[attr]
			nextList := FindAutoIndex(nextSchema, path)
			linkList = append(linkList, nextList...)
		}
	case JsonKey.Array:
		itemDef := attrDef[JsonKey.Items].(map[string]interface{})
		nextList := getItemAutoIndex(schema, attr, itemDef, path)
		linkList = append(linkList, nextList...)
	}
	return linkList
}

func getItemAutoIndex(schema *SchemaDoc.SchemaDoc, attr string, itemDef map[string]interface{}, attrPath string) []*AutoIndex {
	linkList := []*AutoIndex{}
	itemType := itemDef[JsonKey.Type].(string)
	switch itemType {
	case JsonKey.String:
		idx := getStrIndex(schema, attr, itemDef, attrPath)
		if idx != nil {
			linkList = append(linkList, idx)
		}
	case JsonKey.Object:
		nextSchema := schema.SubDocs[attr]
		idxPath := fmt.Sprintf("%s[%s_key]", attrPath, attr)
		nextList := FindAutoIndex(nextSchema, idxPath)
		linkList = append(linkList, nextList...)
	}
	return linkList
}
