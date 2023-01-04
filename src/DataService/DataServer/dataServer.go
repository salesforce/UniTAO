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

package DataServer

import (
	"Data"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"

	"DataService/Common"
	"DataService/Config"
	"DataService/DataHandler"
	"DataService/DataJournal"

	"github.com/salesforce/UniTAO/lib/Schema/Record"
	"github.com/salesforce/UniTAO/lib/Util"
	"github.com/salesforce/UniTAO/lib/Util/Http"
	"github.com/salesforce/UniTAO/lib/Util/Thread"
)

const (
	CONFIG       = "config"
	PORT         = "port"
	PORT_DEFAULT = "8010"
)

type Server struct {
	Id             string
	Port           string
	args           map[string]string
	config         Config.Confuguration
	data           *DataHandler.Handler
	journal        *DataJournal.JournalLib
	journalHandler *DataJournal.JournalHandler
	BackendCtl     *Thread.ThreadCtrl
	logPath        string
	log            *log.Logger
}

func New() (Server, error) {
	srv := Server{
		Port:    PORT_DEFAULT,
		args:    make(map[string]string),
		config:  Config.Confuguration{},
		logPath: "",
	}
	err := srv.init()
	if err != nil {
		return srv, err
	}
	return srv, nil
}

func (srv *Server) Run() {
	srv.log = log.Default()
	if srv.logPath != "" {
		logPath := path.Join(srv.logPath, fmt.Sprintf("%s.log", srv.Id))
		log.Printf("log file: %s", logPath)
		logFile, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("error opening file: %v", err)
		}
		defer logFile.Close()
		mw := io.MultiWriter(os.Stdout, logFile)
		logger := log.New(mw, fmt.Sprintf("%s: ", srv.Id), log.Ldate|log.Ltime|log.Lshortfile)
		srv.log = logger
	}
	srv.BackendCtl = Thread.NewThreadController(srv.log)
	handler, err := DataHandler.New(srv.config, srv.log, Data.ConnectDb)
	if err != nil {
		srv.log.Fatalf("failed to initialize data layer, Err:%s", err)
	}
	srv.data = handler
	journal, err := DataJournal.NewJournalLib(handler.DB, srv.config.DataTable.Data, srv.log)
	if err != nil {
		srv.log.Fatalf("failed to create Journal Library. Error: %s", err)
	}
	srv.journal = journal
	srv.data.AddJournal = srv.journal.AddJournal
	srv.RunJournalHandler()
	srv.RunHttp()
}

func (srv *Server) RunHttp() {
	http.HandleFunc("/", srv.handler)
	srv.log.Printf("Data Server Listen @%s://%s:%s", srv.config.Http.HttpType, srv.config.Http.DnsName, srv.Port)
	srv.log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", srv.Port), nil))
}

func (srv *Server) RunJournalHandler() {
	journal, err := DataJournal.NewJournalHandler(srv.data, srv.journal, srv.log)
	if err != nil {
		srv.log.Fatalf("failed to load Journal Handler. Error:%s", err)
	}
	srv.journalHandler = journal
	worker, err := srv.BackendCtl.AddWorker("journalHandler", srv.journalHandler.Run)
	if err != nil {
		srv.log.Fatalf("failed to create Journal Handler as backend process.")
	}
	srv.journal.HandlerNotify = worker.Notify
	worker.Run()
}

