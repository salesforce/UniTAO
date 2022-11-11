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

package UtilTest

import (
	"testing"

	"github.com/salesforce/UniTAO/lib/Util"
)

func TestParseSimplePath(t *testing.T) {
	pathStr := "path1"
	current, next := Util.ParsePath(pathStr)
	if current != "path1" || next != "" {
		t.Fatal("failed to parse simple path")
	}
	pathStr = "/path1"
	current, next = Util.ParsePath(pathStr)
	if current != "path1" || next != "" {
		t.Fatal("failed to parse simple path")
	}
	pathStr = "//path1"
	current, next = Util.ParsePath(pathStr)
	if current != "path1" || next != "" {
		t.Fatal("failed to parse simple path")
	}
	pathStr = "path1/"
	current, next = Util.ParsePath(pathStr)
	if current != "path1" || next != "" {
		t.Fatal("failed to parse simple path")
	}
	pathStr = "path1//"
	current, next = Util.ParsePath(pathStr)
	if current != "path1" || next != "" {
		t.Fatal("failed to parse simple path")
	}
	pathStr = "/path1/"
	current, next = Util.ParsePath(pathStr)
	if current != "path1" || next != "" {
		t.Fatal("failed to parse simple path")
	}
	pathStr = "//path1//"
	current, next = Util.ParsePath(pathStr)
	if current != "path1" || next != "" {
		t.Fatal("failed to parse simple path")
	}
}

func TestParseStepPath(t *testing.T) {
	pathStr := "path1/path2/path3"
	current, next := Util.ParsePath(pathStr)
	if current != "path1" || next != "path2/path3" {
		t.Fatal("failed to parse two step path")
	}
	pathStr = "/path1/path2/path3"
	current, next = Util.ParsePath(pathStr)
	if current != "path1" || next != "path2/path3" {
		t.Fatal("failed to parse two step path")
	}
	pathStr = "//path1/path2/path3"
	current, next = Util.ParsePath(pathStr)
	if current != "path1" || next != "path2/path3" {
		t.Fatal("failed to parse two step path")
	}
	pathStr = "path1/path2/path3/"
	current, next = Util.ParsePath(pathStr)
	if current != "path1" || next != "path2/path3" {
		t.Fatal("failed to parse two step path")
	}
	pathStr = "path1/path2/path3//"
	current, next = Util.ParsePath(pathStr)
	if current != "path1" || next != "path2/path3" {
		t.Fatal("failed to parse two step path")
	}
	pathStr = "/path1///path2/path3"
	current, next = Util.ParsePath(pathStr)
	if current != "path1" || next != "//path2/path3" {
		t.Fatal("failed to parse two step path")
	}
	pathStr = "/path1/path2//path3"
	current, next = Util.ParsePath(pathStr)
	if current != "path1" || next != "path2//path3" {
		t.Fatal("failed to parse two step path")
	}
}
