package runner

import (
	"context"
	"sync"

	"google.golang.org/grpc/stats"
	"google.golang.org/grpc/status"
)

// StatsHandler is for gRPC stats
type statsHandler struct {
	results chan *callResult

	id     int
	hasLog bool
	log    Logger

	lock   sync.RWMutex
	ignore bool
}

// HandleConn handle the connection
func (c *statsHandler) HandleConn(ctx context.Context, cs stats.ConnStats) {
	// no-op
}

// TagConn exists to satisfy gRPC stats.Handler.
func (c *statsHandler) TagConn(ctx context.Context, cti *stats.ConnTagInfo) context.Context {
	// no-op
	return ctx
}

// HandleRPC implements per-RPC tracing and stats instrumentation.
func (c *statsHandler) HandleRPC(ctx context.Context, rs stats.RPCStats) {
	switch rs := rs.(type) {
	case *stats.End:
		ign := false
		c.lock.RLock()
		ign = c.ignore
		c.lock.RUnlock()

		if !ign {
			duration := rs.EndTime.Sub(rs.BeginTime)

			var st string
			s, ok := status.FromError(rs.Error)
			if ok {
				st = s.Code().String()
			}

			// Extract raw payloads from global map if this request was sampled
			// We store the raw proto.Message objects and marshal them AFTER the benchmark
			var reqPayload interface{}
			var resPayload interface{}

			// Check if this request was sampled (request number in context)
			if reqNumVal := ctx.Value(contextKey("requestNumber")); reqNumVal != nil {
				if reqNum, ok := reqNumVal.(uint64); ok {
					// Retrieve payloads from global map
					sampledPayloadsMutex.RLock()
					if pair, ok := sampledPayloads[reqNum]; ok {
						reqPayload = pair.request
						resPayload = pair.response
					}
					sampledPayloadsMutex.RUnlock()

					if c.hasLog {
						c.log.Debugw("Stats handler: Retrieved payloads", "statsID", c.id, "reqNum", reqNum,
							"hasReq", reqPayload != nil, "hasRes", resPayload != nil)
					}
				}
			}

			c.results <- &callResult{
				err:             rs.Error,
				status:          st,
				duration:        duration,
				timestamp:       rs.EndTime,
				requestPayload:  reqPayload,
				responsePayload: resPayload,
			}

			if c.hasLog {
				c.log.Debugw("Received RPC Stats",
					"statsID", c.id, "code", st, "error", rs.Error,
					"duration", duration, "stats", rs)
			}
		}
	}
}

func (c *statsHandler) Ignore(val bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.ignore = val
}

// TagRPC implements per-RPC context management.
func (c *statsHandler) TagRPC(ctx context.Context, info *stats.RPCTagInfo) context.Context {
	return ctx
}
