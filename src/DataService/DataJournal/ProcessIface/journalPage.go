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

// functions to record all data changes
package ProcessIface

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/salesforce/UniTAO/lib/Util"
)

const (
	KeyDataId   = "dataId"
	KeyDataType = "dataType"
	KeyPage     = "page"
)

type JournalEntry struct {
	Page   int                    `json:"page"`
	Idx    int                    `json:"idx"`
	Time   string                 `json:"time"`
	Before map[string]interface{} `json:"before"`
	After  map[string]interface{} `json:"after"`
}

type JournalPage struct {
	DataType string          `json:"dataType"`
	DataId   string          `json:"dataId"`
	Idx      int             `json:"idx"`
	Active   []*JournalEntry `json:"active"`
	Archived []*JournalEntry `json:"archived"`
}

func NewPage(dataType string, dataId string, idx int) *JournalPage {
	page := JournalPage{
		DataType: dataType,
		DataId:   dataId,
		Idx:      idx,
		Active:   []*JournalEntry{},
		Archived: []*JournalEntry{},
	}
	return &page
}

func (page *JournalPage) Id() string {
	return PageId(page.DataType, page.DataId, page.Idx)
}

func (page *JournalPage) LoadMap(data map[string]interface{}) error {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return page.LoadStr(string(dataBytes))
}

func (page *JournalPage) LoadStr(pageStr string) error {
	err := json.Unmarshal([]byte(pageStr), page)
	if err != nil {
		return err
	}
	return nil
}

func (page *JournalPage) LastEntry() int {
	return len(page.Active) + len(page.Archived)
}

func PageId(dataType string, dataId string, idx int) string {
	return fmt.Sprintf("%s:%s_%s:%s_%s:%d", KeyDataType, dataType, KeyDataId, dataId, KeyPage, idx)
}

func ParseJournalId(journalId string) (string, string, int, error) {
	typeTag := fmt.Sprintf("%s:", KeyDataType)
	idTag := fmt.Sprintf("_%s:", KeyDataId)
	pageTag := fmt.Sprintf("_%s:", KeyPage)
	if !strings.HasPrefix(journalId, typeTag) {
		return "", "", 0, fmt.Errorf("invalid journalId=[%s], missing prefix=[%s]", journalId, KeyDataType)
	}
	typeStr, idStr := Util.ParseCustomPath(journalId, idTag)
	dataType := typeStr[len(typeTag):]
	if idStr == "" {
		if strings.Contains(dataType, pageTag) {
			return "", "", 0, fmt.Errorf("invalid journalId=[%s], missing tag=[%s]", journalId, idTag)
		}
		return dataType, "", 0, nil
	}
	dataId, pageStr := Util.ParseCustomPath(idStr, pageTag)
	if pageStr == "" {
		return dataType, dataId, 0, nil
	}
	idx, err := strconv.Atoi(pageStr)
	if err != nil {
		return "", "", 0, fmt.Errorf("invalid journalId=[%s], failed to convert pageNo=[%s] to int, Error:%s", journalId, pageStr, err)
	}
	return dataType, dataId, idx, nil
}
