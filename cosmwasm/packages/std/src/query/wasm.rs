use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

use crate::Binary;

#[non_exhaustive]
#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub enum WasmQuery {
    /// this queries the public API of another contract at a known address (with known ABI)
    /// return value is whatever the contract returns (caller should know)
    Private {
        contract_addr: String,
        /// code_hash is the hex encoded hash of the code. This is used by trst to harden against replaying the contract
        /// It is used to bind the request to a destination contract in a stronger way than just the contract address which can be faked
        code_hash: String,
        /// msg is the json-encoded QueryMsg struct
        msg: Binary,
    },
    /// this queries the raw kv-store of the contract.
    /// returns the raw, unparsed data stored at that key (or `Ok(Err(StdError:NotFound{}))` if missing)
    Public {
        contract_addr: String,
        /// code_hash is the hex encoded hash of the code. This is used by trst to harden against replaying the contract
        /// It is used to bind the request to a destination contract in a stronger way than just the contract address which can be faked
        //code_hash: String,
        /// Key is the key used in the public contract's Storage
        key: String,
    },
     /// this queries the raw kv-store of the contract.
    /// returns the raw, unparsed data stored at that key (or `Ok(Err(StdError:NotFound{}))` if missing)
    PublicForAddr {
        contract_addr: String,
        account_addr: String,
        /// Key is the key used in the public contract's Storage
        key: String,
    },
    /// returns a ContractInfoResponse with metadata on the contract from the runtime
    ContractInfo { contract_addr: String },
}


#[non_exhaustive]
#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct ContractInfoResponse {
    pub code_id: u64,
    /// address that instantiated this contract
    pub creator: String,
    /// if set, the contract is pinned to the cache, and thus uses less gas when called
    pub pinned: bool,
    /// set if this contract has bound an IBC port
    pub ibc_port: Option<String>,
}

impl ContractInfoResponse {
    /// Convenience constructor for tests / mocks
    #[doc(hidden)]
    pub fn new(code_id: u64, creator: impl Into<String>) -> Self {
        Self {
            code_id,
            creator: creator.into(),
            pinned: false,
            ibc_port: None,
        }
    }
}