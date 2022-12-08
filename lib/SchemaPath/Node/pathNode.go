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
	"strconv"

	"github.com/salesforce/UniTAO/lib/Schema/JsonKey"
	"github.com/salesforce/UniTAO/lib/Schema/SchemaDoc"
	"github.com/salesforce/UniTAO/lib/SchemaPath/Data"
	"github.com/salesforce/UniTAO/lib/Util"
	"github.com/salesforce/UniTAO/lib/Util/Http"
)

type PathNode struct {
	Id       string
	Conn     *Data.Connection
	Schema   *SchemaDoc.SchemaDoc
	DataType string
	DataId   string
	AttrName string
	AttrDef  map[string]interface{}
	Idx      string
	Prev     *PathNode
	Next     []*PathNode
	Data     interface{}
}

const (
	All = "*"
)

func New(conn *Data.Connection, dataType string, dataId string) (*PathNode, *Http.HttpError) {
	node, err := newRecordNode(conn, dataType, dataId, nil)
	if err != nil {
		return nil, err
	}
	err = node.Sync()
	if err != nil {
		return nil, err
	}
	return node, nil
}

func newRecordNode(conn *Data.Connection, dataType string, dataId string, prev *PathNode) (*PathNode, *Http.HttpError) {
	if prev == nil {
		if dataType == "" {
			return nil, Http.NewHttpError("invalid node parameter, need dataType when prev is nil", http.StatusBadRequest)
		}
	}
	if dataType != "" {
		if dataId == "" {
			return nil, Http.NewHttpError("invalid node parameter, need dataId to query data record from connection.", http.StatusBadRequest)
		}
	}
	node := PathNode{
		Id:       "",
		Conn:     conn,
		Prev:     prev,
		DataType: dataType,
		DataId:   dataId,
		Next:     []*PathNode{},
	}
	return &node, nil
}

func (p *PathNode) FullPath() string {
	var path string
	if p.Prev != nil {
		path = fmt.Sprintf("%s%s", p.Prev.FullPath(), p.Id)
	} else {
		path = fmt.Sprintf("/%s/%s", p.DataType, p.DataId)
	}
	return path
}

func (p *PathNode) Sync() *Http.HttpError {
	if p.IsRecord() {
		return p.syncFromConn()
	}
	if p.AttrDef != nil {
		attrType := p.AttrDef[JsonKey.Type]
		attrName := p.AttrName
		if attrName == "" {
			attrName = p.Prev.AttrName
		}
		p.Schema = p.Prev.Schema
		if attrType == JsonKey.Object && !SchemaDoc.IsMap(p.AttrDef) {
			p.Schema = p.Prev.Schema.SubDocs[attrName]
		}
	}
	return nil
}

func (p *PathNode) IsRecord() bool {
	return p.DataType != ""
}

func (p *PathNode) syncFromConn() *Http.HttpError {
	if p.DataId == "" {
		return Http.NewHttpError(fmt.Sprintf("missing data Id in path @[%s]", p.FullPath()), http.StatusBadRequest)
	}
	schemaRecord, err := p.Conn.GetRecord(JsonKey.Schema, p.DataType)
	if err != nil {
		return Http.WrapError(err, fmt.Sprintf("failed to get schema, @path=[%s]", p.FullPath()), http.StatusBadRequest)
	}
	record, err := p.Conn.GetRecord(p.DataType, p.DataId)
	if err != nil {
		return Http.WrapError(err, fmt.Sprintf("failed to get record @path=[%s]", p.FullPath()), http.StatusNotFound)
	}
	p.Data = record.Data
	if record.Version != schemaRecord.Version {
		archivedType := SchemaDoc.ArchivedSchemaId(p.DataType, record.Version)
		schemaRecord, err = p.Conn.GetRecord(JsonKey.Schema, archivedType)
		if err != nil {
			return err
		}
	}
	schema, ex := SchemaDoc.New(schemaRecord.Data)
	if ex != nil {
		return Http.WrapError(ex, fmt.Sprintf("failed to create SchemaDoc @path=[%s]", p.FullPath()), http.StatusInternalServerError)
	}
	if p.DataType != schema.Id {
		return Http.NewHttpError(fmt.Sprintf("invalid schema, %s=[%s] and schema=[%s] not match", JsonKey.Name, p.DataType, schema.Id), http.StatusInternalServerError)
	}
	p.Schema = schema
	return nil
}

