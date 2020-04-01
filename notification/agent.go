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
	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog"
)

type Agent interface {
	SendEvent(event Event)
	Name() string
}

func NewWebhookAgent(url string, events []EventType, logger zerolog.Logger, restyClient *resty.Client) *WebhookAgent {
	return &WebhookAgent{
		Logger:      logger,
		restyClient: restyClient,
		Url:         url,
		events:      events,
	}
}

type WebhookAgent struct {
	Logger      zerolog.Logger
	restyClient *resty.Client
	Url         string
	events      []EventType
}

func (a WebhookAgent) initRequest() *resty.Request {
	return a.restyClient.R().
		SetHeaders(map[string]string{
			"Content-Type": "application/json",
		})
}

func (a WebhookAgent) IsSubscribe(eventType EventType) bool {

	if len(a.events) == 0 {
		return true
	}

	for _, et := range a.events {
		if et == eventType {
			return true
		}
	}
	return false
}

func (a WebhookAgent) SendMessage(event Event, message interface{}) {
	a.Logger.Debug().Interface("event", event).Interface("message", message).Msg("Sending event")
	if message != nil {
		resp, err := a.
			initRequest().
			SetBody(message).
			Post(a.Url)
		if err != nil {
			a.Logger.Error().Err(err).Msgf("Error sending event")
		}
		if resp.IsError() {
			a.Logger.Error().Str("response", string(resp.Body())).Msgf("Error sending event")
		}
	}
}
