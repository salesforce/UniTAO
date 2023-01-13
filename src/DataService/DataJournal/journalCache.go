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
package DataJournal

import (
	"DataService/DataJournal/ProcessIface"
	"fmt"
	"log"

	"github.com/salesforce/UniTAO/lib/Util/HashLock"
)

type JournalCache struct {
	log      *log.Logger
	DataType string
	DataId   string
	Head     *ProcessIface.JournalPage
	Tail     *ProcessIface.JournalPage
	Lock     *HashLock.ChanLock
}

func NewCache(dataType string, dataId string, logger *log.Logger) *JournalCache {
	if logger == nil {
		logger = log.Default()
	}
	cache := JournalCache{
		log:      logger,
		DataType: dataType,
		DataId:   dataId,
		Head:     ProcessIface.NewPage(dataType, dataId, -1),
		Tail:     ProcessIface.NewPage(dataType, dataId, -1),
		Lock:     HashLock.NewChanLock(logger),
	}
	return &cache
}

func (cache *JournalCache) Key() string {
	return fmt.Sprintf("journal[%s/%s]", cache.DataType, cache.DataId)
}

func (cache *JournalCache) ListPages() []string {
	pageCount := cache.Tail.Idx - cache.Head.Idx + 1
	pageList := make([]string, 0, pageCount)
	idx := cache.Head.Idx
	for idx <= cache.Tail.Idx {
		pageList = append(pageList, ProcessIface.PageId(cache.DataType, cache.DataId, idx))
	}
	return pageList
}
