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

package Node

import (
	"fmt"
	"net/http"

	"github.com/salesforce/UniTAO/lib/Schema/JsonKey"
	"github.com/salesforce/UniTAO/lib/Schema/SchemaDoc"
	"github.com/salesforce/UniTAO/lib/SchemaPath/Data"
	"github.com/salesforce/UniTAO/lib/SchemaPath/Error"
	"github.com/salesforce/UniTAO/lib/Util"
)

type PathNode struct {
	Conn     *Data.Connection
	Schema   *SchemaDoc.SchemaDoc
	DataType string
	DataId   string
	fullPath string
	NextPath string
	AttrName string
	AttrDef  map[string]interface{}
	Idx      string
	Prev     *PathNode
	Next     *PathNode
}

func ParseRecordId(dataPath string) (string, string, string, error) {
	if dataPath == "" {
		return "", "", "", nil
	}
	dataType, nextPath := Util.ParsePath(dataPath)
	if nextPath == "" {
		return dataType, "", "", &Error.SchemaPathErr{
			Code:    http.StatusBadRequest,
			PathErr: fmt.Errorf("missing data id in path. dataType=[%s]", dataType),
		}
	}
	dataId, nextPath := Util.ParsePath(nextPath)
	return dataType, dataId, nextPath, nil

}

func New(conn *Data.Connection, dataType string, dataId string, nextPath string, prev *PathNode, schema *SchemaDoc.SchemaDoc) (*PathNode, *Error.SchemaPathErr) {
	fullPath := ""
	if prev != nil {
		fullPath = prev.FullPath()
	}
	node := PathNode{
		Conn:     conn,
		Schema:   schema,
		fullPath: fullPath,
		NextPath: nextPath,
		Prev:     prev,
		DataType: dataType,
		DataId:   dataId,
	}
	err := node.SyncSchema()
	if err != nil {
		return nil, err
	}
	err = node.BuildPath()
	if err != nil {
		return nil, err
	}
	return &node, nil
}

func (p *PathNode) FullPath() string {
	if p.DataType == "" {
		return p.fullPath
	}
	idStr := "{need dataId}"
	if p.DataId != "" {
		idStr = p.DataId
	}
	fullPath := fmt.Sprintf("%s/%s/%s", p.fullPath, p.DataType, idStr)
	return fullPath
}

func (p *PathNode) SyncSchema() *Error.SchemaPathErr {
	if p.DataType != "" {
		schema, err := p.Conn.GetSchema(p.DataType)
		if err != nil {
			return Error.AppendErr(err, fmt.Sprintf("failed to get schema, @path=[%s]", p.FullPath()))
		}
		p.Schema = schema
		if p.DataId == "" && p.Prev == nil {
			return &Error.SchemaPathErr{
				Code:    http.StatusBadRequest,
				PathErr: fmt.Errorf("missing data Id in path @[%s], next=[%s]", p.FullPath(), p.NextPath),
			}
		}
		return nil
	}
	if p.Prev == nil {
		return &Error.SchemaPathErr{
			Code:    http.StatusInternalServerError,
			PathErr: fmt.Errorf("root node should container dataType and dataId"),
		}
	}
	return nil
}

func (p *PathNode) GetRecordData() (interface{}, *Error.SchemaPathErr) {
	if p.DataType != "" {
		if p.DataId != "" {
			record, err := p.Conn.GetRecord(p.DataType, p.DataId)
			if err != nil {
				return nil, Error.AppendErr(err, fmt.Sprintf("failed get record, @path=[%s]", p.FullPath()))
			}
			return record.Data, nil
		}
		return nil, &Error.SchemaPathErr{
			Code:    http.StatusInternalServerError,
			PathErr: fmt.Errorf("please set CMT dataId from Prev before calling GetRecordData"),
		}
	}
	return nil, nil
}

// build path node chain with only schema link to validate path correctness
func (p *PathNode) BuildPath() *Error.SchemaPathErr {
	if p.NextPath == "" {
		return nil
	}
	if p.Prev != nil {
		prevType := p.Prev.AttrDef[JsonKey.Type].(string)
		if prevType == JsonKey.Array || (prevType == JsonKey.Object && SchemaDoc.IsMap(p.Prev.AttrDef)) {
			if p.Prev.Idx == "" {
				return &Error.SchemaPathErr{
					Code:    http.StatusBadRequest,
					PathErr: fmt.Errorf("invalid path. need to specify key to walk into array/map. @path=[%s]", p.Prev.FullPath()),
				}
			}
		}
	}
	attrPath, nextPath := Util.ParsePath(p.NextPath)
	attrName, attrIdx, err := Util.ParseArrayPath(attrPath)
	if err != nil {
		return Error.AppendErr(err, fmt.Sprintf("failed to parse ArrayPath @[path]=[%s]", p.FullPath()))
	}
	if attrIdx != "" {
		p.Idx = attrIdx
	}
	p.AttrName = attrName
	iAttrDef, ok := p.Schema.Data[JsonKey.Properties].(map[string]interface{})[attrName]
	if !ok {
		if nextPath != "" {
			return &Error.SchemaPathErr{
				Code:    http.StatusBadRequest,
				PathErr: fmt.Errorf("attr=[%s] is not a defined attribute. cannot walk any further @[path]=[%s] nextPath=[%s]", attrName, p.FullPath(), nextPath),
			}
		}
		return nil
	}
	attrDef := iAttrDef.(map[string]interface{})
	p.AttrDef = attrDef
	attrType := attrDef[JsonKey.Type].(string)
	if attrIdx != "" && !(attrType == JsonKey.Array || (attrType == JsonKey.Object && SchemaDoc.IsMap(attrDef))) {
		return &Error.SchemaPathErr{
			Code:    http.StatusBadRequest,
			PathErr: fmt.Errorf("only [%s, %s] support idx path, attr [%s] type=[%s], @[path]=[%s]", JsonKey.Array, JsonKey.Map, attrName, attrType, p.FullPath()),
		}
	}
	if attrIdx == "" && attrType == JsonKey.Object && SchemaDoc.IsMap(attrDef) {
		attrIdx, nextPath = Util.ParsePath(nextPath)
		p.Idx = attrIdx
	}
	// even nextPath == "", we still want schema for the next level.
	return p.buildNextPathType(p.AttrDef, nextPath)
}

