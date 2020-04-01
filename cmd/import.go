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

package cmd

import (
	"github.com/lightglitch/seekerr/importer"
	"github.com/lightglitch/seekerr/notification"
	"github.com/lightglitch/seekerr/notification/gotify"
	"github.com/lightglitch/seekerr/provider"
	"github.com/lightglitch/seekerr/provider/imdb"
	"github.com/lightglitch/seekerr/provider/rss"
	traktprovider "github.com/lightglitch/seekerr/provider/trakt"
	"github.com/lightglitch/seekerr/services/guessit"
	"github.com/lightglitch/seekerr/services/omdb"
	"github.com/lightglitch/seekerr/services/radarr"
	"github.com/lightglitch/seekerr/services/trakt"
	"github.com/lightglitch/seekerr/utils/http"
	"github.com/lightglitch/seekerr/utils/logger"
	"github.com/spf13/viper"

	"github.com/spf13/cobra"
)

var listName string

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import the movies found in the lists to radarr.",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {

		if viper.ConfigFileUsed() != "" {
			restyClient := http.GetRestyClient(viper.Sub("services.resty"))

			radarr := radarr.NewClient(viper.Sub("services.radarr"), logger.GetLogger(), restyClient)
			omdb := omdb.NewClient(viper.Sub("services.omdb"), logger.GetLogger(), restyClient)
			gessit := guessit.NewClient(viper.Sub("services.guessIt"), logger.GetLogger(), restyClient)
			trakt := trakt.NewClient(viper.Sub("services.trakt"), logger.GetLogger(), restyClient)

			registry := provider.NewProviderRegistry()

			registry.RegisterProvider(provider.RSS, rss.NewProvider(gessit, logger.GetLogger(), restyClient))
			registry.RegisterProvider(provider.IMDB, imdb.NewProvider(logger.GetLogger(), restyClient))
			registry.RegisterProvider(provider.TRAKT, traktprovider.NewProvider(trakt, logger.GetLogger()))

			dispatcher := notification.NewNotificationDispatcher(logger.GetLogger())

			dispatcher.RegisterAgent(gotify.NewGotifyAgent(viper.Sub("notifications.gotify"), logger.GetLogger(), restyClient))

			importer := importer.NewImporter(viper.Sub("importer"), logger.GetLogger(), radarr, omdb, registry, dispatcher)

			if listName != "" && listName != "all" {
				importer.ProcessList(listName)
			} else {
				importer.ProcessLists()
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(importCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// importCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// importCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	importCmd.Flags().StringVarP(&listName, "list", "l", "all", "The name of the list to import")
}
