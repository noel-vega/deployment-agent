.PHONY: dev

dev:
	@ADMIN_USERNAME=admin \
	ADMIN_PASSWORD=devpass123 \
	JWT_ACCESS_SECRET=dev-access-secret \
	JWT_REFRESH_SECRET=dev-refresh-secret \
  ENVIRONMENT=development \
	REGISTRY_URL=http://localhost:3000 \
	REGISTRY_URL=http://localhost:3000 \
  PROJECTS_ROOT_PATH=./tmp/projects \
	air
