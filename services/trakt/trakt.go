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
	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	"strconv"
	"strings"
)

const (
	TRAKT_URL        = "https://api.trakt.tv/"
	TRAKT_PAGE_LIMIT = 100
)

func NewClient(config *viper.Viper, logger *zerolog.Logger, restyClient *resty.Client) *Client {
	if config.GetString("apiKey") == "" {
		logger.Error().Msg("Missing trakt api key configuration.")
		return nil
	}

	/*
		if config.GetString("apiKSecret") == "" {
			logger.Error().Msg("Missing trakt api secret configuration.")
			return nil
		}
	*/

	return &Client{
		logger:      logger.With().Str("Component", "Trakt").Logger(),
		restyClient: restyClient,
		url:         TRAKT_URL,
		apiKey:      config.GetString("apiKey"),
	}
}

type Client struct {
	logger      zerolog.Logger
	restyClient *resty.Client
	url         string
	apiKey      string
	// apiSecret string
}

type MovieItem struct {
	Watchers  int   `json:"watchers"`
	UserCount int   `json:"user_count"`
	Movie     Movie `json:"movie"`
}

// Movie struct for the Trakt v2 API
type Movie struct {
	IDs struct {
		Imdb  string `json:"imdb"`
		Slug  string `json:"slug"`
		Tmdb  int    `json:"tmdb"`
		Trakt int    `json:"trakt"`
	} `json:"ids"`
	Title string `json:"title"`
	Year  int    `json:"year"`
}

func (c *Client) initRequest() *resty.Request {
	return c.restyClient.R().
		SetHeaders(map[string]string{
			"Content-Type":      "application/json",
			"trakt-api-version": "2",
			"trakt-api-key":     c.apiKey,
		})
}

func (c *Client) fetchMovieItemPagedList(url string, queryParams map[string]string) ([]MovieItem, error) {

	c.logger.Debug().Interface("params", queryParams).Msgf("Fetching trakt list: %s", url)

	result := []MovieItem{}

	resp, err := c.
		initRequest().
		SetQueryParams(queryParams).
		SetResult([]MovieItem{}).
		Get(url)

	if resp != nil && resp.IsSuccess() {
		result = *resp.Result().(*[]MovieItem)
	}

	return result, err
}

func (c *Client) fetchMoviePagedList(url string, queryParams map[string]string) ([]Movie, error) {

	c.logger.Debug().Interface("params", queryParams).Msgf("Fetching trakt list: %s", url)

	result := []Movie{}

	resp, err := c.
		initRequest().
		SetQueryParams(queryParams).
		SetResult([]Movie{}).
		Get(url)

	if resp != nil && resp.IsSuccess() {
		result = *resp.Result().(*[]Movie)
	}

	return result, err
}

func (c *Client) FetchList(url string, limit int) ([]MovieItem, error) {
	result := []MovieItem{}

	page := 1
	pageLimit := TRAKT_PAGE_LIMIT
	if pageLimit > limit {
		pageLimit = limit
	}
	currentCount := pageLimit

	for ok := true; ok; ok = len(result) < limit && currentCount == pageLimit {

		params := map[string]string{
			"page":  strconv.Itoa(page),
			"limit": strconv.Itoa(pageLimit),
		}


		movieItems := []MovieItem{}
		if strings.HasSuffix(url, "/movies/popular") || strings.Contains(url, "/movies/recommended/") {
			movies, err := c.fetchMoviePagedList(url, params)

			if err != nil {
				c.logger.Error().Err(err).Interface("params", params).Msg("Fetching paged movies")
			}

			currentCount = len(movies)
			if currentCount > 0 {
				for _, movie := range movies {
					movieItems = append(movieItems, MovieItem{
						Movie:    movie,
					})
				}
			}
		} else {
			var err error = nil
			movieItems, err = c.fetchMovieItemPagedList(url, params)

			if err != nil {
				c.logger.Error().Err(err).Interface("params", params).Msg("Fetching paged movies")
			}
		}
		currentCount = len(movieItems)
		if currentCount > 0 {
			result = append(result, movieItems...)
		}

		page++
	}

	return result, nil
}
