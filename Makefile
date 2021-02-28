ci:
	cd backend && go test ./app

docker:
	docker build .
