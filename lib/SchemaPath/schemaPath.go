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

package SchemaPath

import (
	"fmt"
	"log"
	"strings"

	"github.com/salesforce/UniTAO/lib/Schema/JsonKey"
	"github.com/salesforce/UniTAO/lib/Schema/SchemaDoc"
	"github.com/salesforce/UniTAO/lib/Util"
)

const (
	Ref       = "$"
	CmdPrefix = "?"
	CmdRef    = "?ref"
	CmdSchema = "?schema"
	CmdFlat   = "?flat"
)

type SchemaPath struct {
	conn     *Connection
	schema   *SchemaDoc.SchemaDoc
	fullPath string
	nextPath string
	prev     *SchemaPath
	data     map[string]interface{}
}

func New(conn *Connection, schema *SchemaDoc.SchemaDoc, fullPath string, nextPath string, prev *SchemaPath, data map[string]interface{}) *SchemaPath {
	schemaPath := SchemaPath{
		conn:     conn,
		schema:   schema,
		fullPath: fullPath,
		nextPath: nextPath,
		prev:     prev,
		data:     data,
	}
	return &schemaPath
}

func NewFromPath(conn *Connection, path string, prev *SchemaPath) (*SchemaPath, error) {
	fullPath := ""
	if prev != nil {
		fullPath = prev.fullPath
	}
	dataType, nextPath := Util.ParsePath(path)
	if nextPath == "" {
		return nil, fmt.Errorf("missing id from path=[%s/%s]", fullPath, dataType)
	}
	schema, err := conn.GetSchema(dataType)
	if err != nil {
		return nil, fmt.Errorf("failed to get schema [type]=[%s] @[path]=[%s]", dataType, fullPath)
	}
	fullPath = fmt.Sprintf("%s/%s", fullPath, dataType)

	dataId, nextPath := Util.ParsePath(nextPath)
	dataId, pathCmd, err := ParsePathCmd(dataId)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Path Cmd on path=[%s], Error: %s", path, err)
	}
	nextPath, err = ValidatePathCmd(nextPath, pathCmd)
	if err != nil {
		return nil, fmt.Errorf("invalid PathCmd format path=[%s], Error: %s", path, err)
	}
	record, err := conn.GetRecord(dataType, dataId)
	if err != nil {
		return nil, fmt.Errorf("failed to get record [type/id]=[%s/%s] @[path]=[%s]", dataType, dataId, fullPath)
	}
	fullPath = fmt.Sprintf("%s/%s", fullPath, dataId)

	return New(conn, schema, fullPath, nextPath, prev, record.Data), nil
}

func ParseArrayPath(path string) (string, string) {
	if path[len(path)-1:] != "]" {
		return path, ""
	}
	keyIdx := strings.Index(path, "[")
	if keyIdx < 1 {
		return path, ""
	}
	attrName := path[:keyIdx]
	key := path[keyIdx+1 : len(path)-1]
	return attrName, key
}

func ParsePathCmd(path string) (string, string, error) {
	qIdx := strings.Index(path, CmdPrefix)
	if qIdx < 0 {
		return path, "", nil
	}
	attrName := path[:qIdx]
	qPath := path[qIdx+1:]
	dupIdx := strings.Index(qPath[1:], CmdPrefix)
	if dupIdx > -1 {
		return "", "", fmt.Errorf("invalid format of PathCmd, more than 1 ? in path. path=[%s]", path)
	}
	return attrName, fmt.Sprintf("%s%s", CmdPrefix, qPath), nil
}

func ValidatePathCmd(nextPath string, cmd string) (string, error) {
	if nextPath != "" && cmd != "" {
		return "", fmt.Errorf("invalid path=[%s/%s],Path Command has to be end of path as query", cmd, nextPath)
	}
	if cmd == "" {
		return nextPath, nil
	}
	for _, c := range []string{CmdRef, CmdFlat, CmdSchema} {
		if c == cmd {
			return cmd, nil
		}
	}
	return "", fmt.Errorf("unknown path cmd=[%s]", cmd)
}

