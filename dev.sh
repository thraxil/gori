#!/bin/bash
export GORI_DB_URL=postgres://pguser:foo@localhost/gori?sslmode=disable
export GORI_MEDIA_DIR=media/
export GORI_PORT=8890

./gori -config=dev.conf
