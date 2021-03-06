package resolvers

import (
	"context"
	log "github.com/sirupsen/logrus"
	"gitlab.com/olaris/olaris-server/metadata/db"
	"math/rand"
	"time"
)

type graphqlLibrarySubscriber struct {
	resolver *Resolver
}

func (h graphqlLibrarySubscriber) MovieAdded(movie *db.Movie) {
	e := &MovieAddedEvent{movie: &MovieResolver{*movie}}

	go func() {
		select {
		case h.resolver.movieAddedEvents <- e:
		case <-time.After(1 * time.Second):
			warnDroppedEvent("Movie", movie.Title)
		}
	}()
	return
}

func (h graphqlLibrarySubscriber) EpisodeAdded(episode *db.Episode) {
	e := &EpisodeAddedEvent{episode: &EpisodeResolver{*episode}}
	go func() {
		select {
		case h.resolver.episodeAddedEvents <- e:
		case <-time.After(1 * time.Second):
			warnDroppedEvent("Episode", episode.Name)
		}
	}()
	return
}

func (h graphqlLibrarySubscriber) SeriesAdded(series *db.Series) {
	e := &SeriesAddedEvent{series: &SeriesResolver{*series}}
	go func() {
		select {
		case h.resolver.seriesAddedEvents <- e:
		case <-time.After(1 * time.Second):
			warnDroppedEvent("Series", series.Name)
		}
	}()
	return
}

func (h graphqlLibrarySubscriber) SeasonAdded(season *db.Season) {
	e := &SeasonAddedEvent{season: &SeasonResolver{*season}}
	go func() {
		select {
		case h.resolver.seasonAddedEvents <- e:
		case <-time.After(1 * time.Second):
			warnDroppedEvent("Season", season.Name)
		}
	}()
	return
}

// SeriesAdded creates a subscription for all SeriesAdded events
func (r *Resolver) SeriesAdded(ctx context.Context) <-chan *SeriesAddedEvent {
	log.Debugln("Adding subscription to SeriesAddedEvent")
	c := make(chan *SeriesAddedEvent)
	r.subscriberChan <- &graphqlSubscriber{seriesAddedEventChan: c, stop: ctx.Done()}

	return c
}

// SeasonAdded creates a subscription for all SeasonAdded events
func (r *Resolver) SeasonAdded(ctx context.Context) <-chan *SeasonAddedEvent {
	log.Debugln("Adding subscription to SeasonAddedEvent")
	c := make(chan *SeasonAddedEvent)
	r.subscriberChan <- &graphqlSubscriber{seasonAddedEventChan: c, stop: ctx.Done()}

	return c
}

// MovieAdded creates a subscription for all MovieAdded events
func (r *Resolver) MovieAdded(ctx context.Context) <-chan *MovieAddedEvent {
	log.Debugln("Adding subscription to MovieAddedEvent")
	c := make(chan *MovieAddedEvent)
	r.subscriberChan <- &graphqlSubscriber{movieAddedEventChan: c, stop: ctx.Done()}

	return c
}

// EpisodeAdded creates a subscription for all MovieAdded events
func (r *Resolver) EpisodeAdded(ctx context.Context) <-chan *EpisodeAddedEvent {
	log.Debugln("Adding subscription to EpisodeAddedEvent")
	c := make(chan *EpisodeAddedEvent)
	r.subscriberChan <- &graphqlSubscriber{episodeAddedEventChan: c, stop: ctx.Done()}

	return c
}

type graphqlSubscriber struct {
	stop                  <-chan struct{}
	episodeAddedEventChan chan<- *EpisodeAddedEvent
	movieAddedEventChan   chan<- *MovieAddedEvent
	seriesAddedEventChan  chan<- *SeriesAddedEvent
	seasonAddedEventChan  chan<- *SeasonAddedEvent
}

