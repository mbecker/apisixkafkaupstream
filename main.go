package main

import (
	"log"
	"os"
	"path/filepath"

	"go.uber.org/zap/zapcore"

	"github.com/apache/apisix-go-plugin-runner/pkg/runner"
	_ "github.com/mbecker/apisixkafkaupstream/plugins"
)

const (
	LogFilePath = "/usr/local/apisix/logs/runner.log"
)

func openFileToWrite(name string) (*os.File, error) {
	dir := filepath.Dir(name)
	if dir != "." {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return nil, err
		}
	}
	f, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func main() {
	log.Print("go-plugin main()")
	cfg := runner.RunnerConfig{}
	cfg.LogLevel = zapcore.DebugLevel
	f, err := openFileToWrite(LogFilePath)
	if err != nil {
		log.Fatalf("failed to open log: %s", err)
	}
	cfg.LogOutput = f
	runner.Run(cfg)
}
