package logging

import (
	"github.com/SENERGY-Platform/go-service-base/srv-base"
	"github.com/y-du/go-log-level"
	"os"
)

var Logger *log_level.Logger

func InitLogger(config srv_base.LoggerConfig) (out *os.File, err error) {
	Logger, out, err = srv_base.NewLogger(config)
	return
}
