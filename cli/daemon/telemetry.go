package daemon

import (
	"context"

	"encr.dev/cli/internal/telemetry"
	daemonpb "encr.dev/proto/encore/daemon"
)

func (s *Server) Telemetry(ctx context.Context, req *daemonpb.TelemetryRequest) (*daemonpb.TelemetryResponse, error) {
	if req.Enabled != nil {
		err := telemetry.SetEnabled(*req.Enabled)
		if err != nil {
			return nil, err
		}
	}
	return &daemonpb.TelemetryResponse{
		Enabled: telemetry.IsEnabled(),
		AnonId:  telemetry.GetAnonID(),
	}, nil
}
