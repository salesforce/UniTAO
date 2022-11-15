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
	defaultInterval = 10 * time.Second
)

type JournalHandler struct {
	lib        *JournalLib
	OpsCtrl    *Thread.ThreadCtrl
	interval   time.Duration
	log        *log.Logger
	processMap map[string][]ProcessIface.JournalProcess
}

func getProcesses(logger *log.Logger) []ProcessIface.JournalProcess {
	processList := []ProcessIface.JournalProcess{}
	schema := Process.NewSchemaProcess(logger)
	processList = append(processList, schema)
	return processList
}

func loadJournalProcesses(logger *log.Logger) map[string][]ProcessIface.JournalProcess {
	processMap := map[string][]ProcessIface.JournalProcess{}
	processList := getProcesses(logger)
	for _, p := range processList {
		typeMap := p.HandleTypes()
		for pType := range typeMap {
			if _, ok := processMap[pType]; !ok {
				processMap[pType] = []ProcessIface.JournalProcess{}
			}
			processMap[pType] = append(processMap[pType], p)
		}
	}
	return processMap
}

func NewJournalHandler(lib *JournalLib, interval time.Duration, logger *log.Logger) *JournalHandler {
	if interval == 0 {
		interval = defaultInterval
	}
	if logger == nil {
		logger = log.Default()
	}
	return &JournalHandler{
		lib:        lib,
		OpsCtrl:    Thread.NewThreadController(),
		interval:   interval,
		log:        logger,
		processMap: loadJournalProcesses(logger),
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
	processList, ok := o.processMap[dataType]
	if !ok {
		processList = []ProcessIface.JournalProcess{}
	}
	if worker == nil {
		jWorker := JournalWorker{
			lib:       o.lib,
			dataType:  dataType,
			dataId:    dataId,
			log:       o.log,
			processes: processList,
		}
		worker, err := o.OpsCtrl.AddWorker(workerId, jWorker.Run)
		if err != nil {
			// TODO: log error and exit without throw anything.
			return
		}
		worker.Run()
		return
	}
	if worker.Stopped() {
		worker.Run()
	}
}
