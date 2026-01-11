package runner

import (
	"context"
	"sync"
	"time"

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

			// Extract payloads from global store if present
			var reqPayload []byte
			var resPayload []byte
			if reqIDVal := ctx.Value(contextKey("requestID")); reqIDVal != nil {
				if reqID, ok := reqIDVal.(uint64); ok {
					if data, ok := payloadStore.Load(reqID); ok {
						if pd, ok := data.(*payloadData); ok {
							// Wait for response to be ready (with timeout)
							select {
							case <-pd.ready:
								// Response is ready, read it safely
								pd.mu.RLock()
								reqPayload = pd.request
								resPayload = pd.response
								pd.mu.RUnlock()
							case <-time.After(5 * time.Second):
								// Timeout - just read what we have
								pd.mu.RLock()
								reqPayload = pd.request
								resPayload = pd.response
								pd.mu.RUnlock()
							}
						}
						// Clean up the entry
						payloadStore.Delete(reqID)
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
