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
		--debug \
		--output ./output \
		-f ./test_fixtures/testmodel/modelfile \
		-t kapsule.io/nicholasjackson/mistral:tune \
		--format ollama \
		./test_fixtures/testmodel

test_build_local:
	go run ./cmd build \
		--debug \
		--output ./output \
		-f ./test_fixtures/testmodel/modelfile \
		-t kapsule.io/nicholasjackson/mistral:tune \
		./test_fixtures/testmodel

test_build_local_encrypted:
	go run ./cmd build \
		--debug \
		--output ./output \
		--encryption-key ./test_fixtures/keys/public.key \
		-f ./test_fixtures/testmodel/modelfile \
		-t kapsule.io/nicholasjackson/mistral:tune \
		./test_fixtures/testmodel

test_push_docker:
	go run ./cmd build \
		--debug \
		-f ./test_fixtures/testmodel/modelfile \
		-t docker.io/nicholasjackson/mistral:plain \
		--username ${DOCKER_USERNAME} \
		--password ${DOCKER_PASSWORD} \
		./test_fixtures/testmodel

test_push_docker_encrypted:
	go run ./cmd build \
		--debug \
		-f ./test_fixtures/testmodel/modelfile \
		-t auth.container.local.jmpd.in:5001/testmodel:enc \
		--encryption-key ./test_fixtures/keys/public.key \
		--username admin \
		--password password \
		--insecure \
		./test_fixtures/testmodel

test_pull_oci:
	go run ./cmd pull \
		--debug \
		--output ./output \
		--username ${DOCKER_USERNAME} \
		--password ${DOCKER_PASSWORD} \
		docker.io/nicholasjackson/mistral:plain

test_pull_oci_encrypted:
	go run ./cmd pull \
		--debug \
		--output ./output \
		--decryption-key ./test_fixtures/keys/private.key \
		--username ${DOCKER_USERNAME} \
		--password ${DOCKER_PASSWORD} \
		docker.io/nicholasjackson/mistral:encrypted

test_pull_ollama:
	go run ./cmd pull \
		--debug \
		--output ./output \
		--format ollama \
		--username ${DOCKER_USERNAME} \
		--password ${DOCKER_PASSWORD} \
		docker.io/nicholasjackson/mistral:plain

test_pull_ollama_encrypted:
	go run ./cmd pull \
		--debug \
		--output ./output \
		--format ollama \
		--username ${DOCKER_USERNAME} \
		--password ${DOCKER_PASSWORD} \
		--decryption-key ./test_fixtures/keys/private.key \
		docker.io/nicholasjackson/mistral:encrypted

test_run_acc:
	jumppad up ./jumppad
	TEST_ACC=1 go test -v -run "TestACC.*" ./... 
	jumppad down --force