// Package goclient provides a Go client for GoChat bots.
//
// A Session owns both the bot REST client and the bot gateway connection. REST
// calls use Authorization: Bot <token>, and Open connects to /bot/ws, identifies
// the shard, maintains heartbeats, and dispatches typed events through
// AddHandler.
package goclient
