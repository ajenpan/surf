package logrotate

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type RotateLog struct {
	file *os.File

	logPath            string
	rotateTime         time.Duration
	maxAge             time.Duration
	deleteFileWildcard string
	timeFormat         string
	timePlaceholder    string

	mutex  *sync.Mutex
	rotate <-chan time.Time // notify rotate event
	close  chan struct{}    // close file and write goroutine
}

func NewRoteteLog(logPath string, opts ...Option) (*RotateLog, error) {
	rl := &RotateLog{
		mutex:   &sync.Mutex{},
		close:   make(chan struct{}, 1),
		logPath: logPath,
	}
	for _, opt := range opts {
		opt(rl)
	}

	if err := os.Mkdir(filepath.Dir(rl.logPath), 0755); err != nil && !os.IsExist(err) {
		return nil, err
	}

	if err := rl.rotateFile(time.Now()); err != nil {
		return nil, err
	}

	if rl.rotateTime != 0 {
		go rl.handleEvent()
	}

	return rl, nil
}

func (r *RotateLog) Write(b []byte) (int, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	n, err := r.file.Write(b)
	r.file.Write([]byte("\n"))
	return n, err
}

func (r *RotateLog) WriteJson(v any) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	err := json.NewEncoder(r.file).Encode(v)
	return err
}

func (r *RotateLog) Close() error {
	select {
	case r.close <- struct{}{}:
	case <-r.close:
		return nil
	}

	return r.file.Close()
}

func (r *RotateLog) handleEvent() {
	for {
		select {
		case <-r.close:
			return
		case now := <-r.rotate:
			r.rotateFile(now)
		}
	}
}

func (r *RotateLog) rotateFile(now time.Time) error {
	if r.rotateTime != 0 {
		nextRotateTime := CalRotateTimeDuration(now, r.rotateTime)
		r.rotate = time.After(nextRotateTime)
	}

	latestLogPath := r.getLatestLogPath(now)
	r.mutex.Lock()
	defer r.mutex.Unlock()
	file, err := os.OpenFile(latestLogPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	if r.file != nil {
		r.file.Close()
	}
	r.file = file

	if r.maxAge > 0 && len(r.deleteFileWildcard) > 0 { // at present
		go r.deleteExpiredFile(now)
	}

	return nil
}

// Judege expired by laste modify time
func (r *RotateLog) deleteExpiredFile(now time.Time) {
	cutoffTime := now.Add(-r.maxAge)
	matches, err := filepath.Glob(r.deleteFileWildcard)
	if err != nil {
		return
	}

	toUnlink := make([]string, 0, len(matches))
	for _, path := range matches {
		fileInfo, err := os.Stat(path)
		if err != nil {
			continue
		}

		if r.maxAge > 0 && fileInfo.ModTime().After(cutoffTime) {
			continue
		}

		toUnlink = append(toUnlink, path)
	}

	for _, path := range toUnlink {
		os.Remove(path)
	}
}

func (r *RotateLog) getLatestLogPath(t time.Time) string {
	return strings.ReplaceAll(r.logPath, r.timePlaceholder, t.Format(r.timeFormat))
}
