build_all:
	dagger -m ./dagger call all \
		--output=./output \
		--src=. \
		--github-token=GITHUB_TOKEN \
		--notorize-cert=${QUILL_SIGN_P12} \
		--notorize-cert-password=QUILL_SIGN_PASSWORD \
		--notorize-key=${QUILL_NOTARY_KEY} \
		--notorize-id=${QUILL_NOTARY_KEY_ID} \
		--notorize-issuer=${QUILL_NOTARY_ISSUER}

test_build_ollama:
	go run ./cmd build \
		--output ./output \
		-f ./test_fixtures/testmodel/modelfile \
		-t kapsule.io/nicholasjackson/mistral:tune \
		--format ollama \
		./test_fixtures/testmodel

test_push_docker:
	go run ./cmd build \
		--output ./output \
		-f ./test_fixtures/testmodel/modelfile \
		-t docker.io/nicholasjackson/mistral:tuned \
		--username ${DOCKER_USERNAME} \
		--password ${DOCKER_PASSWORD} \
		./test_fixtures/testmodel

test_pull_ollama:
	go run ./cmd pull \
		--output ./output \
		--format ollama \
		--username ${DOCKER_USERNAME} \
		--password ${DOCKER_PASSWORD} \
		docker.io/nicholasjackson/mistral:tuned