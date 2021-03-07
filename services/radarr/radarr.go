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

package radarr

import (
	"errors"
	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	"net/url"
	"strings"
)

func NewClient(config *viper.Viper, logger *zerolog.Logger, restyClient *resty.Client) *Client {

	if config.GetString("url") == "" {
		logger.Error().Msg("Missing radarr url configuration.")
		return nil
	}

	if config.GetString("apiKey") == "" {
		logger.Error().Msg("Missing radarr api key configuration.")
		return nil
	}

	if config.GetString("rootFolder") == "" {
		logger.Error().Msg("Missing radarr root folder configuration.")
		return nil
	}

	if config.GetString("quality") == "" {
		logger.Error().Msg("Missing radarr quality configuration.")
		return nil
	}

	if config.GetString("minimumAvailability") == "" {
		logger.Error().Msg("Missing radarr minimum_availability configuration.")
		return nil
	}

	url, err := url.Parse(config.GetString("url"))
	if err != nil {
		logger.Err(err)
		return nil
	}

	c := &Client{
		logger:              logger.With().Str("Component", "Radarr").Logger(),
		restyClient:         restyClient,
		url:                 url.String(),
		apiKey:              config.GetString("apiKey"),
		rootFolder:          config.GetString("rootFolder"),
		quality:             config.GetString("quality"),
		minimumAvailability: config.GetString("minimumAvailability"),
		monitored:           config.GetBool("monitored"),
		searchForMovie:      config.GetBool("searchForMovie"),
	}

	if ok, err := c.validateApiKey(); !ok {
		if err != nil {
			logger.Error().Err(err).Msg("Invalid API key.")
			return nil
		}
	}

	qualityId, err := c.getProfileId()
	if err != nil {
		logger.Error().Err(err).Msg("Invalid Profile.")
		return nil
	}

	c.qualityId = qualityId

	return c
}

type Client struct {
	logger              zerolog.Logger
	restyClient         *resty.Client
	url                 string
	apiKey              string
	rootFolder          string
	quality             string
	qualityId           int
	minimumAvailability string
	monitored           bool
	searchForMovie      bool
}

// Movie ...
type Movie struct {
	Title               string `json:"title"`
	TitleSlug           string `json:"titleSlug"`
	Overview            string `json:"overview"`
	QualityProfileID    int    `json:"qualityProfileId"`
	RootFolderPath      string `json:"rootFolderPath"`
	Year                int    `json:"year"`
	Monitored           bool   `json:"monitored"`
	ImdbId              string `json:"imdbId"`
	TmdbID              int    `json:"tmdbId"`
	MinimumAvailability string `json:"minimumAvailability"`
	Images              []struct {
		CoverType string `json:"coverType"`
		URL       string `json:"url"`
	} `json:"images"`
	IsAvailable bool                   `json:"isAvailable"`
	AddOptions  map[string]interface{} `json:"addOptions"`
}

type ExcludedMovie struct {
	ID         int    `json:"id"`
	MovieTitle string `json:"movieTitle"`
	MovieYear  int    `json:"movieYear"`
	TmdbID     int    `json:"tmdbId"`
}

type Errors []struct {
	PropertyName                      string        `json:"propertyName"`
	ErrorMessage                      string        `json:"errorMessage"`
	AttemptedValue                    int           `json:"attemptedValue"`
	FormattedMessageArguments         []interface{} `json:"formattedMessageArguments"`
	FormattedMessagePlaceholderValues struct {
		PropertyName  string `json:"propertyName"`
		PropertyValue int    `json:"propertyValue"`
	} `json:"formattedMessagePlaceholderValues"`
}

func (c *Client) getEndpointUrl(endpoint string) string {
	return c.url + endpoint
}

func (c *Client) initRequest() *resty.Request {
	return c.restyClient.R().
		SetHeaders(map[string]string{
			"Content-Type": "application/json",
			"X-Api-Key":    c.apiKey,
		})
}

