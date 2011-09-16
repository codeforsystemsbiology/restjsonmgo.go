/*
   Copyright (C) 2003-2011 Institute for Systems Biology
                           Seattle, Washington, USA.

   This library is free software; you can redistribute it and/or
   modify it under the terms of the GNU Lesser General Public
   License as published by the Free Software Foundation; either
   version 2.1 of the License, or (at your option) any later version.

   This library is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
   Lesser General Public License for more details.

   You should have received a copy of the GNU Lesser General Public
   License along with this library; if not, write to the Free Software
   Foundation, Inc., 59 Temple Place, Suite 330, Boston, MA 02111-1307  USA

*/
package main

import (
	"crypto/rand"
	"fmt"
	"http"
	"json"
	"os"
	"strings"
	"url"
)

type RestJsonMongo struct {
	Store     *JsonStore
	Domain    string
	JsonParam string
	Target    *url.URL
}

// GET /<domain>
func (this *RestJsonMongo) Index(rw http.ResponseWriter, values url.Values, header http.Header) {
	logger.Debug("Index():%v,%v", values, header)
	items, err := this.Store.Find(values)
	if err != nil {
		http.Error(rw, err.String(), http.StatusBadRequest)
		return
	}

	handle := ItemsHandle{Items: items, NumberOfItems: len(items)}
	if err := json.NewEncoder(rw).Encode(handle); err != nil {
		http.Error(rw, err.String(), http.StatusBadRequest)
	}
}
// POST /<domain>
func (this *RestJsonMongo) Create(rw http.ResponseWriter, r *http.Request) (item interface{}) {
	logger.Debug("Create()")

	if err := this.LoadJson(r, item); err != nil {
		http.Error(rw, err.String(), http.StatusBadRequest)
		return
	}

	itemId := UniqueId()

	if err := this.Store.Create(itemId, item); err != nil {
		http.Error(rw, err.String(), http.StatusBadRequest)
		return
	}

	this.Find(rw, itemId)
	return
}
// GET /<domain>/id
func (this *RestJsonMongo) Find(rw http.ResponseWriter, id string) {
	logger.Debug("Find(%v)", id)
	storedItem, err := this.Store.Get(id)
	if storedItem == nil {
		http.Error(rw, fmt.Sprintf("item %v not found", id), http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(rw, err.String(), http.StatusBadRequest)
		return
	}

	if err := json.NewEncoder(rw).Encode(storedItem); err != nil {
		http.Error(rw, err.String(), http.StatusBadRequest)
	}
}
// POST /jobs/id/stop or POST /jobs/id/kill
func (this *RestJsonMongo) Act(rw http.ResponseWriter, parts []string, r *http.Request) (item interface{}) {
	logger.Debug("Act(%v):%v", r.URL.Path, parts)

	if len(parts) < 2 {
		if err := this.LoadJson(r, item); err != nil {
			http.Error(rw, err.String(), http.StatusBadRequest)
			return
		} else {
			id := parts[1]
			if err := this.Store.Update(id, item); err != nil {
				http.Error(rw, err.String(), http.StatusBadRequest)
				return
			}
		}
	}

	if this.Target == nil {
		http.Error(rw, fmt.Sprintf("service not found %v", r.URL.Path), http.StatusNotImplemented)
		return
	}

	preq, _ := http.NewRequest(r.Method, r.URL.Path, r.Body)
	proxy := http.NewSingleHostReverseProxy(this.Target)
	go proxy.ServeHTTP(rw, preq)

	return
}

func (this *RestJsonMongo) LoadJson(r *http.Request, item interface{}) (err os.Error) {
	logger.Debug("LoadJson(%v)", r.URL.Path)

	if err = r.ParseForm(); err != nil {
		logger.Warn(err)
		return
	}

	jsonfile := r.Form[this.JsonParam]
	if err = json.NewDecoder(strings.NewReader(jsonfile[0])).Decode(&item); err != nil {
		logger.Warn(err)
	}
	return
}

type ItemsHandle struct {
	Items         []interface{}
	NumberOfItems int
}

func UniqueId() string {
	logger.Debug("UniqueId")
	subId := make([]byte, 16)
	if _, err := rand.Read(subId); err != nil {
		logger.Warn(err)
	}
	return fmt.Sprintf("%x", subId)
}
