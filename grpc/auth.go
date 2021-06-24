package grpc

import (
	"context"
	"encoding/base64"
	"encoding/json"

	"golang.org/x/xerrors"
	"google.golang.org/grpc/metadata"
)

const JWTMetadataName = "x-endpoint-api-userinfo"

type GoogleIDToken struct {
	Aud           string `json:"aud"`
	Azp           string `json:"azp"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Exp           int    `json:"exp"`
	Iat           int    `json:"iat"`
	Iss           string `json:"iss"`
	Sub           string `json:"sub"`
}

func decodeBase64URLEncoding(enc string) ([]byte, error) {
	// see: https://github.com/GoogleCloudPlatform/esp-v2/issues/178
	payload, err := base64.URLEncoding.DecodeString(enc)
	if err == nil {
		return payload, nil
	}

	return base64.RawURLEncoding.DecodeString(enc)
}

func AuthenticateEmail(ctx context.Context, email string) error {
	requesterEmail, err := Email(ctx)
	if err != nil {
		return err
	}
	if requesterEmail == email {
		return nil
	}
	return xerrors.New("mismatch email")
}

func Email(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", xerrors.New("not exists gRPC metadata in context")
	}
	base64EncodedJWTPayloads := md.Get(JWTMetadataName)

	if len(base64EncodedJWTPayloads) != 1 {
		return "", nil
	}

	payload, err := decodeBase64URLEncoding(base64EncodedJWTPayloads[0])
	if err != nil {
		return "", xerrors.Errorf("failed to base64 url decode given JWT: %w", err)
	}

	var idToken GoogleIDToken
	if err := json.Unmarshal(payload, &idToken); err != nil {
		return "", xerrors.Errorf("failed to decode token payload as json: %w", err)
	}

	if !idToken.EmailVerified {
		return "", xerrors.New("idToken not verified")
	}

	return idToken.Email, nil
}

