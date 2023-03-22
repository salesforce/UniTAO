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

package DbConfig

type DatabaseConfig struct {
	DbType     string           `json:"type"`
	Dynamodb   DynmoDbConfig    `json:"dynamodb"`
	Mongodb    MongoDbConfig    `json:"mongodb"`
	SysDirFile SysDirFileConfig `json:"sysdirfile"`
}

type DynmoDbConfig struct {
	Region      string `json:"region"`
	EndPoint    string `json:"endpoint"`
	AccessKey   string `json:"ACCESS_KEY"`
	SecretKey   string `json:"SECRET_KEY"`
	AccessToken string `json:"ACCESS_TOKEN"`
}

type MongoDbConfig struct {
	EndPoint string `json:"endpoint"`
	Database string `json:"database"`
	UserName string `json:"user"`
	Password string `json:"password"`
}

type SysDirFileConfig struct {
	Path string `json:"path"`
}
