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

package DataServiceTest

import (
	"DataService/DataLock"
	"testing"
	"time"
)

func TestSingleLock(t *testing.T) {
	lockHash := DataLock.NewHash(5)
	r := lockHash.NewRequest("", "testType/testId", 0)
	handle, err := lockHash.Lock(r)
	if err != nil {
		t.Fatalf("failed to generate a simple lock with testType/testId")
	}
	lock, ok := lockHash.Hash[handle]
	if !ok {
		t.Fatalf("generated lock=[%s] does not exists", handle)
	}
	if handle != lock.Handle {
		t.Fatalf("invalid lock generated, handle:%s != key:%s", lock.Handle, handle)
	}
}

func TestQueryLock(t *testing.T) {
	lockHash := DataLock.NewHash(5)
	r := lockHash.NewRequest("testKey", "testType/testId", 0)
	handle, err := lockHash.Lock(r)
	if err != nil {
		t.Fatalf("failed to generate a simple lock with testType/testId")
	}
	lock, err := lockHash.QueryLock("testKey", []string{"testType", "testId"})
	if err != nil {
		t.Fatalf("failed to query lock")
	}
	if lock.Handle != handle {
		t.Fatalf("failed to query the right lock from hash")
	}
	lock.ExpireTime = 123
	if lockHash.Hash[handle].ExpireTime != 123 {
		t.Fatalf("failed to update Expire time to 123")
	}
}

func TestLockWithGivenOwnerKey(t *testing.T) {
	lockHash := DataLock.NewHash(5)
	r := lockHash.NewRequest("testKey", "testType/testId", 0)
	if r.UserId != "testKey" {
		t.Fatalf("generated request UserId:%s != given Key:testKey", r.UserId)
	}
	handle, err := lockHash.Lock(r)
	if err != nil {
		t.Fatalf("failed to generate a simple lock with testType/testId")
	}
	if lockHash.Hash[handle].OwnerId != r.UserId {
		t.Fatalf("generated lock ownerId:%s != request UserId:%s", lockHash.Hash[handle].OwnerId, r.UserId)
	}
}

func TestConflictLockRequest(t *testing.T) {
	lockHash := DataLock.NewHash(5)
	r1 := lockHash.NewRequest("", "testType/testId/attr", 5)
	r2 := lockHash.NewRequest("", "testType/testId/attr", 5)
	r3 := lockHash.NewRequest("", "testType/testId", 5)
	r4 := lockHash.NewRequest("", "testType/testId/attr/subAttr", 5)
	if r1.UserId == r2.UserId {
		t.Fatalf("random UserId1: %s != UserId2: %s", r1.UserId, r2.UserId)
	}
	_, err := lockHash.Lock(r1)
	if err != nil {
		t.Fatalf("failed to generate lock on requst r1, Error:%s", err)
	}
	_, err = lockHash.Lock(r2)
	if err == nil {
		t.Fatalf("failed to catch the lock conflict on r2, Error:%s", err)
	}
	_, err = lockHash.Lock(r3)
	if err == nil {
		t.Fatalf("failed to catch the lock conflict on r3, Error:%s", err)
	}
	_, err = lockHash.Lock(r4)
	if err == nil {
		t.Fatalf("failed to catch the lock conflict on r4, Error:%s", err)
	}
}

func TestMatchUserId(t *testing.T) {
	lockHash := DataLock.NewHash(5)
	r1 := lockHash.NewRequest("test1", "testType/testId", 5)
	r2 := lockHash.NewRequest("test1", "testType/testId", 5)
	if r1.UserId != r2.UserId {
		t.Fatalf("random UserId1: %s != UserId2: %s", r1.UserId, r2.UserId)
	}
	hdl1, err := lockHash.Lock(r1)
	if err != nil {
		t.Fatalf("failed to generate lock on requst r1, Error:%s", err)
	}
	exp1 := lockHash.Hash[hdl1].ExpireTime
	time.Sleep(1 * time.Second)
	hdl2, err := lockHash.Lock(r2)
	if err != nil {
		t.Fatalf("failed to generate lock on requst r2, Error:%s", err)
	}
	if hdl1 != hdl2 {
		t.Fatalf("failed to get the same handle with same user and path")
	}
	exp2 := lockHash.Hash[hdl2].ExpireTime
	if exp1 == exp2 {
		t.Fatalf("failed to update expire time with the second lock")
	}
}

func TestMatchUserIdOverWrite(t *testing.T) {
	lockHash := DataLock.NewHash(5)
	r1 := lockHash.NewRequest("test1", "testType/testId/attr", 5)
	r2 := lockHash.NewRequest("test1", "testType/testId/attr/subAttr", 5)
	r3 := lockHash.NewRequest("test1", "testType/testId", 5)
	if r1.UserId != r2.UserId {
		t.Fatalf("random UserId1: %s != UserId2: %s", r1.UserId, r2.UserId)
	}
	hdl1, err := lockHash.Lock(r1)
	if err != nil {
		t.Fatalf("failed to generate lock on requst r1, Error:%s", err)
	}
	if lockHash.Hash[hdl1].Path != "testType/testId/attr" {
		t.Fatalf("failed to retrieve lock path=[testType/testId/attr]")
	}
	hdl2, err := lockHash.Lock(r2)
	if err != nil {
		t.Fatalf("failed to generate lock on requst r2, Error:%s", err)
	}
	if lockHash.Hash[hdl2].Path != "testType/testId/attr" {
		t.Fatalf("failed to retrieve lock path=[testType/testId/attr]")
	}
	hdl3, err := lockHash.Lock(r3)
	if err != nil {
		t.Fatalf("failed to generate lock on requst r3, Error:%s", err)
	}
	if lockHash.Hash[hdl3].Path != "testType/testId" {
		t.Fatalf("failed to over load path [%s] != [testType/testId]", lockHash.Hash[hdl2].Path)
	}
}
