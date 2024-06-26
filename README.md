# Kapsule WORK IN PROGRESS

[<img src="./images/kapsule_logo.png" width="250"/>](./images/kapsule_logo.png)

Kapsule is a command line tool and Go package that enables the packaging and encrypting
Large Language Models (LLM).  Models are defined using the 
[OCI image format](https://github.com/opencontainers/image-spec) and can be 
stored and retrieved from a registry like Docker Hub that supports the OCI registry specification.

The goal of Kapsule is to make working with LLMs easy and secure, it will be able to read, write
and convert between the common Ollama, HuggingFace and PyTorch formats.

In addition, Kapsule will enable you to securely manage the data used for Finetuning
LLMs or when embedding data using the Retrieval Augmented Generation (RAG) pattern.

It is the goal of Kapsule to be easy to provide a similar workflow that developers are
already used to.

## What does Kapsule NOT do?

It is not within the scope of Kapsule to convert models between their differing formats,
for example converting a Hugging Face model into a gguf model. Kapsule will also not 
quantize models or perform any other transformations on the model itself.

## Modelfile

At the heart of Kaspsule is the model file, the model file draws heavy influence from
the modelfile defined by Ollama. It is also familliar to developers who have been working
with Docker.

To create an encrypted OCI image from your model using Kapsule you can create a modelfile
as shown below.

```dockerfile
FROM ./dummy.gguf

TEMPLATE """[INST] {{ .System }} {{ .Prompt }} [/INST]"""

PARAMETER stop [/INST]
PARAMETER stop [INST]
PARAMETER temperature 1

SYSTEM You are brain from Pinky and the Brain, acting as an assitant.
```

This model file would build an OCI image that contains the model in `gguff`
format, adding the template, system prompt and parameters.

## Building images with Kapsule

To compose an image from the previous model and to push it to an OCI registry
the following command can be used. This pushes to the registry in plain format.

```bash
kapsule build \
	--debug \
	-f ./test_fixtures/testmodel/modelfile \
	-t docker.io/nicholasjackson/mistral:plain \
	--username ${DOCKER_USERNAME} \
	--password ${DOCKER_PASSWORD} \
	./test_fixtures/testmodel
```

To push an encrypted image to the registry, you can use the `--encrypt-key` flag
to specify the path to the RSA public key. Kapsule uses OCIEncrypt to encrypt the
layers of the image using asymetic encryption.

```bash
kapsule build \
	--debug \
	-f ./test_fixtures/testmodel/modelfile \
	-t docker.io/nicholasjackson/mistral:encrypted \
	--encryption-key ./test_fixtures/keys/public.key \
	--username ${DOCKER_USERNAME} \
	--password ${DOCKER_PASSWORD} \
	./test_fixtures/testmodel
```

### Full command list

```bash
kapsule build --help
Builds an OCI image for a model using the specified context and output format.

Usage:
  kapsule build [flags]

Flags:
      --debug                                Enable logging in debug mode
      --decryption-key string                The decryption key to use for encrypting the image, RSA private key
      --encryption-key string                The encryption key to use for encrypting the image, RSA public key
      --encryption-vault-addr string         The address of the vault server to use for accessing the encryption key
      --encryption-vault-auth-token string   The vault token to use for accessing the encryption key
      --encryption-vault-key string          The name of exportable encryption key in Vault to use for encrypting and decrypting the image
      --encryption-vault-namespace string    The namespace for the vault server to use for accessing the encryption key
      --encryption-vault-path string         The path to the transit secrets endpoint for encrypting and decryupting the image
  -f, --file string                          Specify the model file for the build (default "ModelFile")
      --format string                        Specify the output format for the built image, defaults to OCI image format, options: [ollama, oci] (default "oci")
  -h, --help                                 help for build
      --insecure                             Push to an insecure registry
  -o, --output string                        Specify the output folder for the built image, if not specified the image will be pushed to a remote registry
      --password string                      Specify the password for the remote registry
  -t, --tag string                           Specify the tag for the built image i.e. docker.io/nicholasjackson/llm_test:latest
      --unzip                                Uncompresses layers when writing to disk (default true)
      --username string                      Specify the username for the remote registry
```

## Pulling images with Kapsule

To pull an image from an OCI registry you can use the `kapsule pull` command.
the following command would downlaod the image and write it in OCI format to the
output directory.

```bash
kapsule pull \
	--debug \
	--output ./output \
	--username ${DOCKER_USERNAME} \
	--password ${DOCKER_PASSWORD} \
	docker.io/nicholasjackson/mistral:plain
```

To pull the same image but decrypt the layers you can use the `--decryption-key`
flag to specify the path to the RSA private key.

```bash
kapsule pull \
	--debug \
	--output ./output \
	--decryption-key ./test_fixtures/keys/private.key \
	--username ${DOCKER_USERNAME} \
	--password ${DOCKER_PASSWORD} \
	docker.io/nicholasjackson/mistral:encrypted
```

## Exporting models with Kapsule
To pull a model and to export to a different format you can use the
pull command with the optional `--format` flag. The following command
would pull the model and export it to the Ollama format.

```bash
kapsule pull \
	--debug \
	--output ./output \
	--format ollama \
	--username ${DOCKER_USERNAME} \
	--password ${DOCKER_PASSWORD} \
	docker.io/nicholasjackson/mistral:plain
```

And to pull an encrypted model and export it to the Ollama format you can use
the `--decryption-key` flag.

```bash
kapsule pull \
	--debug \
	--output ./output \
	--format ollama \
	--username ${DOCKER_USERNAME} \
	--password ${DOCKER_PASSWORD} \
	--decryption-key ./test_fixtures/keys/private.key \
	docker.io/nicholasjackson/mistral:encrypted
```

### Full command list

```bash
kapsule pull --help
Pull an OCI image from a remote registry

Usage:
  kapsule pull [flags]

Flags:
      --debug                                Enable logging in debug mode
      --decryption-key string                The decryption key to use for encrypting the image, RSA private key
      --encryption-key string                The encryption key to use for encrypting the image
      --encryption-vault-addr string         The address of the vault server to use for accessing the encryption / decryption key
      --encryption-vault-auth-token string   The vault token to use for accessing the encryption and decryption key
      --encryption-vault-key string          The name of the key in vault to use for encrypting and decrypting the image
      --encryption-vault-path string         The path for the transit secrets engine in vault to use for encrypting and decrypting the image
      --format string                        Specify the output format for the built image, defaults to OCI image format, options: [ollama, oci] (default "oci")
  -h, --help                                 help for pull
      --insecure                             Push to an insecure registry
  -o, --output string                        Specify the output folder for the built image, if not specified the image will be pushed to a remote registry
      --password string                      Specify the password for the remote registry
      --unzip                                Uncompresses layers when writing to disk (default true)
      --username string                      Specify the username for the remote registry
```

## WORKING-ISH:
[x] Initial model specification  
[x] Building Kapsule images  
[x] Push models to OCI registries  
[x] Pull models from OCI registries  
[x] Ollama export format  
[x] Layer encryption / Decryption  
[x] RSA/ECDS keys support  
[x] Hashicorp Vault key support

## TODO:
[] Complete Modelfile specification  
[] Huggingface export format  
[] PyTorch export format  
