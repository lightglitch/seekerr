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

package importer

import (
	"fmt"
	"github.com/gosimple/slug"
	"github.com/lightglitch/seekerr/importer/validator"
	"github.com/lightglitch/seekerr/notification"
	"github.com/lightglitch/seekerr/provider"
	"github.com/lightglitch/seekerr/services/omdb"
	"github.com/lightglitch/seekerr/services/radarr"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	"strconv"
	"strings"
)

func NewImporter(config *viper.Viper, logger *zerolog.Logger,
	radarrClient *radarr.Client, omdbClient *omdb.Client,
	registry *provider.Registry, dispatcher *notification.Dispatcher) *Importer {
	importer := &Importer{
		logger:     logger.With().Str("Component", "Importer").Logger(),
		config:     config,
		radarr:     radarrClient,
		omdb:       omdbClient,
		registry:   registry,
		dispatcher: dispatcher,
		validator:  validator.NewRuleValidatior(logger),
		processed:  map[string]bool{},
		added:      map[string]bool{},
		excluded:   map[string]bool{},
	}

	importer.initCache()
	return importer
}

type Importer struct {
	logger     zerolog.Logger
	config     *viper.Viper
	radarr     *radarr.Client
	omdb       *omdb.Client
	registry   *provider.Registry
	validator  *validator.RuleValidatior
	dispatcher *notification.Dispatcher
	processed  map[string]bool
	added      map[string]bool
	excluded   map[string]bool
}

func (i *Importer) initCache() {
	movies, err := i.radarr.GetMovies()
	excluded, err := i.radarr.GetExcludedMovies()
	if err == nil {
		i.logger.Info().Int("Count", len(*movies)).Msg("Init radarr cache")
		i.logger.Info().Int("Count", len(*excluded)).Msg("Excluded movies in radarr")
		for _, movie := range *movies {
			i.added[movie.ImdbId] = true
		}
		for _, excluded := range *excluded {
			i.excluded[excluded.MovieTitle] = true
			i.excluded[fmt.Sprintf("tmdb:%d", excluded.TmdbID)] = true
		}
	} else {
		i.logger.Error().Err(err).Msg("Can't load radarr movies.")
	}
}

func (i *Importer) getListsConfigurations() map[string]provider.ListConfig {
	lists := map[string]provider.ListConfig{}

	if i.config.IsSet("filter") {
		i.logger.Debug().Interface("config", i.config.GetStringMap("filter")).Msgf("Debugging lists global filters.")
	}

	if !i.config.IsSet("lists") {
		return lists
	}

	listsConfig := i.config.Sub("lists")

	for listName, _ := range listsConfig.AllSettings() {
		i.logger.Info().Msgf("Processing list '%s' configuration.", listName)

		listConfig := listsConfig.Sub(listName)

		if i.config.IsSet("filter") {
			filterConfig := i.config.Sub("filter")

			if listConfig.IsSet("filter") {
				filterConfig.MergeConfigMap(listConfig.GetStringMap("filter"))
			}

			listConfig.MergeConfigMap(map[string]interface{}{
				"filter": filterConfig.AllSettings(),
			})
		}

		config := provider.ListConfig{}

		_ = listConfig.Unmarshal(&config)

		i.logger.Debug().Interface("config", config).Msgf("Debugging list '%s' configuration.", listName)
		lists[listName] = config
	}

	return lists
}

func (i *Importer) populateExtraInfo(item *provider.ListItem) {

	var movieResult *omdb.Result
	if item.Imdb != "" {
		movieResult, _ = i.omdb.GetMovieById(item.Imdb, map[string]string{})
	} else {
		movieResult, _ = i.omdb.GetMovieByTitle(item.Title, map[string]string{
			"y": strconv.Itoa(item.Year),
		})
	}
	if movieResult != nil {
		item.Imdb = movieResult.ImdbID
		item.ImdbVotes, _ = strconv.Atoi(strings.ReplaceAll(movieResult.ImdbVotes, ",", ""))
		item.Genre = strings.Split(movieResult.Genre, ", ")
		item.Language = strings.Split(movieResult.Language, ", ")
		item.Runtime, _ = strconv.Atoi(strings.TrimSuffix(movieResult.Runtime, " min"))
		item.CountRatings = len(movieResult.Ratings)

		for _, rating := range movieResult.Ratings {
			if rating.Source == omdb.OMDB_IMDB_SOURCE {
				value, _ := strconv.ParseFloat(strings.Replace(rating.Value, "/10", "", 1), 64)
				item.Ratings.Imdb = value
			}
			if rating.Source == omdb.OMDB_METACRITIC_SOURCE {
				value, _ := strconv.Atoi(strings.Replace(rating.Value, "/100", "", 1))
				item.Ratings.Metacritic = value
			}
			if rating.Source == omdb.OMDB_ROTTEN_TOMATOES_SOURCE {
				value, _ := strconv.Atoi(strings.Replace(rating.Value, "%", "", 1))
				item.Ratings.RottenTomatoes = value
			}
		}
	}

}

