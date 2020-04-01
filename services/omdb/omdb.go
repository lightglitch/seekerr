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

package omdb

import (
	"errors"
	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

const (
	OMDB_URL = "http://www.omdbapi.com/"
	OMDB_IMDB_SOURCE = "Internet Movie Database"
	OMDB_METACRITIC_SOURCE = "Metacritic"
	OMDB_ROTTEN_TOMATOES_SOURCE = "Rotten Tomatoes"
)

type MovieResult struct {
	Title      string
	Year       string
	Rated      string
	Released   string
	Runtime    string
	Genre      string
	Director   string
	Writer     string
	Actors     string
	Plot       string
	Country    string
	Awards     string
	Poster     string
	Ratings    []Rating
	Metascore  string
	ImdbRating string
	ImdbVotes  string
	ImdbID     string
	Type       string
	BoxOffice  string
	Production string
	Website    string
	Response   string
	Error      string
}

type Rating struct {
	Source string
	Value  string
}

type SearchResponse struct {
	Search       []SearchResult
	Response     string
	Error        string
	TotalResults string
}

type SearchResult struct {
	Title  string
	Year   string
	ImdbID string
	Type   string
	Poster string
}

func NewClient(config *viper.Viper, logger *zerolog.Logger, restyClient *resty.Client) *Client {

	if config.GetString("apiKey") == "" {
		logger.Error().Msg("Missing omdb api key configuration.")
		return nil
	}

	c := &Client{
		logger:      logger.With().Str("Component", "OMDB").Logger(),
		url:         OMDB_URL,
		restyClient: restyClient,
		apiKey:      config.GetString("apiKey"),
	}

	return c
}

type Client struct {
	logger      zerolog.Logger
	restyClient *resty.Client
	url         string
	apiKey      string
}

func (c Client) initRequest() *resty.Request {
	return c.restyClient.R().
		SetHeaders(map[string]string{
			"Content-Type": "application/json",
		})
}

func (c Client) GetMovieById(imdbId string, params map[string]string) (*MovieResult, error) {

	queryParams := map[string]string{
		"i":      imdbId,
		"type":   "movie",
		"apikey": c.apiKey,
	}

	for k, v := range params {
		queryParams[k] = v
	}

	resp, err := c.initRequest().
		SetQueryParams(queryParams).
		SetResult(&MovieResult{}).
		Get(c.url)

	if resp != nil && resp.IsSuccess() {
		title := resp.Result().(*MovieResult)
		if title.Response == "False" {
			c.logger.Error().Interface("result", title).Msg("Fetching movie info")
			return nil, errors.New(title.Error)
		}
		return title, nil
	}

	c.logger.Error().Err(err).Msg("Fetching movie info")
	return nil, err
}

func (c Client) GetMovieByTitle(title string, params map[string]string) (*MovieResult, error) {

	queryParams := map[string]string{
		"t":      title,
		"type":   "movie",
		"apikey": c.apiKey,
	}

	for k, v := range params {
		queryParams[k] = v
	}

	resp, err := c.initRequest().
		SetQueryParams(queryParams).
		SetResult(&MovieResult{}).
		Get(c.url)

	if resp != nil && resp.IsSuccess() {
		title := resp.Result().(*MovieResult)
		if title.Response == "False" {
			c.logger.Error().Err(errors.New(title.Error)).Msg("Fetching movie info")
			return nil, errors.New(title.Error)
		}
		return title, nil
	}

	c.logger.Error().Err(err).Msg("Fetching movie info")
	return nil, err
}

func (c Client) SearchMovieByTitle(title string, params map[string]string) (*SearchResponse, error) {

	queryParams := map[string]string{
		"s":      title,
		"type":   "movie",
		"apikey": c.apiKey,
	}

	for k, v := range params {
		queryParams[k] = v
	}

	resp, err := c.initRequest().
		SetQueryParams(queryParams).
		SetResult(SearchResponse{}).
		Get(c.url)

	if resp != nil && resp.IsSuccess() {
		search := resp.Result().(*SearchResponse)
		return search, nil
	}

	return nil, err
}
