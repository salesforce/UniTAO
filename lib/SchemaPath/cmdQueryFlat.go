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
	"net/http"
	"reflect"

	"github.com/salesforce/UniTAO/lib/Schema/JsonKey"
	"github.com/salesforce/UniTAO/lib/Schema/SchemaDoc"
	"github.com/salesforce/UniTAO/lib/SchemaPath/Data"
	"github.com/salesforce/UniTAO/lib/SchemaPath/Node"
	"github.com/salesforce/UniTAO/lib/SchemaPath/PathCmd"
	"github.com/salesforce/UniTAO/lib/Util"
	"github.com/salesforce/UniTAO/lib/Util/Http"
)

type CmdQueryFlat struct {
	p *Node.PathNode
}

func NewFlatQuery(conn *Data.Connection, dataType string, dataId string, path string) (*CmdQueryFlat, *Http.HttpError) {
	node, err := BuildNodePath(conn, dataType, dataId, path)
	if err != nil {
		return nil, err
	}
	return &CmdQueryFlat{
		p: node,
	}, nil
}

func (c *CmdQueryFlat) Name() string {
	return PathCmd.CmdFlat
}

func (c *CmdQueryFlat) WalkValue() (interface{}, *Http.HttpError) {
	valueList, err := c.GetNodeValue(c.p)
	if err != nil {
		return nil, err
	}
	if len(valueList) == 1 {
		return valueList[0], nil
	}
	return valueList, nil
}

func (c *CmdQueryFlat) GetNodeValue(node *Node.PathNode) ([]interface{}, *Http.HttpError) {
	nodeList := c.getNodeList([]*Node.PathNode{node})
	result := []interface{}{}
	for _, node := range nodeList {
		valueList, err := c.getSingleNodeValue(node)
		if err != nil {
			return nil, err
		}
		result = append(result, valueList...)
	}
	if len(result) > 0 && reflect.TypeOf(result[0]).Kind() == reflect.String {
		dedupeList, ex := Util.DeDupeList(result)
		if ex != nil {
			return nil, Http.WrapError(ex, fmt.Sprintf("failed to merge dup result @path=[%s]", node.FullPath()), http.StatusInternalServerError)
		}
		return dedupeList, nil
	}
	return result, nil
}

func (c *CmdQueryFlat) getSingleNodeValue(node *Node.PathNode) ([]interface{}, *Http.HttpError) {
	if node.IsRecord() {
		flatObj, err := c.FlatObject(node)
		if err != nil {
			return nil, err
		}
		return []interface{}{flatObj}, nil
	}
	if node.AttrDef[JsonKey.Type].(string) == JsonKey.Object && !SchemaDoc.IsMap(node.AttrDef) {
		if node.Prev.Idx == Node.All {
			return []interface{}{node.Idx}, nil
		}
		flatObj, err := c.FlatObject(node)
		if err != nil {
			return nil, err
		}
		return []interface{}{flatObj}, nil
	}

	dataType := node.AttrDef[JsonKey.Type].(string)
	if dataType == JsonKey.Array || dataType == JsonKey.Object {
		err := node.BuildIdx(Node.All)
		if err != nil {
			return nil, err
		}
		if dataType == JsonKey.Array {
			return c.FlatArray(node)
		}
		flatMap, err := c.FlatMap(node)
		if err != nil {
			return nil, err
		}
		return []interface{}{flatMap}, nil
	}
	return []interface{}{node.Data}, nil
}

func (c *CmdQueryFlat) getNodeList(nodeList []*Node.PathNode) []*Node.PathNode {
	if len(nodeList[0].Next) == 0 {
		return nodeList
	}
	newList := []*Node.PathNode{}
	strMap := map[string]int{}
	canDeDupe := false
	for _, node := range nodeList {
		for _, next := range node.Next {
			if !canDeDupe {
				if !next.IsRecord() && next.AttrDef[JsonKey.Type].(string) == JsonKey.String {
					canDeDupe = true
				}
			}
			if canDeDupe {
				if _, ok := strMap[next.Data.(string)]; ok {
					continue
				}
				strMap[next.Data.(string)] = 1
			}
			newList = append(newList, next)
		}
	}
	return c.getNodeList(newList)
}

func (c CmdQueryFlat) FlatArray(node *Node.PathNode) ([]interface{}, *Http.HttpError) {
	ary := make([]interface{}, 0, len(node.Next))
	for _, next := range node.Next {
		itemType := next.AttrDef[JsonKey.Type].(string)
		if itemType == JsonKey.Object {
			ary = append(ary, next.Idx)
		} else {
			ary = append(ary, next.Data)
		}
	}
	return ary, nil
}

func (c CmdQueryFlat) FlatMap(node *Node.PathNode) (map[string]interface{}, *Http.HttpError) {
	flatMap := map[string]interface{}{}
	for _, next := range node.Next {
		itemType := next.AttrDef[JsonKey.Type].(string)
		if itemType == JsonKey.Object {
			itemKey, ex := next.Schema.BuildKey(next.Data.(map[string]interface{}))
			if ex != nil {
				return nil, Http.WrapError(ex, fmt.Sprintf("failed to get key @path=[%s]", next.FullPath()), http.StatusInternalServerError)
			}
			flatMap[next.Idx] = itemKey
		} else {
			flatMap[next.Idx] = next.Data
		}
	}
	return flatMap, nil
}

func (c *CmdQueryFlat) FlatObject(node *Node.PathNode) (map[string]interface{}, *Http.HttpError) {
	flatObj := map[string]interface{}{}
	for attrName, attrDef := range node.Schema.Data[JsonKey.Properties].(map[string]interface{}) {
		attrData, ok := node.Data.(map[string]interface{})[attrName]
		if !ok {
			continue
		}
		attrType := attrDef.(map[string]interface{})[JsonKey.Type].(string)
		switch attrType {
		case JsonKey.Object:
			attrList := []interface{}{}
			for key := range attrData.(map[string]interface{}) {
				attrList = append(attrList, key)
			}
			flatObj[attrName] = attrList
		case JsonKey.Array:
			err := node.BuildPath(fmt.Sprintf("%s[%s]", attrName, Node.All))
			if err != nil {
				return nil, err
			}
			flatAry, err := c.FlatArray(node.Next[0])
			if err != nil {
				return nil, err
			}
			flatObj[attrName] = flatAry
		default:
			flatObj[attrName] = attrData
		}
		node.Next = []*Node.PathNode{}
	}
	return flatObj, nil
}
