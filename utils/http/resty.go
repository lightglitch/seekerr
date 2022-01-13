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

package http

import (
	"github.com/go-resty/resty/v2"
	"github.com/lightglitch/seekerr/utils/logger"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	"net/http"
	"time"
)

type restyLogger struct {
	logger *zerolog.Logger
}

func (l *restyLogger) Errorf(format string, v ...interface{}) {
	l.logger.Error().Msgf(format, v)
}

func (l *restyLogger) Warnf(format string, v ...interface{}) {
	l.logger.Warn().Msgf(format, v)
}

func (l *restyLogger) Debugf(format string, v ...interface{}) {
	l.logger.Debug().Msgf(format, v)
}

func GetRestyClient(config *viper.Viper) *resty.Client {
	c := resty.New()

	c.SetDebug(false)
	c.SetTimeout(20 * time.Second)
	c.SetRetryCount(0)
	c.SetRetryWaitTime(1 * time.Second)
	c.SetRetryMaxWaitTime(10 * time.Second)

	if config != nil {
		c.SetDebug(config.GetBool("debug"))
		c.SetTimeout(config.GetDuration("timeout"))
		c.SetRetryCount(config.GetInt("retry"))
		c.SetRetryWaitTime(config.GetDuration("retryWaitTime"))
		c.SetRetryMaxWaitTime(config.GetDuration("retryMaxWaitTime"))
	}

	c.SetLogger(&restyLogger{logger: logger.GetLogger()})

	c.AddRetryCondition(func(r *resty.Response, err error) bool {
		return r.StatusCode() == http.StatusTooManyRequests
	})

	//TODO support rate limit in the future

	return c
}
