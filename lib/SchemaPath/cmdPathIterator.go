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
	"github.com/salesforce/UniTAO/lib/SchemaPath/Error"
	"github.com/salesforce/UniTAO/lib/SchemaPath/Node"
	"github.com/salesforce/UniTAO/lib/SchemaPath/PathCmd"
	"github.com/salesforce/UniTAO/lib/Util"
)

const (
	QueryPath      = "queryPath"
	QueryResults   = "queryResults"
	QueryData      = "data"
	QueryIterators = "iterators"
)

type QueryResult struct {
	Data      interface{} `json:"data"`
	Iterators []string    `json:"iterators"`
}

type IteratorResult struct {
	Path         string        `json:"queryPath"`
	QueryResults []QueryResult `json:"queryResults"`
}

type CmdPathIterator struct {
	path string
	p    *Node.PathNode
}

func (c *CmdPathIterator) Name() string {
	return PathCmd.CmdIter
}

func (c *CmdPathIterator) WalkValue() (interface{}, *Error.SchemaPathErr) {
	/*
		TODO:
		walk as the path but only record the * Idx on Array or Map.
		record the format as following:

	*/
	queryResult, err := c.GetNodeValue(c.p, nil)
	if err != nil {
		return nil, err
	}
	iterResult := IteratorResult{
		Path:         c.path,
		QueryResults: queryResult,
	}
	result, cErr := Util.JsonCopy(iterResult)
	if cErr != nil {
		return nil, &Error.SchemaPathErr{
			Code:    http.StatusInternalServerError,
			PathErr: fmt.Errorf("failed to convert struct [IteratorResult] to json, Error:%s", cErr),
		}
	}
	return result, nil
}

func (c *CmdPathIterator) GetNodeValue(node *Node.PathNode, parentData interface{}) ([]QueryResult, *Error.SchemaPathErr) {
	nodeValue, err := Node.GetNodeValue(node, parentData)
	if err != nil {
		return nil, err
	}
	if node.Next == nil {
		return c.BuildResult(nodeValue), nil
	}
	if node.Idx == PathCmd.ALL {
		valueType := node.AttrDef[JsonKey.Type].(string)
		if valueType == JsonKey.Array {
			return c.GetNodeArrayAll(node, nodeValue)
		}
		return c.GetNodeMapAll(node, nodeValue)
	}
	queryResult, err := c.GetNodeValue(node.Next, nodeValue)
	if err != nil {
		return nil, err
	}
	return queryResult, nil
}

func (c *CmdPathIterator) GetNodeArrayAll(node *Node.PathNode, nodeValue interface{}) ([]QueryResult, *Error.SchemaPathErr) {
	parentValues, ok := nodeValue.([]interface{})
	if !ok {
		return nil, &Error.SchemaPathErr{
			Code:    http.StatusInternalServerError,
			PathErr: fmt.Errorf("idx=[%s] didn't return array on function[Node.GetNodeValue], @path=[%s]", PathCmd.ALL, node.FullPath()),
		}
	}
	result := []QueryResult{}
	for idx, item := range parentValues {
		nextResult, err := c.GetNodeValue(node.Next, item)
		if err != nil {
			if err.Code == http.StatusNotFound {
				continue
			}
			return nil, Error.AppendErr(err, fmt.Sprintf("failed to get %s[%d] @path=[%s]", node.AttrName, idx, node.FullPath()))
		}
		itemKey, err := GetItemKey(node, item)
		if err != nil {
			return nil, err
		}
		nextResult = AppendIterator(nextResult, itemKey)
		result = append(result, nextResult...)
	}
	return result, nil
}

func (c *CmdPathIterator) GetNodeMapAll(node *Node.PathNode, nodeValue interface{}) ([]QueryResult, *Error.SchemaPathErr) {
	parentValues, ok := nodeValue.(map[string]interface{})
	if !ok {
		return nil, &Error.SchemaPathErr{
			Code:    http.StatusInternalServerError,
			PathErr: fmt.Errorf("idx=[%s] didn't return map on function[Node.GetNodeValue], @path=[%s]", PathCmd.ALL, node.FullPath()),
		}
	}
	result := []QueryResult{}
	for key, item := range parentValues {
		nextResult, err := c.GetNodeValue(node.Next, item)
		if err != nil {
			if err.Code == http.StatusNotFound {
				continue
			}
			return nil, Error.AppendErr(err, fmt.Sprintf("failed to get %s[%s] @path=[%s]", node.AttrName, key, node.FullPath()))
		}
		AppendIterator(nextResult, key)
		result = append(result, nextResult...)
	}
	return result, nil
}

func (c *CmdPathIterator) BuildResult(value interface{}) []QueryResult {
	result := []QueryResult{
		QueryResult{
			Data:      value,
			Iterators: []string{},
		},
	}
	return result
}

func GetItemKey(node *Node.PathNode, item interface{}) (string, *Error.SchemaPathErr) {
	itemType := node.AttrDef[JsonKey.Items].(map[string]interface{})[JsonKey.Type].(string)
	if itemType == JsonKey.Object {
		itemKey, err := node.Next.Schema.BuildKey(item.(map[string]interface{}))
		if err != nil {
			return "", &Error.SchemaPathErr{
				Code:    http.StatusInternalServerError,
				PathErr: fmt.Errorf("failed to build key at %s", node.Next.FullPath()),
			}
		}
		return itemKey, nil
	}
	return item.(string), nil
}

func AppendIterator(results []QueryResult, key string) []QueryResult {
	result := []QueryResult{}
	for _, r := range results {
		newIterators := append([]string{key}, r.Iterators...)
		r.Iterators = newIterators
		result = append(result, r)
	}
	return result
}
