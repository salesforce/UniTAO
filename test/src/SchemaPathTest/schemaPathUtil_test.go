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

package SchemaPathTest

import (
	"fmt"
	"testing"

	"github.com/salesforce/UniTAO/lib/SchemaPath/PathCmd"
)

func TestPathParseCmdAll(t *testing.T) {
	for _, cmd := range []string{PathCmd.CmdFlat, PathCmd.CmdIter, PathCmd.CmdRef, PathCmd.CmdSchema, PathCmd.CmdValue} {
		path := fmt.Sprintf("/test/123%s", cmd)
		qPath, qCmd, err := PathCmd.Parse(path)
		if err != nil {
			t.Errorf("failed to parse path=[%s], Error:%s", path, err)
		}
		if qCmd != cmd {
			t.Errorf("invalid cmd parsed [%s], expect [%s]", qCmd, cmd)
		}
		if qPath != "/test/123" {
			t.Errorf("invalid path parsed [%s], expect [/test/123]", qPath)
		}
	}

}

func TestPathParseRefPath(t *testing.T) {
	path := "/test/123/$"
	qPath, qCmd, err := PathCmd.Parse(path)
	if err != nil {
		t.Errorf("failed to parse path=[%s], Error:%s", path, err)
	}
	if qCmd != PathCmd.CmdRef {
		t.Errorf("invalid cmd parsed [%s], expect [%s]", qCmd, PathCmd.CmdRef)
	}
	if qPath != "/test/123" {
		t.Errorf("invalid path parsed [%s], expect [/test/123]", qPath)
	}
}

func TestPathParseValueDefault(t *testing.T) {
	path := "/test/123"
	qPath, qCmd, err := PathCmd.Parse(path)
	if err != nil {
		t.Errorf("failed to parse path=[%s], Error:%s", path, err)
	}
	if qCmd != PathCmd.CmdValue {
		t.Errorf("invalid cmd parsed [%s], expect [%s]", qCmd, PathCmd.CmdValue)
	}
	if qPath != "/test/123" {
		t.Errorf("invalid path parsed [%s], expect [/test/123]", qPath)
	}
}
