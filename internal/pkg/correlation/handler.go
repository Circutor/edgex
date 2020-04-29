package correlation

import (
	"context"

	"gitlab.circutor.com/EDS/edgex-go/pkg/clients"
)

func FromContext(ctx context.Context) string {
	hdr, ok := ctx.Value(clients.CorrelationHeader).(string)
	if !ok {
		hdr = ""
	}
	return hdr
}
