package util

import (
	"os"
	"syscall"
	"time"
)

func GetFileCreatedAt(path string) string {
	stat, err := os.Stat(path)
	if err != nil {
		return "-"
	}

	statSys := stat.Sys().(*syscall.Stat_t)
	creationTime := time.Unix(statSys.Birthtimespec.Sec, int64(statSys.Birthtimespec.Nsec))

	return FormatTime(creationTime)
}
