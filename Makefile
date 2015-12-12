build:
	GOARCH="amd64" GOOS="linux" go install github.com/degiz/lingvobot
	cp ${GOPATH}/bin/linux_amd64/lingvobot .
	docker build -t degiz/lingvobot .
	rm lingvobot

run:
	-docker rm -f lingvobot
	docker run --restart=always -d --name lingvobot \
	-e TELEGRAM_BOT_TOKEN=${TELEGRAM_BOT_TOKEN} \
	-e IVONA_ACCESS_KEY=${IVONA_ACCESS_KEY} \
	-e IVONA_SECRET_KEY=${IVONA_SECRET_KEY} degiz/lingvobot

clean:
	-docker rm -f lingvobot

.PHONY: build run clean