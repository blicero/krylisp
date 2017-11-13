// /home/krylon/go/src/krylisp/common/common.go
// -*- mode: go; coding: utf-8; -*-
// Created on 11. 11. 2017 by Benjamin Walkenhorst
// (c) 2017 Benjamin Walkenhorst
// Time-stamp: <2017-11-11 02:52:34 krylon>

package common

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/hashicorp/logutils"
)

// TimestampFormat is the template used throughout the application for rendering
// timestamp values to a human-readable format.
// AppName is the name that application identifies itself with to the user.
// Version is the version number the app displays to the user.
// Debug, if true, causes the application to perform additional sanity checks
// and emit additional message regarding the state of teh interpreter.
const (
	TimestampFormat = "2006-01-02 15:04:05"
	AppName         = "kryLisp"
	Version         = "0.0.1"
	Debug           = true
)

// LogLevels are the names of the log levels supported by the logger.
var LogLevels = []logutils.LogLevel{
	"TRACE",
	"DEBUG",
	"INFO",
	"WARN",
	"ERROR",
	"CRITICAL",
	"CANTHAPPEN"}

// GetHomeDirectory determines the user's home directory.
// The reason this is even a thing is that Unix-like systems
// store this path in the environment variable "HOME",
// whereas Windows uses the environment variable "USERPROFILE",
// Hence this function.
func GetHomeDirectory() string {
	if runtime.GOOS == "windows" {
		return os.Getenv("USERPROFILE")
	}

	return os.Getenv("HOME")
} // func GetHomeDirectory() string

// BaseDir is the folder where all application-specific files are stored.
// It defaults to $HOME/.Kuang2.d
var BaseDir = filepath.Join(
	GetHomeDirectory(),
	fmt.Sprintf(".%s.d", strings.ToLower(AppName)))

// LogPath is the filename of the log file.
var LogPath = filepath.Join(BaseDir, fmt.Sprintf("%s.log", AppName))

// InitApp performs some basic preparations for the application to run.
// Currently, this means creating the BaseDir folder.
func InitApp() error {
	err := os.Mkdir(BaseDir, 0700)
	if err != nil {
		if !os.IsExist(err) {
			msg := fmt.Sprintf("Error creating BASE_DIR %s: %s", BaseDir, err.Error())
			return errors.New(msg)
		}
	}

	return nil
} // func InitApp() error

// SetBaseDir sets the application's base directory. This should only be
// done during initialization.
// Once the log file and the database are opened, this
// is useless at best and opens a world of confusion at worst.
func SetBaseDir(path string) error {
	fmt.Printf("Setting BASE_DIR to %s\n", path)

	BaseDir = path
	LogPath = filepath.Join(BaseDir, fmt.Sprintf("%s.log", strings.ToLower(AppName)))

	if err := InitApp(); err != nil {
		var msg = fmt.Sprintf("Error initializing application environment: %s\n",
			err.Error())
		fmt.Println(msg)
		return errors.New(msg)
	}

	return nil
} // func SetBaseDir(path string)

// GetLogger tries to create a named logger instance and return it.
// If the directory to hold the log file does not exist, try to create it.
func GetLogger(name string) (*log.Logger, error) {
	var err error

	if err = InitApp(); err != nil {
		return nil, fmt.Errorf("Error initializing application environment: %s", err.Error())
	}

	logName := fmt.Sprintf("%s.%s",
		AppName,
		name)

	var logfile *os.File

	if logfile, err = os.OpenFile(LogPath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0600); err != nil {
		msg := fmt.Sprintf("Error opening log file: %s\n", err.Error())
		fmt.Println(msg)
		return nil, errors.New(msg)
	}

	var writer io.Writer

	if Debug {
		writer = io.MultiWriter(os.Stdout, logfile)
	} else {
		writer = logfile
	}

	var lvl logutils.LogLevel
	if Debug {
		lvl = logutils.LogLevel("DEBUG")
	} else {
		lvl = logutils.LogLevel("INFO")
	}

	filter := &logutils.LevelFilter{
		Levels:   LogLevels,
		MinLevel: lvl,
		Writer:   writer,
	}

	logger := log.New(filter, logName, log.Ldate|log.Ltime|log.Lshortfile)
	return logger, nil
} // func GetLogger(name string) (*log.Logger, error)

// GetLoggerStdout returns a Logger that will log only to stdout, but not to disk.
func GetLoggerStdout(name string) (*log.Logger, error) {
	var (
		writer  io.Writer = os.Stdout
		lvl     logutils.LogLevel
		logName = fmt.Sprintf("%s.%s",
			AppName,
			name)
	)

	if Debug {
		fmt.Println("Just a heads up - there'll be many debug messages.")
		lvl = logutils.LogLevel("DEBUG")
	} else {
		lvl = logutils.LogLevel("INFO")
	}

	filter := &logutils.LevelFilter{
		Levels:   LogLevels,
		MinLevel: lvl,
		Writer:   writer,
	}

	logger := log.New(filter, logName, log.Ldate|log.Ltime|log.Lshortfile)
	return logger, nil
} // func GetLoggerStdout(name string) (*log.Logger, error)
