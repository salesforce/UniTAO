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

package DataServiceTest

import (
	"Data/DbConfig"
	"Data/DbIface"
)

type MockDatabase struct {
	config DbConfig.DatabaseConfig
	get    func(queryArgs map[string]interface{}) ([]map[string]interface{}, error)
}

func (db MockDatabase) Create(table string, data interface{}) error {
	return nil
}

func (db MockDatabase) CreateTable(name string, data map[string]interface{}) error {
	return nil
}

func (db MockDatabase) ListTable() ([]*string, error) {
	return nil, nil
}

func (db MockDatabase) DeleteTable(name string) error {
	return nil
}
func (db MockDatabase) Get(queryArgs map[string]interface{}) ([]map[string]interface{}, error) {
	if db.get != nil {
		return db.get(queryArgs)
	}
	return nil, nil
}

func (db MockDatabase) Update(table string, keys map[string]interface{}, cmd DbIface.UpdateCommand) (map[string]interface{}, error) {
	return nil, nil
}
func (db MockDatabase) Replace(table string, keys map[string]interface{}, data interface{}) error {
	return nil
}
func (db MockDatabase) Delete(table string, keys map[string]interface{}) error {
	return nil
}
