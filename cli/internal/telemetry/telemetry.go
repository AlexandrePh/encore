package telemetry

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"github.com/hasura/go-graphql-client"
	"github.com/rs/zerolog/log"

	"encore.dev/types/uuid"
	"encr.dev/internal/conf"
	"encr.dev/pkg/fns"
	"encr.dev/pkg/xos"
)

var singleton = func() *telemetry {
	t := &telemetry{
		client: graphql.NewClient(conf.APIBaseURL+"/graphql", conf.DefaultClient),
	}
	path, err := configPath()
	if err != nil {
		return t
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			// If the file does not exist, telemetry is enabled by default
			t.cfg.Enabled = os.Getenv("DISABLE_ENCORE_TELEMETRY") != "1"
			t.cfg.AnonID = uuid.Must(uuid.NewV4()).String()
			t.cfg.SentEvents = make(map[string]struct{})
			_ = t.saveConfig()
			err = nil
		}
		return t
	}
	err = json.Unmarshal(data, &t.cfg)
	return t
}()

type telemetry struct {
	mu     sync.Mutex
	cfg    telemetryCfg
	client *graphql.Client
}

type TelemetryMessage struct {
	Event       string         `json:"event"`
	AnonymousId string         `json:"anonymousId"`
	Properties  map[string]any `json:"properties"`
}

func (t *telemetry) sendOnce(event string, props ...map[string]any) {
	t.mu.Lock()
	if _, ok := t.cfg.SentEvents[event]; ok {
		t.mu.Unlock()
		return
	}
	t.cfg.SentEvents[event] = struct{}{}
	if err := t.saveConfig(); err != nil {
		log.Debug().Err(err).Msg("failed to save telemetry config")
	}
	t.mu.Unlock()
	if err := t.send(event, props...); err != nil {
		log.Debug().Err(err).Msg("failed to send telemetry message")
		t.mu.Lock()
		delete(t.cfg.SentEvents, event)
		t.mu.Unlock()
	}
}

func (t *telemetry) send(event string, props ...map[string]any) error {
	var m struct {
		Result bool `graphql:"telemetry(msg: $msg)"`
	}
	err := t.client.Mutate(context.Background(), &m, map[string]any{
		"msg": TelemetryMessage{
			Event:       event,
			AnonymousId: t.cfg.AnonID,
			Properties:  fns.MergeMaps(props...),
		}})
	if !m.Result {
		return errors.New("failed to send telemetry message")
	}
	return err
}

func (t *telemetry) trySend(event string, props ...map[string]any) {
	if err := t.send(event, props...); err != nil {
		log.Debug().Msg("failed to send telemetry message")
	}
}

func (t *telemetry) saveConfig() error {
	// Write the telemetry configuration to a file
	path, err := configPath()
	if err != nil {
		return err
	}
	data, err := json.Marshal(t.cfg)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return xos.WriteFile(path, data, 0644)
}

type telemetryCfg struct {
	Enabled      bool                `json:"enabled"`
	AnonID       string              `json:"anon_id"`
	SentEvents   map[string]struct{} `json:"sent_events"`
	ShownWarning bool                `json:"shown_warning"`
}

func IsEnabled() bool {
	if singleton.cfg.Enabled && os.Getenv("DISABLE_ENCORE_TELEMETRY") == "1" {
		return false
	}
	return singleton.cfg.Enabled
}

func SetEnabled(enabled bool) error {
	singleton.cfg.Enabled = enabled
	return singleton.saveConfig()
}

func ShouldShownWarning() bool {
	return singleton.cfg.ShownWarning && IsEnabled()
}

func SetShownWarning() {
	singleton.cfg.ShownWarning = true
	if err := singleton.saveConfig(); err != nil {
		log.Debug().Err(err).Msg("failed to save telemetry config")
	}
}

func SendOnce(event string, props ...map[string]any) {
	if !IsEnabled() {
		return
	}
	go singleton.sendOnce(event, props...)
}

func Send(event string, props ...map[string]any) {
	if !IsEnabled() {
		return
	}
	go singleton.trySend(event, props...)
}

func SendSync(event string, props ...map[string]any) {
	if !IsEnabled() {
		return
	}
	singleton.trySend(event, props...)
}

func configPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "encore", "telemetry.json"), nil
}

func GetAnonID() string {
	return singleton.cfg.AnonID
}
