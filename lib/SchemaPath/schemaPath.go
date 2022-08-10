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
	"net/http"
	"reflect"
	"strings"

	"github.com/salesforce/UniTAO/lib/Schema/JsonKey"
	"github.com/salesforce/UniTAO/lib/Schema/SchemaDoc"
	"github.com/salesforce/UniTAO/lib/Util"
)

const (
	All          = "*"
	Ref          = "$"
	CmdPrefix    = "?"
	CmdRef       = "?ref"
	CmdSchema    = "?schema"
	CmdFlat      = "?flat"
	SchemaFakeId = "#SchemaFakeId#"
)

type SchemaPath struct {
	Conn     *Connection
	Schema   *SchemaDoc.SchemaDoc
	FullPath string
	NextPath string
	PathCmd  string
	Prev     *SchemaPath
	Data     map[string]interface{}
}

type SchemaPathErr struct {
	Code    int
	PathErr error
}

func (e *SchemaPathErr) Error() string {
	return fmt.Sprintf("Code=[%d], Error: %s", e.Code, e.PathErr)
}

func New(conn *Connection, schema *SchemaDoc.SchemaDoc, fullPath string, nextPath string, prev *SchemaPath, data map[string]interface{}) (*SchemaPath, error) {
	nextPath, pathCmd, err := ParsePathCmd(nextPath)
	if err != nil {
		return nil, err
	}
	if pathCmd == "" && prev != nil {
		pathCmd = prev.PathCmd
	}
	schemaPath := SchemaPath{
		Conn:     conn,
		Schema:   schema,
		FullPath: fullPath,
		NextPath: nextPath,
		PathCmd:  pathCmd,
		Prev:     prev,
		Data:     data,
	}
	return &schemaPath, nil
}

func AppendErr(err error, newMsg string) error {
	newErr := fmt.Errorf("%s Error: %s", newMsg, err)
	pathErr, ok := err.(*SchemaPathErr)
	if ok {
		newErr = &SchemaPathErr{
			Code:    pathErr.Code,
			PathErr: newErr,
		}
	}
	return newErr
}

func NewFromPath(conn *Connection, path string, prev *SchemaPath) (*SchemaPath, error) {
	fullPath := ""
	if prev != nil {
		fullPath = prev.FullPath
	}
	return New(conn, nil, fullPath, path, prev, nil)
}

func ParseArrayPath(path string) (string, string, error) {
	if path[len(path)-1:] != "]" {
		return path, "", nil
	}
	keyIdx := strings.Index(path, "[")
	if keyIdx < 1 {
		return path, "", nil
	}
	attrName := path[:keyIdx]
	key := path[keyIdx+1 : len(path)-1]
	if key == "" {
		return "", "", fmt.Errorf("invalid array path=[%s], key empty", path)
	}
	return attrName, key, nil
}

func ParsePathCmd(path string) (string, string, error) {
	qIdx := strings.Index(path, CmdPrefix)
	if qIdx < 0 {
		return path, "", nil
	}
	qPath := path[:qIdx]
	qCmd := path[qIdx:]
	dupIdx := strings.Index(qCmd[1:], CmdPrefix)
	if dupIdx > -1 {
		return "", "", fmt.Errorf("invalid format of PathCmd, more than 1 ? in path. path=[%s]", path)
	}
	err := ValidatePathCmd(qCmd)
	if err != nil {
		return "", "", fmt.Errorf("path command validate failed. Error: %s, @path=[%s]", err, path)
	}
	return qPath, qCmd, nil
}

func ValidatePathCmd(cmd string) error {
	for _, c := range []string{CmdRef, CmdFlat, CmdSchema} {
		if c == cmd {
			return nil
		}
	}
	return fmt.Errorf("unknown path cmd=[%s]", cmd)
}

func (p *SchemaPath) SyncCurrentData() error {
	if p.Schema == nil {
		dataType, nextPath := Util.ParsePath(p.NextPath)
		schema, err := p.Conn.GetSchema(dataType)
		if err != nil {
			return AppendErr(err, fmt.Sprintf("failed to get schema, type=[%s], @path=[%s]", dataType, p.FullPath))
		}
		p.Schema = schema
		p.FullPath = fmt.Sprintf("%s/%s", p.FullPath, dataType)
		p.NextPath = nextPath
	}
	if p.Data == nil {
		dataId, nextPath := Util.ParsePath(p.NextPath)
		if dataId == "" {
			return &SchemaPathErr{
				Code:    http.StatusBadRequest,
				PathErr: fmt.Errorf("missing data Id in path @[%s]", p.FullPath),
			}
		}
		if p.PathCmd != CmdSchema {
			record, err := p.Conn.GetRecord(p.Schema.Id, dataId)
			if err != nil {
				return AppendErr(err, fmt.Sprintf("failed get record=[%s/%s], @path=[%s]", p.Schema.Id, dataId, p.FullPath))
			}
			p.Data = record.Data
		}
		p.FullPath = fmt.Sprintf("%s/%s", p.FullPath, dataId)
		p.NextPath = nextPath
	}
	return nil
}

