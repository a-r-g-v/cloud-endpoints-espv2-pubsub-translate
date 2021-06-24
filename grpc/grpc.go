package grpc

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	"golang.org/x/xerrors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/a-r-g-v/cloud-endpoints-espv2-pubsub-translate/pb/testv1"
	"github.com/a-r-g-v/cloud-endpoints-espv2-pubsub-translate/pubsub"
	"github.com/a-r-g-v/cloud-endpoints-espv2-pubsub-translate/task"
)

type testV1API struct {
	pc *pubsub.Client
	ct *task.Client
	testv1.TestAPIServer
}

func NewServer(pc *pubsub.Client, ct *task.Client) (*grpc.Server, error) {
	server := grpc.NewServer()

	reflection.Register(server)
	testv1.RegisterTestAPIServer(server, NewTestV1API(pc, ct))

	return server, nil
}


func NewTestV1API(pc *pubsub.Client, ct *task.Client) testv1.TestAPIServer {
	return &testV1API{
		pc:            pc,
		ct: ct,
	}
}

func (t *testV1API) HandleTestMessageRPC(ctx context.Context, req *testv1.PubSubRequest) (*emptypb.Empty, error) {
	if err := AuthenticateEmail(ctx, "44903873603-compute@developer.gserviceaccount.com"); err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

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


func (t *testV1API) CreateTestTask(ctx context.Context, req *testv1.CreateTestTaskRequest) (*emptypb.Empty, error) {
	m := &testv1.HandleTestTaskRequest{
		Data:   req.Name,
		Amount: rand.Int63(),
	}

	url := "https://pubsub-translate-gateway-g7ag5sxmgq-an.a.run.app/v1/handle_test_task"
	if err := t.ct.CreateTask(ctx, "test-queue", strconv.FormatInt(int64(uuid.New().ID()), 10), url, m); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (t *testV1API) HandleTestTask(ctx context.Context, req *testv1.HandleTestTaskRequest) (*emptypb.Empty, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	var taskName string
	if ok {
		taskName = strings.Join(md.Get("X-CloudTasks-TaskName"), ",")
	}

	if err := AuthenticateEmail(ctx, "44903873603-compute@developer.gserviceaccount.com"); err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	fmt.Fprintf(os.Stderr, "[INFO] got tasks. name: '%s', data: '%s', amount: '%d'\n", taskName, req.Data, req.Amount)

	return &emptypb.Empty{}, nil
}

