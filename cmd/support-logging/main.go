//
// Copyright (c) 2018
// Cavium
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

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
	"github.com/Circutor/edgex/internal/pkg/correlation"
	"github.com/Circutor/edgex/internal/pkg/startup"
	"github.com/Circutor/edgex/internal/pkg/usage"
	"github.com/Circutor/edgex/internal/support/logging"
	"github.com/Circutor/edgex/pkg/clients/logger"
)

func main() {
	start := time.Now()
	var useProfile string

	flag.StringVar(&useProfile, "profile", "", "Specify a profile other than default.")
	flag.StringVar(&useProfile, "p", "", "Specify a profile other than default.")
	flag.Usage = usage.HelpCallback
	flag.Parse()

	params := startup.BootParams{UseProfile: useProfile, BootTimeout: internal.BootTimeoutDefault}
	startup.Bootstrap(params, logging.Retry, logBeforeInit)

	ok := logging.Init()
	if !ok {
		time.Sleep(time.Millisecond * time.Duration(15))
		logBeforeInit(fmt.Errorf("%s: Service bootstrap failed!", internal.SupportLoggingServiceKey))
		os.Exit(1)
	}
	logging.LoggingClient.Info("Service dependencies resolved...")
	logging.LoggingClient.Info(fmt.Sprintf("Starting %s %s", internal.SupportLoggingServiceKey, edgex.Version))

	errs := make(chan error, 2)
	listenForInterrupt(errs)

	// Time it took to start service
	logging.LoggingClient.Info("Service started in: " + time.Since(start).String())
	logging.LoggingClient.Info("Listening on port: " + strconv.Itoa(logging.Configuration.Service.Port))
	startHTTPServer(errs)

	c := <-errs
	logging.Destruct()
	logging.LoggingClient.Warn(fmt.Sprintf("terminated %v", c))

	os.Exit(0)
}

func logBeforeInit(err error) {
	l := logger.NewClient(internal.SupportLoggingServiceKey, false, "", logger.InfoLog)
	l.Error(err.Error())
}

func listenForInterrupt(errChan chan error) {
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT)
		errChan <- fmt.Errorf("%s", <-c)
	}()
}

func startHTTPServer(errChan chan error) {
	go func() {
		correlation.LoggingClient = logging.LoggingClient
		p := fmt.Sprintf(":%d", logging.Configuration.Service.Port)
		errChan <- http.ListenAndServe(p, logging.HttpServer())
	}()
}