func (i *Importer) lookupMovie(item *provider.ListItem) (movieResult *radarr.Movie, err error) {
	if item.Tmdb != 0 {
		movieResult, err = i.radarr.LookupMovieByTmdb(strconv.Itoa(item.Tmdb))
	} else {
		movieResult, err = i.radarr.LookupMovieByImdb(item.Imdb)
	}
	return movieResult, err
}

func (i *Importer) processProviderItem(listName string, item *provider.ListItem) (approved bool, added bool) {
	itemSlug := fmt.Sprintf("%s-%d", slug.Make(item.Title), item.Year)

	approved, added = false, false
	_, processed := i.processed[item.Imdb]
	_, exist := i.added[item.Imdb]
	_, excluded := i.excluded[item.Title]
	if !excluded {
		_, excluded = i.excluded[fmt.Sprintf("tmdb:%d", item.Tmdb)]
	}
	if exist {
		i.logger.Info().Msgf("Movie already '%s (%d)' to radarr.", item.Title, item.Year)
	}
	if excluded {
		i.logger.Info().Msgf("Movie '%s (%d)' excluded from radarr.", item.Title, item.Year)
	}
	if !processed && !exist && !excluded {
		i.logger.Info().Str("ImdbId", item.Imdb).Str("slug", itemSlug).
			Msgf("Processing list item '%s (%d)'.", item.Title, item.Year)
		i.processed[item.Imdb] = true

		i.populateExtraInfo(item)

		// validate filters
		if approved = i.validator.IsItemApproved(item); approved {

			movieResult, err := i.lookupMovie(item)
			if err == nil {
				if err = i.radarr.AddMovie(movieResult); err == nil {
					i.logger.Info().Msgf("[ADDED] Movie '%s (%d)' added to radarr.", item.Title, item.Year)
					added = true
					i.dispatcher.SendEventAddMovie(listName, item, movieResult)
				} else {
					i.logger.Error().Err(err).Msg("Adding movie to radarr")
				}
			} else {
				i.logger.Error().Err(err).Msg("Looking movie in radarr")
			}
		} else if i.config.GetBool("revision") {
			if revision := i.validator.IsItemForRevision(item); revision {
				movieResult, _ := i.lookupMovie(item)
				i.dispatcher.SendEventRevisionMovie(listName, item, movieResult)
			}
		}
	}
	return approved, added
}

func (i *Importer) processProviderList(listName string, config provider.ListConfig) (approvedCount int, addedCount int) {

	i.logger.Info().
		Interface("Type", config.Type).
		Interface("filter", config.Filter).
		Msgf("Processing list '%s'.", listName)

	i.dispatcher.SendEventStartFeed(listName)
	approvedCount = 0
	addedCount = 0
	if provider, ok := i.registry.GetProvider(config.Type); ok {

		i.validator.InitRules(config)

		items, _ := provider.GetItems(config)
		for _, item := range items {
			approved, added := i.processProviderItem(listName, &item)
			if approved {
				approvedCount++
			}
			if added {
				addedCount++
			}
		}
	}
	i.logger.Info().Int("Approved", approvedCount).Int("Added", addedCount).Msgf("Finish list '%s'.", listName)
	i.dispatcher.SendEventEndFeed(listName, approvedCount, addedCount)
	return approvedCount, addedCount
}

func (i *Importer) ProcessList(listName string) {

	configurations := i.getListsConfigurations()

	if config, ok := configurations[strings.ToLower(listName)]; ok {
		i.processProviderList(listName, config)
	} else {
		i.logger.Error().Msgf("Can't find the configuration for list '%s'", listName)
	}
}

func (i *Importer) ProcessLists() {

	configurations := i.getListsConfigurations()

	approvedCount := 0
	addedCount := 0
	for listName, config := range configurations {
		approved, added := i.processProviderList(listName, config)
		approvedCount += approved
		addedCount += added
	}

	i.dispatcher.SendEventEndAllFeeds(approvedCount, addedCount)
	i.logger.Info().Int("Approved", approvedCount).Int("Added", addedCount).Msg("Finish processing lists.")
}