func (srv *Server) init() error {
	var port string
	var configPath string
	var serverId string
	var logPath string
	flag.StringVar(&serverId, "id", "", "Data Server Id")
	flag.StringVar(&port, "port", "", "Data Server Listen Port")
	flag.StringVar(&configPath, "config", "", "Data Server Configuration JSON path")
	flag.StringVar(&logPath, "log", "", "path that hold log")
	flag.Parse()
	srv.args[PORT] = port
	if serverId == "" {
		flag.Usage()
		return fmt.Errorf("missing parameter id")
	}
	srv.Id = serverId
	if configPath == "" {
		flag.Usage()
		return fmt.Errorf("missing parameter config")
	}
	if logPath != "" {
		err := os.MkdirAll(logPath, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create log path: %s", logPath)
		}
		srv.logPath = logPath
	}
	srv.args[CONFIG] = configPath
	err := Config.Read(srv.args[CONFIG], &srv.config)
	if err != nil {
		return err
	}
	if port != "" {
		srv.Port = port

	} else if srv.config.Http.Port != "" {
		srv.Port = srv.config.Http.Port
	}
	return nil
}

func (srv *Server) handler(w http.ResponseWriter, r *http.Request) {
	dataType, idPath := Util.ParsePath(r.URL.Path)
	if dataType == Record.KeyRecord {
		Http.ResponseJson(w, Http.HttpError{
			Status: http.StatusBadRequest,
			Message: []string{
				fmt.Sprintf("data type=[%s] is not supported", dataType),
			},
		}, http.StatusBadRequest, srv.config.Http)
		return
	}
	if _, ok := Common.ReadOnlyTypes[dataType]; ok && r.Method != http.MethodGet {
		Http.ResponseJson(w, Http.HttpError{
			Status: http.StatusBadRequest,
			Message: []string{
				fmt.Sprintf("update on data type=[%s] is not supported", dataType),
			},
		}, http.StatusBadRequest, srv.config.Http)
		return
	}
	switch r.Method {
	case http.MethodGet:
		srv.handleGet(w, dataType, idPath)
	case http.MethodPost:
		srv.handlePost(w, r, dataType, idPath)
	case http.MethodDelete:
		srv.handleDelete(w, dataType, idPath)
	case http.MethodPut:
		srv.handlePut(w, r, dataType, idPath)
	case http.MethodPatch:
		srv.handlePatch(w, r, dataType, idPath)
	default:
		Http.ResponseJson(w, Http.HttpError{
			Status: http.StatusMethodNotAllowed,
			Message: []string{
				fmt.Sprintf("method [%s] not supported", r.Method),
			},
		}, http.StatusMethodNotAllowed, srv.config.Http)
	}
}

func (srv *Server) handleGet(w http.ResponseWriter, dataType string, dataId string) {
	if dataId == "" {
		idList, err := srv.data.List(dataType)
		if err != nil {
			Http.ResponseJson(w, err, err.Status, srv.config.Http)
			return
		}
		Http.ResponseJson(w, idList, http.StatusOK, srv.config.Http)
		return
	}
	var result interface{}
	var err *Http.HttpError
	switch dataType {
	case Common.KeyJournal:
		result, err = srv.journal.GetJournal(dataId)
	default:
		result, err = srv.data.Get(dataType, dataId)
	}
	if err != nil {
		Http.ResponseJson(w, err, err.Status, srv.config.Http)
		return
	}
	Http.ResponseJson(w, result, http.StatusOK, srv.config.Http)
}

func (srv *Server) BuildRecord(payload map[string]interface{}, dataType string, dataId string) (*Record.Record, *Http.HttpError) {
	if dataType == "" {
		return nil, Http.NewHttpError(fmt.Sprintf("empty data type in path. [%s/%s]=''", Record.DataType, Record.DataId), http.StatusBadRequest)
	}
	if dataId == "" {
		return nil, Http.NewHttpError(fmt.Sprintf("empty data id in path. [%s/%s]=''", Record.DataType, Record.DataId), http.StatusBadRequest)
	}
	schema, err := srv.data.LocalSchema(dataType, "")
	if err != nil {
		return nil, err
	}
	record := Record.NewRecord(dataType, schema.Schema.Version, dataId, payload)
	return record, nil
}

