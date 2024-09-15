package utils

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const MAX_FILE_COUNT = 1000
const MAX_FILE_BYTES = 10 * 1024 * 1024
const MAX_MESSAGE_BYTES = 10 * 1024

type MyLogger struct {
	StdOutLogger     *log.Logger
	FileLogger       *log.Logger
	logLevel         logLevel
	cumBytes         int64
	rolloverCallback func()
}

type logLevel uint8

const (
	logLevelTrace logLevel = iota
	logLevelDebug
	logLevelInfo
	logLevelWarning
	logLevelError
)

type nowFunc func() string

var now nowFunc = func() string {
	return time.Now().UTC().Format("2006-01-02 15:04:05.000")
}

func newLogger(out io.Writer) *log.Logger {
	return log.New(out, "", 0)
}

var Logger *MyLogger = &MyLogger{newLogger(os.Stdout), nil, logLevel(0), 0, func() {}}

func openLogFile() *os.File {
	directory := filepath.Join(GetExecDirectory("."), "logs")
	name := "server.log"
	path := filepath.Join(directory, generateFilename(name, directory))
	logFile, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		err = os.MkdirAll(directory, 0755)
		if err != nil {
			panic(fmt.Sprintf("failed to create log directory: %s", err))
		}

		logFile, err = os.Create(path)
		if err != nil {
			panic(fmt.Sprintf("failed to create log file: %s", err))
		}
	}
	return logFile
}

func generateFilename(name, directory string) string {
	filenameArr := strings.SplitN(name, ".", 2)
	basename := filenameArr[0]
	extension := ".log"
	if len(filenameArr) > 1 && filenameArr[1] != "" {
		extension = "." + filenameArr[1]
	}
	basename += "-" + time.Now().UTC().Format("20060102")

	count := 1
	for {
		filename := fmt.Sprint(basename, "-", count, extension)
		path := filepath.Join(directory, filename)
		_, err := os.Stat(path)
		if err != nil {
			return filename
		}
		count++
		if count > MAX_FILE_COUNT {
			panic(fmt.Sprint(
				"reached maximum allowed number of log files: ",
				MAX_FILE_COUNT,
			))
		}
	}
}

func InitLogger() {
	Logger.logLevel = logLevel(0)
	Logger.FileLogger = newLogger(openLogFile())
	LogTrace("file logger initialised")
}

func SetRolloverCallback(cb func()) {
	Logger.rolloverCallback = cb
}

func CloseLogger() {
	if Logger.FileLogger == nil {
		return
	}
	if file, ok := Logger.FileLogger.Writer().(*os.File); ok {
		err := file.Close()
		if err != nil {
			panic(fmt.Sprint("failed to close log file: ", err))
		}
		LogTrace("log file closed")
	}
}

func logMessage(level string, ansiColour string, message string) {
	reset := "\033[0m"
	time := now()
	if len(message) > MAX_MESSAGE_BYTES {
		message = message[:MAX_MESSAGE_BYTES]
	}
	Logger.StdOutLogger.Printf("%s %s[%s]%s %s", time, ansiColour, level, reset, message)
	if Logger.FileLogger != nil {
		if Logger.cumBytes += int64(len(message)); Logger.cumBytes > MAX_FILE_BYTES {
			rollover()
		}
		Logger.FileLogger.Printf("%s [%s] %s", time, level, message)
	}
}

func rollover() {
	if file, ok := Logger.FileLogger.Writer().(*os.File); ok {
		if stats, err := file.Stat(); err == nil && stats.Size() > MAX_FILE_BYTES {
			Logger.cumBytes = 0
			LogTrace(fmt.Sprintf(
				"reached maximum log file size (%d bytes), rolling over",
				MAX_FILE_BYTES,
			)) // can cause infinite loop if MAX_FILE_BYTES is too low
			CloseLogger()
			Logger.FileLogger = newLogger(openLogFile())
			Logger.rolloverCallback()
			LogTrace("new log file opened")
		}
	}
}

func LogTrace(message string) {
	if Logger.logLevel <= logLevelTrace {
		logMessage("TRACE", "", message)
	}
}

func LogDebug(message string) {
	if Logger.logLevel <= logLevelDebug {
		logMessage("DEBUG", "\033[34m", message)
	}
}

func LogInfo(message string) {
	if Logger.logLevel <= logLevelInfo {
		logMessage("INFO", "\033[36m", message)
	}
}

func LogWarning(message string) {
	if Logger.logLevel <= logLevelWarning {
		logMessage("WARNING", "\033[33;1m", message)
	}
}

func LogError(message string) {
	if Logger.logLevel <= logLevelError {
		buf := make([]byte, 1<<16)
		n := runtime.Stack(buf, false)
		stackTrace := strings.ReplaceAll(string(buf[:n-1]), "\n", "\n\t")
		message += fmt.Sprintf("\n\tstack trace: %s", stackTrace)
		logMessage("ERROR", "\033[31;1m", message)
	}
}
