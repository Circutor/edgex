/*******************************************************************************
 * Copyright 2018 Dell Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *******************************************************************************/
package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/Circutor/edgex"
	"github.com/Circutor/edgex/internal"
	"github.com/Circutor/edgex/internal/core/command"
	"github.com/Circutor/edgex/internal/pkg/correlation"
	"github.com/Circutor/edgex/internal/pkg/startup"
	"github.com/Circutor/edgex/internal/pkg/usage"
	"github.com/Circutor/edgex/pkg/clients/logger"
	"github.com/gorilla/context"
)

func main() {
	start := time.Now()
	var useProfile string

	flag.StringVar(&useProfile, "profile", "", "Specify a profile other than default.")
	flag.StringVar(&useProfile, "p", "", "Specify a profile other than default.")
	flag.Usage = usage.HelpCallback
	flag.Parse()

	params := startup.BootParams{UseProfile: useProfile, BootTimeout: internal.BootTimeoutDefault}
	startup.Bootstrap(params, command.Retry, logBeforeInit)

	ok := command.Init()
	if !ok {
		logBeforeInit(fmt.Errorf("%s: Service bootstrap failed!", internal.CoreCommandServiceKey))
		os.Exit(1)
	}

	command.LoggingClient.Info("Service dependencies resolved...")
	command.LoggingClient.Info(fmt.Sprintf("Starting %s %s ", internal.CoreCommandServiceKey, edgex.Version))

	http.TimeoutHandler(nil, time.Millisecond*time.Duration(command.Configuration.Service.Timeout), "Request timed out")
	command.LoggingClient.Info(command.Configuration.Service.StartupMsg)

	errs := make(chan error, 2)
	listenForInterrupt(errs)
	startHttpServer(errs, command.Configuration.Service.Port)

	// Time it took to start service
	command.LoggingClient.Info("Service started in: " + time.Since(start).String())
	command.LoggingClient.Info("Listening on port: " + strconv.Itoa(command.Configuration.Service.Port))
	c := <-errs
	command.Destruct()
	command.LoggingClient.Warn(fmt.Sprintf("terminating: %v", c))

	os.Exit(0)
}

func logBeforeInit(err error) {
	command.LoggingClient = logger.NewClient(internal.CoreCommandServiceKey, false, "", logger.InfoLog)
	command.LoggingClient.Error(err.Error())
}

func listenForInterrupt(errChan chan error) {
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errChan <- fmt.Errorf("%s", <-c)
	}()
}

func startHttpServer(errChan chan error, port int) {
	go func() {
		correlation.LoggingClient = command.LoggingClient
		r := command.LoadRestRoutes()
		errChan <- http.ListenAndServe(":"+strconv.Itoa(port), context.ClearHandler(r))
	}()
}
