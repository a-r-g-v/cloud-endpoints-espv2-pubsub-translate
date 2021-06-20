package grpc

import (
	"context"
	"fmt"
	"os"

	"github.com/golang/protobuf/proto"
	"golang.org/x/xerrors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/a-r-g-v/cloud-endpoints-espv2-pubsub-translate/pb/testv1"
	"github.com/a-r-g-v/cloud-endpoints-espv2-pubsub-translate/pubsub"
)

type testV1API struct {
	pc *pubsub.Client
	testv1.TestAPIServer
}

func NewServer(pc *pubsub.Client) (*grpc.Server, error) {
	server := grpc.NewServer()

	reflection.Register(server)
	testv1.RegisterTestAPIServer(server, NewTestV1API(pc))

	return server, nil
}


func NewTestV1API(pc *pubsub.Client) testv1.TestAPIServer {
	return &testV1API{
		pc:            pc,
	}
}

func (t *testV1API) HandleTestMessageRPC(ctx context.Context, req *testv1.PubSubRequest) (*emptypb.Empty, error) {
	var m testv1.TestMessage
	if err := proto.Unmarshal(req.Message.Data, &m); err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] proto.Unmarshal failed. err: %v\n", err)
		return &emptypb.Empty{}, nil // ACK
	}

	if m.Data == "nack" {
		return nil, status.Error(codes.InvalidArgument, "nack") // NACK
	}

	fmt.Fprintf(os.Stderr, "[INFO] got request. data: '%s', created_at: '%s'\n", m.Data, m.CreatedAt.String())
	return &emptypb.Empty{}, nil // ACK
}

func (t *testV1API) PublishTestMessageRPC(ctx context.Context, req *testv1.PublishTestMessageRequest) (*emptypb.Empty, error) {
	m := &testv1.TestMessage{
		Data:      req.Data,
		CreatedAt: timestamppb.Now(),
	}

	b, err := proto.Marshal(m)
	if err != nil {
		return nil, xerrors.Errorf("proto.Marshal failed: %w", err)
	}

	if err := t.pc.Publish(ctx, "test_topic", b); err != nil {
		return nil, xerrors.Errorf("Publish failed: %w", err)
	}
	return &emptypb.Empty{}, nil
}
