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

package slack

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

type SlackMessage struct {
	Blocks []SlackBlock `json:"blocks"`
}

type SlackBlock struct {
	Type     string    `json:"type"`
	Text     *SlackText `json:"text,omitempty"`
	ImageURL string    `json:"image_url,omitempty"`
	AltText  string    `json:"alt_text,omitempty"`
}

type SlackText struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func NewSlackAgent(config *viper.Viper, logger *zerolog.Logger, restyClient *resty.Client) *SlackAgent {

	url, err := url.Parse(config.GetString("webhook"))
	if err != nil {
		logger.Err(err)
		return nil
	}

	events := []notification.EventType{}
	_ = config.UnmarshalKey("events", &events)

	return &SlackAgent{
		*notification.NewWebhookAgent(url.String(), events, logger.With().Str("Component", "SlackAgent").Logger(), restyClient),
	}
}

type SlackAgent struct {
	notification.WebhookAgent
}

func (g *SlackAgent) Name() string {
	return "slack"
}

func (g *SlackAgent) getMessage(event notification.Event) interface{} {
	g.WebhookAgent.Logger.Debug().Interface("event", event).Msg("Processing message")
	message := SlackMessage{Blocks: []SlackBlock{}}

	switch event.Type {
	case notification.START_FEED:
		message.Blocks = append(message.Blocks, SlackBlock{
			Type: "section",
			Text: &SlackText{
				Type: "mrkdwn",
				Text: fmt.Sprintf("Start processing feed %s", event.Data["name"]),
			},
		})
	case notification.FINISH_FEED:
		message.Blocks = append(message.Blocks, SlackBlock{
			Type: "section",
			Text: &SlackText{
				Type: "mrkdwn",
				Text: fmt.Sprintf("Finish processing feed %s, added %d movies", event.Data["name"], event.Data["added"]),
			},
		})
	case notification.FINISH_ALL_FEEDS:
		message.Blocks = append(message.Blocks, SlackBlock{
			Type: "section",
			Text: &SlackText{
				Type: "mrkdwn",
				Text: fmt.Sprintf("Finish processing all feeds, added %d movies", event.Data["added"]),
			},
		})
	case notification.ADDED_MOVIE:
		movie := event.Data["movie"].(*radarr.Movie)
		item := event.Data["item"].(*provider.ListItem)
		url := fmt.Sprintf("https://www.themoviedb.org/movie/%d-%s", movie.TmdbID, movie.TitleSlug)

		message.Blocks = append(message.Blocks, SlackBlock{
			Type: "section",
			Text: &SlackText{
				Type: "mrkdwn",
				Text: fmt.Sprintf("Added new movie found in feed *%s*.", event.Data["name"]),
			},
		}, SlackBlock{
			Type: "section",
			Text: &SlackText{
				Type: "mrkdwn",
				Text: fmt.Sprintf("<%s|%s (%d)>\n%s\n\n IMDB: *%.1f*/10 | METACRITIC: *%d*/100 | ROTTEN TOMATOES: *%d%%*",
					url, movie.Title, movie.Year, movie.Overview, item.Ratings.Imdb, item.Ratings.Metacritic, item.Ratings.RottenTomatoes),
			},
		})
		if len(movie.Images) > 0 {
			message.Blocks = append(message.Blocks, SlackBlock{
				Type: "image",
				ImageURL: movie.Images[0].URL,
				AltText: movie.Title,
			})
		}

	default:
		g.WebhookAgent.Logger.Error().Interface("event", event).Msg("Invalid event type")
		return nil
	}

	return message
}

func (a *SlackAgent) SendEvent(event notification.Event) {
	if a.IsSubscribe(event.Type) {
		a.SendMessage(event, a.getMessage(event))
	}
}
