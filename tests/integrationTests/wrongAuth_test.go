package integrationtests

import (
	"strconv"
	"testing"
	"time"

	"DataConsumer/internal/natsutil"

	"github.com/nats-io/nats.go"
)

func TestNATSConnection_WrongCredsPath_ReturnsConnectError(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test skipped in short mode")
	}

	cfg := loadIntegrationNATSConfig()
	if err := ensureNATSReachable(cfg); err != nil {
		t.Skipf("NATS non raggiungibile con la configurazione valida (%v), integrazione saltata", err)
	}

	url := "nats://" + cfg.Address + ":" + strconv.Itoa(cfg.Port)
	options := []nats.Option{
		natsutil.CredsFileAuth("/tmp/definitely-missing.creds"),
		natsutil.CAPemAuth(cfg.CAPemPath),
		nats.Timeout(3 * time.Second),
		nats.MaxReconnects(0),
	}

	nc, err := nats.Connect(url, options...)
	if err == nil {
		_ = nc.Drain()
		nc.Close()
		t.Fatal("expected connection error with wrong creds path, got nil")
	}
}
