package correlation

import (
	"context"
	"github.com/Circutor/edgex/pkg/clients"
)

func FromContext(ctx context.Context) string {
	hdr, ok := ctx.Value(clients.CorrelationHeader).(string)
	if !ok {
		hdr = ""
	}
	return hdr
}
