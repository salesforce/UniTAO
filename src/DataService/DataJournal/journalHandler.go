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
	"DataService/DataHandler"
	"DataService/DataJournal/Process"
	"DataService/DataJournal/ProcessIface"
	"fmt"
	"log"
	"os"
	"syscall"
	"time"

	"github.com/salesforce/UniTAO/lib/Util/Thread"
)

const (
	defaultInterval = 60 * time.Second
)

type JournalHandler struct {
	Data        *DataHandler.Handler
	lib         *JournalLib
	OpsCtrl     *Thread.ThreadCtrl
	interval    time.Duration
	log         *log.Logger
	processList []ProcessIface.JournalProcess
}

func LoadProcesses(data *DataHandler.Handler, log *log.Logger) ([]ProcessIface.JournalProcess, error) {
	processList := []ProcessIface.JournalProcess{}
	schema, err := Process.NewSchemaProcess(data, log)
	if err != nil {
		return nil, err
	}
	processList = append(processList, schema)
	cmtIdx, err := Process.NewCmtIndexProcess(data, log)
	if err != nil {
		return nil, err
	}
	processList = append(processList, cmtIdx)
	return processList, nil
}

func (o *JournalHandler) getProcessList() error {
	processList, err := LoadProcesses(o.Data, o.log)
	if err != nil {
		return err
	}
	o.processList = processList
	return nil
}

func NewJournalHandler(data *DataHandler.Handler, lib *JournalLib, interval time.Duration, logger *log.Logger) (*JournalHandler, error) {
	if interval == 0 {
		interval = defaultInterval
	}
	if logger == nil {
		logger = log.Default()
	}
	journal := JournalHandler{
		Data:     data,
		lib:      lib,
		OpsCtrl:  Thread.NewThreadController(logger),
		interval: interval,
		log:      logger,
	}
	err := journal.getProcessList()
	if err != nil {
		return nil, err
	}
	return &journal, nil
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
		case <-time.After(o.interval):
			o.ProcessAllJournals()
		}
	}
}

func (o *JournalHandler) ProcessAllJournals() {
	o.log.Print("start threads process all journals")
	dataTypeList := o.lib.ListJournalTypes()
	for _, dataType := range dataTypeList {
		o.ProcessJournalByType(dataType)
	}
}

func (o *JournalHandler) ProcessJournalByType(dataType string) {
	dataIdList := o.lib.ListJournalIds(dataType)
	for _, dataId := range dataIdList {
		o.ProcessJournalById(dataType, dataId)
	}
}

func (o *JournalHandler) ProcessJournalById(dataType string, dataId string) {
	workerId := fmt.Sprintf("%s_%s", dataType, dataId)
	worker := o.OpsCtrl.GetWorker(workerId)
	if worker != nil {
		return
	}
	jWorker := JournalWorker{
		lib:       o.lib,
		dataType:  dataType,
		dataId:    dataId,
		log:       o.log,
		processes: o.processList,
	}
	worker, err := o.OpsCtrl.AddWorker(workerId, jWorker.Run)
	if err != nil {
		// TODO: log error and exit without throw anything.
		return
	}
	worker.Run()
}