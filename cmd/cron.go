/*
 * Copyright © 2021 Mário Franco
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

package cmd

import (
	"fmt"
	"github.com/lightglitch/seekerr/utils/logger"
	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// cronCmd represents the cron command
var cronCmd = &cobra.Command{
	Use:   "cron",
	Short: "Import the movies found in the lists to radarr using the schedule on the config file.",
	Long:  ``,
	PreRun: func(cmd *cobra.Command, args []string) {
		initConfig()
		logger.InitLogger()
	},
	Run: func(cmd *cobra.Command, args []string) {
		spec := viper.GetString("cron")
		schedule, err := cron.ParseStandard(spec)
		if err != nil {
			fmt.Printf("Invalid cron schedule: %s, %s \n", err)
			return
		}
		c := cron.New(cron.WithLogger(zeroLogger {
			logger: logger.GetLogger(),
		}))

		c.Schedule(schedule, cron.FuncJob(func() {
			importCmd.Run(cmd, args)
		}))

		c.Run()
	},
}

func init() {
	rootCmd.AddCommand(cronCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// importCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// importCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	cronCmd.Flags().StringP("schedule", "s", "", "Run with this cron schedule")
	viper.BindPFlag("cron", cronCmd.Flags().Lookup("schedule"))
}

type zeroLogger struct {
	logger *zerolog.Logger
}

func (zl zeroLogger) Info(msg string, keysAndValues ...interface{}) {
	ev := zl.logger.Info()
	for key, value := range keysAndValues {
		ev.Interface(fmt.Sprintf("%v", key), value)
	}
	ev.Msg(msg)
}

func (zl zeroLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	ev := zl.logger.Error().Err(err)
	for key, value := range keysAndValues {
		ev.Interface(fmt.Sprintf("%v", key), value)
	}
	ev.Msg(msg)
}
