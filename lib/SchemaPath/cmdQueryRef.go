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
	node, err := BuildNodePath(conn, dataType, dataId, path)
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
	dataList, err := c.GetNodeValue(c.p)
	if err != nil {
		return nil, err
	}
	if len(dataList) == 1 {
		return dataList[0], nil
	}
	return dataList, nil
}

func (c *CmdQueryRef) GetNodeValue(node *Node.PathNode) ([]interface{}, *Http.HttpError) {
	if len(node.Next) == 0 {
		if node.IsRecord() && node.Prev != nil {
			return []interface{}{node.Prev.Data}, nil
		}
		dataType := node.AttrDef[JsonKey.Type].(string)
		if node.Idx != "" && dataType == JsonKey.Object {
			ref, err := node.Schema.BuildKey(node.Data.(map[string]interface{}))
			if err != nil {
				return nil, Http.WrapError(err, fmt.Sprintf("failed to build id @path=[%s]", node.FullPath()), http.StatusInternalServerError)
			}
			return []interface{}{ref}, nil
		}
		return []interface{}{node.Data}, nil
	}
	dataList := []interface{}{}
	for _, next := range node.Next {
		valueList, ex := c.GetNodeValue(next)
		if ex != nil {
			return nil, ex
		}
		dataList = append(dataList, valueList...)
	}
	return dataList, nil
}
