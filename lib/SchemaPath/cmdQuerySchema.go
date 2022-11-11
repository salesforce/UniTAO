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
	"github.com/salesforce/UniTAO/lib/Schema/JsonKey"
	"github.com/salesforce/UniTAO/lib/Schema/SchemaDoc"
	"github.com/salesforce/UniTAO/lib/SchemaPath/Data"
	"github.com/salesforce/UniTAO/lib/SchemaPath/Node"
	"github.com/salesforce/UniTAO/lib/SchemaPath/PathCmd"
	"github.com/salesforce/UniTAO/lib/Util/Http"
)

type CmdQuerySchema struct {
	p *Node.PathNode
}

func NewSchemaQuery(conn *Data.Connection, dataType string, dataId string, path string) (*CmdQuerySchema, *Http.HttpError) {
	node, err := BuildNodePath(conn, dataType, dataId, path)
	if err != nil {
		return nil, err
	}
	return &CmdQuerySchema{
		p: node,
	}, nil
}

func (c *CmdQuerySchema) Name() string {
	return PathCmd.CmdSchema
}

func (c *CmdQuerySchema) WalkValue() (interface{}, *Http.HttpError) {
	dataList := c.GetNodeSchema(c.p)
	if len(dataList) == 1 {
		return dataList[0], nil
	}
	return dataList, nil
}

func (c *CmdQuerySchema) GetNodeSchema(node *Node.PathNode) []interface{} {
	if len(node.Next) > 0 {
		schemaList := []interface{}{}
		for _, next := range node.Next {
			valueList := c.GetNodeSchema(next)
			schemaList = append(schemaList, valueList...)
		}
		return schemaList
	}
	if node.IsRecord() || node.AttrDef[JsonKey.Type].(string) == JsonKey.Object && !SchemaDoc.IsMap(node.AttrDef) {
		return []interface{}{node.Schema.RAW}
	}
	if node.Idx != "" && node.Idx != Node.All {
		attrDefRaw := node.Schema.RAW[JsonKey.Properties].(map[string]interface{})[node.Prev.AttrName]
		itemDefRaw := attrDefRaw.(map[string]interface{})[JsonKey.Items]
		return []interface{}{itemDefRaw}
	}
	attrDefRaw := node.Schema.RAW[JsonKey.Properties].(map[string]interface{})[node.AttrName]
	return []interface{}{attrDefRaw}
}
