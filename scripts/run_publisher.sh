#!/usr/bin/env bash
set -euo pipefail

export MQTT_BROKER=${MQTT_BROKER:-tcp://localhost:1883}
export VEHICLE_ID=${VEHICLE_ID:-B1234XYZ}
export BASE_LAT=${BASE_LAT:--6.2088}
export BASE_LON=${BASE_LON:-106.8456}

go run ./cmd/publisher
