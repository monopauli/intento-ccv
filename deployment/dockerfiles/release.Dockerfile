# Base image
FROM rust-go-base-image AS build-env-rust-go

# Final image
FROM enigmampc/enigma-sgx-base:2004-1.1.3

# wasmi-sgx-test script requirements
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    #### Base utilities ####
    jq \
    openssl \
    curl \
    wget \
    bash-completion && \
    rm -rf /var/lib/apt/lists/*

RUN echo "source /etc/profile.d/bash_completion.sh" >> ~/.bashrc

ARG SGX_MODE=SW
ENV SGX_MODE=${SGX_MODE}

ARG SECRET_NODE_TYPE=BOOTSTRAP
ENV SECRET_NODE_TYPE=${SECRET_NODE_TYPE}

ENV SCRT_ENCLAVE_DIR=/usr/lib/

# workaround because paths seem kind of messed up
RUN cp /opt/sgxsdk/lib64/libsgx_urts_sim.so /usr/lib/libsgx_urts_sim.so
RUN cp /opt/sgxsdk/lib64/libsgx_uae_service_sim.so /usr/lib/libsgx_uae_service_sim.so

# Install ca-certificates
WORKDIR /root

# Copy over binaries from the build-env
COPY --from=build-env-rust-go /go/src/github.com/enigmampc/SecretNetwork/go-cosmwasm/target/release/libgo_cosmwasm.so /usr/lib/
COPY --from=build-env-rust-go /go/src/github.com/enigmampc/SecretNetwork/go-cosmwasm/librust_cosmwasm_enclave.signed.so /usr/lib/
COPY --from=build-env-rust-go /go/src/github.com/enigmampc/SecretNetwork/tppd /usr/bin/tppd

COPY deployment/docker/bootstrap/bootstrap_init.sh .
COPY deployment/docker/node/node_init.sh .
COPY deployment/docker/startup.sh .
COPY deployment/docker/node_key.json .

RUN chmod +x /usr/bin/tppd
RUN chmod +x bootstrap_init.sh
RUN chmod +x startup.sh
RUN chmod +x node_init.sh

RUN tppd completion > /root/tppd_completion

RUN echo 'source /root/tppd_completion' >> ~/.bashrc

RUN mkdir -p /root/.tppd/.compute/
RUN mkdir -p /root/.sgx_secrets/
RUN mkdir -p /root/.tppd/.node/

COPY deployment/docker/bootstrap/config.toml /root/.tppd/config/config-cli.toml

COPY x/compute/internal/keeper/testdata/erc20.wasm /root/erc20.wasm
COPY deployment/docker/sanity-test.sh /root/

RUN chmod +x /root/sanity-test.sh

####### Node parameters
ARG MONIKER=default
ARG CHAINID=secret-testnet-1
ARG GENESISPATH=https://raw.githubusercontent.com/enigmampc/SecretNetwork/master/secret-testnet-genesis.json
ARG PERSISTENT_PEERS=201cff36d13c6352acfc4a373b60e83211cd3102@bootstrap.southuk.azure.com:26656

ENV GENESISPATH="${GENESISPATH}"
ENV CHAINID="${CHAINID}"
ENV MONIKER="${MONIKER}"
ENV PERSISTENT_PEERS="${PERSISTENT_PEERS}"

#ENV LD_LIBRARY_PATH=/opt/sgxsdk/libsgx-enclave-common/:/opt/sgxsdk/lib64/

# Run tppd by default, omit entrypoint to ease using container with tppd
ENTRYPOINT ["/bin/bash", "sanity-test.sh"]