func (p *PathNode) buildAttrNode(attrName string) *Http.HttpError {
	if attrName == "" {
		return nil
	}
	attrData, ok := p.Data.(map[string]interface{})[attrName]
	if !ok {
		return Http.NewHttpError(fmt.Sprintf("attr=[%s] does not exists, @path=[%s]", attrName, p.FullPath()), http.StatusNotFound)
	}
	attrNode := PathNode{
		Id:       attrName,
		Conn:     p.Conn,
		Prev:     p,
		DataType: "",
		DataId:   "",
		Next:     []*PathNode{},
		AttrName: attrName,
		Data:     attrData,
	}
	attrDef, attrDefined := p.Schema.Data[JsonKey.Properties].(map[string]interface{})[attrName]
	if attrDefined {
		attrNode.AttrDef = attrDef.(map[string]interface{})
		attrNode.Sync()
		err := attrNode.buildCmtNode()
		if err != nil {
			return err
		}
	}
	p.Next = append(p.Next, &attrNode)
	return nil
}

func (p *PathNode) newIdxNode(idx string, itemDef map[string]interface{}, data interface{}) *Http.HttpError {
	idxNode := PathNode{
		Id:       fmt.Sprintf("[%s]", idx),
		Conn:     p.Conn,
		Prev:     p,
		DataType: "",
		DataId:   "",
		Next:     []*PathNode{},
		AttrName: "",
		Idx:      idx,
		AttrDef:  itemDef,
		Data:     data,
	}
	err := idxNode.Sync()
	if err != nil {
		return err
	}
	p.Next = append(p.Next, &idxNode)
	return nil
}