func checkAndSendEvent(id string, s *graphqlSubscriber, unsubChan chan string, event interface{}) {
	// Check the stop chan first, otherwise it might never be processed if events are always
	// available.
	// See https://github.com/matiasanaya/go-graphql-subscription-example/issues/4#issuecomment-424604826
	select {
	case <-s.stop:
		unsubChan <- id
		return
	default:
	}

	if episodeAddedEvent, ok := event.(*EpisodeAddedEvent); ok {
		select {
		case <-s.stop:
			unsubChan <- id
		case s.episodeAddedEventChan <- episodeAddedEvent:
		case <-time.After(time.Second):
			log.Warningln(
				"Timed out trying to deliver to deliver EpisodeAddedEvent:",
				episodeAddedEvent)
		}
		return
	}

	if movieEvent, ok := event.(*MovieAddedEvent); ok {
		select {
		case <-s.stop:
			unsubChan <- id
		case s.movieAddedEventChan <- movieEvent:
		case <-time.After(time.Second):
			log.Warningln(
				"Timed out trying to deliver to deliver MovieAddedEvent:",
				movieEvent)
		}
		return
	}

	if seriesEvent, ok := event.(*SeriesAddedEvent); ok {
		select {
		case <-s.stop:
			unsubChan <- id
		case s.seriesAddedEventChan <- seriesEvent:
		case <-time.After(time.Second):
			log.Warningln(
				"Timed out trying to deliver to deliver SeriesAddedEvent:",
				seriesEvent)
		}
		return
	}

	if seasonEvent, ok := event.(*SeasonAddedEvent); ok {
		select {
		case <-s.stop:
			unsubChan <- id
		case s.seasonAddedEventChan <- seasonEvent:
		case <-time.After(time.Second):
			log.Warningln(
				"Timed out trying to deliver to deliver SeasonAddedEvent:",
				seasonEvent)
		}
		return
	}

	log.Errorln("Got an event that could not be cast")
}

func (r *Resolver) startGraphQLSubscriptionManager(exitChan chan bool) {
	unsubscribe := make(chan string)
	subscriptions := map[string]*graphqlSubscriber{}

	for {
		select {
		case exit := <-exitChan:
			if exit {
				log.Debugln("Shutting down GraphQLSubscriptionManager")
				break
			}
		case id := <-unsubscribe:
			log.WithFields(log.Fields{"id": id}).Debugln("Received unscribe event via channel")
			delete(subscriptions, id)
		case s := <-r.subscriberChan:
			id := randomID()
			subscriptions[id] = s
			log.WithFields(log.Fields{"id": id}).Debugln("Added subscription")
		case e := <-r.episodeAddedEvents:
			log.Debugln("Received episode event")
			for id, s := range subscriptions {
				if s.episodeAddedEventChan != nil {
					go checkAndSendEvent(id, s, unsubscribe, e)
				}
			}
		case e := <-r.movieAddedEvents:
			log.Debugln("Received movie event")
			for id, s := range subscriptions {
				if s.movieAddedEventChan != nil {
					go checkAndSendEvent(id, s, unsubscribe, e)
				}
			}
		case e := <-r.seriesAddedEvents:
			log.Debugln("Received series event")
			for id, s := range subscriptions {
				if s.seriesAddedEventChan != nil {
					go checkAndSendEvent(id, s, unsubscribe, e)
				}
			}
		case e := <-r.seasonAddedEvents:
			log.Debugln("Received season event")
			for id, s := range subscriptions {
				if s.seasonAddedEventChan != nil {
					go checkAndSendEvent(id, s, unsubscribe, e)
				}
			}
		}
	}
}

func randomID() string {
	var letter = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	b := make([]rune, 16)
	for i := range b {
		b[i] = letter[rand.Intn(len(letter))]
	}
	return string(b)
}

// SeriesAddedEvent is fired when a new episode has been found and correctly identified
type SeriesAddedEvent struct {
	series *SeriesResolver
}

// Series is a resolver for the series object
func (s *SeriesAddedEvent) Series() *SeriesResolver {
	return s.series
}

// SeasonAddedEvent is fired when a new episode has been found and correctly identified
type SeasonAddedEvent struct {
	season *SeasonResolver
}

// Season is a resolver for the season object
func (s *SeasonAddedEvent) Season() *SeasonResolver {
	return s.season
}

// EpisodeAddedEvent is fired when a new episode has been found and correctly identified
type EpisodeAddedEvent struct {
	episode *EpisodeResolver
}

// Episode is a resolver for the episode struct
func (e *EpisodeAddedEvent) Episode() *EpisodeResolver {
	return e.episode
}

// MovieAddedEvent adds an event when a movie has been found and correctly identified
type MovieAddedEvent struct {
	movie *MovieResolver
}

// Movie is a resolver for the movie struct
func (m *MovieAddedEvent) Movie() *MovieResolver {
	return m.movie
}

func warnDroppedEvent(m string, n string) {
	log.WithFields(log.Fields{"eventType": m, "name": n}).Warnln("Subscription event could not be pushed into channel. Events might be missed.")
}
