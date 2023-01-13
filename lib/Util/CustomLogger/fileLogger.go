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

package CustomLogger

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"
)

func AddLogParam(param *flag.FlagSet, logPath *string) {
	param.StringVar(logPath, "log", "", "log file path")
}

func ParseLogFilePathInArgs() string {
	for idx, arg := range os.Args {
		if arg == "-log" {
			if idx+1 == len(os.Args) {
				return ""
			}
			logPath := os.Args[idx+1]
			if strings.HasPrefix(logPath, "-") {
				return ""
			}
			return logPath
		}
	}
	return ""
}

func FileLoger(logPath string, logId string) (*os.File, *log.Logger, error) {
	if logPath == "" {
		return nil, log.Default(), nil
	}

	err := os.MkdirAll(logPath, os.ModePerm)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create log path: %s", logPath)
	}
	logPath = path.Join(logPath, fmt.Sprintf("%s.log", logId))
	log.Printf("log file: %s", logPath)
	logFile, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	mw := io.MultiWriter(os.Stdout, logFile)
	logger := log.New(mw, fmt.Sprintf("%s: ", logId), log.Ldate|log.Ltime|log.Lshortfile)
	return logFile, logger, nil
}
