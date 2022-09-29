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

	"github.com/salesforce/UniTAO/lib/Schema/JsonKey"
	"github.com/salesforce/UniTAO/lib/SchemaPath/Data"
	"github.com/salesforce/UniTAO/lib/SchemaPath/Node"
	"github.com/salesforce/UniTAO/lib/SchemaPath/PathCmd"
	"github.com/salesforce/UniTAO/lib/Util/Http"
)

type CmdQueryRef struct {
	p *Node.PathNode
}

func NewRefQuery(conn *Data.Connection, dataType string, dataId string, path string) (*CmdQueryRef, *Http.HttpError) {
	node, err := Node.New(conn, dataType, dataId, path, nil, nil)
	if err != nil {
		return nil, err
	}
	err = node.BuildPath()
	if err != nil {
		return nil, err
	}
	return &CmdQueryRef{
		p: node,
	}, nil
}

func (c *CmdQueryRef) Name() string {
	return PathCmd.CmdRef
}

func (c *CmdQueryRef) WalkValue() (interface{}, *Http.HttpError) {
	return c.GetNodeValue(c.p, nil)
}

func (c *CmdQueryRef) GetNodeValue(node *Node.PathNode, nodeValue interface{}) (interface{}, *Http.HttpError) {
	nodeValue, err := Node.GetNodeValue(node, nodeValue)
	if err != nil {
		return nil, err
	}
	if node.Next == nil {
		return nil, Http.NewHttpError(fmt.Sprintf("path does not contain Ref attribute. @path=[%s]", node.FullPath()), http.StatusBadRequest)
	}
	if node.Next.NextPath == "" {
		return c.GetNodeRef(node, nodeValue)
	}
	if node.Idx == PathCmd.ALL {
		valueType := node.AttrDef[JsonKey.Type].(string)
		if valueType == JsonKey.Array {
			return c.GetNodeArrayAll(node, nodeValue)
		}
		return c.GetNodeMapAll(node, nodeValue)
	}
	nodeValue, err = c.GetNodeValue(node.Next, nodeValue)
	if err != nil {
		return nil, err
	}
	return nodeValue, nil
}

func (c *CmdQueryRef) GetNodeRef(node *Node.PathNode, nodeValue interface{}) (interface{}, *Http.HttpError) {
	itemType := node.AttrDef[JsonKey.Type].(string)
	// Array, map, object and string
	switch itemType {
	case JsonKey.Array:
		return c.GetNodeArrayRef(node, nodeValue)
	case JsonKey.Object:
		return c.GetNodeMapRef(node, nodeValue)
	default:
		return nodeValue, nil
	}
}

func (c *CmdQueryRef) GetNodeArrayRef(node *Node.PathNode, nodeValue interface{}) (interface{}, *Http.HttpError) {
	itemType := node.AttrDef[JsonKey.Items].(map[string]interface{})[JsonKey.Type].(string)
	if node.Idx == PathCmd.ALL {
		valueList, ok := nodeValue.([]interface{})
		if !ok {
			return nil, Http.NewHttpError(fmt.Sprintf("invalid node value cannot convert to []interface{}, @path=[%s]", node.FullPath()), http.StatusInternalServerError)
		}
		if itemType == JsonKey.String {
			return valueList, nil
		}
		result := make([]string, 0, len(valueList))
		for idx, itemObj := range valueList {
			itemKey, err := node.Next.Schema.BuildKey(itemObj.(map[string]interface{}))
			if err != nil {
				return nil, Http.WrapError(err, fmt.Sprintf("failed to get key, %s[%d], @path=[%s]", node.AttrName, idx, node.FullPath()), http.StatusInternalServerError)
			}
			result = append(result, itemKey)
		}
		return result, nil
	}
	if itemType == JsonKey.String {
		return nodeValue, nil
	}
	return node.Idx, nil
}

func (c *CmdQueryRef) GetNodeMapRef(node *Node.PathNode, nodeValue interface{}) (interface{}, *Http.HttpError) {
	itemType := node.AttrDef[JsonKey.Items].(map[string]interface{})[JsonKey.Type].(string)
	if node.Idx == PathCmd.ALL {
		valueMap, ok := nodeValue.(map[string]interface{})
		if !ok {
			return nil, Http.NewHttpError(fmt.Sprintf("invalid node value cannot convert to map[string]interface{}, @path=[%s]", node.FullPath()), http.StatusInternalServerError)
		}
		result := make([]string, 0, len(valueMap))
		for key, keyValue := range valueMap {
			if itemType == JsonKey.String {
				result = append(result, keyValue.(string))
			} else {
				itemKey, err := node.Next.Schema.BuildKey(keyValue.(map[string]interface{}))
				if err != nil {
					return nil, Http.WrapError(err, fmt.Sprintf("failed to get key, %s[%s], @path=[%s]", node.AttrName, key, node.FullPath()), http.StatusInternalServerError)
				}
				result = append(result, itemKey)
			}
		}
		return result, nil
	}
	if itemType == JsonKey.String {
		return nodeValue, nil
	}
	itemKey, err := node.Next.Schema.BuildKey(nodeValue.(map[string]interface{}))
	if err != nil {
		return nil, Http.WrapError(err, fmt.Sprintf("failed to get key, %s[%s], @path=[%s]", node.AttrName, node.Idx, node.FullPath()), http.StatusInternalServerError)
	}
	return itemKey, nil
}

func (c *CmdQueryRef) GetNodeArrayAll(node *Node.PathNode, nodeValue interface{}) (interface{}, *Http.HttpError) {
	parentValues, ok := nodeValue.([]interface{})
	if !ok {
		return nil, Http.NewHttpError(fmt.Sprintf("idx=[%s] didn't return array on function[Node.GetNodeValue], @path=[%s]", PathCmd.ALL, node.FullPath()), http.StatusInternalServerError)
	}
	result := make([]interface{}, 0, len(parentValues))
	for idx, item := range parentValues {
		itemValue, err := c.GetNodeValue(node.Next, item)
		if err != nil {
			if err.Status == http.StatusNotFound {
				continue
			}
			return nil, Http.WrapError(err, fmt.Sprintf("failed to get %s[%d] @path=[%s]", node.AttrName, idx, node.FullPath()), err.Status)
		}
		result = append(result, itemValue)
	}
	return result, nil
}

func (c *CmdQueryRef) GetNodeMapAll(node *Node.PathNode, nodeValue interface{}) (interface{}, *Http.HttpError) {
	parentValues, ok := nodeValue.(map[string]interface{})
	if !ok {
		return nil, Http.NewHttpError(fmt.Sprintf("idx=[%s] didn't return map on function[Node.GetNodeValue], @path=[%s]", PathCmd.ALL, node.FullPath()), http.StatusInternalServerError)
	}
	result := make([]interface{}, 0, len(parentValues))
	for key, item := range parentValues {
		itemValue, err := c.GetNodeValue(node.Next, item)
		if err != nil {
			if err.Status == http.StatusNotFound {
				continue
			}
			return nil, Http.WrapError(err, fmt.Sprintf("failed to get %s[%s] @path=[%s]", node.AttrName, key, node.FullPath()), err.Status)
		}
		result = append(result, itemValue)
	}
	return result, nil
}
