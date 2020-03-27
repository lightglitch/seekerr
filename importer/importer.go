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
	registry *provider.Registry) *Importer {
	importer := &Importer{
		logger:    logger.With().Str("Component", "Importer").Logger(),
		config:    config,
		radarr:    radarrClient,
		omdb:      omdbClient,
		registry:  registry,
		processed: map[string]bool{},
		cache:     map[string]*radarr.Movie{},
	}

	importer.initCache()
	return importer
}

type Importer struct {
	logger    zerolog.Logger
	config    *viper.Viper
	radarr    *radarr.Client
	omdb      *omdb.Client
	registry  *provider.Registry
	processed map[string]bool
	cache     map[string]*radarr.Movie
}

func (i Importer) initCache() {
	movies, err := i.radarr.GetMovies()
	if err == nil {
		i.logger.Info().Int("Count", len(*movies)).Msg("Init radarr cache")
		for _, movie := range *movies {
			i.cache[movie.ImdbId] = &movie
		}
	} else {
		i.logger.Error().Err(err).Msg("Can't load radarr movies.")
	}
}

func (i Importer) getListsConfigurations() map[string]provider.ListConfig {
	lists := map[string]provider.ListConfig{}

	i.logger.Debug().Interface("config", i.config.GetStringMap("filter")).Msgf("Debugging lists global filters.")

	listsConfig := i.config.Sub("lists")

	//listsConfig.MergeConfigMap( filterConfig. )

	for listName, _ := range listsConfig.AllSettings() {
		i.logger.Info().Msgf("Processing list '%s' configuration.", listName)

		listConfig := listsConfig.Sub(listName)
		filterConfig := i.config.Sub("filter")

		filterConfig.MergeConfigMap(listConfig.GetStringMap("filter"))

		listConfig.MergeConfigMap(map[string]interface{}{
			"filter": filterConfig.AllSettings(),
		})

		config := provider.ListConfig{}

		_ = listConfig.Unmarshal(&config)

		i.logger.Debug().Interface("config", config).Msgf("Debugging list '%s' configuration.", listName)
		lists[listName] = config
	}

	return lists
}

func (i Importer) filterRulesValid(item *provider.ListItem, config *provider.ListConfig) bool {

	approved := true

	if config.Filter.Ratings.Imdb > 0 || config.Filter.Ratings.Metacritic > 0 || config.Filter.Ratings.RottenTomatoes > 0 {

		var movieResult *omdb.MovieResult
		var err error
		if item.Imdb != "" {
			movieResult, err = i.omdb.GetMovieById(item.Imdb, map[string]string{})
		} else {
			movieResult, err = i.omdb.GetMovieByTitle(item.Title, map[string]string{
				"y": strconv.Itoa(item.Year),
			})
			if movieResult != nil {
				item.Imdb = movieResult.ImdbID
			}
		}

		if err != nil {
			i.logger.Error().Err(err).Msg("Fetching omdb ratings")
			return false
		}

		if movieResult != nil {

			i.logger.Info().Str("votes", movieResult.ImdbVotes).Interface("ratings", movieResult.Ratings).Msgf("Processing item ratings '%s (%d)'.", item.Title, item.Year)

			initValue := false

			if config.Filter.MatchAllRatings {
				initValue = config.Filter.IgnoreMissingRatings >= 3-len(movieResult.Ratings)
			}

			approvedImdb, approvedMetacritic, approvedRottenTomatoes := initValue, initValue, initValue

			for _, rating := range movieResult.Ratings {
				if rating.Source == omdb.OMDB_IMDB_SOURCE {
					value, _ := strconv.ParseFloat(strings.Replace(rating.Value, "/10", "", 1), 64)
					if config.Filter.Ratings.Imdb > 0 {
						approvedImdb = value >= config.Filter.Ratings.Imdb
					}
					if config.Filter.ImdbVotes > 0 {
						votes, _ := strconv.Atoi(strings.ReplaceAll(movieResult.ImdbVotes, ",", ""))
						approvedImdb = approvedImdb && votes >= config.Filter.ImdbVotes
					}
				}
				if rating.Source == omdb.OMDB_METACRITIC_SOURCE {
					value, _ := strconv.Atoi(strings.Replace(rating.Value, "/100", "", 1))
					if config.Filter.Ratings.Metacritic > 0 {
						approvedMetacritic = value >= config.Filter.Ratings.Metacritic
					}
				}
				if rating.Source == omdb.OMDB_ROTTEN_TOMATOES_SOURCE {
					value, _ := strconv.Atoi(strings.Replace(rating.Value, "%", "", 1))
					if config.Filter.Ratings.RottenTomatoes > 0 {
						approvedRottenTomatoes = value >= config.Filter.Ratings.RottenTomatoes
					}
				}
			}

			if config.Filter.MatchAllRatings {
				approved = approved && (approvedImdb && approvedMetacritic && approvedRottenTomatoes)
			} else {
				approved = approved && (approvedImdb || approvedMetacritic || approvedRottenTomatoes)
			}
			i.logger.Info().Bool("approved", approved).Msgf("Result item ratings '%s (%d)'.", item.Title, item.Year)
		} else {
			i.logger.Error().Err(err).Msg("Fetching omdb ratings")
			return false
		}
	}

	//TODO process more filters

	return approved
}

