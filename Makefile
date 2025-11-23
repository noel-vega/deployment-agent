.PHONY: dev

dev:
	@ADMIN_USERNAME=admin \
	ADMIN_PASSWORD=devpass123 \
	JWT_ACCESS_SECRET=dev-access-secret \
	JWT_REFRESH_SECRET=dev-refresh-secret \
	air
