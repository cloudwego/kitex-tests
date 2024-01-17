package common

import (
	"net"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
)

func WaitServer(hostPort string) {
	for begin := time.Now(); time.Since(begin) < time.Second; {
		if _, err := net.Dial("tcp", hostPort); err == nil {
			klog.Infof("server %s is up", hostPort)
			return
		}
		time.Sleep(time.Millisecond * 10)
	}
}
