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

package trakt

import (
	"github.com/lightglitch/seekerr/provider"
	"github.com/lightglitch/seekerr/services/trakt"
	"github.com/rs/zerolog"
	"regexp"
	"strings"
)

const (
	TRAKT_URL_PROTOCOL = "trakt://"
	TRAKT_URL_PREFIX   = "https://trakt.tv/users/"
)

func NewProvider(trakt *trakt.Client, logger *zerolog.Logger) *Provider {
	return &Provider{
		trakt:  trakt,
		logger: logger.With().Str("Component", "Trakt Provider").Logger(),
	}
}

type Provider struct {
	logger zerolog.Logger
	trakt  *trakt.Client
}

func (p Provider) GetItems(config provider.ListConfig) ([]provider.ListItem, error) {

	limit := config.Filter.Limit
	if limit == 0 {
		limit = 1000
	}
	result := []provider.ListItem{}

	url := config.Url
	if strings.HasPrefix(url, TRAKT_URL_PROTOCOL) {
		url = trakt.TRAKT_URL + strings.TrimPrefix(url, TRAKT_URL_PROTOCOL)
	}
	if strings.HasPrefix(url, TRAKT_URL_PREFIX) {
		user := strings.TrimPrefix(regexp.MustCompile(`/users/([^/]*)`).FindString(url), "/users/")
		list := strings.TrimPrefix(regexp.MustCompile(`/lists/([^/?]*)`).FindString(url), "/lists/")

		url = trakt.TRAKT_URL + "users/" + user + "/lists/" + list + "/items/movies"
		p.logger.Debug().Msgf("Finding list url %s", url)
	}

	movies, _ := p.trakt.FetchList(url, limit)

	for _, item := range movies {
		result = append(result, provider.ListItem{
			Title: item.Movie.Title,
			Year:  item.Movie.Year,
			Imdb:  item.Movie.IDs.Imdb,
			Tmdb:  item.Movie.IDs.Tmdb,
		})
	}

	return result, nil
}
