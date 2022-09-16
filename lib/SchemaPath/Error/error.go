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

package Error

import (
	"fmt"
	"net/http"
)

type SchemaPathErr struct {
	Code    int
	PathErr error
}

func IsSchemaPathErr(err error) bool {
	_, ok := err.(*SchemaPathErr)
	return ok
}

func (e *SchemaPathErr) Error() string {
	return fmt.Sprintf("Code=[%d], Error: %s", e.Code, e.PathErr)
}

func AppendErr(err error, newMsg string) *SchemaPathErr {
	newErr := fmt.Errorf("%s Error: %s", newMsg, err)
	if !IsSchemaPathErr(err) {
		return &SchemaPathErr{
			Code:    http.StatusInternalServerError,
			PathErr: newErr,
		}
	}
	return &SchemaPathErr{
		Code:    err.(*SchemaPathErr).Code,
		PathErr: newErr,
	}
}
