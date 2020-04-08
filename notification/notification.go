/*
 * Copyright © 2020 Mário Franco
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
 * THE SOFTWARE.
 */

package notification

import (
	"github.com/lightglitch/seekerr/provider"
	"github.com/lightglitch/seekerr/services/radarr"
	"github.com/rs/zerolog"
)

type EventType string

const (
	START_FEED       = "START_FEED"
	FINISH_FEED      = "FINISH_FEED"
	FINISH_ALL_FEEDS = "FINISH_ALL_FEEDS"
	ADDED_MOVIE      = "ADDED_MOVIE"
	REVISION_MOVIE   = "REVISION_MOVIE"
)

type Event struct {
	Type EventType
	Data map[string]interface{}
}

func NewNotificationDispatcher(logger *zerolog.Logger) *Dispatcher {
	return &Dispatcher{
		logger: logger.With().Str("Component", "Notification Dispatcher").Logger(),
		agents: map[string]Agent{},
	}
}

type Dispatcher struct {
	logger zerolog.Logger
	agents map[string]Agent
}

func (d *Dispatcher) RegisterAgent(agent Agent) {
	if agent != nil {
		d.logger.Info().Str("agent", agent.Name()).Msg("Register agent")
		d.agents[agent.Name()] = agent
	}
}

func (d *Dispatcher) SendEvent(event Event) {
	d.logger.Info().Interface("event", event).Int("agents", len(d.agents)).Msg("Broadcast event")
	for _, agent := range d.agents {
		agent.SendEvent(event)
	}
}

func (d *Dispatcher) SendEventAddMovie(name string, item *provider.ListItem, movie *radarr.Movie) {
	d.SendEvent(Event{
		Type: ADDED_MOVIE,
		Data: map[string]interface{}{
			"name":  name,
			"item":  item,
			"movie": movie,
		},
	})
}

func (d *Dispatcher) SendEventRevisionMovie(name string, item *provider.ListItem, movie *radarr.Movie) {
	d.SendEvent(Event{
		Type: REVISION_MOVIE,
		Data: map[string]interface{}{
			"name":  name,
			"item":  item,
			"movie": movie,
		},
	})
}

func (d *Dispatcher) SendEventEndFeed(name string, approved int, added int) {
	d.SendEvent(Event{
		Type: FINISH_FEED,
		Data: map[string]interface{}{
			"name":     name,
			"approved": approved,
			"added":    added,
		},
	})
}

func (d *Dispatcher) SendEventEndAllFeeds(approved int, added int) {
	d.SendEvent(Event{
		Type: FINISH_ALL_FEEDS,
		Data: map[string]interface{}{
			"approved": approved,
			"added":    added,
		},
	})
}

func (d *Dispatcher) SendEventStartFeed(name string) {
	d.SendEvent(Event{
		Type: START_FEED,
		Data: map[string]interface{}{
			"name": name,
		},
	})
}
