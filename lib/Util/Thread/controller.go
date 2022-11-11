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
	"fmt"
	"sync"

	"github.com/google/uuid"
)

type ThreadCtrl struct {
	wg      *sync.WaitGroup
	workers map[string]*Worker
	lock    sync.Mutex
}

func NewThreadController() *ThreadCtrl {
	c := ThreadCtrl{
		wg:      &sync.WaitGroup{},
		workers: map[string]*Worker{},
		lock:    sync.Mutex{},
	}
	return &c
}

func (c *ThreadCtrl) AddWorker(workerId string, funcToCall func(chan interface{})) (*Worker, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if workerId == "" {
		workerId = uuid.NewString()
	}
	if _, ok := c.workers[workerId]; ok {
		return nil, fmt.Errorf("worker %s already exists", workerId)
	}
	worker := NewWorker(workerId, c.wg, funcToCall)
	c.workers[workerId] = worker
	return worker, nil
}

func (c *ThreadCtrl) GetWorker(workerId string) *Worker {
	worker, ok := c.workers[workerId]
	if !ok {
		return nil
	}
	return worker
}

func (c *ThreadCtrl) Broadcast(event interface{}) {
	for _, worker := range c.workers {
		worker.Notify(event)
	}
}
