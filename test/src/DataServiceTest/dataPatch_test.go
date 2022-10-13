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
	"Data/DbIface"
	"encoding/json"
	"testing"
)

func TestUpdateAttr1Level(t *testing.T) {
	dataStr := `{
		"__id": "test01",
		"__type": "testData",
		"__ver": "0.0.1",
		"data": {
			"attr01": "test"
		}
	}`
	data := map[string]interface{}{}
	err := json.Unmarshal([]byte(dataStr), &data)
	if err != nil {
		t.Fatalf(err.Error())
	}
	subData, attrPath, err := DbIface.GetDataOnPath(data, "data/attr01", "")
	if err != nil {
		t.Fatalf(err.Error())
	}
	if attrPath != "attr01" {
		t.Fatalf("failed to get the valid attrPath. expect [attr01]!=[%s]", attrPath)
	}
	err = DbIface.SetPatchData(subData, attrPath, "testOk")
	if err != nil {
		t.Fatalf(err.Error())
	}
	if subData["attr01"].(string) != "testOk" {
		t.Fatalf("failed to set attr01 to testOk")
	}
	if data["data"].(map[string]interface{})["attr01"].(string) != "testOk" {
		t.Fatalf("failed to update data in data")
	}
	subData, attrPath, err = DbIface.GetDataOnPath(data, "data/attr02", "")
	if err != nil {
		t.Fatalf(err.Error())
	}
	if attrPath != "attr02" {
		t.Fatalf("failed to get the valid attrPath. expect [attr02]!=[%s]", attrPath)
	}
	err = DbIface.SetPatchData(subData, attrPath, "testOk")
	if err != nil {
		t.Fatalf(err.Error())
	}
	if subData["attr02"].(string) != "testOk" {
		t.Fatalf("failed to set attr01 to testOk")
	}
	if data["data"].(map[string]interface{})["attr02"].(string) != "testOk" {
		t.Fatalf("failed to update data in data")
	}
	subData, attrPath, err = DbIface.GetDataOnPath(data, "data/attr01", "")
	if err != nil {
		t.Fatalf(err.Error())
	}
	if attrPath != "attr01" {
		t.Fatalf("failed to get the valid attrPath. expect [attr02]!=[%s]", attrPath)
	}
	err = DbIface.SetPatchData(subData, attrPath, nil)
	if err != nil {
		t.Fatalf(err.Error())
	}
	_, ok := subData["attr01"]
	if ok {
		t.Fatalf("failed to delete attr01 in subData")
	}
	_, ok = data["data"].(map[string]interface{})["attr01"]
	if ok {
		t.Fatalf("failed to delete attr01 in data")
	}
}

func TestUpdateAttr2Level(t *testing.T) {
	dataStr := `{
		"__id": "test01",
		"__type": "testData",
		"__ver": "0.0.1",
		"data": {
			"attr02": {
				"attr02-01": "test"
			}
		}
	}`
	data := map[string]interface{}{}
	err := json.Unmarshal([]byte(dataStr), &data)
	if err != nil {
		t.Fatalf(err.Error())
	}
	subData, attrPath, err := DbIface.GetDataOnPath(data, "data/attr02/attr02-01", "")
	if err != nil {
		t.Fatalf(err.Error())
	}
	if attrPath != "attr02-01" {
		t.Fatalf("failed to get the valid attrPath. expect [attr02-01]!=[%s]", attrPath)
	}
	err = DbIface.SetPatchData(subData, attrPath, "testOk")
	if err != nil {
		t.Fatalf(err.Error())
	}
	if subData["attr02-01"].(string) != "testOk" {
		t.Fatalf("failed to set attr02-01 to testOk")
	}
	if data["data"].(map[string]interface{})["attr02"].(map[string]interface{})["attr02-01"].(string) != "testOk" {
		t.Fatalf("failed to update data in data")
	}
	subData, attrPath, err = DbIface.GetDataOnPath(data, "data/attr02/attr02-02", "")
	if err != nil {
		t.Fatalf(err.Error())
	}
	if attrPath != "attr02-02" {
		t.Fatalf("failed to get the valid attrPath. expect [attr02-02]!=[%s]", attrPath)
	}
	err = DbIface.SetPatchData(subData, attrPath, "testOk")
	if err != nil {
		t.Fatalf(err.Error())
	}
	if subData["attr02-02"].(string) != "testOk" {
		t.Fatalf("failed to set attr02-02 to testOk")
	}
	if data["data"].(map[string]interface{})["attr02"].(map[string]interface{})["attr02-02"].(string) != "testOk" {
		t.Fatalf("failed to update data in data")
	}
	subData, attrPath, err = DbIface.GetDataOnPath(data, "data/attr02/attr02-01", "")
	if err != nil {
		t.Fatalf(err.Error())
	}
	if attrPath != "attr02-01" {
		t.Fatalf("failed to get the valid attrPath. expect [attr02-02]!=[%s]", attrPath)
	}
	err = DbIface.SetPatchData(subData, attrPath, nil)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if _, ok := subData["attr02-01"]; ok {
		t.Fatalf("failed to delete attr02-01")
	}
	if _, ok := data["data"].(map[string]interface{})["attr02"].(map[string]interface{})["attr02-01"]; ok {
		t.Fatalf("failed to update data in data")
	}
}