func (p *PathNode) buildIdxNodes(idx string) *Http.HttpError {
	if idx == "" {
		return nil
	}
	if idx == All {
		p.Idx = All
	}
	attrType := p.AttrDef[JsonKey.Type].(string)
	var err *Http.HttpError
	switch attrType {
	case JsonKey.Array:
		err = p.buildArrayIdxNode(idx)
	case JsonKey.Map:
		err = p.buildMapIdxNode(idx)
	case JsonKey.Object:
		if !SchemaDoc.IsMap(p.AttrDef) {
			return Http.NewHttpError(fmt.Sprintf("invalid schema type=[%s] not a map for idx=[%s] @path=[%s]", attrType, p.Idx, p.FullPath()), http.StatusBadRequest)
		}
		err = p.buildMapIdxNode(idx)
	default:
		return Http.NewHttpError(fmt.Sprintf("invalid schema type=[%s] for idx=[%s] @path=[%s]", attrType, p.Idx, p.FullPath()), http.StatusBadRequest)
	}
	if p.Idx != All && len(p.Next) == 0 {
		return Http.NewHttpError(fmt.Sprintf("invalid idx, [%s] not found @path=[%s]", p.Idx, p.FullPath()), http.StatusNotFound)
	}
	if err != nil {
		return err
	}
	for _, next := range p.Next {
		err := next.buildCmtNode()
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *PathNode) BuildIdx(idx string) *Http.HttpError {
	if idx == "" {
		return nil
	}
	nextList := []*PathNode{}
	var lastErr *Http.HttpError
	if len(p.Next) > 0 {
		for _, next := range p.Next {
			err := next.BuildIdx(idx)
			if err != nil {
				lastErr = err
				continue
			}
			nextList = append(nextList, next)
		}
		p.Next = nextList
		if len(p.Next) == 0 {
			return lastErr
		}
		return nil
	}
	return p.buildIdxNodes(idx)
}

func (p *PathNode) BuildPath(nextPath string) *Http.HttpError {
	if nextPath == "" {
		return nil
	}
	nextList := []*PathNode{}
	var lastErr *Http.HttpError
	if len(p.Next) > 0 {
		for _, next := range p.Next {
			err := next.BuildPath(nextPath)
			if err != nil {
				lastErr = err
				continue
			}
			nextList = append(nextList, next)
		}
		p.Next = nextList
		if len(p.Next) == 0 {
			return lastErr
		}
		return nil
	}
	if p.Schema == nil {
		return Http.NewHttpError(fmt.Sprintf("cannot walk further with undefined attr=[%s] @path=[%s]", p.AttrName, p.FullPath()), http.StatusBadRequest)
	}
	attrName, attrIdx, err := Util.ParseArrayPath(nextPath)
	if err != nil {
		return Http.WrapError(err, fmt.Sprintf("failed to parse path @[path]=[%s]", p.FullPath()), http.StatusBadRequest)
	}
	if p.AttrDef != nil {
		attrType := p.AttrDef[JsonKey.Type].(string)
		if attrType == JsonKey.Array {
			return Http.NewHttpError(fmt.Sprintf("invalid path, missing array idx @path=[%s]", p.FullPath()), http.StatusBadRequest)
		}
		if attrType == JsonKey.Map || (attrType == JsonKey.Object && SchemaDoc.IsMap(p.AttrDef)) {
			if attrIdx != "" {
				return Http.NewHttpError(fmt.Sprintf("invalid path, missing array key @path=[%s]", p.FullPath()), http.StatusBadRequest)
			}
			idxErr := p.buildIdxNodes(attrName)
			if idxErr != nil {
				return idxErr
			}
			return nil
		}
	}
	e := p.buildAttrNode(attrName)
	if e != nil {
		return e
	}
	e = p.Next[0].buildIdxNodes(attrIdx)
	if e != nil {
		return e
	}
	return nil
}

func (p *PathNode) buildArrayIdxNode(idx string) *Http.HttpError {
	itemDef := p.AttrDef[JsonKey.Items].(map[string]interface{})
	itemType := itemDef[JsonKey.Type].(string)
	arrayData, isArray := p.Data.([]interface{})
	if !isArray {
		return Http.NewHttpError(fmt.Sprintf("data cannot convert to array. @path=[%s]", p.FullPath()), http.StatusBadRequest)
	}
	for i, item := range arrayData {
		selected := false
		var itemKey string
		switch itemType {
		case JsonKey.Object:
			iSchema := p.Schema.SubDocs[p.AttrName]
			key, ex := iSchema.BuildKey(item.(map[string]interface{}))
			if ex != nil {
				return Http.WrapError(ex, fmt.Sprintf("failed to generate key from item @path=[%s[%d]]", p.FullPath(), i), http.StatusInternalServerError)
			}
			itemKey = key
		case JsonKey.String:
			itemKey = item.(string)
		default:
			itemKey = strconv.Itoa(i)
		}
		if idx != All && idx != itemKey {
			continue
		}
		err := p.newIdxNode(itemKey, itemDef, item)
		if err != nil {
			return err
		}
		if selected && idx != All {
			break
		}
	}
	return nil
}

func (p *PathNode) buildMapIdxNode(idx string) *Http.HttpError {
	itemDef, ok := p.AttrDef[JsonKey.AdditionalProperties].(map[string]interface{})
	if !ok {
		return Http.NewHttpError(fmt.Sprintf("missing field=[%s] in schema, for attr=[%s], @path=[%s]", JsonKey.Items, p.AttrName, p.FullPath()), http.StatusBadRequest)
	}
	itemType := itemDef[JsonKey.Type].(string)
	if itemType == JsonKey.Array || (itemType == JsonKey.Object && SchemaDoc.IsMap(itemDef)) {
		return Http.NewHttpError(fmt.Sprintf("walk on [%s,%s] in [%s] is not supported. attr=[%s], @path=[%s]", JsonKey.Array, JsonKey.Map, JsonKey.Map, p.AttrName, p.FullPath()), http.StatusBadRequest)
	}
	mapData, isMap := p.Data.(map[string]interface{})
	if !isMap {
		return Http.NewHttpError(fmt.Sprintf("data cannot convert to array. @path=[%s]", p.FullPath()), http.StatusBadRequest)
	}
	if idx != All {
		filterData, ok := mapData[idx]
		if !ok {
			return Http.NewHttpError(fmt.Sprintf("data key=[%s] does not exists @path=[%s]", p.Idx, p.FullPath()), http.StatusNotFound)
		}
		mapData = map[string]interface{}{
			idx: filterData,
		}
	}
	for key, item := range mapData {
		err := p.newIdxNode(key, itemDef, item)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *PathNode) buildCmtNode() *Http.HttpError {
	attrType := p.AttrDef[JsonKey.Type].(string)
	if attrType != JsonKey.String {
		return nil
	}
	_, ok := p.AttrDef[JsonKey.ContentMediaType].(string)
	if !ok {
		return nil
	}
	attrName := p.AttrName
	if attrName == "" {
		attrName = p.Prev.AttrName
	}
	ref, ok := p.Schema.CmtRefs[attrName]
	if !ok {
		return Http.NewHttpError(fmt.Sprintf("failed to find Cmt @path=[%s]", p.FullPath()), http.StatusBadRequest)
	}

	cmtNode, err := newRecordNode(p.Conn, ref.ContentType, p.Data.(string), p)
	if err != nil {
		return err
	}
	err = cmtNode.Sync()
	if err != nil {
		return err
	}
	p.Next = append(p.Next, cmtNode)
	return nil
}
