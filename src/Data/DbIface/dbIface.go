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

package DbIface

import "fmt"

const (
	DbType  = "type"
	Table   = "table"
	command = "command"
	payload = "payload"
)

type UpdateCommand struct {
	Command string
	Payload map[string]interface{}
}

type Database interface {
	ListTable() ([]*string, error)
	CreateTable(name string, data map[string]interface{}) error
	DeleteTable(name string) error
	Get(queryArgs map[string]interface{}) ([]map[string]interface{}, error)
	Create(table string, data interface{}) error
	Update(table string, keys map[string]interface{}, cmd UpdateCommand) (map[string]interface{}, error)
	Replace(table string, keys map[string]interface{}, data interface{}) error
	Delete(table string, keys map[string]interface{}) error
}

func CreateUpdateCommand(data map[string]interface{}) (UpdateCommand, error) {
	update := UpdateCommand{}
	cmd, ok := data[command].(string)
	if !ok {
		return update, fmt.Errorf("missing field %s", command)
	}
	update.Command = cmd
	payload, ok := data[payload].(map[string]interface{})
	if !ok {
		return update, fmt.Errorf("missing field %s", payload)
	}
	update.Payload = payload
	return update, nil
}