func (p *SchemaPath) FlatValue() (interface{}, error) {
	flat := map[string]interface{}{}
	for attrName, val := range p.Data {
		attrDef, attrDefined := p.Schema.Data[JsonKey.Properties].(map[string]interface{})[attrName].(map[string]interface{})
		if val == nil {
			if attrDefined && attrDef[JsonKey.Type].(string) == JsonKey.Array {
				flat[attrName] = []interface{}{}
			}
			continue
		}
		if !attrDefined {
			if reflect.TypeOf(val).Kind() == reflect.Slice {
				return val, nil
			}
		}
		attrType := attrDef[JsonKey.Type].(string)
		switch attrType {
		case JsonKey.Array:
			itemDef := attrDef[JsonKey.Items].(map[string]interface{})
			if itemDef[JsonKey.Type] != JsonKey.Object {
				flat[attrName] = val
				continue
			}
			flatVal, err := p.WalkArray(attrName, "", attrDef, val, "")
			if err != nil {
				return nil, err
			}
			flat[attrName] = flatVal
		case JsonKey.Object:
			keyMap := val.(map[string]interface{})
			keyList := make([]string, 0, len(keyMap))
			for key, _ := range keyMap {
				keyList = append(keyList, key)
			}
			flat[attrName] = keyList
		default:
			flat[attrName] = val
		}
	}
	return flat, nil
}

func (p *SchemaPath) WalkValue() (interface{}, error) {
	err := p.SyncCurrentData()
	if err != nil {
		return nil, err
	}
	// no more path to walk. return current value
	if p.NextPath == "" {
		switch p.PathCmd {
		case CmdSchema:
			return p.Schema.RAW, nil
		case CmdFlat:
			return p.FlatValue()
		default:
			return p.Data, nil
		}
	}
	attrPath, nextPath := Util.ParsePath(p.NextPath)
	attrName, attrIdx, err := ParseArrayPath(attrPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ArrayPath @[path]=[%s]", p.FullPath)
	}
	attrDef, ok := p.Schema.Data[JsonKey.Properties].(map[string]interface{})[attrName].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("attr=[%s] is not a defined attribute. @[path]=[%s/%s]", attrName, p.FullPath, attrPath)
	}
	var attrValue interface{}
	if p.PathCmd != CmdSchema {
		attrValue, ok = p.Data[attrName]
		if !ok || attrValue == nil {
			log.Printf("data does not exists or is nil. @[path]=[%s/%s]", p.FullPath, attrName)
			return nil, nil
		}
	}
	attrType := attrDef[JsonKey.Type].(string)
	if attrIdx != "" && attrType != JsonKey.Array {
		return nil, fmt.Errorf("only [%s] support idx path, attr [%s] type=[%s], @[path]=[%s]", JsonKey.Array, attrName, attrType, p.FullPath)
	}
	switch attrType {
	case JsonKey.Array:
		log.Printf("walk into array")
		return p.WalkArray(attrName, attrIdx, attrDef, attrValue, nextPath)
	case JsonKey.Object:
		log.Printf("walk into object @[path]=[%s/%s]", p.FullPath, attrName)
		if SchemaDoc.IsMap(attrDef) {
			return p.WalkMap(attrName, attrDef, attrValue, nextPath)
		}
		return p.WalkObject(attrName, attrValue, nextPath)
	case JsonKey.String:
		log.Printf("walk into CMT ref")
		if nextPath != "" || p.PathCmd != CmdFlat {
			attrValue, err = p.WalkCMTIdx(attrName, attrValue.(string), nextPath)
			if err != nil {
				return nil, fmt.Errorf("failed to get CMDIdx. [%s]=[%s], @[path]=[%s]", attrName, attrValue, p.FullPath)
			}
		}
		return attrValue, nil
	default:
		if nextPath != "" {
			return nil, fmt.Errorf("[type]=[%s] does not support walk further, @[path]=[%s/%s]", attrType, p.FullPath, attrName)
		}
		switch p.PathCmd {
		case CmdSchema:
			return p.Schema.RAW[JsonKey.Properties].(map[string]interface{})[attrName], nil
		case CmdRef:
			return nil, fmt.Errorf("[type]=[%s] does not support ref work, @[path]=[%s/%s]", attrType, p.FullPath, attrName)
		default:
			return attrValue, nil
		}
	}
}