func (c *Client) validateApiKey() (bool, error) {
	resp, err := c.
		initRequest().
		SetResult(map[string]interface{}{}).
		Get(c.getEndpointUrl("api/system/status"))
	if err != nil {
		return false, err
	}

	if resp.IsSuccess() {
		status := *resp.Result().(*map[string]interface{})
		version, ok := status["version"]

		if !ok {
			return ok, errors.New("Can't access radarr system status. Check you API key.")
		}

		c.logger.Info().Msgf("Accessing radarr with version %s.", version)
		return ok, nil
	}

	return false, nil
}

func (c *Client) getProfileId() (int, error) {
	resp, err := c.
		initRequest().
		SetResult([]map[string]interface{}{}).
		Get(c.getEndpointUrl("api/profile"))
	if err != nil {
		return 0, err
	}

	if resp.IsSuccess() {
		profiles := *resp.Result().(*[]map[string]interface{})
		for _, profile := range profiles {
			if strings.ToLower(profile["name"].(string)) == strings.ToLower(c.quality) {
				c.logger.Debug().Msgf("Found Quality Profile ID for '%s': %d", c.quality, int(profile["id"].(float64)))
				return int(profile["id"].(float64)), nil
			}
		}
	}

	return 0, nil
}

func (c *Client) AddMovie(movie *Movie) error {

	movie.QualityProfileID = c.qualityId
	movie.Monitored = c.monitored
	movie.RootFolderPath = c.rootFolder
	movie.MinimumAvailability = c.minimumAvailability

	if c.searchForMovie {
		movie.AddOptions = map[string]interface{}{
			"searchForMovie": c.searchForMovie,
		}
	}

	resp, err := c.
		initRequest().
		SetBody(movie).
		SetError(Errors{}).
		Post(c.getEndpointUrl("api/movie"))

	if resp.StatusCode() == 400 {
		message := ""
		errorsList := resp.Error().(*Errors)
		for _, e := range *errorsList {
			message += e.ErrorMessage + " "
		}
		return errors.New(message)
	}

	return err
}

func (c *Client) LookupMovieByImdb(imdbId string) (*Movie, error) {

	resp, err := c.
		initRequest().
		SetQueryParams(map[string]string{
			"imdbId": imdbId,
		}).
		SetResult(Movie{}).
		Get(c.getEndpointUrl("api/movie/lookup/imdb"))

	if resp != nil && resp.IsSuccess() {
		c.logger.Debug().RawJSON("response", []byte(resp.String())).Msg("Lookup search")
		result := resp.Result().(*Movie)
		return result, nil
	}

	return nil, err
}

func (c *Client) LookupMovieByTmdb(tmdbId string) (*Movie, error) {

	resp, err := c.
		initRequest().
		SetQueryParams(map[string]string{
			"tmdbId": tmdbId,
		}).
		SetResult(Movie{}).
		Get(c.getEndpointUrl("api/movie/lookup/tmdb"))

	if resp != nil && resp.IsSuccess() {
		c.logger.Debug().RawJSON("response", []byte(resp.String())).Msg("Lookup search")
		result := resp.Result().(*Movie)
		return result, nil
	}

	return nil, err
}

func (c *Client) GetMovies() (*[]Movie, error) {

	resp, err := c.
		initRequest().
		SetResult([]Movie{}).
		Get(c.getEndpointUrl("api/movie"))

	if resp != nil && resp.IsSuccess() {
		result := resp.Result().(*[]Movie)
		return result, nil
	}

	return nil, err
}

func (c *Client) GetExcludedMovies() (*[]ExcludedMovie, error) {

	resp, err := c.
		initRequest().
		SetResult([]ExcludedMovie{}).
		Get(c.getEndpointUrl("api/exclusions"))

	if resp != nil && resp.IsSuccess() {
		result := resp.Result().(*[]ExcludedMovie)
		return result, nil
	}

	return nil, err
}
