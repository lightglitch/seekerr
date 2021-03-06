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

package provider

type ListType string

const (
	RSS   = "rss"
	IMDB  = "imdb"
	TRAKT = "trakt"
)

type ListFilter struct {
	Limit    int
	Exclude  []string
	Revision []string
}

type ListConfig struct {
	Url     string
	Type    ListType
	GuessIt bool
	Filter  ListFilter
}

type ListItem struct {
	Title        string
	Year         int
	Imdb         string
	Tmdb         int
	ImdbVotes    int
	Genre        []string
	Language     []string
	Runtime      int
	Ratings      Ratings
	CountRatings int
}

type Ratings struct {
	RottenTomatoes int
	Imdb           float64
	Metacritic     int
}

type ListProvider interface {
	GetItems(config ListConfig) ([]ListItem, error)
}

func NewProviderRegistry() *Registry {
	return &Registry{
		providers: map[ListType]ListProvider{},
	}
}

type Registry struct {
	providers map[ListType]ListProvider
}

func (r *Registry) RegisterProvider(listType ListType, provider ListProvider) {
	r.providers[listType] = provider
}

func (r *Registry) GetProvider(listType ListType) (provider ListProvider, ok bool) {
	provider, ok = r.providers[listType]
	return provider, ok
}
