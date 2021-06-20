gcloud config set run/region asia-northeast1
APP_IMAGE="asia.gcr.io/arg-vc/cloud-endpoints-espv2-pubsub-translate/cloud-endpoints-espv2-pubsub-translate@sha256:e3e9f752c3526e11e71ce39de65dc2e107fc520fc6ae5d7fb0bba5d95d439453"
GATEWAY_IMAGE="gcr.io/arg-vc/endpoints-runtime-serverless:2.28.0-pubsub-translate-gateway-g7ag5sxmgq-an.a.run.app-2021-06-19r1"

gcloud alpha run deploy pubsub-translate-app --image="$APP_IMAGE" --allow-unauthenticated --platform managed --project arg-vc --use-http2 --allow-unauthenticated 

gcloud alpha run deploy pubsub-translate-gateway --image="$GATEWAY_IMAGE" --allow-unauthenticated --platform managed --project arg-vc --use-http2 --allow-unauthenticated --update-env-vars "ESPv2_ARGS=^++^--http_request_timeout_s=300++--transcoding_always_print_primitive_fields++--transcoding_always_print_enums_as_ints++--tracing_sample_rate=1.0"

