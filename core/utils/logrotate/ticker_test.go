package logrotate

import (
	"testing"
	"time"
)

func Test_CalRotateTimeDuration(t *testing.T) {
	t.Skip()
	d := CalRotateTimeDuration(time.Now(), time.Hour*2)
	t.Logf("now:%s, duration:%s\n", time.Now(), d)
}
