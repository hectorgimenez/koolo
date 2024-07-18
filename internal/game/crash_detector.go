package game

import (
	"log/slog"
	"time"

	"golang.org/x/sys/windows"
)

type CrashDetector struct {
	pid         int32
	supervisor  string
	hwnd        uintptr
	logger      *slog.Logger
	restartFunc func()
	stopChan    chan struct{}
}

func NewCrashDetector(sup string, pid int32, hwnd uintptr, logger *slog.Logger, restartFunc func()) *CrashDetector {
	return &CrashDetector{
		supervisor:  sup,
		pid:         pid,
		hwnd:        hwnd,
		logger:      logger,
		restartFunc: restartFunc,
		stopChan:    make(chan struct{}),
	}
}

func (cd *CrashDetector) Start() {
	cd.logger.Info("Starting Crash Detector ...", slog.Int("PID", int(cd.pid)), slog.String("Supervisor", cd.supervisor))
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-cd.stopChan:
			cd.logger.Info("Crash Detector stopped.", slog.Int("PID", int(cd.pid)), slog.String("Supervisor", cd.supervisor))
			return
		case <-ticker.C:
			if !cd.isProcessRunning() {
				cd.logger.Error("Client crash detected ...", slog.Int("PID", int(cd.pid)), slog.String("Supervisor", cd.supervisor))
				if cd.restartFunc != nil {
					cd.logger.Info("Attempting to restart client ...", slog.String("Supervisor", cd.supervisor))
					cd.restartFunc()
				}
				return
			}
		}
	}
}

func (cd *CrashDetector) Stop() {
	cd.logger.Info("Stopping Crash Detector", slog.Int("PID", int(cd.pid)), slog.String("Supervisor", cd.supervisor))
	close(cd.stopChan)
}

func (cd *CrashDetector) isProcessRunning() bool {
	handle, err := windows.OpenProcess(windows.PROCESS_QUERY_INFORMATION, false, uint32(cd.pid))
	if err != nil {
		cd.logger.Debug("Failed to open process", slog.Int("PID", int(cd.pid)), slog.String("err", err.Error()))
		return false
	}
	defer windows.CloseHandle(handle)

	var exitCode uint32
	err = windows.GetExitCodeProcess(handle, &exitCode)
	if err != nil {
		cd.logger.Debug("Failed to get exit code", slog.Int("PID", int(cd.pid)), slog.String("error", err.Error()))
		return false
	}

	isRunning := exitCode == 259 // STILL_ACTIVE

	return isRunning
}