func TestUpdateAttrArray(t *testing.T) {
	dataStr := `{
		"__id": "test01",
		"__type": "testData",
		"__ver": "0.0.1",
		"data": {
			"attr03": [
				{
					"attr03-01": "test"
				}
			]
		}
	}`
	data := map[string]interface{}{}
	err := json.Unmarshal([]byte(dataStr), &data)
	if err != nil {
		t.Fatalf(err.Error())
	}
	subData, attrPath, err := DbIface.GetDataOnPath(data, "data/attr03[0]/attr03-01", "")
	if err != nil {
		t.Fatalf(err.Error())
	}
	if attrPath != "attr03-01" {
		t.Fatalf("failed to get the valid attrPath. expect [attr03-01]!=[%s]", attrPath)
	}
	err = DbIface.SetPatchData(subData, attrPath, "testOk")
	if err != nil {
		t.Fatalf(err.Error())
	}
	if subData["attr03-01"].(string) != "testOk" {
		t.Fatalf("failed to set attr03-01 to testOk")
	}
	if data["data"].(map[string]interface{})["attr03"].([]interface{})[0].(map[string]interface{})["attr03-01"].(string) != "testOk" {
		t.Fatalf("failed to update data in data")
	}
	subData, attrPath, err = DbIface.GetDataOnPath(data, "data/attr03[0]/attr03-02", "")
	if err != nil {
		t.Fatalf(err.Error())
	}
	if attrPath != "attr03-02" {
		t.Fatalf("failed to get the valid attrPath. expect [attr03-02]!=[%s]", attrPath)
	}
	err = DbIface.SetPatchData(subData, attrPath, "testOk")
	if err != nil {
		t.Fatalf(err.Error())
	}
	if subData["attr03-02"].(string) != "testOk" {
		t.Fatalf("failed to set attr03-02 to testOk")
	}
	if data["data"].(map[string]interface{})["attr03"].([]interface{})[0].(map[string]interface{})["attr03-02"].(string) != "testOk" {
		t.Fatalf("failed to update data in data")
	}
	subData, attrPath, err = DbIface.GetDataOnPath(data, "data/attr03[0]/attr03-01", "")
	if err != nil {
		t.Fatalf(err.Error())
	}
	if attrPath != "attr03-01" {
		t.Fatalf("failed to get the valid attrPath. expect [attr03-01]!=[%s]", attrPath)
	}
	err = DbIface.SetPatchData(subData, attrPath, nil)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if _, ok := subData["attr03-01"]; ok {
		t.Fatalf("failed to delete attr02-01")
	}
	if _, ok := data["data"].(map[string]interface{})["attr03"].([]interface{})[0].(map[string]interface{})["attr03-01"]; ok {
		t.Fatalf("failed to update data in data")
	}
	subData, attrPath, err = DbIface.GetDataOnPath(data, "data/attr03[-1]", "")
	if err != nil {
		t.Fatalf(err.Error())
	}
	if attrPath != "attr03[-1]" {
		t.Fatalf("failed to get the valid attrPath. expect [attr03[-1]]!=[%s]", attrPath)
	}
	newData := map[string]interface{}{"attr03-03": "testOk"}
	err = DbIface.SetPatchData(subData, attrPath, newData)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if subData["attr03"].([]interface{})[0].(map[string]interface{})["attr03-03"].(string) != "testOk" {
		t.Fatalf("failed to add new item in array")
	}
	if data["data"].(map[string]interface{})["attr03"].([]interface{})[0].(map[string]interface{})["attr03-03"].(string) != "testOk" {
		t.Fatalf("failed to add new item in array in overall data")
	}
	subData, attrPath, err = DbIface.GetDataOnPath(data, "data/attr03[2]", "")
	if err != nil {
		t.Fatalf(err.Error())
	}
	if attrPath != "attr03[2]" {
		t.Fatalf("failed to get the valid attrPath. expect [attr03[2]]!=[%s]", attrPath)
	}
	newData = map[string]interface{}{"attr03-04": "testOk"}
	err = DbIface.SetPatchData(subData, attrPath, newData)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if subData["attr03"].([]interface{})[2].(map[string]interface{})["attr03-04"].(string) != "testOk" {
		t.Fatalf("failed to add new item in array")
	}
	if data["data"].(map[string]interface{})["attr03"].([]interface{})[2].(map[string]interface{})["attr03-04"].(string) != "testOk" {
		t.Fatalf("failed to add new item in array in overall data")
	}
	subData, attrPath, err = DbIface.GetDataOnPath(data, "data/attr03[1]", "")
	if err != nil {
		t.Fatalf(err.Error())
	}
	if attrPath != "attr03[1]" {
		t.Fatalf("failed to get the valid attrPath. expect [attr03[1]]!=[%s]", attrPath)
	}
	newData = map[string]interface{}{"attr03-05": "testOk"}
	err = DbIface.SetPatchData(subData, attrPath, newData)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if subData["attr03"].([]interface{})[1].(map[string]interface{})["attr03-05"].(string) != "testOk" {
		t.Fatalf("failed to add new item in array")
	}
	if data["data"].(map[string]interface{})["attr03"].([]interface{})[1].(map[string]interface{})["attr03-05"].(string) != "testOk" {
		t.Fatalf("failed to add new item in array in overall data")
	}
	subData, attrPath, err = DbIface.GetDataOnPath(data, "data/attr03[1]", "")
	if err != nil {
		t.Fatalf(err.Error())
	}
	if attrPath != "attr03[1]" {
		t.Fatalf("failed to get the valid attrPath. expect [attr03[1]]!=[%s]", attrPath)
	}
	err = DbIface.SetPatchData(subData, attrPath, nil)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if len(subData["attr03"].([]interface{})) != 2 {
		t.Fatalf("failed to delete the second item in array")
	}
	if subData["attr03"].([]interface{})[0].(map[string]interface{})["attr03-03"].(string) != "testOk" {
		t.Fatalf("failed to delete the right item.")
	}
	if subData["attr03"].([]interface{})[1].(map[string]interface{})["attr03-04"].(string) != "testOk" {
		t.Fatalf("failed to delete the right item.")
	}
	if len(data["data"].(map[string]interface{})["attr03"].([]interface{})) != 2 {
		t.Fatalf("failed to delete the second item in array")
	}
	if data["data"].(map[string]interface{})["attr03"].([]interface{})[0].(map[string]interface{})["attr03-03"].(string) != "testOk" {
		t.Fatalf("failed to delete the right item.")
	}
	if data["data"].(map[string]interface{})["attr03"].([]interface{})[1].(map[string]interface{})["attr03-04"].(string) != "testOk" {
		t.Fatalf("failed to add new item in array in overall data")
	}
}

