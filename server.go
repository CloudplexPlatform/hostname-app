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
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	log      *logrus.Logger
	port     = "3550"
	DirPath  = ""
	fileName = "callerInfo"
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

		TimestampFormat: time.RFC3339,
	}
	log.SetReportCaller(true)
	log.Out = os.Stdout
}
func mapEnv(target *string, envKey string) {
	v := os.Getenv(envKey)
	if v != "" {
		//panic(fmt.Sprintf("environment variable %q not set", envKey))
		*target = v
	}
}
func main() {
	mapEnv(&port, "PORT")
	log.Infof("starting http server at :%s", port)
	mapEnv(&DirPath, "DIR_PATH")
	log.Infof("storing caller info at :%s", DirPath)
	//gin gonic for http requests
	g := gin.Default()
	g.GET("/hostname", getHostname)
	g.GET("/callerinfo", getCallerInfo)
	panic(g.Run(fmt.Sprintf(":%s", port)))
}

func getHostname(g *gin.Context) {
	writeLastCallerInfo(g.Request)
	fqdn, err := GetFQDN()
	if err != nil {
		g.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	g.JSON(http.StatusOK, fmt.Sprintf("my hostname is %s", fqdn))
	return
}

func getCallerInfo(g *gin.Context) {
	data, err := readCallerInfo()
	if err != nil {
		g.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	fqdn, err := GetFQDN()
	if err != nil {
		g.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	g.JSON(http.StatusOK, gin.H{"hostname": fqdn, "caller": data})
	return
}

//go-fqdn
func GetFQDN() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return "", fmt.Errorf("hostname not found: Error: %s", err.Error())
	}

	addrs, err := net.LookupIP(hostname)
	if err != nil {
		return hostname, nil
	}

	for _, addr := range addrs {
		if ipv4 := addr.To4(); ipv4 != nil {
			ip, err := ipv4.MarshalText()
			if err != nil {
				return hostname, nil
			}
			hosts, err := net.LookupAddr(string(ip))
			if err != nil || len(hosts) == 0 {
				return hostname, nil
			}
			fqdn := hosts[0]
			return strings.TrimSuffix(fqdn, "."), nil // return fqdn without dot
		}
	}
	return hostname, nil
}

type CallerRequest struct {
	Host       string `json:"host"`
	RemoteAddr string `json:"remote_addr"`
}

func writeLastCallerInfo(data *http.Request) {
	//if no directory is specified
	if DirPath == "" {
		return
	}
	req := CallerRequest{
		Host:       data.Host,
		RemoteAddr: data.RemoteAddr,
	}
	raw, err := json.Marshal(req)
	if err != nil {
		log.Errorf("unable to marshal caller info data: %s", err.Error())
		return
	}
	err = ioutil.WriteFile(fmt.Sprintf("%s/%s", DirPath, fileName), raw, 0644)
	if err != nil {
		log.Errorf("unable to write caller info data: %s", err.Error())
		return
	}
}

func readCallerInfo() (*CallerRequest, error) {
	//if no directory is specified
	if DirPath == "" {
		return nil, nil
	}
	rawData, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", DirPath, fileName))
	if err != nil {
		log.Errorf("unable to marshal caller info data: %s", err.Error())
		return nil, err
	}
	callerInfo := CallerRequest{}
	err = json.Unmarshal(rawData, &callerInfo)
	if err != nil {
		log.Errorf("unable to unmarshal caller info data: %s", err.Error())
		return nil, err
	}
	return &callerInfo, nil
}
