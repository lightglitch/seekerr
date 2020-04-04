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

package guessit

import (
	"encoding/json"
	"errors"
	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	"net/url"
	"os/exec"
)

const (
	WEBSERVICE = "webservice"
	COMMAND    = "command"
)

type GuessResult struct {
	AudioChannels string `json:"audioChannels"`
	AudioCodec    string `json:"audioCodec"`
	Container     string `json:"container"`
	EpisodeNumber int64  `json:"episodeNumber"`
	Format        string `json:"format"`
	MimeType      string `json:"mimetype"`
	ReleaseGroup  string `json:"releaseGroup"`
	ScreenSize    string `json:"screenSize"`
	Season        int64  `json:"season"`
	Series        string `json:"series"`
	Title         string `json:"title"`
	Type          string `json:"type"`
	VideoCodec    string `json:"videoCodec"`
	Year          int    `json:"year"`
}

func NewClient(config *viper.Viper, logger *zerolog.Logger, restyClient *resty.Client) *Client {

	serviceType := config.GetString("type")
	if serviceType == "" || (serviceType != WEBSERVICE && serviceType != COMMAND) {
		logger.Error().Msg("Missing guessit type configuration.")
		return nil
	}

	if serviceType == WEBSERVICE && config.GetString("url") == "" {
		logger.Error().Msg("Missing guessit url configuration.")
		return nil
	}

	urlString := config.GetString("url")
	if serviceType == WEBSERVICE {
		// Test the url
		_, err := url.Parse(config.GetString("url"))
		if err != nil {
			logger.Err(err)
			return nil
		}
	}

	path := config.GetString("path")
	if serviceType == COMMAND {
		cmd := exec.Command(path)
		_, err := cmd.Output()
		if err != nil {
			logger.Log().Err(err).Msg("Testing GuessIt command")
			return nil
		}
	}

	return &Client{
		logger:      logger.With().Str("Component", "GuessIt").Logger(),
		restyClient: restyClient,
		serviceType: serviceType,
		url:         urlString,
		path:        path,
	}
}

type Client struct {
	logger      zerolog.Logger
	restyClient *resty.Client
	serviceType string
	path        string
	url         string
}

func (c *Client) initRequest() *resty.Request {
	return c.restyClient.R().
		SetHeaders(map[string]string{
			"Content-Type": "application/json",
		})
}

func (c *Client) guessItUsingWebService(title string) (*GuessResult, error) {

	resp, err := c.initRequest().
		SetQueryParams(map[string]string{
			"filename": title,
		}).
		SetResult(&GuessResult{}).
		Get(c.url)

	if resp != nil && resp.IsSuccess() {
		result := resp.Result().(*GuessResult)
		return result, nil
	}

	return nil, err
}

func (c *Client) guessItUsingCommand(title string) (*GuessResult, error) {

	cmd := exec.Command(c.path, title, "--json")
	stdout, err := cmd.Output()
	result := GuessResult{}
	if err == nil {
		if err = json.Unmarshal(stdout, &result); err != nil {
			return nil, err
		}
		return &result, nil
	}

	return nil, err
}

func (c *Client) GuessIt(title string) (*GuessResult, error) {

	if c.serviceType == WEBSERVICE {
		return c.guessItUsingWebService(title)
	}
	if c.serviceType == COMMAND {
		return c.guessItUsingCommand(title)
	}
	return nil, errors.New("Service not found")
}
