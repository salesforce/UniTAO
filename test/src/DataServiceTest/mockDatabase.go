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
	"encoding/json"
	"fmt"

	"github.com/salesforce/UniTAO/lib/Schema/Record"
)

type MockDatabase struct {
	config DbConfig.DatabaseConfig
	Data   map[string]interface{}
}

func NewMockDb(config DbConfig.DatabaseConfig, dataStr string) (*MockDatabase, error) {
	data := map[string]interface{}{}
	err := json.Unmarshal([]byte(dataStr), &data)
	if err != nil {
		return nil, err
	}
	db := MockDatabase{
		config: config,
		Data:   data,
	}
	return &db, nil
}

func (db MockDatabase) Create(table string, data interface{}) error {
	record, err := Record.LoadMap(data.(map[string]interface{}))
	if err != nil {
		return err
	}
	keys := map[string]interface{}{
		Record.DataType: record.Type,
		Record.DataId:   record.Id,
	}
	err = db.Replace("", keys, record.Map())
	if err != nil {
		return err
	}
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
	dataType, ok := queryArgs[Record.DataType].(string)
	if !ok {
		return nil, fmt.Errorf("invalid queryArgs. missing=[%s]", Record.DataType)
	}
	typeMap, ok := db.Data[dataType].(map[string]interface{})
	if !ok {
		return []map[string]interface{}{}, nil
	}
	dataId, ok := queryArgs[Record.DataId].(string)
	if !ok {
		result := make([]map[string]interface{}, 0, len(typeMap))
		for _, data := range typeMap {
			result = append(result, data.(map[string]interface{}))
		}
		return result, nil
	}
	data, ok := typeMap[dataId].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("type=[%s] id=[%s] not found", dataType, dataId)
	}
	return []map[string]interface{}{data}, nil
}

func (db MockDatabase) Update(table string, keys map[string]interface{}, data interface{}) (map[string]interface{}, error) {
	return nil, nil
}
func (db MockDatabase) Replace(table string, keys map[string]interface{}, data interface{}) error {
	dataType, ok := keys[Record.DataType].(string)
	if !ok {
		return fmt.Errorf("invalid queryArgs. missing=[%s]", Record.DataType)
	}
	dataId, ok := keys[Record.DataId].(string)
	if !ok {
		return fmt.Errorf("invalid queryArgs. missing=[%s]", Record.DataId)
	}
	typeMap, ok := db.Data[dataType].(map[string]interface{})
	if !ok {
		typeMap = map[string]interface{}{}
		db.Data[dataType] = typeMap
	}
	typeMap[dataId] = data
	return nil
}

func (db MockDatabase) Delete(table string, keys map[string]interface{}) error {
	dataType, ok := keys[Record.DataType].(string)
	if !ok {
		return fmt.Errorf("invalid queryArgs. missing=[%s]", Record.DataType)
	}
	dataId, ok := keys[Record.DataId].(string)
	if !ok {
		return fmt.Errorf("invalid queryArgs. missing=[%s]", Record.DataId)
	}
	typeMap, ok := db.Data[dataType].(map[string]interface{})
	if !ok {
		return nil
	}
	delete(typeMap, dataId)
	return nil
}
