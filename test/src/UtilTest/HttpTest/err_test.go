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

package HttpErrorTest

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/salesforce/UniTAO/lib/Util/Http"
)

func TestIsHttpError(t *testing.T) {
	err := fmt.Errorf("test")
	if Http.IsHttpError(err) {
		t.Fatalf("should not be a HttpError")
	}
	httpErr := Http.NewHttpError("test", http.StatusBadRequest)
	if !Http.IsHttpError(httpErr) {
		t.Fatalf("fail to detect HttpError")
	}
}

func TestNewHttpErr(t *testing.T) {
	msg := `123
	234`
	err := Http.NewHttpError(msg, http.StatusOK)
	if err.Status != http.StatusOK {
		t.Fatalf("created the wrong err. status=[%d], expect [%d]", err.Status, http.StatusOK)
	}
	if len(err.Message) != 2 {
		t.Fatalf("invalid split err msg. line number [%d] expect [2]", len(err.Message))
	}
}

func TestAppendErr(t *testing.T) {
	err01 := Http.NewHttpError("test", http.StatusOK)
	err02 := Http.NewHttpError("testTab", http.StatusOK)
	Http.AppendError(err01, err02)
	if len(err01.Message) != 2 {
		t.Fatalf("failed to append message to 2")
	}
	if len(err01.Context) != 1 {
		t.Fatalf("failed to append err02 into Context")
	}
}

func TestWrapErr(t *testing.T) {
	err01 := fmt.Errorf("test01")
	err := Http.WrapError(err01, "wrapTest", http.StatusBadRequest)
	if len(err.Message) != 2 {
		t.Fatalf("failed to wrap message as 2 line")
	}
	if err.Message[0] != "wrapTest" {
		t.Fatalf("failed to add new title message")
	}
}
