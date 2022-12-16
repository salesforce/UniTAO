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
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/salesforce/UniTAO/lib/SchemaPath/Data"
	"github.com/salesforce/UniTAO/lib/SchemaPath/PathCmd"
	"github.com/salesforce/UniTAO/lib/Util/Http"
)

const (
	PathName = "pathname"
)

type PathData struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

func LoadPathDataMap(data map[string]interface{}) (*PathData, *Http.HttpError) {
	raw, _ := json.Marshal(data)
	pData := PathData{}
	err := json.Unmarshal(raw, &pData)
	if err != nil {
		httpEr := Http.WrapError(err, "failed to load data as PathData", http.StatusBadRequest)
		return nil, httpEr
	}
	return &pData, nil
}

func IsCmdPathName(cmd string) bool {
	cmdPrefix := fmt.Sprintf("%s=", PathCmd.CmdPathName)
	return strings.HasPrefix(cmd, cmdPrefix)
}

func NewPathQuery(conn *Data.Connection, dataType string, idPath string, pathCmd string) (PathCmd.QueryIface, *Http.HttpError) {
	if !IsCmdPathName(pathCmd) {
		return nil, Http.NewHttpError(fmt.Sprintf("invalid pathCmd in Url, expect format [{dataType}/{dataId}%s={pathname}]", PathCmd.CmdPathName), http.StatusBadRequest)
	}
	prefixStr := fmt.Sprintf("%s=", PathCmd.CmdPathName)
	pathName := pathCmd[len(prefixStr):]
	pathRecord, err := conn.FuncRecord(PathName, pathName)
	if err != nil {
		return nil, err
	}
	pathData, e := LoadPathDataMap(pathRecord.Data)
	if e != nil {
		return nil, Http.NewHttpError(e.Error(), http.StatusInternalServerError)
	}
	var pathTemp string
	if strings.HasPrefix(pathData.Path, PathCmd.CmdPrefix) {
		pathTemp = "%s%s"
	} else {
		pathTemp = "%s/%s"
	}
	idPath = fmt.Sprintf(pathTemp, idPath, pathData.Path)
	qPath, e := CreateQuery(conn, dataType, idPath)
	if e != nil {
		newE := Http.WrapError(e, fmt.Sprintf("failed on create query with path [%s/%s]", dataType, idPath), e.Status)
		pathRecStr, err := json.MarshalIndent(pathRecord, "", "     ")
		if err != nil {
			newE.AppendError(e)
		} else {
			newE.Context = append(newE.Context, string(pathRecStr))
		}

		return nil, newE
	}
	return qPath, nil
}
