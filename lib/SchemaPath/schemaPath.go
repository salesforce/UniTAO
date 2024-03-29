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
	"github.com/salesforce/UniTAO/lib/Util"
	"github.com/salesforce/UniTAO/lib/Util/Http"
)

func BuildNodePath(conn *Data.Connection, dataType string, dataId string, dataPath string) (*Node.PathNode, *Http.HttpError) {
	node, err := Node.New(conn, dataType, dataId)
	if err != nil {
		return nil, err
	}
	for dataPath != "" {
		stepPath, stepNext := Util.ParsePath(dataPath)
		err = node.BuildPath(stepPath)
		if err != nil {
			return nil, err
		}
		dataPath = stepNext
	}
	return node, nil
}

func CreateQuery(conn *Data.Connection, dataType string, dataPath string) (PathCmd.QueryIface, *Http.HttpError) {
	qPath, qCmd, pErr := PathCmd.Parse(dataPath)
	if pErr != nil {
		return nil, pErr
	}
	dataId, nextPath := Util.ParsePath(qPath)
	switch qCmd {
	case PathCmd.CmdSchema:
		return NewSchemaQuery(conn, dataType, dataId, nextPath)
	case PathCmd.CmdFlat:
		return NewFlatQuery(conn, dataType, dataId, nextPath)
	case PathCmd.CmdRef:
		return NewRefQuery(conn, dataType, dataId, nextPath)
	case PathCmd.CmdIter:
		return NewIteratorQuery(conn, dataType, dataId, nextPath)
	default:
		if IsCmdPathName(qCmd) {
			return NewPathQuery(conn, dataType, qPath, qCmd)
		}
		return NewValueQuery(conn, dataType, dataId, nextPath)
	}
}