func TestUpdateAttrMap(t *testing.T) {
	dataStr := `{
		"__id": "test01",
		"__type": "testData",
		"__ver": "0.0.1",
		"data": {
			"attr04": {
				"attr04-01": {
					"attr04-01-01": "test"
				}
			}
		}
	}`
	data := map[string]interface{}{}
	err := json.Unmarshal([]byte(dataStr), &data)
	if err != nil {
		t.Fatalf(err.Error())
	}
	subData, attrPath, err := DbIface.GetDataOnPath(data, "data/attr04[attr04-01]/attr04-01-01", "")
	if err != nil {
		t.Fatalf(err.Error())
	}
	if attrPath != "attr04-01-01" {
		t.Fatalf("failed to get the valid attrPath. expect [attr04-01-01]!=[%s]", attrPath)
	}
	err = DbIface.SetPatchData(subData, attrPath, "testOk")
	if err != nil {
		t.Fatalf(err.Error())
	}
	if subData["attr04-01-01"].(string) != "testOk" {
		t.Fatalf("failed to set attr03-01 to testOk")
	}
	if data["data"].(map[string]interface{})["attr04"].(map[string]interface{})["attr04-01"].(map[string]interface{})["attr04-01-01"].(string) != "testOk" {
		t.Fatalf("failed to update data in data")
	}
	subData, attrPath, err = DbIface.GetDataOnPath(data, "data/attr04[attr04-01]/attr04-01-02", "")
	if err != nil {
		t.Fatalf(err.Error())
	}
	if attrPath != "attr04-01-02" {
		t.Fatalf("failed to get the valid attrPath. expect [attr04-01-02]!=[%s]", attrPath)
	}
	err = DbIface.SetPatchData(subData, attrPath, "testOk")
	if err != nil {
		t.Fatalf(err.Error())
	}
	if subData["attr04-01-02"].(string) != "testOk" {
		t.Fatalf("failed to set attr03-02 to testOk")
	}
	if data["data"].(map[string]interface{})["attr04"].(map[string]interface{})["attr04-01"].(map[string]interface{})["attr04-01-02"].(string) != "testOk" {
		t.Fatalf("failed to update data in data")
	}
	subData, attrPath, err = DbIface.GetDataOnPath(data, "data/attr04[attr04-01]/attr04-01-01", "")
	if err != nil {
		t.Fatalf(err.Error())
	}
	if attrPath != "attr04-01-01" {
		t.Fatalf("failed to get the valid attrPath. expect [attr04-01-01]!=[%s]", attrPath)
	}
	err = DbIface.SetPatchData(subData, attrPath, nil)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if _, ok := subData["attr04-01-01"]; ok {
		t.Fatalf("failed to delete attr02-01")
	}
	if _, ok := data["data"].(map[string]interface{})["attr04"].(map[string]interface{})["attr04-01"].(map[string]interface{})["attr04-01-01"]; ok {
		t.Fatalf("failed to update data in data")
	}
	subData, attrPath, err = DbIface.GetDataOnPath(data, "data/attr04[attr04-02]", "")
	if err != nil {
		t.Fatalf(err.Error())
	}
	if attrPath != "attr04[attr04-02]" {
		t.Fatalf("failed to get the valid attrPath. expect [attr04[attr04-02]]!=[%s]", attrPath)
	}
	newData := map[string]interface{}{"attr04-02-01": "testOk"}
	err = DbIface.SetPatchData(subData, attrPath, newData)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if subData["attr04"].(map[string]interface{})["attr04-02"].(map[string]interface{})["attr04-02-01"].(string) != "testOk" {
		t.Fatalf("failed to add new item in array")
	}
	if data["data"].(map[string]interface{})["attr04"].(map[string]interface{})["attr04-02"].(map[string]interface{})["attr04-02-01"].(string) != "testOk" {
		t.Fatalf("failed to add new item in array in overall data")
	}
	subData, attrPath, err = DbIface.GetDataOnPath(data, "data/attr04[attr04-01]", "")
	if err != nil {
		t.Fatalf(err.Error())
	}
	if attrPath != "attr04[attr04-01]" {
		t.Fatalf("failed to get the valid attrPath. expect [attr04[attr04-01]]!=[%s]", attrPath)
	}
	newData = map[string]interface{}{"attr04-01-03": "testOk"}
	err = DbIface.SetPatchData(subData, attrPath, newData)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if subData["attr04"].(map[string]interface{})["attr04-01"].(map[string]interface{})["attr04-01-03"].(string) != "testOk" {
		t.Fatalf("failed to add new item in array")
	}
	if data["data"].(map[string]interface{})["attr04"].(map[string]interface{})["attr04-01"].(map[string]interface{})["attr04-01-03"].(string) != "testOk" {
		t.Fatalf("failed to add new item in array in overall data")
	}
	subData, attrPath, err = DbIface.GetDataOnPath(data, "data/attr04[attr04-01]", "")
	if err != nil {
		t.Fatalf(err.Error())
	}
	if attrPath != "attr04[attr04-01]" {
		t.Fatalf("failed to get the valid attrPath. expect [attr04[attr04-01]]!=[%s]", attrPath)
	}
	err = DbIface.SetPatchData(subData, attrPath, nil)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if len(subData["attr04"].(map[string]interface{})) != 1 {
		t.Fatalf("failed to delete the second item in array")
	}
	if subData["attr04"].(map[string]interface{})["attr04-02"].(map[string]interface{})["attr04-02-01"].(string) != "testOk" {
		t.Fatalf("failed to delete the right item.")
	}
	if _, ok := subData["attr04"].(map[string]interface{})["attr04-01"]; ok {
		t.Fatalf("failed to delete map item key=[attr04-01]")
	}
	if len(data["data"].(map[string]interface{})["attr04"].(map[string]interface{})) != 1 {
		t.Fatalf("failed to delete the second item in array")
	}
	if data["data"].(map[string]interface{})["attr04"].(map[string]interface{})["attr04-02"].(map[string]interface{})["attr04-02-01"].(string) != "testOk" {
		t.Fatalf("failed to delete the right item.")
	}
	if _, ok := data["data"].(map[string]interface{})["attr04"].(map[string]interface{})["attr04-01"]; ok {
		t.Fatalf("failed to delete map item key=[attr04-01]")
	}
}
