# Webhook Demonstration

- A WebService which exposes a Web Hook at `/log` for accumulating log entries and dumping them to a third service in the background.

## Steps to Run

- Build the docker image using `docker build -t batch-logger`
- Run the docker image using `docker run --env-file .env batch-logger`
- Do not forget to provide a file of environment variables or environment variables via command line

### Environment Variables
- BATCH_SIZE: amount after which sync to POST_ENDPOINT should be triggered
- BATCH_INTERVAL: time in seconds after which sync to POST_ENDPOINT should be triggered
- POST_ENDPOINT: http endpoint where POST request should be made to dump log entries
## Endpoints

- `/log`: post log entries here
- `/healthz`: check health of the application

## Sync

- should send a request with all the accumulated log entries as an array to $POST_ENDPOINT
- log a message containing the `batch size`, `result status code`, `duration of request` after a POST request
- on Failure, `retry 3 times` with a `2 second wait` and `log and exit after 3 failures`
- on Success, clear the in-memory entries

## When to sync

- When total entries received is equal to $BATCH_SIZE - in the endpoint
- When $BATCH_INTERVAL time has passed - in the scheduled job
