package simtest

import (
	"testing"

	_ "modernc.org/sqlite"

	"github.com/jeremy/mlogger-fd/internal/handler"
	"github.com/jeremy/mlogger-fd/internal/ws"
)

func TestSimulation(t *testing.T) {
	_ = ws.NewHub()
	_ = handler.HealthCheck
	t.Error("simulation test not yet implemented")
}
