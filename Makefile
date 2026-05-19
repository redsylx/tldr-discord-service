IMAGE      ?= ghcr.io/redsylx/tldr-discord-service
VERSION    ?= v0.1.0

.PHONY: build push deploy test vet clean

build:
	docker build -t $(IMAGE):$(VERSION) -t $(IMAGE):latest .

push:
	docker push $(IMAGE):$(VERSION)
	docker push $(IMAGE):latest

deploy:
	gcloud run deploy tldr-discord-service \
		--image $(IMAGE):$(VERSION) \
		--region asia-southeast1 \
		--max-instances 1 \
		--no-cpu-throttling \
		--set-env-vars "GCS_BUCKET_NAME=...,DISCORD_WEBHOOK_URL=..."

test:
	go test ./... -v -count=1

vet:
	go vet ./...

clean:
	rm -f tldr-discord-service
	docker rmi $(IMAGE):$(VERSION) 2>/dev/null || true