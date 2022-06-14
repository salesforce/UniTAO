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

package Data

import (
	"UniTao/Data/DbConfig"
	"UniTao/Data/DbDynamoDb"
	"UniTao/Data/DbIface"
	"UniTao/Data/SysDirFile"
	"fmt"
	"log"
)

func ConnectDb(config DbConfig.DatabaseConfig) (DbIface.Database, error) {
	switch config.DbType {
	case DbDynamoDb.Name:
		db, err := DbDynamoDb.Connect(config)
		if err != nil {
			newError := fmt.Errorf("failed to connect to DynamoDB. Error:%s", err.Error())
			return nil, newError
		}
		log.Printf("dynamodb connected")
		return db, nil
	case SysDirFile.Name:
		db, err := SysDirFile.Connect(config)
		if err != nil {
			err = fmt.Errorf("failed to connect to SysDirFile. Error:%s", err)
			return nil, err
		}
		return db, nil
	default:
		return nil, fmt.Errorf("unknown dbType:%s, Don't know how to connect", config.DbType)
	}
}
