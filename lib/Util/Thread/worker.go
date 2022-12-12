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

package Thread

import (
	"log"
	"sync"
	"syscall"
	"time"
)

const (
	DefaultInt = 5 * time.Second // by default sleep 5 second if error
)

type Worker struct {
	Id       string
	wg       *sync.WaitGroup
	stopped  bool
	interval time.Duration
	log      *log.Logger
	event    chan interface{}
	run      func(chan interface{})
	cleanup  func(id string) error
}

func NewWorker(workerId string, interval time.Duration, logger *log.Logger, run func(chan interface{}), cleanup func(id string) error) *Worker {
	if interval <= 0 {
		interval = DefaultInt
	}
	if logger == nil {
		logger = log.Default()
	}
	worker := Worker{
		Id:       workerId,
		wg:       &sync.WaitGroup{},
		interval: interval,
		log:      logger,
		event:    make(chan interface{}),
		run:      run,
		cleanup:  cleanup,
	}
	return &worker
}

func (w *Worker) setup() {
	w.stopped = false
	w.wg.Add(1)
}

func (w *Worker) postRun() {
	for {
		err := w.cleanup(w.Id)
		if err == nil {
			return
		}
		w.log.Printf("failed to run cleanup function, sleep and try again. Error: %s", err)
		time.Sleep(w.interval)
	}
}

func (w *Worker) workerRoutine() {
	w.setup()
	defer w.postRun()
	w.run(w.event)
}

func (w *Worker) Run() {
	go w.workerRoutine()
}

func (w *Worker) Notify(event interface{}) {
	w.event <- event
}

func (w *Worker) Stop() {
	w.event <- syscall.SIGINT
}
