package service_test

import (
	"context"
	"eventstore/internal/pb"
	"eventstore/internal/pkg/server"
	"log"
	"net"
	"testing"

	"github.com/matryer/is"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

func init() {
	// bufconn allows the server to call itself
	// great for testing across whole infrastructure
	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()
	eventstore := server.MakeEventStoreServer()
	pb.RegisterEventStoreServer(s, eventstore)
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func TestCreateAndGet(t *testing.T) {
	is := is.New(t)
	ctx := context.Background()

	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	client := pb.NewEventStoreClient(conn)

	event := &pb.Event{
		EventId:       "2345",
		EventType:     "UserCreated",
		AggregateId:   "user",
		AggregateType: "service",
		EventData:     "name: imie",
		// ChannelName:   "user",
	}
	createResponse, err := client.Create(ctx, &pb.CreateEventRequest{
		Event: event,
	})

	if !createResponse.IsSuccess {
		t.Fatalf("Failed to create event, err: %v", err)
	}

	resp, err := client.Get(ctx, &pb.GetEventsRequest{
		EventId:     "2345",
		AggregateId: "user",
	})
	if err != nil {
		t.Fatalf("Failed to get: %v", err)
	}

	want := &pb.GetEventsResponse{
		Event: event,
	}
	is.Equal(resp, want)
}
