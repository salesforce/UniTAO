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
	"UniTao/DataService/lib/Config"
	"UniTao/DataService/lib/DataHandler"
	"flag"
	"fmt"
	"github.com/salesforce/UniTAO/lib/Util"
	"log"
	"net/http"
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
	handler, err := DataHandler.New(srv.config)
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
	dataType, dataId := Util.ParsePath(r.URL.Path)
	switch r.Method {
	case "GET":
		srv.handleGet(w, dataType, dataId)
	case "POST":
		srv.handlePost(w, r, dataType, dataId)
	case "DELETE":
		srv.handleDelete(w, dataType, dataId)
	case "PUT":
		srv.handlerPut(w, r, dataType, dataId)
	default:
		http.Error(w, fmt.Sprintf("method [%s] not supported", r.Method), http.StatusMethodNotAllowed)
	}
}

func (srv *Server) handleGet(w http.ResponseWriter, dataType string, dataId string) {
	if dataId == "" {
		idList, code, err := srv.data.List(dataType)
		if err != nil {
			http.Error(w, err.Error(), code)
			return
		}
		Util.ResponseJson(w, idList, code)
		return
	}
	result, code, err := srv.data.Get(dataType, dataId)
	if err != nil {
		http.Error(w, err.Error(), code)
		return
	}
	Util.ResponseJson(w, result, code)
}

func (srv *Server) handlePost(w http.ResponseWriter, r *http.Request, dataType string, dataId string) {
	payload := make(map[string]interface{})
	code, err := Util.LoadJSONPayload(r, payload)
	if err != nil {
		http.Error(w, err.Error(), code)
		return
	}
	code, err = srv.data.Add(dataType, dataId, payload)
	if err != nil {
		http.Error(w, err.Error(), code)
		return
	}
	Util.ResponseJson(w, payload, http.StatusCreated)
}

func (srv *Server) handlerPut(w http.ResponseWriter, r *http.Request, dataType string, dataId string) {
	payload := make(map[string]interface{})
	code, err := Util.LoadJSONPayload(r, payload)
	if err != nil {
		http.Error(w, err.Error(), code)
		return
	}
	code, err = srv.data.Set(dataType, dataId, payload)
	if err != nil {
		http.Error(w, err.Error(), code)
		return
	}
	Util.ResponseJson(w, payload, http.StatusCreated)
}

func (srv *Server) handleDelete(w http.ResponseWriter, dataType string, dataId string) {
	code, err := srv.data.Delete(dataType, dataId)
	if err != nil {
		http.Error(w, err.Error(), code)
	}
	result := map[string]string{
		"result": fmt.Sprintf("item [type/id]=[%s/%s] deleted", dataType, dataId),
	}
	Util.ResponseJson(w, result, code)
}
