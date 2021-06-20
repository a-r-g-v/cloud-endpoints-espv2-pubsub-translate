package pubsub

import (
	"context"

	"cloud.google.com/go/pubsub"
	"golang.org/x/xerrors"
)

type Client struct {
	client *pubsub.Client
}

func NewClient(ctx context.Context, projectID string) (*Client, error) {
	pc, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, xerrors.Errorf("pubsub.NewClient: %w", err)
	}

	return &Client{
		client: pc,
	}, nil
}

func (c *Client) Publish(ctx context.Context, topic string, data []byte) error {
	t := c.client.Topic(topic)
	defer t.Stop()

	pr := t.Publish(ctx, &pubsub.Message{Data: data})
	_, err := pr.Get(ctx)
	return err
}
