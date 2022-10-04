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

package DataLock

import (
	"time"

	"github.com/google/uuid"
	"github.com/salesforce/UniTAO/lib/Util"
)

type Lock struct {
	Handle     string
	DataType   string
	DataId     string
	DataPath   []string
	OwnerId    string
	Path       string
	ExpireTime int64
}

func ParsePathList(path string) []string {
	pathList := []string{}
	pathItem, path := Util.ParsePath(path)
	for pathItem != "" {
		pathList = append(pathList, pathItem)
		pathItem, path = Util.ParsePath(path)
	}
	return pathList
}

// Create Lock in lockHash
// returns:
//  - lockHandle: string
//  - baseTime: int
func NewLock(UserId string, dataPath string, lockTime int) *Lock {
	pathList := ParsePathList(dataPath)
	lock := Lock{
		Handle:     uuid.NewString(),
		DataType:   pathList[0],
		DataId:     pathList[1],
		DataPath:   pathList[2:],
		OwnerId:    UserId,
		Path:       dataPath,
		ExpireTime: time.Now().Unix() + int64(lockTime),
	}
	return &lock
}

func ListMatch(leftList []string, rightList []string) bool {
	for idx, item := range leftList {
		if idx == len(rightList) {
			break
		}
		if item != rightList[idx] {
			return false
		}
	}
	return true
}
