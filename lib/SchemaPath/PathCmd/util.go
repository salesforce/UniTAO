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

package PathCmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/salesforce/UniTAO/lib/Util/Http"
)

var CmdList = []string{CmdRef, CmdFlat, CmdSchema, CmdValue, CmdIter, CmdPathName}

func Parse(path string) (string, string, *Http.HttpError) {
	if strings.HasSuffix(path, CmdFlatPath) {
		qPath := path[:len(path)-len(CmdFlatPath)]
		return qPath, CmdRef, nil
	}
	qIdx := strings.Index(path, CmdPrefix)
	if qIdx < 0 {
		return path, CmdValue, nil
	}
	qPath := path[:qIdx]
	qCmd := path[qIdx:]
	dupIdx := strings.Index(qCmd[1:], CmdPrefix)
	if dupIdx > -1 {
		return "", "", Http.NewHttpError(fmt.Sprintf("invalid format of PathCmd, more than 1 ? in path. path=[%s]", path), http.StatusBadRequest)

	}
	err := Validate(qCmd)
	if err != nil {
		return "", "", err

	}
	return qPath, qCmd, nil
}

func Validate(cmd string) *Http.HttpError {
	for _, c := range CmdList {
		if c == cmd {
			return nil
		}
	}
	if strings.HasPrefix(cmd, fmt.Sprintf("%s=", CmdPathName)) {
		return nil
	}
	e := Http.NewHttpError(fmt.Sprintf("unknown path cmd=[%s]", cmd), http.StatusBadRequest)
	cmdListStr, _ := json.MarshalIndent(CmdList, "", "     ")
	e.Context = append(e.Context, fmt.Sprintf("available options\n%s", cmdListStr))
	return e
}
