package router

import (
	"edgelog/internal/buffer"
)

// FlowController decides whether a source must be throttled based on
// backpressure and quota state.
type FlowController struct {
	throttled map[string]bool
}

// NewFlowController creates the flow controller.
func NewFlowController() *FlowController {
	return &FlowController{throttled: make(map[string]bool)}
}

// ShouldThrottle reports whether a source is currently throttled or its
// buffer has dropped more lines than the backpressure threshold.
func (c *FlowController) ShouldThrottle(sourceID string, buf *buffer.RingBuffer) bool {
	if c.throttled[sourceID] {
		return true
	}
	if buf != nil && buf.Dropped() > 0 {
		return true
	}
	return false
}

// SetThrottled marks a source as throttled.
func (c *FlowController) SetThrottled(sourceID string, value bool) {
	c.throttled[sourceID] = value
}

// ResetLimiter clears the throttled flag of a source.
func (c *FlowController) ResetLimiter(sourceID string) {
	delete(c.throttled, sourceID)
}

// IsThrottled reports whether the limiter flag of a source is set.
func (c *FlowController) IsThrottled(sourceID string) bool {
	return c.throttled[sourceID]
}
