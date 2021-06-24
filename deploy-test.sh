gcloud config set run/region asia-northeast1
APP_IMAGE="asia.gcr.io/arg-vc/cloud-endpoints-espv2-pubsub-translate/cloud-endpoints-espv2-pubsub-translate@sha256:d5a1c7091abb2e9110969fde88dad5041c0f5f035ac920b682f0e942befdefcd"
GATEWAY_IMAGE="gcr.io/arg-vc/endpoints-runtime-serverless:2.28.0-pubsub-translate-gateway-g7ag5sxmgq-an.a.run.app-2021-06-24r2"

gcloud alpha run deploy pubsub-translate-app --image="$APP_IMAGE" --allow-unauthenticated --platform managed --project arg-vc --use-http2 --allow-unauthenticated 

gcloud alpha run deploy pubsub-translate-gateway --image="$GATEWAY_IMAGE" --allow-unauthenticated --platform managed --project arg-vc --use-http2 --allow-unauthenticated --update-env-vars "ESPv2_ARGS=^++^--http_request_timeout_s=300++--transcoding_always_print_primitive_fields++--transcoding_always_print_enums_as_ints++--tracing_sample_rate=1.0++--enable_debug"