func (p *PathNode) buildNextPathType(valueDef map[string]interface{}, nextPath string) *Error.SchemaPathErr {
	valueType := valueDef[JsonKey.Type].(string)
	switch valueType {
	case JsonKey.Array:
		return p.buildNextPathArray(nextPath)
	case JsonKey.Object:
		if SchemaDoc.IsMap(valueDef) {
			return p.buildNextPathMap(nextPath)
		}
		return p.buildNextPathObj(nextPath)
	case JsonKey.String:
		return p.buildNextPathCmt(nextPath)
	default:
		if nextPath == "" {
			return nil
		}
		return &Error.SchemaPathErr{
			Code:    http.StatusBadRequest,
			PathErr: fmt.Errorf("attr=[%s], type=[%s] does not support walk further. @path=[%s]", p.AttrName, valueType, p.FullPath()),
		}
	}
}

func (p *PathNode) buildNextPathArray(nextPath string) *Error.SchemaPathErr {
	itemDef, ok := p.AttrDef[JsonKey.Items].(map[string]interface{})
	if !ok {
		return &Error.SchemaPathErr{
			Code:    http.StatusBadRequest,
			PathErr: fmt.Errorf("missing field=[%s] in schema, for attr=[%s], @path=[%s]", JsonKey.Items, p.AttrName, p.FullPath()),
		}
	}
	itemType := itemDef[JsonKey.Type].(string)
	if itemType == JsonKey.Array || (itemType == JsonKey.Object && SchemaDoc.IsMap(itemDef)) {
		return &Error.SchemaPathErr{
			Code:    http.StatusBadRequest,
			PathErr: fmt.Errorf("walk on [%s,%s] in [%s] is not supported. attr=[%s], @path=[%s]", JsonKey.Array, JsonKey.Map, JsonKey.Array, p.AttrName, p.FullPath()),
		}
	}
	if p.Idx == "" {
		if nextPath != "" {
			return &Error.SchemaPathErr{
				Code:    http.StatusBadRequest,
				PathErr: fmt.Errorf("cannot walk into array attr=[%s] with empty idx", p.AttrName),
			}
		}
		return nil
	}
	return p.buildNextPathType(itemDef, nextPath)
}

func (p *PathNode) buildNextPathMap(nextPath string) *Error.SchemaPathErr {
	itemDef, ok := p.AttrDef[JsonKey.AdditionalProperties].(map[string]interface{})
	if !ok {
		return &Error.SchemaPathErr{
			Code:    http.StatusBadRequest,
			PathErr: fmt.Errorf("missing field=[%s] in schema, for attr=[%s], @path=[%s]", JsonKey.Items, p.AttrName, p.FullPath()),
		}
	}
	itemType := itemDef[JsonKey.Type].(string)
	if itemType == JsonKey.Array || (itemType == JsonKey.Object && SchemaDoc.IsMap(itemDef)) {
		return &Error.SchemaPathErr{
			Code:    http.StatusBadRequest,
			PathErr: fmt.Errorf("walk on [%s,%s] in [%s] is not supported. attr=[%s], @path=[%s]", JsonKey.Array, JsonKey.Map, JsonKey.Map, p.AttrName, p.FullPath()),
		}
	}
	if p.Idx == "" {
		// for map, if p.Idx == "", it means nextPath is already empty
		// because buildPath will try to use one step if nextPath as idx
		return nil
	}
	return p.buildNextPathType(itemDef, nextPath)
}

func (p *PathNode) buildNextPathObj(nextPath string) *Error.SchemaPathErr {
	itemDoc, ok := p.Schema.SubDocs[p.AttrName]
	if !ok {
		return &Error.SchemaPathErr{
			Code:    http.StatusBadRequest,
			PathErr: fmt.Errorf("path not walkable, attr value has no definition @path=[%s/%s]", p.FullPath(), p.AttrName),
		}
	}
	nextNode, err := New(p.Conn, "", "", nextPath, p, itemDoc)
	if err != nil {
		return err
	}
	p.Next = nextNode
	return p.Next.BuildPath()
}

func (p *PathNode) buildNextPathCmt(nextPath string) *Error.SchemaPathErr {
	cmtRef, ok := p.Schema.CmtRefs[p.AttrName]
	if !ok {
		// just a normal string attribute. not a CMT ref.
		return nil
	}
	if cmtRef.CmtType != JsonKey.Inventory {
		return &Error.SchemaPathErr{
			Code:    http.StatusInternalServerError,
			PathErr: fmt.Errorf("schema path does not support [%s]=[%s/%s]", JsonKey.ContentMediaType, cmtRef.CmtType, cmtRef.ContentType),
		}
	}
	nextNode, err := New(p.Conn, cmtRef.ContentType, "", nextPath, p, nil)
	if err != nil {
		return err
	}
	p.Next = nextNode
	return p.Next.BuildPath()
}
