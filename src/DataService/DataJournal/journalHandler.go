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
	intervalStep = 5
	maxInterval  = 60
)

type JournalHandler struct {
	Data        *DataHandler.Handler
	lib         *JournalLib
	OpsCtrl     *Thread.ThreadCtrl
	log         *log.Logger
	processList []ProcessIface.JournalProcess
}

func LoadProcesses(data *DataHandler.Handler, log *log.Logger) ([]ProcessIface.JournalProcess, error) {
	processList := []ProcessIface.JournalProcess{}
	schema, err := Process.NewSchemaProcess(data, log)
	if err != nil {
		return nil, err
	}
	log.Printf("process [%s] created", schema.Name())
	processList = append(processList, schema)
	cmtIdx, err := Process.NewCmtIndexProcess(data, log)
	if err != nil {
		return nil, err
	}
	log.Printf("process [%s] created", cmtIdx.Name())
	processList = append(processList, cmtIdx)
	return processList, nil
}

func (o *JournalHandler) Log(message string) {
	o.log.Printf("Journal Handler: %s", message)
}

func (o *JournalHandler) getProcessList() error {
	processList, err := LoadProcesses(o.Data, o.log)
	if err != nil {
		return err
	}
	o.processList = processList
	return nil
}

func NewJournalHandler(data *DataHandler.Handler, lib *JournalLib, logger *log.Logger) (*JournalHandler, error) {
	if logger == nil {
		logger = log.Default()
	}
	journal := JournalHandler{
		Data:    data,
		lib:     lib,
		OpsCtrl: Thread.NewThreadController(logger),
		log:     logger,
	}
	journal.Log("Created")
	err := journal.getProcessList()
	if err != nil {
		return nil, err
	}
	return &journal, nil
}

func (o *JournalHandler) Run(notify chan interface{}) error {
	interval := 0
	o.Log("Start Journal Handler Loop")
	for {
		select {
		case event := <-notify:
			o.Log("got an event")
			signal, ok := event.(os.Signal)
			if ok && signal == syscall.SIGINT {
				o.Log("exit signal")
				o.OpsCtrl.Broadcast(event)
				return nil
			}
			journalEvent, ok := event.(ProcessIface.JournalEvent)
			if ok {
				workerId := WorkId(journalEvent.DataType, journalEvent.DataId)
				o.Log(fmt.Sprintf("Journal Event [%s]", workerId))
				worker := o.OpsCtrl.GetWorker(workerId)
				if worker != nil {
					o.Log(fmt.Sprintf("found the worker, pass event to worker:[%s]", workerId))
					worker.Notify(event)
				} else {
					o.Log(fmt.Sprintf("worker not exists, create a new one. [%s]", workerId))
					added := o.AddJournalWorker(workerId, journalEvent.DataType, journalEvent.DataId)
					if !added {
						o.Log(fmt.Sprintf("failed to add Journal Worker [%s] from event, wait %d seconds for next loop to try again", workerId, interval))
					}
				}
			}
		case <-time.After(time.Duration(interval) * time.Second):
			if !o.ProcessAllJournals() {
				if interval < maxInterval {
					interval += intervalStep
				}
			} else {
				interval = intervalStep
			}
			o.Log(fmt.Sprintf("sleep for %d seconds", interval))
		}
	}
}

func (o *JournalHandler) ProcessAllJournals() bool {
	o.Log("start threads process all journals")
	hasChange := false
	dataTypeList := o.lib.ListJournalTypes()
	for _, dataType := range dataTypeList {
		o.Log(fmt.Sprintf("process journal for type:[%s]", dataType))
		if o.ProcessJournalByType(dataType) {
			hasChange = true
		}
	}
	return hasChange
}

func (o *JournalHandler) ProcessJournalByType(dataType string) bool {
	dataIdList := o.lib.ListJournalIds(dataType)
	hasChange := false
	for _, dataId := range dataIdList {
		o.Log(fmt.Sprintf("process journal for data[%s/%s]", dataType, dataId))
		if o.ProcessJournalById(dataType, dataId) {
			hasChange = true
		}
	}
	return hasChange
}

func (o *JournalHandler) ProcessJournalById(dataType string, dataId string) bool {
	workerId := WorkId(dataType, dataId)
	worker := o.OpsCtrl.GetWorker(workerId)
	if worker != nil {
		return false
	}
	added := o.AddJournalWorker(workerId, dataType, dataId)
	if !added {
		o.Log(fmt.Sprintf("failed to add Journal Worker [%s] from loop, wait for next loop to try again", workerId))
	}
	return true
}

func (o *JournalHandler) AddJournalWorker(workerId string, dataType string, dataId string) bool {
	jWorker := NewJournalWorker(o.lib, dataType, dataId, o.log, o.processList)
	o.Log(fmt.Sprintf("worker created: [%s]", workerId))
	worker, err := o.OpsCtrl.AddWorker(workerId, jWorker.Run)
	if err != nil {
		// TODO: log error and exit without throw anything.
		o.Log(fmt.Sprintf("failed to add workder[%s], Error:%s", workerId, err))
		return true
	}
	worker.Run()
	return true
}

func WorkId(dataType string, dataId string) string {
	return fmt.Sprintf("%s_%s", dataType, dataId)
}
