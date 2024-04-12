banner-run:
	go run ./cmd/banner-shift/main.go

postgres-run:
	sudo docker run --rm --name postgres -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=postgres -p 5432:5432 -d postgres
	sleep 5

postgres-stop: 
	@if [ $$(sudo docker ps -q -f name=postgres) ]; then sudo docker stop postgres; fi

postgres-remove:
	@if [ $$(sudo docker ps -q -f name=postgres) ]; then sudo docker remove postgres; fi

banner-clean:
	rm -f banner-shift

banner-build: postgres-stop postgres-remove banner-clean postgres-run banner-run