func (srv *Server) handlePost(w http.ResponseWriter, r *http.Request, dataType string, dataId string) {
	reqBody, err := Http.LoadRequest(r)
	if err != nil {
		Http.ResponseJson(w, err, err.Status, srv.config.Http)
		return
	}
	payload, ok := reqBody.(map[string]interface{})
	if !ok {
		Http.ResponseJson(w, "failed to parse request into JSON object", http.StatusBadRequest, srv.config.Http)
		return
	}
	var record *Record.Record
	var ex error
	if len(r.Header.Values(Record.NotRecord)) == 0 {
		if dataType != "" {
			Http.ResponseJson(w, Http.NewHttpError("data type expect to be empty for action=[POST]", http.StatusBadRequest), http.StatusBadRequest, srv.config.Http)
			return
		}
		if dataId != "" {
			Http.ResponseJson(w, Http.NewHttpError("data id expect to be empty for action=[POST]", http.StatusBadRequest), http.StatusBadRequest, srv.config.Http)
			return
		}
		record, ex = Record.LoadMap(payload)
		if ex != nil {
			Http.ResponseJson(w, Http.WrapError(ex, "failed to load payload as Record", http.StatusBadRequest), http.StatusBadRequest, srv.config.Http)
			return
		}
	} else {
		record, err = srv.BuildRecord(payload, dataType, dataId)
		if err != nil {
			Http.ResponseJson(w, err, err.Status, srv.config.Http)
			return
		}
	}
	err = srv.data.Add(record)
	if err != nil {
		Http.ResponseJson(w, err, err.Status, srv.config.Http)
		return
	}
	Http.ResponseText(w, []byte(record.Id), http.StatusCreated, srv.config.Http)
}

func (srv *Server) handlePut(w http.ResponseWriter, r *http.Request, dataType string, dataId string) {
	reqBody, err := Http.LoadRequest(r)
	if err != nil {
		Http.ResponseJson(w, err, err.Status, srv.config.Http)
		return
	}
	payload, ok := reqBody.(map[string]interface{})
	if !ok {
		Http.ResponseJson(w, "failed to parse request into JSON object", http.StatusBadRequest, srv.config.Http)
	}
	var record *Record.Record
	var ex error
	if len(r.Header.Values(Record.NotRecord)) == 0 {
		record, ex = Record.LoadMap(payload)
		if ex != nil {
			Http.ResponseJson(w, Http.WrapError(ex, "failed to load payload as Record", http.StatusBadRequest), http.StatusBadRequest, srv.config.Http)
			return
		}
	} else {
		record, err = srv.BuildRecord(payload, dataType, dataId)
		if err != nil {
			Http.ResponseJson(w, err, err.Status, srv.config.Http)
			return
		}
	}
	err = srv.data.Set(dataType, dataId, record)
	if err != nil {
		Http.ResponseJson(w, err, err.Status, srv.config.Http)
		return
	}
	Http.ResponseText(w, []byte(record.Id), http.StatusCreated, srv.config.Http)
}

func (srv *Server) handleDelete(w http.ResponseWriter, dataType string, dataId string) {
	err := srv.data.Delete(dataType, dataId)
	if err != nil {
		Http.ResponseJson(w, err, err.Status, srv.config.Http)
	}
	result := map[string]string{
		"result": fmt.Sprintf("item [type/id]=[%s/%s] deleted", dataType, dataId),
	}
	Http.ResponseJson(w, result, http.StatusAccepted, srv.config.Http)
}

func (srv *Server) handlePatch(w http.ResponseWriter, r *http.Request, dataType string, idPath string) {
	payload, e := Http.LoadRequest(r)
	if e != nil {
		Http.ResponseJson(w, e, e.Status, srv.config.Http)
		return
	}
	headers := Http.ParseHeaders(r)
	response, e := srv.data.Patch(dataType, idPath, headers, payload)
	if e != nil {
		Http.ResponseJson(w, e, e.Status, srv.config.Http)
		return
	}
	Http.ResponseJson(w, response, http.StatusAccepted, srv.config.Http)
}
