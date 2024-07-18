package eventstore

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/krixlion/dev_forum-article/pkg/entity"
	"github.com/krixlion/dev_forum-lib/event"
	"github.com/krixlion/dev_forum-lib/tracing"

	"github.com/EventStore/EventStore-Client-Go/v4/esdb"
)

func addArticlesPrefix(v string) string {
	return fmt.Sprintf("%s-%s", "article", v)
}

func (db Eventstore) Create(ctx context.Context, article entity.Article) error {
	ctx, span := db.tracer.Start(ctx, "esdb.Create")
	defer span.End()

	e, err := event.MakeEvent(event.ArticleAggregate, event.ArticleCreated, article)
	if err != nil {
		tracing.SetSpanErr(span, err)
		return err
	}

	data, err := json.Marshal(e)
	if err != nil {
		tracing.SetSpanErr(span, err)
		return err
	}

	eventData := esdb.EventData{
		ContentType: esdb.ContentTypeJson,
		EventType:   string(e.Type),
		Data:        data,
	}

	streamID := addArticlesPrefix(article.Id)
	if _, err := db.client.AppendToStream(ctx, streamID, esdb.AppendToStreamOptions{}, eventData); err != nil {
		tracing.SetSpanErr(span, err)
		return err
	}

	return nil
}

func (db Eventstore) Update(ctx context.Context, article entity.Article) error {
	ctx, span := db.tracer.Start(ctx, "esdb.Update")
	defer span.End()

	e, err := event.MakeEvent(event.ArticleAggregate, event.ArticleUpdated, article)
	if err != nil {
		tracing.SetSpanErr(span, err)
		return err
	}

	data, err := json.Marshal(e)
	if err != nil {
		tracing.SetSpanErr(span, err)
		return err
	}

	lastEvent, err := db.getLastRevision(ctx, article.Id)
	if err != nil {
		return err
	}

	appendOpts := esdb.AppendToStreamOptions{
		ExpectedRevision: esdb.Revision(lastEvent.OriginalEvent().EventNumber),
	}

	eventData := esdb.EventData{
		ContentType: esdb.ContentTypeJson,
		EventType:   string(e.Type),
		Data:        data,
	}
	streamID := addArticlesPrefix(article.Id)

	_, err = db.client.AppendToStream(ctx, streamID, appendOpts, eventData)
	if err != nil {
		tracing.SetSpanErr(span, err)
		return err
	}

	return nil
}

func (db Eventstore) Delete(ctx context.Context, id string) error {
	ctx, span := db.tracer.Start(ctx, "esdb.Delete")
	defer span.End()

	e, err := event.MakeEvent(event.ArticleAggregate, event.ArticleDeleted, id)
	if err != nil {
		tracing.SetSpanErr(span, err)
		return err
	}

	data, err := json.Marshal(e)
	if err != nil {
		tracing.SetSpanErr(span, err)
		return err
	}

	eventData := esdb.EventData{
		ContentType: esdb.ContentTypeJson,
		EventType:   string(e.Type),
		Data:        data,
	}
	streamID := addArticlesPrefix(id)

	if _, err := db.client.AppendToStream(ctx, streamID, esdb.AppendToStreamOptions{}, eventData); err != nil {
		tracing.SetSpanErr(span, err)
		return err
	}

	return nil
}

func (db Eventstore) getLastRevision(ctx context.Context, articleId string) (*esdb.ResolvedEvent, error) {
	ctx, span := db.tracer.Start(ctx, "esdb.lastRevision")
	defer span.End()

	readOpts := esdb.ReadStreamOptions{
		Direction: esdb.Backwards,
		From:      esdb.End{},
	}

	streamID := addArticlesPrefix(articleId)

	stream, err := db.client.ReadStream(ctx, streamID, readOpts, 1)
	if err != nil {
		tracing.SetSpanErr(span, err)
		return nil, err
	}
	defer stream.Close()

	lastEvent, err := stream.Recv()
	if err != nil {
		tracing.SetSpanErr(span, err)
		return nil, err
	}

	return lastEvent, nil
}
