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
	"http"
	"os"
	"url"
	"github.com/codeforsystemsbiology/rest.go"
	"github.com/codeforsystemsbiology/verboselogger.go"
	"goconf.googlecode.com/hg"
)

var logger *log4go.VerboseLogger
var configFile *conf.ConfigFile

// starts HTTP server with REST resources for configured domains
func main() {
	InitConfigFile()
	InitLogger()
	StartREST()
	StartHtmlHandler()
	ListenAndServe()
}

// sets up global config file
// looks for "restjsonmgo.config" in execution PATH
func InitConfigFile() {
	if cf, err := conf.ReadConfigFile("restjsonmgo.config"); err != nil {
		logger.Fatal(err.String())
	} else {
		configFile = cf
	}
}

// sets global logger based on verbosity level in configuration
// optional parameter:  default.verbose (defaults to true if not present or incorrectly set)
func InitLogger() {
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
func StartREST() {
	domains := configFile.GetSections()
	for _, domain := range domains {
		dbhost := GetRequiredString(configFile, domain, "dbHost")
		dbstore := GetRequiredString(configFile, domain, "dbName")
		jsonParameter := GetRequiredString(configFile, domain, "jsonParameter")
		proxyTarget, err := configFile.GetString(domain, "serviceProxy")
		if err != nil {
			logger.Warn(err)
			logger.Print("no service proxy configured")
		}

		targetUrl, _ := url.Parse(proxyTarget)
		store := &JsonStore{Domain: domain, Host: dbhost, Store: dbstore}
		rest.Resource(domain, RestJsonMongo{Store: store, Target: targetUrl, JsonParam: jsonParameter})

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
func StartHtmlHandler() {
	if contentDir, _ := configFile.GetString("default", "contentDirectory"); contentDir != "" {
		logger.Printf("StartHtmlHandler(): serving HTML content from [%v]", contentDir)
		http.Handle("/html/", http.StripPrefix("/html/", http.FileServer(http.Dir(contentDir))))
		http.Handle("/", http.RedirectHandler("/html/", http.StatusTemporaryRedirect))
	}
}

// starts HTTP server based on hostname in configuration file
func ListenAndServe() {
	hostname := GetRequiredString(configFile, "default", "hostname")
	if err := http.ListenAndServe(hostname, nil); err != nil {
		logger.Fatal(err.String())
	}
}

func GetRequiredString(config *conf.ConfigFile, section string, key string) (value string) {
	var err os.Error
	if value, err = config.GetString(section, key); err != nil {
		logger.Fatalf("[CONFIG] [%v] required in section [%v]", key, section)
	}
	return value
}
