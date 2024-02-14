LAMBDA_BUILD_DIR=dist/functions/
GITHUB_USER=pevin

build:
	for d in handlers/* ; do \
		LAMBDA_DIR=$${d#*/} ;\
		echo $(LAMBDA_BUILD_DIR)$${LAMBDA_DIR} ; \
		GOOS=linux GOARCH=amd64 go build -o $(LAMBDA_BUILD_DIR)$${LAMBDA_DIR} handlers/$${LAMBDA_DIR}/main.go ; \
	done

deploy: build deploy-only

deploy-only:
	bin/deploy.sh

clean:
	rm -rf dist

TEST_DIRS="./handlers/create_post/..."
test:
	go test $(TEST_DIRS)
