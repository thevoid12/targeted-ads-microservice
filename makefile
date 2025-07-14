
include .env
export $(shell sed 's/=.*//' .env)

migration-create:
	@echo "**************************** migration create ***************************************"
	goose -dir migrations create targeted_ads sql
	@echo "******************************************************************************"

migrate-up:
	@echo "**************************** migration up ***************************************"
	@command="goose -dir migrations postgres \"host=${PG_HOST} port=${PG_PORT} user=${PG_USER} password=${PG_PASSWORD} dbname=${PG_DB} sslmode=${PG_SSLMODE}\" up"; \
	echo $$command; \
	result=$$(eval $$command); \
	echo "$$result"
	@echo "******************************************************************************"
migrate-down:
	@echo "**************************** migration down ***************************************"
	@command="goose -dir migrations postgres \"host=${PG_HOST} port=${PG_PORT} user=${PG_USER} password=${PG_PASSWORD} dbname=${PG_DB} sslmode=${PG_SSLMODE}\" down"; \
	echo $$command; \
	result=$$(eval $$command); \
	echo "$$result"
	@echo "******************************************************************************"

bootstrap:
	@echo "**************************** migration down ***************************************"
	chmod +x ./docker/bootstrap.sh
	./docker/bootstrap.sh; 
	@echo "******************************************************************************"