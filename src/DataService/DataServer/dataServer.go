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
	"log"
	"net/http"

	"DataService/Config"
	"DataService/DataHandler"

	"github.com/salesforce/UniTAO/lib/Schema/Record"
	"github.com/salesforce/UniTAO/lib/Util"
	"github.com/salesforce/UniTAO/lib/Util/Http"
)

const (
	CONFIG       = "config"
	PORT         = "port"
	PORT_DEFAULT = "8010"
)

type Server struct {
	Port   string
	args   map[string]string
	config Config.Confuguration
	data   *DataHandler.Handler
}

func New() (Server, error) {
	srv := Server{
		Port:   PORT_DEFAULT,
		args:   make(map[string]string),
		config: Config.Confuguration{},
	}
	err := srv.init()
	if err != nil {
		return srv, err
	}
	return srv, nil
}

func (srv *Server) Run() {
	handler, err := DataHandler.New(srv.config, Data.ConnectDb)
	if err != nil {
		log.Fatalf("failed to initialize data layer, Err:%s", err)
	}
	srv.data = handler
	http.HandleFunc("/", srv.handler)
	log.Printf("Data Server Listen @%s://%s:%s", srv.config.Http.HttpType, srv.config.Http.DnsName, srv.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", srv.Port), nil))
}

func (srv *Server) init() error {
	var port string
	var configPath string
	flag.StringVar(&port, "port", "", "Data Server Listen Port")
	flag.StringVar(&configPath, "config", "", "Data Server Configuration JSON path")
	flag.Parse()
	srv.args[PORT] = port
	if configPath == "" {
		flag.Usage()
		return fmt.Errorf("missing parameter config")
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
	switch r.Method {
	case Http.GET:
		srv.handleGet(w, dataType, idPath)
	case Http.POST:
		srv.handlePost(w, r, dataType, idPath)
	case Http.DELETE:
		srv.handleDelete(w, dataType, idPath)
	case Http.PUT:
		srv.handlePut(w, r, dataType, idPath)
	case Http.PATCH:
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
	result, err := srv.data.Get(dataType, dataId)
	if err != nil {
		Http.ResponseJson(w, err, err.Status, srv.config.Http)
		return
	}
	Http.ResponseJson(w, result, http.StatusOK, srv.config.Http)
}

func ParseRecord(noRecordList []string, payload map[string]interface{}, dataType string, dataId string) (*Record.Record, *Http.HttpError) {
	if len(noRecordList) == 0 {
		record, err := Record.LoadMap(payload)
		if err != nil {
			return nil, Http.WrapError(err, "failed to load JSON payload from request", http.StatusBadRequest)
		}
		if dataType != "" {
			return nil, Http.NewHttpError("data type expect to be empty for action=[POST, PUT]", http.StatusBadRequest)
		}
		if dataId != "" {
			return nil, Http.NewHttpError("data expect to be empty for action=[POST, PUT]", http.StatusBadRequest)
		}
		return record, nil
	}
	if dataType == "" {
		return nil, Http.NewHttpError(fmt.Sprintf("empty data type in path. [%s/%s]=''", Record.DataType, Record.DataId), http.StatusBadRequest)
	}
	if dataId == "" {
		return nil, Http.NewHttpError(fmt.Sprintf("empty data id in path. [%s/%s]=''", Record.DataType, Record.DataId), http.StatusBadRequest)
	}
	record := Record.NewRecord(dataType, "0_00_00", dataId, payload)
	return record, nil
}

func (srv *Server) handlePost(w http.ResponseWriter, r *http.Request, dataType string, dataId string) {
	payload := make(map[string]interface{})
	err := Http.LoadRequest(r, &payload)
	if err != nil {
		Http.ResponseJson(w, err, err.Status, srv.config.Http)
		return
	}
	record, e := ParseRecord(r.Header.Values(Record.NotRecord), payload, dataType, dataId)
	if e != nil {
		Http.ResponseJson(w, err, e.Status, srv.config.Http)
		return
	}
	e = srv.data.Add(record)
	if e != nil {
		Http.ResponseJson(w, e, e.Status, srv.config.Http)
		return
	}
	Http.ResponseText(w, []byte(record.Id), http.StatusCreated, srv.config.Http)
}

func (srv *Server) handlePut(w http.ResponseWriter, r *http.Request, dataType string, dataId string) {
	payload := make(map[string]interface{})
	e := Http.LoadRequest(r, &payload)
	if e != nil {
		Http.ResponseJson(w, e, e.Status, srv.config.Http)
		return
	}
	record, e := ParseRecord(r.Header.Values(Record.NotRecord), payload, dataType, dataId)
	if e != nil {
		Http.ResponseJson(w, e, e.Status, srv.config.Http)
		return
	}
	e = srv.data.Set(record)
	if e != nil {
		Http.ResponseJson(w, e, e.Status, srv.config.Http)
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
	payload := make(map[string]interface{})
	e := Http.LoadRequest(r, &payload)
	if e != nil {
		Http.ResponseJson(w, e, e.Status, srv.config.Http)
		return
	}
	response, e := srv.data.Patch(dataType, idPath, payload)
	if e != nil {
		Http.ResponseJson(w, e, e.Status, srv.config.Http)
		return
	}
	Http.ResponseJson(w, response, http.StatusAccepted, srv.config.Http)
}
