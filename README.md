# Fleet Management (Transjakarta Backend Test)

This project implements a fleet management backend with:
- MQTT subscriber for vehicle location updates
- PostgreSQL storage
- Gin REST API for latest location and history
- RabbitMQ geofence events
- Docker Compose for full stack

## System Flow

How this system runs end-to-end:

1. `publisher` service generates mock vehicle location data every 2 seconds.
2. `publisher` sends the JSON payload to MQTT topic `/fleet/vehicle/{vehicle_id}/location` via Mosquitto.
3. `api` service subscribes to `/fleet/vehicle/+/location` and receives incoming location messages.
4. `api` validates payload format (`vehicle_id`, latitude/longitude range, and timestamp).
5. Valid data is inserted into PostgreSQL table `vehicle_locations`.
6. `api` checks geofence distance against configured center point and radius.
7. If vehicle is inside geofence radius, `api` publishes `geofence_entry` event to RabbitMQ exchange `fleet.events`.
8. `worker` consumes events from RabbitMQ queue `geofence_alerts` and logs the event message.
9. Client applications can query latest and historical locations through HTTP endpoints:
- `GET /vehicles/{vehicle_id}/location`
- `GET /vehicles/{vehicle_id}/history?start=...&end=...`

Request/processing flow summary:

```text
Mock Publisher -> MQTT (Mosquitto) -> API Subscriber -> PostgreSQL
                                          |
                                          v
                                  Geofence Check
                                          |
                                          v
                                  RabbitMQ Publisher -> Worker Consumer
```

## Services
- **API**: HTTP + MQTT subscriber + geofence publisher
- **Worker**: RabbitMQ consumer for `geofence_alerts`
- **Publisher**: Mock MQTT publisher (every 2 seconds)

## Quick Start (Docker)
1. Build and run all services with live logs (attached mode):

```bash
cd /media/user/New Volume/go/fleet-management
docker compose up --build
```

2. Alternative: run in background (detached mode), then follow logs:

```bash
docker compose up -d --build
docker compose logs -f
```

3. Check service status:

```bash
docker compose ps
```

4. API endpoints:
- `GET http://localhost:8080/vehicles/{vehicle_id}/location`
- `GET http://localhost:8080/vehicles/{vehicle_id}/history?start=...&end=...`

5. RabbitMQ management UI:
- `http://localhost:15672` (user/pass: `guest`/`guest`)

## API Documentation

Base URL:
- `http://localhost:8080`

### Health Check
Endpoint:
- `GET /health`

Example:
```bash
curl --location 'http://localhost:8080/health'
```

Success response (`200`):
```json
{
  "status": "ok"
}
```

### Get Latest Vehicle Location
Endpoint:
- `GET /vehicles/{vehicle_id}/location`

Path params:
- `vehicle_id` (string), example: `B1234XYZ`

Example:
```bash
curl --location 'http://localhost:8080/vehicles/B1234XYZ/location'
```

Success response (`200`):
```json
{
  "vehicle_id": "B1234XYZ",
  "latitude": -6.2088,
  "longitude": 106.8456,
  "timestamp": 1770945000
}
```

Error response (`404`):
```json
{
  "error": "location not found"
}
```

### Get Vehicle Location History
Endpoint:
- `GET /vehicles/{vehicle_id}/history?start={unix_start}&end={unix_end}`

Path params:
- `vehicle_id` (string), example: `B1234XYZ`

Query params:
- `start` (required, Unix timestamp in seconds)
- `end` (required, Unix timestamp in seconds)

Example:
```bash
now=$(date +%s)
curl --location "http://localhost:8080/vehicles/B1234XYZ/history?start=$((now-300))&end=$now"
```

Success response (`200`):
```json
{
  "vehicle_id": "B1234XYZ",
  "items": [
    {
      "vehicle_id": "B1234XYZ",
      "latitude": -6.20891,
      "longitude": 106.84573,
      "timestamp": 1770944800
    }
  ]
}
```

Error response (`400`) for invalid `start`:
```json
{
  "error": "invalid start"
}
```

Error response (`400`) for invalid `end`:
```json
{
  "error": "invalid end"
}
```

Error response (`500`) when database query fails:
```json
{
  "error": "failed to fetch history"
}
```

## Docker Behavior (Updated)
- `postgres` host port is mapped to `5433` (`5433:5432`).
- `postgres` and `rabbitmq` have health checks.
- `api` waits for `postgres` and `rabbitmq` to be healthy before starting.
- `worker` waits for `rabbitmq` to be healthy before starting.
- `api`, `worker`, and `publisher` use `restart: unless-stopped`.

This prevents startup race conditions where `api` or `worker` fails because RabbitMQ/PostgreSQL is not ready yet.

## Useful Commands

```bash
# Follow logs for all services
docker compose logs -f

# Follow logs for selected services
docker compose logs -f api worker publisher

# Restart specific services
docker compose up -d api worker

# Stop everything
docker compose down
```

## Local Run (without Docker)
- Start PostgreSQL, RabbitMQ, Mosquitto
- Export environment variables (see below)
- Run:

```bash
go run ./cmd/api
```

```bash
go run ./cmd/worker
```

```bash
go run ./cmd/publisher
```

## Environment Variables
API/Worker:
- `HTTP_PORT` (default `8080`)
- `POSTGRES_URL`
- `MQTT_BROKER`
- `MQTT_CLIENT_ID`
- `MQTT_USERNAME`
- `MQTT_PASSWORD`
- `RABBIT_URL`
- `GEOFENCE_LAT`
- `GEOFENCE_LON`
- `GEOFENCE_RADIUS_METERS`
- `RABBIT_EXCHANGE`
- `RABBIT_QUEUE`
- `RABBIT_ROUTING_KEY`

Publisher:
- `MQTT_BROKER`
- `VEHICLE_ID`
- `BASE_LAT`
- `BASE_LON`

## Postman Collection
The Postman collection is in:
- `docs/postman_collection.json`

## Notes
- The MQTT subscriber listens on `/fleet/vehicle/+/location`.
- Geofence event format follows the provided specification.

## Example Logs

Sample `worker` logs when geofence events are consumed:

```text
worker-1     | 2026/02/13 01:05:04 geofence event: {"vehicle_id":"B1234XYZ","event":"geofence_entry","location":{"latitude":-6.20849,"longitude":106.845891},"timestamp":1770944704}
worker-1     | 2026/02/13 01:05:34 geofence event: {"vehicle_id":"B1234XYZ","event":"geofence_entry","location":{"latitude":-6.208498,"longitude":106.845365},"timestamp":1770944734}
worker-1     | 2026/02/13 01:06:04 geofence event: {"vehicle_id":"B1234XYZ","event":"geofence_entry","location":{"latitude":-6.208763,"longitude":106.845552},"timestamp":1770944764}
```

