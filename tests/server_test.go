package main_test

import (
	"strings"
	"testing"

	"github.com/AYGA2K/dictago/handlers"
)

func TestInfoCommand(t *testing.T) {
	// Test INFO replication as master
	infoCmdMaster := []string{"INFO", "replication"}
	infoResultMaster := handlers.Info(infoCmdMaster, "", "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb")
	if !strings.Contains(infoResultMaster, "role:master") {
		t.Errorf("INFO replication (master): Expected role:master, got %q", infoResultMaster)
	}

	// Test INFO replication as replica
	infoCmdReplica := []string{"INFO", "replication"}
	infoResultReplica := handlers.Info(infoCmdReplica, "localhost:6380", "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb")
	if !strings.Contains(infoResultReplica, "role:slave") {
		t.Errorf("INFO replication (replica): Expected role:slave, got %q", infoResultReplica)
	}
}
