A Go program that creates a local web server and runs an HTML page to communicate over WebSockets.

In order to connect to time series, the following environment variables must be set
The values included here are for Predix Basic.

TS_INGEST="wss://gateway-predix-data-services.run.aws-usw02-pr.ice.predix.io/v1/stream/messages"
TS_UAA_HOST="https://a8a2ffc4-b04e-4ec1-bfed-7a51dd408725.predix-uaa.run.aws-usw02-pr.ice.predix.io"
TS_UAA_TOKEN_URL="/oauth/token"

TS_UAA_CLIENT_ID=<your client id>

TS_UAA_CLIENT_SECRET=<your client secret>

TS_ZONE_ID=<your time series zone id>
