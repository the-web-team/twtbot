start:
	docker-compose up -d

start-mongo:
	docker-compose up -d mongo

stop:
	docker-compose stop

restart:
	docker-compose restart

logs:
	docker-compose logs -ft --tail 500 app