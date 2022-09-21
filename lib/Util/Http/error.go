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

package Http

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/salesforce/UniTAO/lib/Util"
)

const TAB = "    "

type HttpError struct {
	Status  int           `json:"httpStatus"`
	Message []string      `json:"message"`
	Code    int           `json:"code"`
	Context []interface{} `json:"context"`
	Payload interface{}   `json:"payload"`
}

func (e HttpError) Error() string {
	errTxtBytes, err := json.MarshalIndent(e, "", "    ")
	if err != nil {
		newErr := HttpError{
			Status: http.StatusInternalServerError,
			Message: []string{
				"failed to parse HttpError to string",
				"Error:",
			},
			Code: e.Status,
		}

		return newErr.Error()
	}
	return string(errTxtBytes)
}

func AppendError(srcErr *HttpError, err *HttpError) {
	tabErrMessage := Util.PrefixStrLst(err.Message, TAB)
	srcErr.Message = append(srcErr.Message, tabErrMessage...)
	srcErr.Context = append(srcErr.Context, err)
}

func IsHttpError(err error) bool {
	_, ok := err.(*HttpError)
	return ok
}

func NewHttpError(msg string, status int) *HttpError {
	return &HttpError{
		Status:  status,
		Message: strings.Split(msg, "\n"),
		Context: []interface{}{},
	}
}

func WrapError(err error, newMsg string, newStatus int) *HttpError {
	newErr := NewHttpError(newMsg, newStatus)
	if !IsHttpError(err) {
		AppendError(newErr, NewHttpError(err.Error(), http.StatusInternalServerError))
	} else {
		AppendError(newErr, err.(*HttpError))
	}
	return newErr
}
