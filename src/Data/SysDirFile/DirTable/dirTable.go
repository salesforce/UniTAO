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

package DirTable

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"Data/SysDirFile/FileRecord"
)

type Table struct {
	Name     string
	FullPath string
	records  map[string]*FileRecord.Record
}

func (tbl *Table) List() ([]map[string]interface{}, error) {
	idList, err := FileRecord.List(tbl.FullPath)
	if err != nil {
		return nil, err
	}
	result := make([]map[string]interface{}, 0, len(idList))
	for _, id := range idList {
		record, err := tbl.Get(*id)
		if err != nil {
			return nil, err
		}
		result = append(result, record)
	}
	return result, nil
}

func (tbl *Table) refresh() {
	for dataId := range tbl.records {
		record, err := FileRecord.New(tbl.FullPath, dataId)
		if err != nil {
			delete(tbl.records, dataId)
		}
		tbl.records[dataId] = record
	}
}

func (tbl *Table) Get(id string) (map[string]interface{}, error) {
	tbl.refresh()
	idList, err := FileRecord.List(tbl.FullPath)
	if err != nil {
		return nil, err
	}
	for _, dataId := range idList {
		if *dataId == id {
			_, ok := tbl.records[id]
			if !ok {
				record, err := FileRecord.New(tbl.FullPath, id)
				if err != nil {
					return nil, fmt.Errorf("failed to load data [%s] from table [%s] @path=[%s]", id, tbl.Name, tbl.FullPath)
				}
				tbl.records[id] = record
				return tbl.records[id].Data.(map[string]interface{}), nil
			}
			err = tbl.records[id].Refresh()
			if err != nil {
				return nil, err
			}
			return tbl.records[id].Data.(map[string]interface{}), nil
		}
	}
	return nil, nil
}

func (tbl *Table) Put(id string, payload map[string]interface{}) error {
	return FileRecord.Put(tbl.FullPath, id, payload)
}

func (tbl *Table) Delete(id string) error {
	idList, err := FileRecord.List(tbl.FullPath)
	if err != nil {
		return err
	}
	for _, fileId := range idList {
		if *fileId == id {
			return FileRecord.Delete(tbl.FullPath, id)
		}
	}
	return nil
}

func (tbl *Table) Exists() (bool, error) {
	pathInfo, err := os.Stat(tbl.FullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to stat path [%s], Err:%s", tbl.FullPath, err)
	}
	if !pathInfo.IsDir() {
		return false, fmt.Errorf("table path is not dir, path=[%s], Err:%s", tbl.FullPath, err)
	}
	return true, nil
}

func New(name string, path string) (*Table, error) {
	table := Table{
		Name:     name,
		FullPath: filepath.Join(path, name),
		records:  make(map[string]*FileRecord.Record),
	}
	hasTable, err := table.Exists()
	if err != nil {
		err = fmt.Errorf("failed to check table path exists. Error:%s", err)
		return nil, err
	}
	if !hasTable {
		err = fmt.Errorf("table does not exists, [path]=[%s]", table.FullPath)
		return nil, err
	}
	return &table, nil
}

func List(rootPath string) ([]interface{}, error) {
	result := []interface{}{}
	dirList, err := ioutil.ReadDir(rootPath)
	if err != nil {
		return nil, err
	}
	for _, file := range dirList {
		if file.IsDir() {
			dirName := file.Name()
			result = append(result, dirName)
		}
	}
	return result, nil
}

func Create(rootPath string, name string) error {
	dirPath := filepath.Join(rootPath, name)
	err := os.MkdirAll(dirPath, 0755)
	if err != nil {
		return fmt.Errorf("failed to create table [%s] at path [%s], Err:%s", name, rootPath, err)
	}
	return nil
}

func Delete(rootPath string, name string) error {
	dirPath := filepath.Join(rootPath, name)
	err := os.RemoveAll(dirPath)
	if err != nil {
		return err
	}
	return nil
}
