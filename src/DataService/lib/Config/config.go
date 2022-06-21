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

package Config

import (
	"UniTao/Data/DbConfig"
	"encoding/json"
	"fmt"
	"github.com/salesforce/UniTAO/lib/Util"
	"io/ioutil"
	"os"
)

const (
	DATABASE = "database"
	HTTP     = "http"
)

type Confuguration struct {
	Database DbConfig.DatabaseConfig `json:"database"`
	Http     Util.HttpConfig         `json:"http"`
	Inv      InvConfig               `json:"inventory"`
}

type InvConfig struct {
	Url string `json:"url"`
}

func Read(configPath string, config *Confuguration) error {
	jsonFile, err := os.Open(configPath)
	if err != nil {
		return fmt.Errorf("failed to open Config JSON file: [%s], err:%s", configPath, err)

	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal([]byte(byteValue), config)
	return nil
}
