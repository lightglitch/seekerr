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

package logger

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	"github.com/utahta/go-cronowriter"
)

var (
	logger zerolog.Logger
)

func InitLogger() {

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	zerolog.TimeFieldFormat = viper.GetString("logger.time_format")

	level, err := zerolog.ParseLevel(viper.GetString("logger.level"))
	if err != nil {
		fmt.Println(err)
	} else {
		zerolog.SetGlobalLevel(level)
	}

	consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: viper.GetString("logger.time_format"),  NoColor: !viper.GetBool("logger.color")}
	multi := zerolog.MultiLevelWriter(consoleWriter)

	if viper.GetString("logger.file") != "" {
		filelog := cronowriter.MustNew(viper.GetString("logger.file"))
		multi = zerolog.MultiLevelWriter(consoleWriter, filelog)

		if viper.GetBool("logger.human") {
			fileConsoleWriter := zerolog.ConsoleWriter{Out: filelog, TimeFormat: viper.GetString("logger.time_format"),  NoColor: true}
			multi = zerolog.MultiLevelWriter(consoleWriter, fileConsoleWriter)
		}
	}

	logger = zerolog.New(multi).With().Timestamp().Logger()
}

func GetLogger() *zerolog.Logger {
	return &logger
}
