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