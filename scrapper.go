package main

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/fredshema/rssagg/internal/database"
	"github.com/google/uuid"
)

func startScrapping(
	db *database.Queries,
	concurrency int,
	timeBetweenRequest time.Duration,
) {
	log.Printf("Scrapping on %d goroutines every %s duration\n", concurrency, timeBetweenRequest)
	ticker := time.NewTicker(timeBetweenRequest)
	for ; ; <-ticker.C {
		feeds, err := db.GetNextFeedsToFetch(
			context.Background(),
			int32(concurrency),
		)
		if err != nil {
			log.Println("Cannot fetch feeds to fetch:", err)
			continue
		}
		log.Printf("Found %v feeds to fetch!", len(feeds))

		wg := &sync.WaitGroup{}
		for _, feed := range feeds {
			wg.Add(1)
			go scrapeFeed(db, feed, wg)
		}
		wg.Wait()
	}
}

func scrapeFeed(db *database.Queries, feed database.Feed, wg *sync.WaitGroup) {
	defer wg.Done()
	_, err := db.MarkFeedAsFetched(context.Background(), feed.ID)
	if err != nil {
		log.Printf("Couldn't mark feed %s fetched: %v", feed.Name, err)
		return
	}

	feedData, err := urlToFeed(feed.Url)
	if err != nil {
		log.Printf("Couldn't collect feed %s: %v", feed.Name, err)
		return
	}
	for _, item := range feedData.Channel.Item {
		pubDate, err := time.Parse(time.RFC1123Z, item.PubDate)
		if err != nil {
			log.Printf("Couldn't parse date %s: %v", item.PubDate, err)
			continue
		}
		_, err = db.CreatePost(context.Background(), database.CreatePostParams{
			ID:          uuid.New(),
			Title:       item.Title,
			Url:         item.Link,
			FeedID:      feed.ID,
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
			Description: item.Description[:255],
			PublishedAt: pubDate,
		})
		if err != nil {
			if strings.Contains(err.Error(), "unique constraint") {
				continue
			}
			log.Printf("Couldn't create post %s: %v", item.Title, err)
		}
	}
	log.Printf("Feed %s collected, %v posts found", feed.Name, len(feedData.Channel.Item))
}
