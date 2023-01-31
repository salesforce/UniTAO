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
	"fmt"
	"net/http"

	"github.com/salesforce/UniTAO/lib/Schema/JsonKey"
	"github.com/salesforce/UniTAO/lib/Schema/Record"
	"github.com/salesforce/UniTAO/lib/Schema/SchemaDoc"
	"github.com/salesforce/UniTAO/lib/Util/Http"
	"github.com/salesforce/UniTAO/lib/Util/Json"
)

type RecordFunction func(dataType string, dataId string) (*Record.Record, *Http.HttpError)

type Connection struct {
	FuncRecord RecordFunction
	cache      map[string]TypeCache
}

type TypeCache struct {
	DataType string
	IdCache  map[string]interface{}
}

func (c *Connection) cacheData(dataType string, id string) (interface{}, *Http.HttpError) {
	if c.cache == nil {
		c.cache = map[string]TypeCache{}
	}
	if _, ok := c.cache[dataType]; !ok {
		c.cache[dataType] = TypeCache{
			DataType: dataType,
			IdCache:  make(map[string]interface{}),
		}
	}
	if dataType == JsonKey.Schema {
		schemaId, schemaVer, ex := SchemaDoc.ParseDataType(id)
		if ex != nil {
			return nil, Http.WrapError(ex, fmt.Sprintf("failed to parse data type[%s]", id), http.StatusBadRequest)
		}
		if schemaVer != "" {
			id = SchemaDoc.ArchivedSchemaId(schemaId, schemaVer)
		}
	}
	data, ok := c.cache[dataType].IdCache[id]
	if ok {
		dataCopy, ex := Json.Copy(data)
		if ex != nil {
			return nil, Http.WrapError(ex, "failed to copy cache data", http.StatusInternalServerError)
		}
		recordCopy, _ := Record.LoadMap(dataCopy.(map[string]interface{}))
		return recordCopy, nil
	}
	data, err := c.FuncRecord(dataType, id)
	if err != nil {
		return nil, err
	}
	dataCopy, ex := Json.Copy(data)
	if ex != nil {
		return nil, Http.WrapError(ex, "failed to copy cache data", http.StatusInternalServerError)
	}
	c.cache[dataType].IdCache[id] = dataCopy
	return data, err
}

func (c *Connection) GetRecord(dataType string, dataId string) (*Record.Record, *Http.HttpError) {
	if c.FuncRecord == nil {
		return nil, Http.NewHttpError("field funcRecord is nil", http.StatusInternalServerError)
	}
	data, err := c.cacheData(dataType, dataId)
	if err != nil {
		return nil, err
	}
	record, ok := data.(*Record.Record)
	if !ok {
		return nil, Http.NewHttpError(fmt.Sprintf("function GetRecord return invalid data. failed convert it to Record.Record. [type]=[%s], id=[%s]", dataType, dataId), http.StatusInternalServerError)
	}
	return record, nil
}
