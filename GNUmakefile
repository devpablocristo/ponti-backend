include Makefile

.PHONY: select-ponti-stg-local select-ponti-dev-local up-ponti-stg-local up-ponti-dev-local

LOCAL_STG_ENV := .env.main.local
LOCAL_DEV_ENV := .env.develop.local

select-ponti-stg-local:
	@test -f $(LOCAL_STG_ENV) || (echo "ERROR: falta $(LOCAL_STG_ENV)" >&2; exit 1)
	@cp $(LOCAL_STG_ENV) .env
	@echo "Backend local env activo: $(LOCAL_STG_ENV)"

select-ponti-dev-local:
	@test -f $(LOCAL_DEV_ENV) || (echo "ERROR: falta $(LOCAL_DEV_ENV)" >&2; exit 1)
	@cp $(LOCAL_DEV_ENV) .env
	@echo "Backend local env activo: $(LOCAL_DEV_ENV)"

up-ponti-stg-local: select-ponti-stg-local up-ponti-local

up-ponti-dev-local: select-ponti-dev-local up-ponti-local