func (i Importer) lookupMovie(item provider.ListItem) (movieResult *radarr.Movie, err error) {
	if item.Tmdb != 0 {
		movieResult, err = i.radarr.LookupMovieByTmdb(strconv.Itoa(item.Tmdb))
	} else {
		movieResult, err = i.radarr.LookupMovieByImdb(item.Imdb)
	}
	return movieResult, err
}

func (i Importer) processProviderItem(item provider.ListItem, config provider.ListConfig) (approved bool, added bool) {
	itemSlug := fmt.Sprintf("%s-%d", slug.Make(item.Title), item.Year)

	approved, added = false, false
	_, processed := i.processed[item.Imdb]
	_, exist := i.cache[item.Imdb]
	if exist {
		i.logger.Info().Msgf("Movie already '%s (%d)' to radarr.", item.Title, item.Year)
	}
	if !processed && !exist {
		i.logger.Info().Str("ImdbId", item.Imdb).Str("slug", itemSlug).
			Msgf("Processing list item '%s (%d)'.", item.Title, item.Year)
		i.processed[item.Imdb] = true

		// validate filters
		if i.filterRulesValid(&item, &config) {
			approved = true

			movieResult, err := i.lookupMovie(item)
			if err == nil {
				if err = i.radarr.AddMovie(*movieResult); err == nil {
					i.logger.Info().Msgf("[ADDED] Movie '%s (%d)' added to radarr.", item.Title, item.Year)
					added = true
				} else {
					i.logger.Error().Err(err).Msg("Adding movie to radarr")
				}
			} else {
				i.logger.Error().Err(err).Msg("Looking movie in radarr")
			}
		}
	}
	return approved, added
}

func (i Importer) processProviderList(listName string, config provider.ListConfig) (approvedCount int, addedCount int) {

	i.logger.Info().
		Interface("Type", config.Type).
		Interface("filter", config.Filter).
		Msgf("Processing list '%s'.", listName)

	approvedCount = 0
	addedCount = 0
	if provider, ok := i.registry.GetProvider(config.Type); ok {

		items, _ := provider.GetItems(config)
		for _, item := range items {
			approved, added := i.processProviderItem(item, config)
			if approved {
				approvedCount++
			}
			if added {
				addedCount++
			}
		}
	}
	i.logger.Info().Int("Approved", approvedCount).Int("Added", addedCount).Msgf("Finish list '%s'.", listName)
	return approvedCount, addedCount
}

func (i Importer) ProcessList(listName string) {

	configurations := i.getListsConfigurations()

	if config, ok := configurations[strings.ToLower(listName)]; ok {
		i.processProviderList(listName, config)
	} else {
		i.logger.Error().Msgf("Can't find the configuration for list '%s'", listName)
	}
}

func (i Importer) ProcessLists() {

	configurations := i.getListsConfigurations()

	approvedCount := 0
	addedCount := 0
	for listName, config := range configurations {
		approved, added := i.processProviderList(listName, config)
		approvedCount += approved
		addedCount += added
	}

	i.logger.Info().Int("Approved", approvedCount).Int("Added", addedCount).Msg("Finish processing lists.")
}
