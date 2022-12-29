package eventstore

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/krixlion/dev-forum_article/pkg/entity"
	"github.com/krixlion/dev-forum_article/pkg/event"
	"github.com/krixlion/dev-forum_article/pkg/tracing"

	"github.com/EventStore/EventStore-Client-Go/v3/esdb"
	"go.opentelemetry.io/otel"
)

func (db DB) Create(ctx context.Context, article entity.Article) error {
	ctx, span := otel.Tracer(tracing.ServiceName).Start(ctx, "esdb.Create")
	defer span.End()

	jsonArticle, err := json.Marshal(article)
	if err != nil {
		tracing.SetSpanErr(span, err)
		return err
	}

	e := event.Event{
		AggregateId: "article",
		Type:        event.ArticleCreated,
		Body:        jsonArticle,
		Timestamp:   time.Now(),
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
	streamID := fmt.Sprintf("%s-%s", "article", article.Id)

	_, err = db.client.AppendToStream(ctx, streamID, esdb.AppendToStreamOptions{}, eventData)
	if err != nil {
		tracing.SetSpanErr(span, err)
		return err
	}

	return nil
}

func (db DB) Update(ctx context.Context, article entity.Article) error {
	ctx, span := otel.Tracer(tracing.ServiceName).Start(ctx, "esdb.Update")
	defer span.End()

	jsonArticle, err := json.Marshal(article)
	if err != nil {
		return err
	}

	e := event.Event{
		AggregateId: "article",
		Type:        event.ArticleUpdated,
		Body:        jsonArticle,
		Timestamp:   time.Now(),
	}

	data, err := json.Marshal(e)
	if err != nil {
		tracing.SetSpanErr(span, err)
		return err
	}

	lastEvent, err := db.lastRevision(ctx, article.Id)
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
	streamID := fmt.Sprintf("%s-%s", "article", article.Id)

	_, err = db.client.AppendToStream(ctx, streamID, appendOpts, eventData)
	if err != nil {
		tracing.SetSpanErr(span, err)
		return err
	}

	return nil
}

func (db DB) Delete(ctx context.Context, id string) error {
	ctx, span := otel.Tracer(tracing.ServiceName).Start(ctx, "esdb.Delete")
	defer span.End()

	jsonID, err := json.Marshal(id)
	if err != nil {
		tracing.SetSpanErr(span, err)
		return err
	}

	e := event.Event{
		AggregateId: "article",
		Type:        event.ArticleDeleted,
		Body:        jsonID,
		Timestamp:   time.Now(),
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
	streamID := fmt.Sprintf("%s-%s", "article", id)

	_, err = db.client.AppendToStream(ctx, streamID, esdb.AppendToStreamOptions{}, eventData)

	if err != nil {
		tracing.SetSpanErr(span, err)
		return err
	}

	return nil
}

func (db DB) lastRevision(ctx context.Context, articleId string) (*esdb.ResolvedEvent, error) {
	ctx, span := otel.Tracer(tracing.ServiceName).Start(ctx, "esdb.lastRevision")
	defer span.End()

	readOpts := esdb.ReadStreamOptions{
		Direction: esdb.Backwards,
		From:      esdb.End{},
	}

	streamID := fmt.Sprintf("%s-%s", "article", articleId)

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
