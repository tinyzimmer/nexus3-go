build:
	cd cmd && go get -d ./...
	cd cmd && go build -o ../bin/nexus-cmd .

clean:
	rm -rf bin/

start_nexus:
	cd docker && docker-compose up -d

stop_nexus:
	cd docker && docker-compose down

nexus_log_%:
	cd docker && docker-compose logs --tail=$*
