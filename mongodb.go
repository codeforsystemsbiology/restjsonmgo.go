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
	"os"
	"launchpad.net/mgo"
	"launchpad.net/gobson/bson"
)

type JsonStore struct {
	Domain string
	Host   string
	Store  string
}

func (this *JsonStore) GetCollection() (c mgo.Collection, err os.Error) {
	logger.Debug("GetCollection()")
	session, err := mgo.Mongo(this.Host)
	if err != nil {
		return
	}

	session.SetMode(mgo.Strong, true) // [Safe, Monotonic, Strong] Strong syncs on inserts/updates
	db := session.DB(this.Store)
	c = db.C(this.Domain)
	return
}

func (this *JsonStore) Create(itemId string, item interface{}) (err os.Error) {
	logger.Debug("Create(%v,%v)", itemId, item)
	item["Id"] = itemId

	collection, err := this.GetCollection()
	if err != nil {
		logger.Warn(err)
		return
	}

	if err = collection.Insert(item); err != nil {
		logger.Warn(err)
	}

	return
}

func (this *JsonStore) Get(itemId string) (item interface{}, err os.Error) {
	logger.Debug("Get(%v)", itemId)
	collection, err := this.GetCollection()
	if err != nil {
		logger.Warn(err)
		return
	}

	if err = collection.Find(bson.M{"Id": itemId}).One(&item); err != nil {
		logger.Warn(err)
	}
	return
}

func (this *JsonStore) Update(itemId string, item interface{}) (err os.Error) {
	logger.Debug("Update(%v,%v)", itemId, item)
	if itemId == "" {
		return os.NewError("Existing Item Not Found")
	}

	collection, err := this.GetCollection()
	if err != nil {
		logger.Warn(err)
		return
	}

	return collection.Update(bson.M{"Id": itemId}, item)
}

func (this *JsonStore) Find(m map[string]interface{}) (items []interface{}, err os.Error) {
	//    queryByValues := bson.M{}
	//    for key, val := range values {
	//        queryByValues[key] = val
	//    }
	//    return this.Find(values)
	logger.Debug("Find(%v)", m)
	collection, err := this.GetCollection()
	if err != nil {
		logger.Warn(err)
		return
	}

	iter, err := collection.Find(m).Iter()
	if err != nil {
		logger.Warn(err)
		return
	}

	for {
		item := new(interface{})
		if nexterr := iter.Next(&item); nexterr != nil {
			logger.Warn(nexterr)
			break
		}
		items = append(items, item)
	}
	return
}

func (this *JsonStore) Count(m map[string]interface{}) (int, os.Error) {
	logger.Debug("Count(%v)", m)

	collection, err := this.GetCollection()
	if err != nil {
		logger.Warn(err)
		return 0, err
	}

	return collection.Find(m).Count()
}
