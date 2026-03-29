// Package agents defines HTTP layer for agent coordination.
//
// It intentionally splits external APIs into two independent surfaces:
//   - management API for operators/control-plane integrations;
//   - agent API for backup/verification agents (bearer-token based identity).
package agents
