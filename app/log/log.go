package log

/*
import (
	"encoding/json"
	"sync"

	"go.uber.org/zap"
)

var once sync.Once
var logger *zap.Logger

func init() {
	once.Do(func() {
		rawJSON := []byte(`{
			"level": "info",
			"encoding": "json",
			"outputPaths": ["stdout", "/tmp/logs"],
			"errorOutputPaths": ["stderr"],
			"encoderConfig": {
			  "messageKey": "message",
			  "levelKey": "level",
			  "levelEncoder": "lowercase"
			}
		  }`)

		var cfg zap.Config
		if err := json.Unmarshal(rawJSON, &cfg); err != nil {
			panic(err)
		}
		logger = zap.Must(cfg.Build())
	})
}
func GetLogger() *zap.Logger {
	return logger
}
*/
