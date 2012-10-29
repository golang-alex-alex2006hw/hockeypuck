/*
   Hockeypuck - OpenPGP key server
   Copyright (C) 2012  Casey Marshall

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, version 3.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"code.google.com/p/gorilla/mux"
	"launchpad.net/hockeypuck"
	"launchpad.net/hockeypuck/pq"
)

var pqUser *string = flag.String("user", "", "postgres user")
var pqPassword *string = flag.String("password", "", "postgres password")
var pqHost *string = flag.String("host", "localhost", "postgres host")
var pqPort *int = flag.Int("port", 5432, "postgres port")
var pqDbName *string = flag.String("dbname", "hkp", "postgres database")

var httpBind *string = flag.String("http", ":11371", "http bind port")
var numWorkers *int = flag.Int("workers", runtime.NumCPU(), "number of workers")

func usage() {
	flag.PrintDefaults()
	os.Exit(1)
}

func die(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}

func main() {
	// Create an HTTP request router
	r := mux.NewRouter()
	// Create a new Hockeypuck server, bound to this router
	hkp := hockeypuck.NewHkpServer(r)
	flag.Parse()
	// Resolve flags, get the database connection string
	if *pqUser == "" {
		usage()
	}
	connect := fmt.Sprintf("user=%s password=%s host=%s port=%d dbname=%s",
		*pqUser, *pqPassword, *pqHost, *pqPort, *pqDbName)
	for i := 0; i < *numWorkers; i++ {
		worker := &pq.PqWorker{}
		err := worker.Init(connect)
		if err != nil {
			die(err)
		}
		// Start the worker
		hockeypuck.Start(hkp, worker)
	}
	// Bind the router to the built-in webserver root
	http.Handle("/", r)
	// Start the built-in webserver, run forever
	err := http.ListenAndServe(*httpBind, nil)
	die(err)
}