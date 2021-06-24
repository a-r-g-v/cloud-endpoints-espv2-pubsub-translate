package task

import (
	"context"
	"fmt"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	"golang.org/x/xerrors"
	taskspb "google.golang.org/genproto/googleapis/cloud/tasks/v2"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type Client struct {
	client              *cloudtasks.Client
	queuePath           func(queueID string) string
	serviceAccountEmail string
	audience            string
}

func NewClient(ctx context.Context, projectID string, locationID string, serviceAccountEmail, audience string) (*Client, error) {
	pc, err := cloudtasks.NewClient(ctx)
	if err != nil {
		return nil, xerrors.Errorf("cloudtasks.NewClient: %w", err)
	}

	return &Client{
		client: pc,
		queuePath: func(queueID string) string {
			return fmt.Sprintf("projects/%s/locations/%s/queues/%s", projectID, locationID, queueID)
		},
		serviceAccountEmail: serviceAccountEmail,
		audience: audience,
	}, nil
}

func (c *Client) CreateTask(ctx context.Context, queueID string, taskName string, url string, data proto.Message) error {
	v, err := protojson.Marshal(data)
	if err != nil {
		return xerrors.Errorf("protojson.Marshal failed: %w", err)
	}

	if _, err := c.client.CreateTask(ctx, &taskspb.CreateTaskRequest{
		Parent:       c.queuePath(queueID),
		Task: &taskspb.Task{
			Name: fmt.Sprintf("%s/tasks/%s", c.queuePath(queueID), taskName),
			// https://godoc.org/google.golang.org/genproto/googleapis/cloud/tasks/v2#HttpRequest
			MessageType: &taskspb.Task_HttpRequest{
				HttpRequest: &taskspb.HttpRequest{
					HttpMethod: taskspb.HttpMethod_POST,
					Url:  url,
					Body: v,
					AuthorizationHeader: &taskspb.HttpRequest_OidcToken{
						OidcToken: &taskspb.OidcToken{
							ServiceAccountEmail: c.serviceAccountEmail,
							Audience: c.audience,
						},
					},
				},
			},
		},
		ResponseView: 0,
	}); err != nil {
		return xerrors.Errorf("cloudtask.CreateTask failed: %w", err)
	}
	return nil
}