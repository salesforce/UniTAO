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
	"fmt"
	"os"
	"syscall"
	"time"

	"github.com/salesforce/UniTAO/lib/Util/Thread"
)

const (
	defaultInterval = 10 * time.Second
)

type JournalHandler struct {
	lib      *JournalLib
	OpsCtrl  *Thread.ThreadCtrl
	interval time.Duration
}

func NewJournalHandler(lib *JournalLib, interval time.Duration) *JournalHandler {
	if interval == 0 {
		interval = defaultInterval
	}
	return &JournalHandler{
		lib:      lib,
		OpsCtrl:  Thread.NewThreadController(),
		interval: interval,
	}
}

func (o *JournalHandler) Run(notify chan interface{}) {
	for {
		select {
		case event := <-notify:
			o.OpsCtrl.Broadcast(event)
			signal, ok := event.(os.Signal)
			if ok && signal == syscall.SIGINT {
				return
			}
		default:
			o.ProcessAllJournals()
			time.Sleep(o.interval)
		}
	}
}

func (o *JournalHandler) ProcessAllJournals() {
	dataTypeList, err := o.lib.ListJournalTypes()
	if err != nil {
		// TODO: log error and exit without throw anything.
		return
	}
	for _, dataType := range dataTypeList {
		o.ProcessJournalByType(dataType)
	}
}

func (o *JournalHandler) ProcessJournalByType(dataType string) {
	dataIdList, err := o.lib.ListJournalIds(dataType)
	if err != nil {
		// TODO: log error and exit without throw anything.
		return
	}
	for _, dataId := range dataIdList {
		o.ProcessJournalById(dataType, dataId)
	}
}

func (o *JournalHandler) ProcessJournalById(dataType string, dataId string) {
	workerId := fmt.Sprintf("%s_%s", dataType, dataId)
	worker := o.OpsCtrl.GetWorker(workerId)
	if worker == nil {
		jWorker := JournalWorker{
			lib:      o.lib,
			dataType: dataType,
			dataId:   dataId,
		}
		worker, err := o.OpsCtrl.AddWorker(workerId, jWorker.Run)
		if err != nil {
			// TODO: log error and exit without throw anything.
			return
		}
		worker.Run()
	}
}
