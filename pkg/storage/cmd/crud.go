package cmd

import (
	"context"
	"encoding/json"
	"time"

	"github.com/EventStore/EventStore-Client-Go/v3/esdb"
	"github.com/krixlion/dev-forum_article/pkg/entity"
	"github.com/krixlion/dev-forum_article/pkg/event"
)

func (db DB) Close() error {
	return db.client.Close()
}

func (db DB) Create(ctx context.Context, article entity.Article) error {

	jsonArticle, err := json.Marshal(article)
	if err != nil {
		return err
	}

	e := event.Event{
		Entity:    entity.ArticleEntity,
		Type:      event.Created,
		Body:      jsonArticle,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(e)
	if err != nil {
		return err
	}

	eventData := esdb.EventData{
		ContentType: esdb.ContentTypeJson,
		EventType:   string(e.Type),
		Data:        data,
	}
	_, err = db.client.AppendToStream(ctx, "some-stream", esdb.AppendToStreamOptions{}, eventData)

	return err
}

func (db DB) Update(ctx context.Context, article entity.Article) error {

	jsonArticle, err := json.Marshal(article)
	if err != nil {
		return err
	}

	e := event.Event{
		Entity:    entity.ArticleEntity,
		Type:      event.Updated,
		Body:      jsonArticle,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(e)
	if err != nil {
		return err
	}

	eventData := esdb.EventData{
		ContentType: esdb.ContentTypeJson,
		EventType:   string(e.Type),
		Data:        data,
	}
	_, err = db.client.AppendToStream(ctx, "some-stream", esdb.AppendToStreamOptions{}, eventData)

	return err
}

func (db DB) Delete(ctx context.Context, id string) error {

	jsonID, err := json.Marshal(id)
	if err != nil {
		return err
	}

	e := event.Event{
		Entity:    entity.ArticleEntity,
		Type:      event.Deleted,
		Body:      jsonID,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(e)
	if err != nil {
		return err
	}

	eventData := esdb.EventData{
		ContentType: esdb.ContentTypeJson,
		EventType:   string(e.Type),
		Data:        data,
	}
	_, err = db.client.AppendToStream(ctx, "some-stream", esdb.AppendToStreamOptions{}, eventData)

	return err
}
