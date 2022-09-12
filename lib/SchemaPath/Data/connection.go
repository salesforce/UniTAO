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
	"github.com/salesforce/UniTAO/lib/SchemaPath/Error"
)

type SchemaFunction func(dataType string) (*SchemaDoc.SchemaDoc, *Error.SchemaPathErr)

type RecordFunction func(dataType string, dataId string) (*Record.Record, *Error.SchemaPathErr)

type Connection struct {
	FuncSchema SchemaFunction
	FuncRecord RecordFunction
	cache      map[string]TypeCache
}

type TypeCache struct {
	DataType string
	IdCache  map[string]interface{}
}

func (c *Connection) cacheData(dataType string, id string) (interface{}, *Error.SchemaPathErr) {
	if c.cache == nil {
		c.cache = map[string]TypeCache{}
	}
	if _, ok := c.cache[dataType]; !ok {
		c.cache[dataType] = TypeCache{
			DataType: dataType,
			IdCache:  make(map[string]interface{}),
		}
	}
	data, ok := c.cache[dataType].IdCache[id]
	if ok {
		return data, nil
	}
	var err *Error.SchemaPathErr
	switch dataType {
	case JsonKey.Schema:
		data, err = c.FuncSchema(id)
	default:
		data, err = c.FuncRecord(dataType, id)
	}
	if err != nil {
		return nil, err
	}
	c.cache[dataType].IdCache[id] = data
	return data, err
}

func (c *Connection) GetSchema(dataType string) (*SchemaDoc.SchemaDoc, *Error.SchemaPathErr) {
	if c.FuncSchema == nil {
		return nil, &Error.SchemaPathErr{
			Code:    http.StatusInternalServerError,
			PathErr: fmt.Errorf("field funcSchema is nil"),
		}
	}
	data, err := c.cacheData(JsonKey.Schema, dataType)
	if err != nil {
		return nil, err
	}
	schema, ok := data.(*SchemaDoc.SchemaDoc)
	if !ok {
		return nil, &Error.SchemaPathErr{
			Code:    http.StatusInternalServerError,
			PathErr: fmt.Errorf("function schema return invalid data. failed convert it to SchemaDoc.SchemaDoc. [type]=[%s]", dataType),
		}
	}
	return schema, nil
}

func (c *Connection) GetRecord(dataType string, dataId string) (*Record.Record, *Error.SchemaPathErr) {
	if c.FuncRecord == nil {
		return nil, &Error.SchemaPathErr{
			Code:    http.StatusInternalServerError,
			PathErr: fmt.Errorf("field funcRecord is nil"),
		}
	}
	data, err := c.cacheData(dataType, dataId)
	if err != nil {
		return nil, err
	}
	record, ok := data.(*Record.Record)
	if !ok {
		return nil, &Error.SchemaPathErr{
			Code:    http.StatusInternalServerError,
			PathErr: fmt.Errorf("function schema return invalid data. failed convert it to Record.Record. [type]=[%s], id=[%s]", dataType, dataId),
		}
	}
	return record, nil
}
