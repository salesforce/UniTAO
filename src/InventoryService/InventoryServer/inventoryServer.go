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

package InventoryServer

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"InventoryService/Config"
	"InventoryService/DataHandler"

	"github.com/salesforce/UniTAO/lib/Util"
)

type Server struct {
	Port   string
	args   ServerArgs
	config Config.ServerConfig
	data   *DataHandler.Handler
}

type ServerArgs struct {
	port   string
	config string
}

const (
	CONFIG       = "config"
	PORT         = "port"
	PORT_DEFAULT = "8003"
)

func argHandler() ServerArgs {
	args := ServerArgs{}
	var port string
	var configPath string
	flag.StringVar(&port, "port", "", "Data Server Listen Port")
	flag.StringVar(&configPath, "config", "", "Data Server Configuration JSON path")
	flag.Parse()
	args.port = port
	args.config = configPath
	if args.config == "" {
		flag.Usage()
		log.Fatalf("missing parameter [%s]", CONFIG)
	}
	return args
}

func New() Server {
	log.Println("Create Inventory Service Instance")
	server := Server{
		args: argHandler(),
	}
	err := Config.Read(server.args.config, &server.config)
	if err != nil {
		log.Fatalf("failed to read config=[%s], Err:%s", server.args.config, err)
	}
	if server.args.port == "" {
		if server.config.Http.Port == "" {
			server.Port = PORT_DEFAULT
			return server
		}
		server.Port = server.config.Http.Port
		return server
	}
	server.Port = server.args.port
	return server
}

func (srv *Server) Run() {
	log.Printf("Server Listen on PORT:%s", srv.Port)
	handler, err := DataHandler.New(srv.config.Database)
	if err != nil {
		log.Fatalf("failed to initialize data layer, Err:%s", err)
	}
	srv.data = handler
	http.HandleFunc("/", srv.handler)
	log.Printf("Data Server Listen @%s://%s:%s", srv.config.Http.HttpType, srv.config.Http.DnsName, srv.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", srv.Port), nil))
}

func (srv *Server) handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Inventory Server only support GET method", http.StatusMethodNotAllowed)
		return
	}
	dataType, dataId := Util.ParsePath(r.URL.Path)
	if dataType == "" {
		respObj := make(map[string]string)
		respObj["error message"] = "please use inventory{type}[/{id}]"
		Util.ResponseJson(w, respObj, http.StatusOK)
		return
	}
	if dataId == "" {
		idList, code, err := srv.data.List(dataType)
		if err != nil {
			http.Error(w, err.Error(), code)
			return
		}
		Util.ResponseJson(w, idList, code)
		return
	}
	data, code, err := srv.data.Get(dataType, dataId)
	if err != nil {
		http.Error(w, err.Error(), code)
		return
	}
	Util.ResponseJson(w, data, http.StatusOK)
}
