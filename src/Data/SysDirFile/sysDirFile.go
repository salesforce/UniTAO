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

package SysDirFile

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"Data/DbConfig"
	"Data/DbIface"
	"Data/SysDirFile/DirTable"

	"github.com/salesforce/UniTAO/lib/Schema/Record"
)

const (
	Name  = "sysdirfile"
	Index = "index"
)

type Database struct {
	logger *log.Logger
	Path   string
	config DbConfig.SysDirFileConfig
	tables map[string]*DirTable.Table
}

func (db *Database) Name() string {
	return Name
}

func (db *Database) ListTable() ([]interface{}, error) {
	result, err := DirTable.List(db.Path)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (db *Database) CreateTable(name string, data map[string]interface{}) error {
	tableList, err := DirTable.List(db.Path)
	if err != nil {
		return err
	}
	for _, tableName := range tableList {
		if tableName == name {
			return nil
		}
	}
	err = DirTable.Create(db.Path, name)
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) DeleteTable(name string) error {
	tableList, err := DirTable.List(db.Path)
	if err != nil {
		return err
	}
	for _, tableName := range tableList {
		if tableName == name {
			err = DirTable.Delete(db.config.Path, tableName.(string))
			if err != nil {
				return err
			}
			_, ok := db.tables[tableName.(string)]
			if ok {
				delete(db.tables, tableName.(string))
			}
			return nil
		}
	}
	return nil
}

func (db *Database) Get(queryArgs map[string]interface{}) ([]map[string]interface{}, error) {
	tableName, ok := queryArgs[DbIface.Table].(string)
	if !ok {
		err := fmt.Errorf("missing field [%s] from queryArgs", DbIface.Table)
		log.Print(err)
		return nil, err
	}
	table, err := db.GetTable(tableName)
	if err != nil {
		err = fmt.Errorf("failed to get table, Err: %s", err)
		return nil, err
	}
	dataId, ok := queryArgs[Record.DataId].(string)
	if !ok {
		result, err := table.List()
		if err != nil {
			err = fmt.Errorf("failed to list table [%s], error: %s", tableName, err)
			return nil, err
		}
		return result, nil
	}

	record, err := table.Get(dataId)
	if err != nil {
		err = fmt.Errorf("failed to get data [%s] from table [%s], Err:%s", dataId, tableName, err)
		return nil, err
	}
	var result []map[string]interface{}
	if record != nil {
		result = append(result, record)
	}
	return result, nil
}

func (db *Database) refresh() []error {
	errList := []error{}
	for name, tbl := range db.tables {
		exists, err := tbl.Exists()
		if err != nil {
			err = fmt.Errorf("failed to check table exists. [table]=[%s], [path]=[%s]", name, tbl.FullPath)
			errList = append(errList, err)
			delete(db.tables, name)
			continue
		}
		if !exists {
			delete(db.tables, name)
		}
	}
	return errList
}

func (db *Database) GetTable(name string) (*DirTable.Table, error) {
	db.refresh()
	table, ok := db.tables[name]
	if ok {
		return table, nil
	}
	tableList, err := db.ListTable()
	if err != nil {
		return nil, err
	}
	for _, tableName := range tableList {
		if tableName == name {
			tbl, err := DirTable.New(name, db.Path)
			if err != nil {
				return nil, fmt.Errorf("failed to get DirTable name=[%s] from path=[%s]", name, db.Path)
			}
			db.tables[name] = tbl
			return tbl, nil
		}
	}
	return nil, fmt.Errorf("table [%s] does not exists at path [%s]", name, db.Path)
}

func (db *Database) Create(table string, data interface{}) error {
	payload := data.(map[string]interface{})
	dataId, ok := payload[Record.DataId].(string)
	if !ok {
		err := fmt.Errorf("missing key [%s] from data", Record.DataId)
		log.Print(err)
		return err
	}
	tbl, err := db.GetTable(table)
	if err != nil {
		log.Print(err)
		return err
	}
	record, err := tbl.Get(dataId)
	if err != nil {
		return fmt.Errorf("failed to get data [%s]=[%s]. Error:%s", Record.DataId, dataId, err)
	}
	if record != nil {
		return fmt.Errorf("data [%s]=[%s] already exists", Record.DataId, dataId)
	}
	err = tbl.Put(dataId, payload)
	if err != nil {
		err = fmt.Errorf("failed to put data id=[%s], Error:%s", dataId, err)
		log.Print(err)
		return err
	}
	return nil
}
func (db *Database) Update(table string, keys map[string]interface{}, data interface{}) (map[string]interface{}, error) {
	return nil, nil
}
func (db *Database) Replace(table string, keys map[string]interface{}, data interface{}) error {
	payload := data.(map[string]interface{})
	dataId, ok := payload[Record.DataId].(string)
	if !ok {
		err := fmt.Errorf("missing key [%s] from data", Record.DataId)
		log.Print(err)
		return err
	}
	tbl, err := db.GetTable(table)
	if err != nil {
		log.Print(err)
		return err
	}
	err = tbl.Put(dataId, payload)
	if err != nil {
		err = fmt.Errorf("failed to put data id=[%s], Error:%s", dataId, err)
		log.Print(err)
		return err
	}
	return nil
}

func (db *Database) Delete(table string, keys map[string]interface{}) error {
	tbl, err := db.GetTable(table)
	if err != nil {
		log.Print(err)
		return err
	}
	dataId, ok := keys[Record.DataId].(string)
	if !ok {
		err = fmt.Errorf("missing key field [%s]", Record.DataId)
		return err
	}
	err = tbl.Delete(dataId)
	if err != nil {
		err = fmt.Errorf("failed to put data id=[%s], Err:%s", dataId, err)
		log.Print(err)
		return err
	}
	return nil
}

func Connect(config DbConfig.DatabaseConfig, logger *log.Logger) (DbIface.Database, error) {
	if logger == nil {
		logger = log.Default()
	}
	if config.SysDirFile.Path == "" {
		return nil, fmt.Errorf("missing path from config for %s", Name)
	}
	absPath, err := filepath.Abs(config.SysDirFile.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse path to absolute path, path=[%s], Error:%s", config.SysDirFile.Path, err)
	}
	pathInfo, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("path [%s] does not exists, Err: %s", absPath, err)
		}
		return nil, fmt.Errorf("failed to stat path [%s], Err:%s", config.SysDirFile.Path, err)
	}

	if !pathInfo.IsDir() {
		return nil, fmt.Errorf("inventory path is not dir, path=[%s], Err:%s", config.SysDirFile.Path, err)
	}
	db := Database{
		logger: logger,
		Path:   absPath,
		config: config.SysDirFile,
		tables: make(map[string]*DirTable.Table),
	}
	return &db, nil
}
