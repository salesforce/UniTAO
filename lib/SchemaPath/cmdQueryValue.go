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
	"github.com/salesforce/UniTAO/lib/SchemaPath/Data"
	"github.com/salesforce/UniTAO/lib/SchemaPath/Node"
	"github.com/salesforce/UniTAO/lib/SchemaPath/PathCmd"
	"github.com/salesforce/UniTAO/lib/Util/Http"
)

type CmdQueryValue struct {
	p    *Node.PathNode
	Path string
	Data interface{}
}

func NewValueQuery(conn *Data.Connection, dataType string, dataId string, path string) (*CmdQueryValue, *Http.HttpError) {
	node, err := BuildNodePath(conn, dataType, dataId, path)
	if err != nil {
		return nil, err
	}
	return &CmdQueryValue{
		p:    node,
		Path: path,
	}, nil
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
	dataList := c.GetNodeValue(c.p)
	if len(dataList) == 1 {
		return dataList[0], nil
	}
	return dataList, nil

}

func (c *CmdQueryValue) GetNodeValue(node *Node.PathNode) []interface{} {
	if len(node.Next) == 0 {
		return []interface{}{node.Data}
	}
	dataList := []interface{}{}
	for _, next := range node.Next {
		result := c.GetNodeValue(next)
		if len(result) > 0 {
			// we want to keep the array embeded structure
			if len(result) == 1 {
				dataList = append(dataList, result[0])
			} else {
				dataList = append(dataList, result)
			}
		}
	}
	return dataList
}
