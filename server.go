// Copyright 2020 Cloudplex. Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	log  *logrus.Logger
	port = "3550"
)

func init() {
	log = logrus.New()
	log.Formatter = &logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "severity",
			logrus.FieldKeyFile:  "file",
			logrus.FieldKeyFunc:  "caller",
			logrus.FieldKeyMsg:   "message",
		},
		PrettyPrint:     true,
		TimestampFormat: time.RFC3339,
	}
	log.SetReportCaller(true)
	log.Out = os.Stdout
}

func main() {
	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}
	log.Infof("starting http server at :%s", port)
	//gin gonic for http requests
	g := gin.Default()
	g.GET("/hostname", getHostname)

	panic(g.Run(fmt.Sprintf(":%s", port)))
}

func getHostname(g *gin.Context) {
	g.JSON(http.StatusOK, GetFQDN())
	return
}

//go-fqdn
func GetFQDN() string {
	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Sprintf("hostname not found: Error: %s", err.Error())
	}

	addrs, err := net.LookupIP(hostname)
	if err != nil {
		return hostname
	}

	for _, addr := range addrs {
		if ipv4 := addr.To4(); ipv4 != nil {
			ip, err := ipv4.MarshalText()
			if err != nil {
				return hostname
			}
			hosts, err := net.LookupAddr(string(ip))
			if err != nil || len(hosts) == 0 {
				return hostname
			}
			fqdn := hosts[0]
			return strings.TrimSuffix(fqdn, ".") // return fqdn without dot
		}
	}
	return hostname
}
