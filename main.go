package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"

	"golang.org/x/oauth2/google"
	"golang.org/x/xerrors"

	"github.com/a-r-g-v/cloud-endpoints-espv2-pubsub-translate/grpc"
	"github.com/a-r-g-v/cloud-endpoints-espv2-pubsub-translate/pubsub"
)

func getProjectID(ctx context.Context) (string, error) {
	credentials, err := google.FindDefaultCredentials(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "google: could not find default credentials") {
			return "", nil
		}

		return "", xerrors.Errorf("google.FindDefaultCredentials: %w", err)
	}
	return credentials.ProjectID, nil
}

func main() {
	ctx := context.Background()
	projectID, err := getProjectID(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] getProjectID failed. err: %+v", err)
		os.Exit(1)
	}

	port := os.Getenv("PORT")
	if port == "" {
		fmt.Fprintf(os.Stderr, "[ERROR] failed to get PORT from enviroment variables.")
		os.Exit(1)
	}

	pc, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] pubsub.NewClient failed. err: %+v", err)
		os.Exit(1)
	}

	s, err := grpc.NewServer(pc)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] grpc.NewServer failed. err: %+v", err)
		os.Exit(1)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", port))
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] net.Listen failed. err: %+v", err)
		os.Exit(1)
	}
	defer lis.Close()

	if err := s.Serve(lis); err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Serve failed. err: %+v", err)
	}
}
