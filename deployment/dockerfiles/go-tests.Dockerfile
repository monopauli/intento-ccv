# Simple usage with a mounted data directory:
# > docker build -t enigma .
# > docker run -it -p 26657:26657 -p 26656:26656 -v ~/.trstd:/root/.trstd -v ~/.trstcli:/root/.trstcli enigma trstd init
# > docker run -it -p 26657:26657 -p 26656:26656 -v ~/.trstd:/root/.trstd -v ~/.trstcli:/root/.trstcli enigma trstd start
FROM rust-go-base-image

RUN cp /go/src/github.com/trstlabs/trst/cosmwasm/enclaves/execute/librust_cosmwasm_enclave.signed.so x/compute/internal/keeper
RUN mkdir -p /opt/trustlesshub/.sgx_secrets

RUN rustup target add wasm32-unknown-unknown

COPY scripts/install-wasm-tools.sh .
RUN chmod +x install-wasm-tools.sh
RUN ./install-wasm-tools.sh

RUN make build-test-contract

# workaround because paths seem kind of messed up
# RUN cp /opt/sgxsdk/lib64/libsgx_urts_sim.so /usr/lib/libsgx_urts_sim.so
# RUN cp /opt/sgxsdk/lib64/libsgx_uae_service_sim.so /usr/lib/libsgx_uae_service_sim.so
# RUN cp /go/src/github.com/trstlabs/trst/go-cosmwasm/target/release/libgo_cosmwasm.so /usr/lib/libgo_cosmwasm.so
# RUN cp /go/src/github.com/trstlabs/trst/go-cosmwasm/librust_cosmwasm_enclave.signed.so /usr/lib/librust_cosmwasm_enclave.signed.so

COPY deployment/ci/go-tests.sh .

RUN chmod +x go-tests.sh

ENTRYPOINT ["/bin/bash", "go-tests.sh"]
