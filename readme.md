## 背景

Pub/Sub の 同期 Pull Subscription を使うよりは、 Push Subscription を 使ったほうが以下の観点で好ましい。
* コンピューティングリソースの無駄になる。同期 Pull Subscription は 常時起動するワーカー  が必要となり、メッセージ処理の必要がない場合も Listen する必要があることから。
* APIコーダーの生産性向上及びメンテナンスコストの削減。API を実装するだけで Subscription ハンドラを実装することができるため、Pub/Sub の場合と 通常リクエストを区別する必要がなくなることから。デプロイメントも通常の方法で行うことができるから。

では、Cloud Pub/Sub の Push Subscription を Cloud Endpoints  (ESPv2) + gRPC 構成でどうやって使うのがベストか？検討した。

## Pub/Sub Push Subscription のリクエストペイロード

https://cloud.google.com/pubsub/docs/push/?hl=ja によると、Pub/Sub の Push Subscription は 以下のような payload でリクエストを行う。

```json
{
    "message": {
        "attributes": {
            "key": "value"
        },
        "data": "SGVsbG8gQ2xvdWQgUHViL1N1YiEgSGVyZSBpcyBteSBtZXNzYWdlIQ==",
        "messageId": "2070443601311540",
        "message_id": "2070443601311540",
        "publishTime": "2021-02-26T19:13:55.749Z",
        "publish_time": "2021-02-26T19:13:55.749Z",
    },
   "subscription": "projects/myproject/subscriptions/mysubscription"
}
```

message fields の正式な定義は、https://github.com/googleapis/googleapis/blob/master/google/pubsub/v1/pubsub.proto#L209 にある。

つまり、以下のような protobuf 定義を使えば Envoy の gRPC Transcode ができる。

```proto
message PubSubRequest {
  google.pubsub.v1.PubsubMessage message = 1;
  string subscription = 2;
}

// TestAPI
service TestAPI {
  rpc HandleTestMessageRPC(PubSubRequest) returns (google.protobuf.Empty) {
    option (google.api.http) = {
      post : "/v1/handle_test_message"
      body: "*"
    };
}
```

## gRPC ハンドラ

proto.Unmarshal する必要はあるが、普通の RPC っぽくハンドリングできる
https://cloud.google.com/pubsub/docs/push/?hl=ja によると、`102`, `200`, `201`, `202`, `204` を返す場合は ACK 扱いなので、 gRPC の作法でエラーハンドリングもすればいい。

```go
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
```

## Push 認証

Pub/Sub Push Subscription は オプション設定により JWT を付与して リクエストしてくれるようになる。Subscription Handler は JWT 検証によって リクエスタの真正性確認をすることができる。https://cloud.google.com/pubsub/docs/push/?hl=ja#authentication_and_authorization

JWT クレームは以下のようになるらしい。
```json
{
   "aud":"https://example.com",
   "azp":"113774264463038321964",
   "email":"gae-gcp@appspot.gserviceaccount.com",
   "sub":"113774264463038321964",
   "email_verified":true,
   "exp":1550185935,
   "iat":1550182335,
   "iss":"https://accounts.google.com"
  }
```

aud は自分で設定できる
<img width="1049" alt="スクリーンショット 2021-06-20 9.07.28.png (108.7 kB)" src="https://img.esa.io/uploads/production/attachments/17724/2021/06/20/2924/c84e13cf-000f-404f-bbef-03b64e9f23e5.png">

踏まえて、以下のような service.yaml を作成し ESPv2 の食わせると gateway で JWT検証もできる

```yaml
authentication:
  providers:
    - id: pubsub
      jwks_uri: https://www.googleapis.com/oauth2/v3/certs
      issuer: https://accounts.google.com
      audiences: pubsub-translate-gateway-g7ag5sxmgq-an.a.run.app

  rules:
    - selector: "test.TestAPI.HandleTestMessageRPC"
      requirements:
        - provider_id: pubsub

```