func (p *SchemaPath) WalkArray(attrName string, attrIdx string, attrDef map[string]interface{}, attrValue interface{}, nextPath string) (interface{}, error) {
	itemDef := attrDef[JsonKey.Items].(map[string]interface{})
	if p.PathCmd == CmdSchema {
		// path is arrayAttrOnly
		if nextPath == "" {
			if attrIdx == "" {
				return attrDef, nil
			}
			return itemDef, nil
		}
		attrValue = []interface{}{}
	}
	itemList, ok := attrValue.([]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to convert value to []string")
	}
	itemType := itemDef[JsonKey.Type].(string)
	if p.PathCmd == CmdSchema && attrIdx == "" {
		return p.Schema.RAW[JsonKey.Properties].(map[string]interface{})[attrName], nil
	}
	valueList := []interface{}{}
	switch itemType {
	case JsonKey.Object:
		itemDoc, ok := p.Schema.SubDocs[attrName]
		if !ok {
			if p.PathCmd == CmdSchema {
				return nil, &SchemaPathErr{
					Code:    http.StatusBadRequest,
					PathErr: fmt.Errorf("schema attr=[%s] has missing ref field=[%s]. path=[%s]", attrName, JsonKey.Ref, p.FullPath),
				}
			}
			if nextPath == "" && attrIdx == "" {
				return attrValue, nil
			}
			return nil, fmt.Errorf("failed to load item object schema. @[path]=[%s/%s]", p.FullPath, attrName)
		}
		if attrIdx != "" && itemDoc.KeyTemplate == "" {
			return nil, fmt.Errorf(`
				attr=[%s], type=[%s] has no template definition [%s] in schema, 
				index no key object not supported. @[path]=[%s/%s[%s]]`,
				attrName, itemDoc.Id, JsonKey.Key,
				p.FullPath, attrName, attrIdx)
		}
		if p.PathCmd == CmdSchema {
			item := map[string]interface{}{}
			itemValue, err := p.WalkObject(attrName, item, nextPath)
			if err != nil {
				return nil, err
			}
			return itemValue, nil
		}
		for _, item := range itemList {
			itemKey, err := itemDoc.BuildKey(item.(map[string]interface{}))
			if err != nil {
				return nil, err
			}
			if attrIdx != "" && attrIdx != itemKey && attrIdx != All {
				continue
			}
			if attrIdx == "" && p.PathCmd == CmdFlat {
				valueList = append(valueList, itemKey)
				continue
			}
			itemValue, err := p.WalkObject(attrName, item, nextPath)
			if err != nil {
				if attrIdx == All {
					// if failed to walk next path, skip when idx=*
					continue
				}
				return nil, err
			}
			if itemKey == attrIdx {
				return itemValue, nil
			}
			if attrIdx == All {
				if itemValue == nil {
					// if return value is nil, skip when idx=*
					continue
				}
			}
			valueList = append(valueList, itemValue)
		}
		if attrIdx != "" && attrIdx != All {
			return nil, fmt.Errorf("data not exists. @[path]=[%s/%s[%s]", p.FullPath, attrName, attrIdx)
		}
		if len(valueList) == 0 && attrIdx == All && len(itemList) > 0 {
			// only return empty array, when the array is really empty
			return nil, fmt.Errorf("no existing path match query path @[path]=[%s/%s[%s], nextPath=[%s]", p.FullPath, attrName, attrIdx, nextPath)
		}
		return valueList, nil
	case JsonKey.String:
		if p.PathCmd == CmdSchema {
			cmtValue, err := p.WalkCMTIdx(attrName, SchemaFakeId, nextPath)
			if err != nil {
				return nil, err
			}
			return cmtValue, nil
		}
		if nextPath == "" && attrIdx == "" && p.PathCmd == CmdFlat {
			return attrValue, nil
		}
		for _, item := range attrValue.([]interface{}) {
			key := item.(string)
			if attrIdx != "" && attrIdx != key && attrIdx != All {
				continue
			}
			cmtValue, err := p.WalkCMTIdx(attrName, key, nextPath)
			if err != nil {
				if attrIdx == All {
					// if failed to walk, then keep going to next
					continue
				}
				return nil, err
			}
			if key == attrIdx {
				return cmtValue, nil
			}
			valueList = append(valueList, cmtValue)
		}
		if attrIdx != "" && attrIdx != All {
			return nil, fmt.Errorf("data not exists. @[path]=[%s/%s[%s]", p.FullPath, attrName, attrIdx)
		}
		return valueList, nil
	default:
		if attrIdx != "" {
			return nil, fmt.Errorf("type []%s does not support idx, @[path]=[%s/%s]", itemType, p.FullPath, attrName)
		}
		if nextPath != "" {
			return nil, fmt.Errorf("type []%s does not support walk in, @[path]=[%s/%s]", itemType, p.FullPath, attrName)
		}
		return attrValue, nil
	}
}

