package main

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/copyagent/copyagentd/internal/config"
	"github.com/copyagent/copyagentd/internal/transport/feishu"
)

type Report struct {
	Scenario string           `json:"scenario"`
	PID      int              `json:"pid"`
	Memory   runtime.MemStats `json:"memory"`
}

func main() {
	scenario := "baseline"
	if len(os.Args) > 1 {
		scenario = os.Args[1]
	}
	if scenario == "feishu-client" {
		cfg, _ := config.Load("")
		_ = feishu.NewTransport(cfg.FeishuAppID, cfg.FeishuAppSecret)
	}
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	_ = json.NewEncoder(os.Stdout).Encode(Report{Scenario: scenario, PID: os.Getpid(), Memory: mem})
	fmt.Fprintln(os.Stderr, "ready")
	for {
		time.Sleep(time.Minute)
	}
}
