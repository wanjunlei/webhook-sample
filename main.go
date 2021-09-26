/*
Copyright 2020 The KubeSphere Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"fmt"
	"github.com/emicklei/go-restful"
	"github.com/golang/glog"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

var (
	cn       string
	user     string
	password string
)

func AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&cn, "cn", "", "Common name")
	fs.StringVar(&user, "user", "admin", "user name")
}

func main() {

	cmd := NewServerCommand()

	if err := cmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}

func NewServerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "notification-adapter",
		Long: `The webhook to receive alert from notification manager, and send to socket`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return Run()
		},
	}
	AddFlags(cmd.Flags())
	cmd.Flags().AddGoFlagSet(flag.CommandLine)

	return cmd
}

func Run() error {

	pflag.VisitAll(func(flag *pflag.Flag) {
		glog.Errorf("FLAG: --%s=%q", flag.Name, flag.Value)
	})

	password = uuid.New().String()
	token, err := generateToken(user)
	if err != nil {
		glog.Fatal(err)
	}
	fmt.Printf("User: %s\n", user)
	fmt.Printf("Password: %s\n", password)
	fmt.Printf("Token: %s\n", token)

	go httpsServer()

	return httpServer()
}

func httpServer() error {
	container := restful.NewContainer()
	ws := new(restful.WebService)
	ws.Path("").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)
	ws.Route(ws.POST("/notifications").To(handler))
	ws.Route(ws.GET("/readiness").To(readiness))
	ws.Route(ws.GET("/liveness").To(readiness))
	ws.Route(ws.GET("/preStop").To(preStop))

	container.Add(ws)

	server := &http.Server{
		Addr:    ":8080",
		Handler: container,
	}

	if err := server.ListenAndServe(); err != nil {
		glog.Fatal(err)
	}

	return nil
}

func httpsServer() {
	container := restful.NewContainer()
	ws := new(restful.WebService)
	ws.Path("").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)
	ws.Route(ws.POST("/notifications").To(handler))
	ws.Route(ws.GET("/readiness").To(readiness))
	ws.Route(ws.GET("/liveness").To(readiness))
	ws.Route(ws.GET("/preStop").To(preStop))

	container.Add(ws)

	if cn == "" {
		cn = "webhook-sample"
	}
	_, serverKey, serverCrt, err := CreateCa(cn)
	if err != nil {
		glog.Fatal(err)
	}

	file, err := os.Create("tls.key")
	if err != nil {
		glog.Fatal(err)
	}

	_, err = file.Write(serverKey)
	if err != nil {
		glog.Fatal(err)
	}

	file, err = os.Create("tls.crt")
	if err != nil {
		glog.Fatal(err)
	}

	_, err = file.Write(serverCrt)
	if err != nil {
		glog.Fatal(err)
	}

	server := &http.Server{
		Addr:    ":443",
		Handler: container,
	}

	if err := server.ListenAndServeTLS("tls.crt", "tls.key"); err != nil {
		glog.Fatal(err)
	}

	return
}

func handler(req *restful.Request, resp *restful.Response) {

	authorization := req.HeaderParameter("Authorization")
	if authorization != "" {
		if strings.HasPrefix(authorization, "Basic") {
			u, p, ok := req.Request.BasicAuth()
			if !ok || u != user || p != password {
				_ = resp.WriteHeaderAndEntity(http.StatusUnauthorized, "wrong user or password")
				return
			}
		} else if strings.HasPrefix(authorization, "Bearer") {
			token := strings.TrimPrefix(authorization, "Bearer ")
			if !parseToken(user, token) {
				_ = resp.WriteHeaderAndEntity(http.StatusUnauthorized, "invalid token")
				return
			}
		}
	}

	body, err := ioutil.ReadAll(req.Request.Body)
	if err != nil {
		glog.Errorf("read request body error, %s", err)
		_ = resp.WriteHeaderAndEntity(http.StatusBadRequest, "")
		return
	}

	fmt.Println(string(body))

	_ = resp.WriteHeaderAndEntity(http.StatusOK, "")
}

//readiness
func readiness(_ *restful.Request, resp *restful.Response) {

	responseWithHeaderAndEntity(resp, http.StatusOK, "")
}

//preStop
func preStop(_ *restful.Request, resp *restful.Response) {
	responseWithHeaderAndEntity(resp, http.StatusOK, "")
	glog.Flush()
}

func responseWithHeaderAndEntity(resp *restful.Response, status int, value interface{}) {
	e := resp.WriteHeaderAndEntity(status, value)
	if e != nil {
		glog.Errorf("response error %s", e)
	}
}
