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

package Template

import (
	"fmt"
	"reflect"
	"strings"
)

type StrTemp struct {
	Template string
	Open     string
	Close    string
	Parts    []string
	Vars     []string
}

func ParseStr(template string, open string, close string) (*StrTemp, error) {
	temp, err := NewStrTemp(open, close)
	if err != nil {
		return nil, err
	}
	err = temp.Parse(template)
	if err != nil {
		return nil, err
	}
	return temp, nil
}

func NewStrTemp(open string, close string) (*StrTemp, error) {
	if len(open) == 0 || len(close) == 0 {
		return nil, fmt.Errorf("found empty string in open=[%s], close=[%s]", open, close)
	}
	return &StrTemp{
		Open:  open,
		Close: close,
	}, nil
}

func (t *StrTemp) Parse(template string) error {
	t.Template = template
	t.Parts = []string{}
	t.Vars = []string{}
	varHash := map[string]bool{}
	for len(template) > 0 {
		leftIdx := strings.Index(template, t.Open)
		if leftIdx < 0 {
			if strings.Contains(template, t.Close) {
				return fmt.Errorf("invalid template, extra '%s' spotted", t.Close)
			}
			t.Parts = append(t.Parts, template)
			break
		}
		prefix := template[:leftIdx]
		if strings.Contains(prefix, t.Close) {
			return fmt.Errorf("invalid template. extra '%s' spotted", t.Close)
		}
		if len(prefix) > 0 {
			t.Parts = append(t.Parts, prefix)
		}
		template = template[leftIdx:]
		rightIdx := strings.Index(template, t.Close)
		if rightIdx < 0 {
			return fmt.Errorf("invalid template, extra %s spotted", t.Open)
		}
		varPart := template[:rightIdx+len(t.Close)]
		attrName := t.ParseVar(varPart)
		if attrName == "" {
			return fmt.Errorf("invalid template, need attrname between '%s' and '%s'", t.Open, t.Close)
		}
		if strings.Contains(attrName, t.Open) {
			return fmt.Errorf("invalid template, extra %s spotted", t.Open)
		}
		t.Parts = append(t.Parts, varPart)
		if _, ok := varHash[attrName]; !ok {
			varHash[attrName] = true
			t.Vars = append(t.Vars, attrName)
		}
		template = template[rightIdx+len(t.Close):]
	}
	return nil
}

func (t *StrTemp) ParseVar(varPart string) string {
	if len(varPart) <= len(t.Open)+len(t.Close) {
		return ""
	}
	if !strings.HasPrefix(varPart, t.Open) || !strings.HasSuffix(varPart, t.Close) {
		return ""
	}
	attrName := varPart[len(t.Open) : len(varPart)-len(t.Close)]
	return attrName
}

func (t *StrTemp) BuildVarMap(data map[string]interface{}) (map[string]interface{}, error) {
	varMap := map[string]interface{}{}
	for _, attr := range t.Vars {
		attrValue, ok := data[attr]
		if !ok {
			return nil, fmt.Errorf("missing attr=[%s]", attr)
		}
		if attrValue == nil {
			return nil, fmt.Errorf("attr=[%s] valie is nil", attr)
		}
		varMap[attr] = attrValue
	}
	return varMap, nil
}

func (t *StrTemp) TestValue() string {
	testMap := map[string]interface{}{}
	for _, attr := range t.Vars {
		testMap[attr] = fmt.Sprintf("Test%sValue", attr)
	}
	testStr, _ := t.BuildValue(testMap)
	return testStr
}

func (t *StrTemp) BuildValue(data map[string]interface{}) (string, error) {
	varMap, err := t.BuildVarMap(data)
	if err != nil {
		return "", err
	}
	result := ""
	for _, part := range t.Parts {
		attr := t.ParseVar(part)
		if attr == "" {
			result = fmt.Sprintf("%s%s", result, part)
			continue
		}
		attrStr, ok := varMap[attr].(string)
		if !ok {
			return "", fmt.Errorf("invalid type attr=[%s], filed to convert type=[%s] to string", attr, reflect.TypeOf(varMap[attr]).Kind())
		}
		result = fmt.Sprintf("%s%s", result, attrStr)
	}
	return result, nil
}
