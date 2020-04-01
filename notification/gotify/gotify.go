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

package gotify

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/lightglitch/seekerr/notification"
	"github.com/lightglitch/seekerr/provider"
	"github.com/lightglitch/seekerr/services/radarr"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	"net/url"
)

func NewGotifyAgent(config *viper.Viper, logger *zerolog.Logger, restyClient *resty.Client) *GotifyAgent {

	url, err := url.Parse(config.GetString("webhook"))
	if err != nil {
		logger.Err(err)
		return nil
	}

	events := []notification.EventType{}
	_ = config.UnmarshalKey("events", &events)

	return &GotifyAgent{
		*notification.NewWebhookAgent(url.String(), events, logger.With().Str("Component", "GotifyAgent").Logger(), restyClient),
	}
}

type GotifyAgent struct {
	notification.WebhookAgent
}

func (g GotifyAgent) Name() string {
	return "gotify"
}

func (g GotifyAgent) getMessage(event notification.Event) interface{} {
	g.WebhookAgent.Logger.Info().Interface("event", event).Msg("Processing message")
	message := map[string]string{}

	switch event.Type {
	case notification.START_FEED:
		message["title"] = fmt.Sprintf("Seekerr: %s", event.Data["name"])
		message["message"] = fmt.Sprintf("Start processing feed %s", event.Data["name"])
	case notification.FINISH_FEED:
		message["title"] = fmt.Sprintf("Seekerr: %s", event.Data["name"])
		message["message"] = fmt.Sprintf("Finish processing feed %s, added %d movies", event.Data["name"], event.Data["added"])
	case notification.FINISH_ALL_FEEDS:
		message["title"] = "Seekerr"
		message["message"] = fmt.Sprintf("Finish processing all feeds, added %d movies", event.Data["added"])
	case notification.ADDED_MOVIE:
		message["title"] = fmt.Sprintf("Seekerr: %s", event.Data["name"])
		movie := event.Data["movie"].(*radarr.Movie)
		item := event.Data["item"].(*provider.ListItem)
		message["message"] = fmt.Sprintf("Added new movie '%s (%d)', ratings: imdb %.1f, metacritic %d/100, rotten tomatoes %d%%",
			movie.Title, movie.Year, item.Ratings.Imdb, item.Ratings.Metacritic, item.Ratings.RottenTomatoes)
	default:
		g.WebhookAgent.Logger.Error().Interface("event", event).Msg("Invalid event type")
		return nil
	}

	return message
}

func (a GotifyAgent) SendEvent(event notification.Event) {
	if a.IsSubscribe(event.Type) {
		a.SendMessage(event, a.getMessage(event))
	}
}