func (p *SchemaPath) WalkCMTIdx(attrName string, attrValue string, nextPath string) (interface{}, error) {
	cmtRef, ok := p.Schema.CmtRefs[attrName]
	if !ok {
		log.Printf("[attr]=[%s] has no Cmt Ref @[path]=[%s]", attrName, p.FullPath)
		if p.PathCmd == CmdSchema {
			return p.Schema.RAW[JsonKey.Properties].(map[string]interface{})[attrName], nil
		}
		return attrValue, nil
	}
	// if query path ends with /.
	// then return current CMT reference value
	if nextPath == "$" || (nextPath == "" && p.PathCmd == CmdRef) {
		// if no further path to walk
		return attrValue, nil
	}
	// otherwise we see the CMT as direct to the real object
	cmtPath := fmt.Sprintf("%s/%s", cmtRef.ContentType, attrValue)
	if nextPath != JsonKey.DocRoot {
		cmtPath = fmt.Sprintf("%s/%s", cmtPath, nextPath)
	}
	cmt, err := NewFromPath(p.Conn, cmtPath, p)
	if err != nil {
		return nil, err
	}
	cmtValue, err := cmt.WalkValue()
	if err != nil {
		return nil, err
	}
	return cmtValue, nil
}

func (p *SchemaPath) WalkMap(attrName string, attrDef map[string]interface{}, attrValue interface{}, nextPath string) (interface{}, error) {
	mapDef := attrDef[JsonKey.AdditionalProperties].(map[string]interface{})
	if nextPath == "" {
		if p.PathCmd == CmdSchema {
			return attrDef, nil
		}
		return attrValue, nil
	}
	itemKey, nextPath := Util.ParsePath(nextPath)
	if itemKey == All {
		itemList := []interface{}{}
		for key := range attrValue.(map[string]interface{}) {
			itemPath := fmt.Sprintf("%s/%s", key, nextPath)
			item, err := p.WalkMap(attrName, attrDef, attrValue, itemPath)
			if err == nil {
				itemList = append(itemList, item)
			}
		}
		return itemList, nil
	}
	itemValue, ok := attrValue.(map[string]interface{})[itemKey]
	if !ok || itemValue == nil {
		if p.PathCmd != CmdSchema {
			log.Printf("map [itemKey]=[%s] does not exists. @[path]=[%s/%s]", itemKey, p.FullPath, attrName)
			return nil, nil
		}
		itemValue = nil
	}
	switch mapDef[JsonKey.Type].(string) {
	case JsonKey.String:
		if itemValue == nil {
			itemValue = SchemaFakeId
		}
		cmtValue, err := p.WalkCMTIdx(attrName, itemValue.(string), nextPath)
		if err != nil {
			return nil, err
		}
		return cmtValue, nil
	case JsonKey.Object:
		if itemValue == nil {
			itemValue = make(map[string]interface{})
		}
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
	itemDoc, ok := p.Schema.SubDocs[attrName]
	if !ok {
		return nil, fmt.Errorf("failed to get schema for @[path]=[%s/%s]", p.FullPath, attrName)
	}
	fullPath := fmt.Sprintf("%s/%s", p.FullPath, attrName)
	data := attrValue.(map[string]interface{})
	objPath, err := New(p.Conn, itemDoc, fullPath, nextPath, p, data)
	if err != nil {
		return nil, err
	}
	objValue, err := objPath.WalkValue()
	if err != nil {
		return nil, err
	}
	return objValue, nil
}
