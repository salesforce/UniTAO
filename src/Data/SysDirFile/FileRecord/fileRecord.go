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

package FileRecord

import (
	"encoding/json"
	"fmt"
	"github.com/salesforce/UniTAO/lib/Util"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
)

type Record struct {
	Id             string
	FullPath       string
	lastUpdateTime time.Time
	Data           interface{}
}

func (rec *Record) Refresh() error {
	fileInfo, err := os.Stat(rec.FullPath)
	if err != nil {
		err = fmt.Errorf("file record id=[%s] is not a file @[%s]", rec.Id, rec.FullPath)
		log.Print(err)
		return err
	}
	if fileInfo.Mode().IsDir() {
		err = fmt.Errorf("file record id=[%s] is dir @[%s]", rec.Id, rec.FullPath)
		log.Print(err)
		return err
	}
	fileTime := fileInfo.ModTime()
	if rec.lastUpdateTime.Before(fileTime) {
		data, err := Util.LoadJSONMap(rec.FullPath)
		if err != nil {
			err = fmt.Errorf("failed to parse data from record id=[%s], path=[%s]. Err:%s", rec.Id, rec.FullPath, err)
			log.Print(err)
			return err
		}
		rec.Data = data
		rec.lastUpdateTime = fileTime
	}
	return nil
}

func List(dirPath string) ([]*string, error) {
	result := []*string{}
	fileList, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}
	for _, file := range fileList {
		if !file.IsDir() {
			fileName := file.Name()
			result = append(result, &fileName)
		}
	}
	return result, nil
}

func New(dirPath string, fileName string) (*Record, error) {
	fullPath := filepath.Join(dirPath, fileName)
	record := Record{
		Id:       fileName,
		FullPath: fullPath,
	}
	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		return nil, fmt.Errorf("%s is not a file. \n Error:%s", fullPath, err)
	}
	record.lastUpdateTime = fileInfo.ModTime()
	record.Data, err = Util.LoadJSONMap(fullPath)
	return &record, nil
}

func Put(dirPath string, fileName string, data map[string]interface{}) error {
	filePath := filepath.Join(dirPath, fileName)
	file, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filePath, file, 0644)
	if err != nil {
		return err
	}
	return nil
}

func Delete(dirPath string, fileName string) error {
	filePath := filepath.Join(dirPath, fileName)
	return os.Remove(filePath)

}
