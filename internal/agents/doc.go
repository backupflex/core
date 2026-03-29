// Package agents defines the base coordinator module for backup agents.
//
// The module introduced in this iteration contains:
//   - persistent agent profiles in BadgerDB;
//   - volatile runtime statuses in memory (heartbeat-driven);
//   - registration workflow;
//   - heartbeat processing;
//   - ability to mark agents offline.
//
// It is transport-agnostic and can be used by HTTP, gRPC, or message-based APIs.
package agents
