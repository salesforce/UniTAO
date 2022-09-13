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

	"github.com/salesforce/UniTAO/lib/SchemaPath/Data"
	"github.com/salesforce/UniTAO/lib/SchemaPath/Error"
	"github.com/salesforce/UniTAO/lib/SchemaPath/Node"
	"github.com/salesforce/UniTAO/lib/SchemaPath/PathCmd"
	"github.com/salesforce/UniTAO/lib/Util"
)

func CreateQuery(conn *Data.Connection, dataType string, dataPath string) (PathCmd.QueryIface, *Error.SchemaPathErr) {
	qPath, qCmd, pErr := PathCmd.Parse(dataPath)
	if pErr != nil {
		return nil, &Error.SchemaPathErr{
			Code:    http.StatusBadRequest,
			PathErr: fmt.Errorf("failed to parse path=[%s], Error:%s", dataPath, pErr),
		}
	}
	dataId, nextPath := Util.ParsePath(qPath)
	queryPath, err := Node.New(conn, dataType, dataId, nextPath, nil, nil)
	if err != nil {
		return nil, err
	}
	switch qCmd {
	case PathCmd.CmdSchema:
		return &CmdQuerySchema{
			p: queryPath,
		}, nil
	case PathCmd.CmdFlat:
		return &CmdQueryFlat{
			p: queryPath,
		}, nil
	case PathCmd.CmdRef:
		return &CmdQueryRef{
			p: queryPath,
		}, nil
	case PathCmd.CmdIter:
		return &CmdPathIterator{
			path: qPath,
			p:    queryPath,
		}, nil
	default:
		return &CmdQueryValue{
			p: queryPath,
		}, nil
	}
}
