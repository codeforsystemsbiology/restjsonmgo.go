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
	"flag"
	"http"
	"url"
	"github.com/codeforsystemsbiology/rest.go"
	"github.com/codeforsystemsbiology/verboselogger.go"
	"goconf.googlecode.com/hg"
)

var logger *log4go.VerboseLogger

//parse args and start as master, scribe, addama proxy or worker
func main() {
	var configurationFile string
	flag.StringVar(&configurationFile, "config", "rest-json-mgo.config", "A configuration file for MongoDB REST+JSON Service")
	flag.Parse()

	configFile, err := conf.ReadConfigFile(configurationFile)
	if err != nil {
		panic(err)
	}

	GlobalLogger(configFile)
    StartREST(configFile)
    StartHtmlHandler(configFile)

    hostname := GetRequiredString(configFile, "default", "hostname")
    if err = http.ListenAndServe(hostname, nil); err != nil {
        logger.Fatal(err.String())
    }
}

// sets global logger based on verbosity level in configuration
// optional parameter:  default.verbose (defaults to true if not present or incorrectly set)
func GlobalLogger(configFile *conf.ConfigFile) {
	verbose, err := configFile.GetBool("default", "verbose")
	logger = log4go.NewVerboseLogger(verbose, nil, "")
	if err != nil {
		logger.Warn(err)
		verbose = true
	}
	logger.Printf("verbose set [%v]", verbose)
}

// starts service based on the given configuration file
// required parameters:  sections for each domain object accepted (e.g. jobs for /jobs)
// optional parameters:  [domain_group] contentType=application/json (default)
func StartREST(configFile *conf.ConfigFile) {
    domains := configFile.GetSections()
    for _, domain := range domains {
    	dbhost := GetRequiredString(configFile, domain, "dbHost")
    	dbstore := GetRequiredString(configFile, domain, "dbName")
    	proxyTarget, err := configFile.GetString(domain, "serviceProxy")
    	if err != nil {
    	    logger.Warn(err)
    	    logger.Print("no service proxy configured")
    	}

        targetUrl, _ := url.Parse(proxyTarget)
    	store := &JsonStore{Domain: domain, Host: dbhost, Store: dbstore}
        rest.Resource(domain, RestJsonMongo{Store: store, Target: targetUrl})

        contentType, err := configFile.GetString(domain, "contentType")
        if err != nil {
            logger.Warn(err)
            logger.Print("Defaulting content type to application/json")
            contentType = "application/json"
        }
        rest.ResourceContentType(domain, contentType)
    }
}

// starts http handlers for HTML content based on the given configuration file
// optional parameters:  default.contentDirectory (location of html content to be served at https://example.com/ or https://example.com/html/index.html
func StartHtmlHandler(configFile *conf.ConfigFile) {
	if contentDir, _ := configFile.GetString("default", "contentDirectory"); contentDir != "" {
		logger.Printf("StartHtmlHandler(): serving HTML content from [%v]", contentDir)
		http.Handle("/html/", http.StripPrefix("/html/", http.FileServer(http.Dir(contentDir))))
		http.Handle("/", http.RedirectHandler("/html/", http.StatusTemporaryRedirect))
	}
}

func GetRequiredString(config *conf.ConfigFile, section string, key string) (value string) {
	if value, err := config.GetString(section, key); err != nil {
		logger.Fatalf("[CONFIG] [%v] required in section [%v]", key, section)
	}
	return
}
