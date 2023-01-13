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

package HashLock

import (
	"fmt"
	"log"
	"time"
)

type ChanLock struct {
	log     *log.Logger
	channel chan struct{}
}

func NewChanLock(logger *log.Logger) *ChanLock {
	if logger == nil {
		logger = log.Default()
	}
	return &ChanLock{
		log:     logger,
		channel: make(chan struct{}, 1),
	}
}

func (l *ChanLock) Lock(timeout time.Duration) error {
	select {
	case l.channel <- struct{}{}:
		// lock acquired
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("failed to acquire lock after [%s]", timeout)
	}
}

func (l *ChanLock) Unlock() {
	<-l.channel
}
