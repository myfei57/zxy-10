package router

import (
	"edgelog/internal/buffer"
	"edgelog/internal/config"
)

// Router dispatches committed batches to sinks according to the current
// routing rules and flow-control state.
type Router struct {
	rules   *config.RulesService
	flow    *FlowController
	buffers map[string]*buffer.RingBuffer
}

// NewRouter creates the router.
func NewRouter(rules *config.RulesService, flow *FlowController) *Router {
	return &Router{rules: rules, flow: flow, buffers: make(map[string]*buffer.RingBuffer)}
}

// RegisterBuffer attaches the ring buffer of a source to the router.
func (r *Router) RegisterBuffer(sourceID string, buf *buffer.RingBuffer) {
	r.buffers[sourceID] = buf
}

// Buffer returns the ring buffer of a source.
func (r *Router) Buffer(sourceID string) *buffer.RingBuffer {
	return r.buffers[sourceID]
}
