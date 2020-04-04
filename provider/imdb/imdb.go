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

package imdb

import (
	"errors"
	"github.com/PuerkitoBio/goquery"
	"github.com/go-resty/resty/v2"
	"github.com/lightglitch/seekerr/provider"
	"github.com/rs/zerolog"
	"regexp"
	"strconv"
)

func NewProvider(logger *zerolog.Logger, restyClient *resty.Client) *Provider {
	return &Provider{
		restyClient: restyClient,
		logger:      logger.With().Str("Component", "IMDB Provider").Logger(),
	}
}

type Provider struct {
	logger      zerolog.Logger
	restyClient *resty.Client
}

func (p *Provider) GetItems(config provider.ListConfig) ([]provider.ListItem, error) {

	limit := config.Filter.Limit
	if limit == 0 {
		limit = 1000
	}
	result := []provider.ListItem{}

	resp, err := p.restyClient.R().SetDoNotParseResponse(true).Get(config.Url)

	if err != nil {
		p.logger.Error().Err(err).Msg("Fetching html")
		return []provider.ListItem{}, err
	}

	if resp.IsError() {
		return []provider.ListItem{}, errors.New(resp.Status() + " - " + resp.String())
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(resp.RawBody())
	if err != nil {
		p.logger.Error().Err(err).Msg("Parsing html")
		return result, err
	}

	iddRegex := regexp.MustCompile(`tt\d+`)
	yearRegex := regexp.MustCompile(`\d+`)

	doc.Find("div.lister-list .lister-item").Each(func(index int, s *goquery.Selection) {
		// For each item found, get the band and title
		item := s.Find(".lister-item-header a")
		url := item.AttrOr("href", "")
		imdbId := ""
		if url != "" {
			imdbId = iddRegex.FindString(url)
		}
		yearText := yearRegex.FindString(s.Find(".lister-item-header .lister-item-year").Text())
		year, _ := strconv.Atoi(yearText)
		p.logger.Debug().Interface("ImdbId", imdbId).Interface("Year", year).Msgf("Processing imdb list item %s.", item.Text())

		if index < limit {
			result = append(result, provider.ListItem{
				Title: item.Text(),
				Year:  year,
				Imdb:  imdbId,
			})
		}

	})
	return result, nil
}
