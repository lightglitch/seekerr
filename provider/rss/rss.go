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

package rss

import (
	"errors"
	"github.com/go-resty/resty/v2"
	"github.com/lightglitch/seekerr/provider"
	"github.com/lightglitch/seekerr/services/guessit"
	"github.com/rs/zerolog"
)
import "github.com/mmcdole/gofeed"

func NewProvider(guessit *guessit.Client, logger *zerolog.Logger, restyClient *resty.Client) *Provider {
	return &Provider{
		parser:      gofeed.NewParser(),
		guessit:     guessit,
		restyClient: restyClient,
		logger:      logger.With().Str("Component", "RSS Provider").Logger(),
	}
}

type Provider struct {
	parser      *gofeed.Parser
	logger      zerolog.Logger
	restyClient *resty.Client
	guessit     *guessit.Client
}

func (p *Provider) GetItems(config provider.ListConfig) ([]provider.ListItem, error) {

	limit := config.Filter.Limit
	if limit == 0 {
		limit = 1000
	}
	result := []provider.ListItem{}

	resp, err := p.restyClient.R().Get(config.Url)

	if err != nil {
		p.logger.Error().Err(err).Msg("Fetching feed")
		return []provider.ListItem{}, err
	}

	if resp.IsError() {
		return []provider.ListItem{}, errors.New(resp.Status() + " - " + resp.String())
	}

	feed, err := p.parser.ParseString(resp.String())
	if err != nil {
		p.logger.Error().Err(err).Msg("Parsing feed")
		return result, err
	}

	p.logger.Info().Int("Count", len(feed.Items)).Msg("Found feed items.")

	for index, item := range feed.Items {
		p.logger.Debug().Interface("item", item).Msgf("Processing feed item %s.", item.Title)

		if config.GuessIt && p.guessit != nil {
			guessResult, err := p.guessit.GuessIt(item.Title)
			if err == nil && guessResult != nil && guessResult.Type == "movie" {
				p.logger.Info().Msgf("Guessed feed item %s (%d).", guessResult.Title, guessResult.Year)
				result = append(result, provider.ListItem{
					Title: guessResult.Title,
					Year:  guessResult.Year,
					Imdb:  "",
				})
			}
		} else {

			result = append(result, provider.ListItem{
				Title: item.Title,
				Year:  0,
				Imdb:  "",
			})
		}

		if index >= limit {
			break
		}
	}

	return result, nil
}
