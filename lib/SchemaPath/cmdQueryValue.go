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
	"github.com/salesforce/UniTAO/lib/SchemaPath/Node"
	"github.com/salesforce/UniTAO/lib/SchemaPath/PathCmd"
	"github.com/salesforce/UniTAO/lib/Util/Http"
)

type CmdQueryValue struct {
	p    *Node.PathNode
	Data interface{}
}

func (c *CmdQueryValue) Name() string {
	return PathCmd.CmdRef
}

func (c *CmdQueryValue) WalkValue() (interface{}, *Http.HttpError) {
	/*
		TODO:
		walk as normal value except at last step if it is a
			a, array or map, if CMT, return CMT referenced value
			b, CMT, return CMT referenced value
	*/
	return c.GetNodeValue(c.p, nil)
}

func (c *CmdQueryValue) GetNodeValue(node *Node.PathNode, parentData interface{}) (interface{}, *Http.HttpError) {
	nodeValue, err := Node.GetNodeValue(node, parentData)
	if err != nil {
		return nil, err
	}
	if node.Next == nil {
		return nodeValue, nil
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

func (c *CmdQueryValue) GetNodeArrayAll(node *Node.PathNode, nodeValue interface{}) (interface{}, *Http.HttpError) {
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

func (c *CmdQueryValue) GetNodeMapAll(node *Node.PathNode, nodeValue interface{}) (interface{}, *Http.HttpError) {
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