func (p *SchemaPath) WalkValue() (interface{}, error) {
	// no more path to walk. return current value
	if p.nextPath == "" {
		return p.data, nil
	}
	if p.nextPath[:1] == CmdPrefix {
		switch p.nextPath {
		case CmdSchema:
			return p.schema.RAW, nil
		default:
			return nil, fmt.Errorf("unknown Path command: %s", p.nextPath)
		}
	}
	attrPath, nextPath := Util.ParsePath(p.nextPath)
	attrPath, pathCmd, err := ParsePathCmd(attrPath)
	if err != nil {
		return nil, err
	}
	nextPath, err = ValidatePathCmd(nextPath, pathCmd)
	if err != nil {
		return nil, err
	}

	attrName, attrIdx := ParseArrayPath(attrPath)
	attrDef, ok := p.schema.Data[JsonKey.Properties].(map[string]interface{})[attrName].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("attr=[%s] is not a defined attribute. @[path]=[%s/%s]", attrName, p.fullPath, attrPath)
	}
	attrValue, ok := p.data[attrName]
	if !ok || attrValue == nil {
		log.Printf("data does not exists or is nil. @[path]=[%s/%s]", p.fullPath, attrName)
		return nil, nil
	}
	attrType := attrDef[JsonKey.Type].(string)
	if attrIdx != "" && attrType != JsonKey.Array {
		return nil, fmt.Errorf("only [%s] support idx path, attr [%s] type=[%s], @[path]=[%s]", JsonKey.Array, attrName, attrType, p.fullPath)
	}
	switch attrType {
	case JsonKey.Array:
		log.Printf("walk into array")
		return p.WalkArray(attrName, attrIdx, attrDef[JsonKey.Items].(map[string]interface{}), attrValue, nextPath)
	case JsonKey.Object:
		log.Printf("walk into object @[path]=[%s/%s]", p.fullPath, attrName)
		if SchemaDoc.IsMap(attrDef) {
			return p.WalkMap(attrName, attrDef[JsonKey.AdditionalProperties].(map[string]interface{}), attrValue, nextPath)
		}
		return p.WalkObject(attrName, attrValue, nextPath)
	case JsonKey.String:
		log.Printf("walk into CMT ref")
		attrValue, err := p.WalkCMTIdx(attrName, attrValue.(string), nextPath)
		if err != nil {
			return nil, fmt.Errorf("failed to get CMDIdx. [%s]=[%s], @[path]=[%s]", attrName, attrValue, p.fullPath)
		}
		return attrValue, nil
	default:
		if nextPath != "" {
			return nil, fmt.Errorf("[type]=[%s] does not support ref work, @[path]=[%s/%s]", attrType, p.fullPath, attrName)
		}
		return attrValue, nil
	}
}

