# seekerr
Tool to add new movies to Radarr based on RSS, IMDB and Trakt lists. 

[![GitHub issues](https://img.shields.io/github/issues/lightglitch/seekerr.svg?maxAge=60&style=flat-square)](https://github.com/lightglitch/seekerr/issues)
[![GitHub pull requests](https://img.shields.io/github/issues-pr/lightglitch/seekerr.svg?maxAge=60&style=flat-square)](https://github.com/lightglitch/seekerr/pulls)
[![MIT](https://img.shields.io/badge/license-MIT-blue.svg?maxAge=60&style=flat-square)](https://opensource.org/licenses/MIT)
[![Copyright 2020-2022](https://img.shields.io/badge/copyright-2022-blue.svg?maxAge=60&style=flat-square)](https://github.com/lightglitch/seekerr)
[![Github Releases](https://img.shields.io/github/downloads/lightglitch/seekerr/total.svg?maxAge=60&style=flat-square)](https://github.com/lightglitch/seekerr/releases/)

---

- [seekerr](#seekerr)
  - [Introduction](#introduction)
  - [Configuration](#configuration)
    - [Sample Configuration](#sample-configuration)
    - [CRON](#cron)
    - [Services](#services)
    - [Filters](#filters)
    - [Lists](#lists)
    - [Notifications](#notifications)
    - [Logger](#logger)
  - [Usage](#usage)
    - [Docker](#docker)
    - [General](#general)
    - [Import](#import)
    - [TODO](#todo)
    - [References and Inspiration](#references-and-inspiration)

## Introduction

Seekerr uses RSS, IMDB and Trakt.tv lists to find movies and adds them to Radarr.

Examples of supported lists:

- RSS
  - [RARBG](https://rarbgprx.org/rssdd_magnet.php?category=44)

- IMDB
  - [New Releases](https://www.imdb.com/list/ls016522954/?sort=list_order,asc&st_dt=&mode=detail&page=1&title_type=movie&user_rating=6.0%2C&ref_=ttls_ref_rt_usr)
- Trakt
  - Official Trakt Lists
    - Trending
    - Popular
    - Anticipated
    - Box Office
  - Public Lists
    - [Movist App](https://trakt.tv/users/movistapp/lists/now-playing?sort=rank,asc)

## Configuration

### Sample Configuration

```yaml
logger:
  level: info # panic,fatal,error,warn,info,debug,trace
  timeFormat: "" # golang time format
  color: true # active color on console
  human: true # store file log as human readable
  file: "var/log/seekerr.%Y%m%d.log" # %Y%m%d is used for rotation. leave empty to disable file log

cron: "0 */2 * * *"

services:
  resty:
    debug: false
    timeout: 20s #golang duration
    retry: 0 # zero for no retries
    retryWaitTime: 1s
    retryMaxWaitTime: 10s

  trakt:
    apiKey: ""

  omdb:
    apiKey: ""

  guessIt:
    type: "command" # webservice
    path: "guessit"
    # url: "http://192.168.1.100:5000/"

  radarr:
    url: "http://192.168.1.100:7878/"
    apiKey: ""
    rootFolder: "/movies/"
    quality: "Bluray"
    minimumAvailability: "inCinemas"
    monitored: true
    searchForMovie: false

notifications:
  gotify:
    webhook: "http://192.168.1.100:8070/message?token=XXXX"
    events: ["ADDED_MOVIE","REVISION_MOVIE","FINISH_ALL_FEEDS"] # START_FEED, FINISH_FEED, FINISH_ALL_FEEDS, ADDED_MOVIE, leave empty for all

  slack:
    webhook: "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX"
    events: ["ADDED_MOVIE","REVISION_MOVIE","FINISH_ALL_FEEDS"] # START_FEED, FINISH_FEED, FINISH_ALL_FEEDS, ADDED_MOVIE, leave empty for all

importer:
  revision: false
  filter:
    limit: 100 # limit the movies to process on each list
    exclude:
      - 'CountRatings < 2 || Runtime < 20 || ImdbVotes < 1000 || Year > Now().Year()'
      - 'Ratings.Imdb != 0 && Ratings.Imdb < 7'
      - 'Ratings.Metacritic != 0 && Ratings.Metacritic < 70'
      - 'Ratings.RottenTomatoes != 0 && Ratings.RottenTomatoes < 75'
    revision:
      - 'CountRatings < 2 || Runtime < 20 || ImdbVotes < 1000 || Year > Now().Year()'
      - 'Ratings.Imdb != 0 && Ratings.Imdb < 6.5'
      - 'Ratings.Metacritic != 0 && Ratings.Metacritic < 60'
      - 'Ratings.RottenTomatoes != 0 && Ratings.RottenTomatoes < 65'

  lists:
    rarbg:
      type: "rss" # rss | trakt | imdb
      url: "https://rarbgprx.org/rssdd_magnet.php?category=44"
      guessIt: true

    imdb:
      type: "imdb" # rss | trakt | imdb
      url: "https://www.imdb.com/list/ls016522954/?sort=list_order,asc&st_dt=&mode=detail&page=1&title_type=movie&user_rating=6.0%2C&ref_=ttls_ref_rt_usr"

    traktTrending:
      type: "trakt" # rss | trakt | imdb
      # special urls trakt://movies/trending, trakt://movies/popular, trakt://movies/anticipated, trakt://movies/boxoffice
      url: "trakt://movies/trending"

    traktPublic:
      type: "trakt" # rss | trakt | imdb
      url: "https://trakt.tv/users/movistapp/lists/now-playing?sort=rank,asc"
```
### CRON

Added a new cron command that runs the import based on the schedule in the configuration:
  ```yaml
  cron: "0 */2 * * *"
  ```

A cron expression represents a set of times, using 5 space-separated fields.

Field name   | Mandatory? | Allowed values  | Allowed special characters
----------   | ---------- | --------------  | --------------------------
Minutes      | Yes        | 0-59            | * / , -
Hours        | Yes        | 0-23            | * / , -
Day of month | Yes        | 1-31            | * / , - ?
Month        | Yes        | 1-12 or JAN-DEC | * / , -
Day of week  | Yes        | 0-6 or SUN-SAT  | * / , - ?

Note: Month and Day-of-week field values are case insensitive. "SUN", "Sun", and "sun" are equally accepted.

Then execute the following command:

```
seekerr cron
```


### Services

- Radarr

  Radarr configuration.
  
  ```yaml
    radarr:
      url: "http://192.168.1.100:7878/"
      apiKey: ""
      rootFolder: "/movies/"
      quality: "Bluray"
      minimumAvailability: "inCinemas"
      monitored: true
      searchForMovie: false
  ```

  `apiKey` - Radarr's API Key.
  
  `quality` - Quality Profile that movies are assigned to.
  
  `minimumAvailability` - The minimum availability the movies are set to.
  
  - Choices are `announced`, `inCinemas`, `released` (Physical/Web), or `predb`.

  `rootFolder` - Root folder for movies.
  
  `url` - Radarr's URL.
  
  - Note: If you have URL Base enabled in Radarr's settings, you will need to add that into the URL as well.
  
- OMDB

  [OMDb](https://www.omdbapi.com/) Authentication info.  
  Needed to fetch movie ratings and filter out movies.

  ```yaml
    omdb:
      apiKey: ""
  ```

- Trakt

  1. Create a Trakt application by going [here](https://trakt.tv/oauth/applications/new)
  
  2. Enter a name for your application; for example `seekerr`
  
  3. Enter `urn:ietf:wg:oauth:2.0:oob` in the `Redirect uri` field.
  
  4. Click "SAVE APP".
  
  5. Open the seekerr configuration file `seekerr.yaml` and insert your Trakt API Key:
  
      ```yaml
        trakt:
          apiKey: "your_trakt_api_key"
      ```

- GuessIt

  GuessIt it's used to parse the title of RSS item and obtain the correct movie name and year.
  By knowing the correct movie we can fetch the ratings from OMDB.  
  It can be configured as a command line or as rest service.
  
  ```yaml
    guessIt:
      type: "command" # webservice
      path: "guessit"
      # url: "http://192.168.1.100:5000/"
  ```

- Resty

  ```yaml
    resty:
      debug: false
      timeout: 20s #golang duration
      retry: 0 # zero for no retries
      retryWaitTime: 1s
      retryMaxWaitTime: 10s
  ```

### Filters

The filters can be configured globally and per list, the list configuration takes precedence over the global filter configuration.

```yaml
  revision: false
  filter:
    limit: 100 # limit the movies to process on each list
    exclude:
      - 'CountRatings < 2 || Runtime < 20 || ImdbVotes < 1000 || Year > Now().Year()'
      - 'Ratings.Imdb != 0 && Ratings.Imdb < 7'
      - 'Ratings.Metacritic != 0 && Ratings.Metacritic < 70'
      - 'Ratings.RottenTomatoes != 0 && Ratings.RottenTomatoes < 75'
    revision:
      - 'CountRatings < 2 || Runtime < 20 || ImdbVotes < 1000 || Year > Now().Year()'
      - 'Ratings.Imdb != 0 && Ratings.Imdb < 6.5'
      - 'Ratings.Metacritic != 0 && Ratings.Metacritic < 60'
      - 'Ratings.RottenTomatoes != 0 && Ratings.RottenTomatoes < 65'
```

  `revision` - Notify me about movies that are not approved but match revision rules

  `limit` - Process only this number movies in the list

  `exclude` - An list of expressions that exclude the movie from being added

### Lists

The base configuration for the lists is:

```yaml
  name_of_list:
    # The type of the feed, support 3 types
    type: rss | trakt | imdb
    # special urls for trakt type trakt://movies/trending, trakt://movies/popular, trakt://movies/anticipated, trakt://movies/boxoffice
    url: "http://feed-url.com"
    guessIt: true # only for rss and if it's necessary to parse the title to get the correct movie name and year
    # you can override the global filters for a specific feed
    filter:  
```

- RSS

```yaml
    rarbg:
      type: "rss"
      url: "https://rarbgprx.org/rssdd_magnet.php?category=44"
      guessIt: true
```
  
- IMDB

```yaml
    imdb:
      type: "imdb"
      url: "https://www.imdb.com/list/ls016522954/?sort=list_order,asc&st_dt=&mode=detail&page=1&title_type=movie&user_rating=6.0%2C&ref_=ttls_ref_rt_usr"
```

- Trakt

```yaml
    traktTrending:
      type: "trakt" # rss | trakt | imdb
      # special urls trakt://movies/trending, trakt://movies/popular, trakt://movies/anticipated, trakt://movies/boxoffice
      url: "trakt://movies/trending"

    traktPublic:
      type: "trakt" # rss | trakt | imdb
      url: "https://trakt.tv/users/movistapp/lists/now-playing?sort=rank,asc"
```

### Notifications

- Gotify

```yaml
notifications:
  gotify:
    webhook: "http://192.168.1.100:8070/message?token=XXXX"
    events: ["ADDED_MOVIE","REVISION_MOVIE","FINISH_ALL_FEEDS"] # START_FEED, FINISH_FEED, FINISH_ALL_FEEDS, ADDED_MOVIE, leave empty for all
```

- Slack

```yaml
  slack:
    webhook: "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX"
    events: ["ADDED_MOVIE","REVISION_MOVIE","FINISH_ALL_FEEDS"] # START_FEED, FINISH_FEED, FINISH_ALL_FEEDS, ADDED_MOVIE, leave empty for all
```

  Slack notification example:
  
  ![slack](https://user-images.githubusercontent.com/196953/78181877-fd144a00-745c-11ea-9832-0cdfbcb3be2c.jpg)

### Logger

```yaml
logger:
  level: info # panic,fatal,error,warn,info,debug,trace
  timeFormat: "" # golang time format
  color: true # active color on console
  human: true # store log as human readable
  file: "var/log/seekerr.%Y%m%d.log" # %Y%m%d is used for rotation. leave empty to disable file log
```

## Usage

### Docker

Using stable releases:

```
docker run -d --name='Seekerr' -v '<path to data>':'/config':'rw' 'lightglitch/seekerr:stable'
```

Using master:

```
docker run -d --name='Seekerr' -v '<path to data>':'/config':'rw' 'lightglitch/seekerr:latest'
```

You need to create your config file before running the docker image.

### General

```
seekerr
```

```
Tool to add new movies to Radarr using internet lists.
You can filter, exclude and define the minimum ratings to add the movies you pretend.

Usage:
  seekerr [command]

Available Commands:
  help        Help about any command
  import      Import the movies found in the lists to radarr.
  version     Print the version number of seekerr

Flags:
      --config string   config file (default is config/seekerr.yaml)
  -h, --help            help for seekerr

Use "seekerr [command] --help" for more information about a command.
```

### Import

```
seekerr import --help
```

```
Import the movies found in the lists to radarr.

Usage:
  seekerr import [flags]

Flags:
  -h, --help          help for import
  -l, --list string   The name of the list to import (default "all")
  -r, --revision      Notify me about movies that are not approved but match revision rules

Global Flags:
      --config string   config file (default is config/seekerr.yaml)
```

`-l`, `--list` -  The name of the list that is configured in the file seekerr.yaml. If empty imports all lists.


### Cron

```
seekerr cron --help
```

```
Import the movies found in the lists to radarr using the schedule on the config file.

Usage:
  seekerr cron [flags]

Flags:
  -h, --help              help for cron
  -s, --schedule string   Run with this cron schedule

Global Flags:
      --config string   config file (default is config/seekerr.yaml)
```

### TODO

- [ ] Tests
- [ ] Support Series and Sonarr

### References and Inspiration

- [traktarr](https://github.com/l3uddz/traktarr/): Traktarr uses Trakt.tv to find shows and movies to add in to Sonarr and Radarr, respectively.
- [Radarr](https://github.com/Radarr/Radarr): A fork of Sonarr to work with movies à la Couchpotato.


- Seekerr logo used icons made by **phatplus** and **freepik** from [www.flaticon.com](https://www.flaticon.com/)
---
