//
// Copyright (c) 2018 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package logging

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Circutor/edgex/pkg/models"
)

const (
	rmFileSuffix string = ".tmp"
)

type fileLog struct {
	filename  string
	maxBytes  int64
	logsCount int
	out       *os.File //io.WriteCloser
}

func (fl *fileLog) closeSession() {
	if fl.out != nil {
		fl.out.Close()
	}
}

func (fl *fileLog) add(le models.LogEntry) error {
	if fl.out == nil {
		var err error
		//First check to see if the specified directory exists
		//File won't be written without directory.
		path := filepath.Dir(fl.filename)
		if _, err = os.Stat(path); os.IsNotExist(err) {
			os.MkdirAll(path, 0755)
		}
		fl.out, err = os.OpenFile(fl.filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			//fmt.Println("Error opening log file: ", fl.filename, err)
			fl.out = nil
			return err
		}
	}
	stat, err := fl.out.Stat()
	if err != nil {
		//fmt.Println("Error reading log file size: ", fl.filename, err)
		return err
	}
	if stat.Size() > fl.maxBytes {
		fl.out.Close()

		for i := fl.logsCount - 1; i > 0; i-- {
			oldFileName := fmt.Sprintf("%s.%d", fl.filename, i)
			newFileName := fmt.Sprintf("%s.%d", fl.filename, i+1)

			os.Rename(oldFileName, newFileName)
		}
		newFileName := fmt.Sprintf("%s.1", fl.filename)
		os.Rename(fl.filename, newFileName)
		fl.out, _ = os.OpenFile(fl.filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	}
	res, err := json.Marshal(le)
	if err != nil {
		return err
	}
	fl.out.Write(res)
	fl.out.Write([]byte("\n"))

	return nil
}

func (fl *fileLog) remove(criteria matchCriteria) (int, error) {
	var allCount int
	for i := fl.logsCount - 1; i > 0; i-- {
		var logPartialFile string
		if i > 0 {
			logPartialFile = fmt.Sprintf("%s.%d", fl.filename, i)
		} else {
			logPartialFile = fl.filename
		}

		tmpFilename := logPartialFile + rmFileSuffix
		tmpFile, err := os.OpenFile(tmpFilename, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			//fmt.Println("Error creating tmp log file: ", tmpFilename, err)
			return 0, err
		}

		defer os.Remove(tmpFilename)

		count := 0
		f, err := os.Open(logPartialFile)
		if err != nil {
			fmt.Println("Error opening log file: ", logPartialFile, err)
			tmpFile.Close()
			return 0, err
		}
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			var le models.LogEntry

			line := scanner.Bytes()
			err := json.Unmarshal(line, &le)
			if err == nil {
				if !criteria.match(le) {
					tmpFile.Write(line)
					tmpFile.Write([]byte("\n"))
				} else {
					count += 1
				}
			}
		}

		tmpFile.Close()
		if count != 0 {
			err = os.Rename(tmpFilename, logPartialFile)
			if err != nil {
				//fmt.Printf("Error renaming %s to %s: %v", tmpFilename, fl.filename, err)
				return 0, err
			}

			// Close old file to open the new one when writing next log
			if fl.out != nil {
				fl.out.Close()
				fl.out = nil
			}

		}
		allCount += count
	}
	return allCount, nil
}

func (fl *fileLog) find(criteria matchCriteria) ([]models.LogEntry, error) {
	var logs []models.LogEntry
	var err error
	// Here we should make a for to include all files in logs
	for i := fl.logsCount - 1; i > 0; i-- {
		var logPartialFile string
		if i > 0 {
			logPartialFile = fmt.Sprintf("%s.%d", fl.filename, i)
		} else {
			logPartialFile = fl.filename
		}
		f, _ := os.Open(logPartialFile)
		/*if err != nil {
			//fmt.Println("Error opening log file: ", fl.filename, err)
			return nil, err
		}*/
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			var le models.LogEntry

			line := scanner.Bytes()
			err := json.Unmarshal(line, &le)
			if err == nil {
				if criteria.match(le) {
					logs = append(logs, le)

					if criteria.Limit != 0 && len(logs) >= criteria.Limit {
						break
					}
				}
			}
		}
	}
	return logs, err
}

func (fl *fileLog) reset() {
	if fl.out != nil {
		fl.out.Close()
		fl.out = nil
	}
	os.Remove(fl.filename)
}