func (p *SchemaPath) WalkArray(attrName string, attrIdx string, attrDef map[string]interface{}, attrValue interface{}, nextPath string) (interface{}, error) {
	itemType := attrDef[JsonKey.Type].(string)
	valueList := []interface{}{}
	switch itemType {
	case JsonKey.Object:
		itemDoc, ok := p.schema.SubDocs[attrName]
		if !ok {
			if nextPath == "" && attrIdx == "" {
				return attrValue, nil
			}
			return nil, fmt.Errorf("failed to load item object schema. @[path]=[%s/%s]", p.fullPath, attrName)
		}
		if attrIdx != "" && itemDoc.KeyTemplate == "" {
			return nil, fmt.Errorf(`
				attr=[%s], type=[%s] has no template definition [%s] in schema, 
				index no key object not supported. @[path]=[%s/%s[%s]]`,
				attrName, itemDoc.Id, JsonKey.Key,
				p.fullPath, attrName, attrIdx)
		}
		for _, item := range attrValue.([]interface{}) {
			itemKey := itemDoc.BuildKey(item.(map[string]interface{}))
			if attrIdx != "" && attrIdx != itemKey {
				continue
			}
			itemValue, err := p.WalkObject(attrName, item, nextPath)
			if err != nil {
				return nil, err
			}
			if itemKey == attrIdx {
				return itemValue, nil
			}
			valueList = append(valueList, itemValue)
		}
		return valueList, nil
	case JsonKey.String:
		for _, item := range attrValue.([]interface{}) {
			key := item.(string)
			if attrIdx != "" && attrIdx != key {
				continue
			}
			cmtValue, err := p.WalkCMTIdx(attrName, key, nextPath)
			if err != nil {
				return nil, err
			}
			if key == attrIdx {
				return cmtValue, nil
			}
			valueList = append(valueList, cmtValue)
		}
		return valueList, nil
	default:
		if attrIdx != "" {
			return nil, fmt.Errorf("type []%s does not support idx, @[path]=[%s/%s]", itemType, p.fullPath, attrName)
		}
		if nextPath != "" {
			return nil, fmt.Errorf("type []%s does not support walk in, @[path]=[%s/%s]", itemType, p.fullPath, attrName)
		}
		return attrValue, nil
	}
}

func (p *SchemaPath) WalkCMTIdx(attrName string, attrValue string, nextPath string) (interface{}, error) {
	cmtRef, ok := p.schema.CmtRefs[attrName]
	if !ok {
		log.Printf("[attr]=[%s] has no Cmt Ref @[path]=[%s]", attrName, p.fullPath)
		return attrValue, nil
	}
	// if query path ends with /.
	// then return current CMT reference value
	if nextPath == "$" || nextPath == CmdRef {
		// if no further path to walk
		return attrValue, nil
	}
	// otherwise we see the CMT as direct to the real object
	cmtPath := fmt.Sprintf("%s/%s", cmtRef.ContentType, attrValue)
	if nextPath != JsonKey.DocRoot {
		cmtPath = fmt.Sprintf("%s/%s", cmtPath, nextPath)
	}
	cmt, err := NewFromPath(p.conn, cmtPath, p)
	if err != nil {
		return nil, err
	}
	cmtValue, err := cmt.WalkValue()
	if err != nil {
		return nil, err
	}
	return cmtValue, nil
}

func (p *SchemaPath) WalkMap(attrName string, mapDef map[string]interface{}, attrValue interface{}, nextPath string) (interface{}, error) {
	if nextPath == "" {
		return attrValue, nil
	}
	itemKey, nextPath := Util.ParsePath(nextPath)
	itemValue, ok := attrValue.(map[string]interface{})[itemKey]
	if !ok {
		log.Printf("map [itemKey]=[%s] does not exists. @[path]=[%s/%s]", itemKey, p.fullPath, attrName)
		return nil, nil
	}
	switch mapDef[JsonKey.Type].(string) {
	case JsonKey.String:
		cmtValue, err := p.WalkCMTIdx(attrName, itemValue.(string), nextPath)
		if err != nil {
			return nil, err
		}
		return cmtValue, nil
	case JsonKey.Object:
		objValue, err := p.WalkObject(attrName, itemValue.(map[string]interface{}), nextPath)
		if err != nil {
			return nil, err
		}
		return objValue, nil
	}
	return itemValue, nil
}

func (p *SchemaPath) WalkObject(attrName string, attrValue interface{}, nextPath string) (interface{}, error) {
	if nextPath == "" {
		return attrValue, nil
	}
	itemDoc, ok := p.schema.SubDocs[attrName]
	if !ok {
		return nil, fmt.Errorf("failed to get schema for @[path]=[%s/%s]", p.fullPath, attrName)
	}
	fullPath := fmt.Sprintf("%s/%s", p.fullPath, attrName)
	data := attrValue.(map[string]interface{})
	objPath := New(p.conn, itemDoc, fullPath, nextPath, p, data)
	objValue, err := objPath.WalkValue()
	if err != nil {
		return nil, err
	}
	return objValue, nil
}
