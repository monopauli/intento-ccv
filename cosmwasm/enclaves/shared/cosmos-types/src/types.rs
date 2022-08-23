use log::*;

use enclave_ffi_types::EnclaveError;
use protobuf::Message;
use serde::{Deserialize, Serialize};

use crate::multisig::MultisigThresholdPubKey;

use enclave_crypto::{
    hash::sha::HASH_SIZE, secp256k1::Secp256k1PubKey, sha_256, traits::VerifyingKey, CryptoError,
};

use cosmos_proto as proto;

use enclave_cosmwasm_types::{
    //todo change to HumanAddr to Addr
    addresses::{CanonicalAddr, HumanAddr},
    coins::Coin,
    encoding::Binary,
    math::Uint128,
};

use crate::traits::CosmosAminoPubkey;

pub fn calc_contract_hash(contract_bytes: &[u8]) -> [u8; HASH_SIZE] {
    sha_256(&contract_bytes)
}

pub struct ContractCode<'code> {
    code: &'code [u8],
    hash: [u8; HASH_SIZE],
}

impl<'code> ContractCode<'code> {
    pub fn new(code: &'code [u8]) -> Self {
        let hash = calc_contract_hash(code);
        Self { code, hash }
    }

    pub fn code(&self) -> &[u8] {
        self.code
    }

    pub fn hash(&self) -> [u8; HASH_SIZE] {
        self.hash
    }
}

#[derive(PartialEq, Clone, Debug)]
pub enum CosmosPubKey {
    Secp256k1(Secp256k1PubKey),
    Multisig(MultisigThresholdPubKey),
}

/// `"/"` + `proto::crypto::multisig::LegacyAminoPubKey::descriptor_static().full_name()`
const TYPE_URL_MULTISIG_LEGACY_AMINO_PUBKEY: &str = "/cosmos.crypto.multisig.LegacyAminoPubKey";
/// `"/"` + `proto::crypto::secp256k1::PubKey::descriptor_static().full_name()`
const TYPE_URL_SECP256K1_PUBKEY: &str = "/cosmos.crypto.secp256k1.PubKey";

impl CosmosPubKey {
    pub fn from_proto(public_key: &protobuf::well_known_types::Any) -> Result<Self, CryptoError> {
        let public_key_parser = match public_key.type_url.as_str() {
            TYPE_URL_SECP256K1_PUBKEY => Self::secp256k1_from_proto,
            TYPE_URL_MULTISIG_LEGACY_AMINO_PUBKEY => Self::multisig_legacy_amino_from_proto,
            _ => {
                warn!("found public key of unsupported type: {:?}", public_key);
                return Err(CryptoError::ParsingError);
            }
        };

        public_key_parser(&public_key.value)
    }

    fn secp256k1_from_proto(public_key_bytes: &[u8]) -> Result<Self, CryptoError> {
        use proto::crypto::secp256k1::PubKey;
        let pub_key = PubKey::parse_from_bytes(public_key_bytes).map_err(|_err| {
            warn!(
                "Could not parse secp256k1 public key from these bytes: {}",
                Binary(public_key_bytes.to_vec())
            );
            CryptoError::ParsingError
        })?;
        Ok(CosmosPubKey::Secp256k1(Secp256k1PubKey::new(pub_key.key)))
    }

    fn multisig_legacy_amino_from_proto(public_key_bytes: &[u8]) -> Result<Self, CryptoError> {
        use proto::crypto::multisig::LegacyAminoPubKey;
        let multisig_key =
            LegacyAminoPubKey::parse_from_bytes(public_key_bytes).map_err(|_err| {
                warn!(
                    "Could not parse multisig public key from these bytes: {}",
                    Binary(public_key_bytes.to_vec())
                );
                CryptoError::ParsingError
            })?;
        let mut pubkeys = vec![];
        for public_key in &multisig_key.public_keys {
            pubkeys.push(CosmosPubKey::from_proto(public_key)?);
        }
        Ok(CosmosPubKey::Multisig(MultisigThresholdPubKey::new(
            multisig_key.threshold,
            pubkeys,
        )))
    }
}

impl CosmosAminoPubkey for CosmosPubKey {
    fn get_address(&self) -> CanonicalAddr {
        match self {
            CosmosPubKey::Secp256k1(pubkey) => pubkey.get_address(),
            CosmosPubKey::Multisig(pubkey) => pubkey.get_address(),
        }
    }

    fn amino_bytes(&self) -> Vec<u8> {
        match self {
            CosmosPubKey::Secp256k1(pubkey) => pubkey.amino_bytes(),
            CosmosPubKey::Multisig(pubkey) => pubkey.amino_bytes(),
        }
    }
}

impl VerifyingKey for CosmosPubKey {
    fn verify_bytes(&self, bytes: &[u8], sig: &[u8]) -> Result<(), CryptoError> {
        match self {
            CosmosPubKey::Secp256k1(pubkey) => pubkey.verify_bytes(bytes, sig),
            CosmosPubKey::Multisig(pubkey) => pubkey.verify_bytes(bytes, sig),
        }
    }
}

// Info of the msg to be signed
#[derive(Deserialize, Clone, Debug, PartialEq)]
pub struct MsgInfo {
    pub code_hash: Binary,
    pub funds: Vec<Coin>,
}

// This type is a copy of the `proto::tx::signing::SignMode` allowing us
// to create a Deserialize impl for it without touching the autogenerated type.
// See: https://serde.rs/remote-derive.html
#[allow(non_camel_case_types)]
#[derive(Deserialize)]
#[serde(remote = "proto::tx::signing::SignMode")]
pub enum SignModeDef {
    SIGN_MODE_UNSPECIFIED = 0,
    SIGN_MODE_DIRECT = 1,
    SIGN_MODE_TEXTUAL = 2,
    SIGN_MODE_LEGACY_AMINO_JSON = 127,
}

#[allow(non_camel_case_types)]
#[derive(Deserialize, Clone, Debug, PartialEq)]
pub enum HandleType {
    HANDLE_TYPE_EXECUTE = 0,
    HANDLE_TYPE_REPLY = 1,
}

impl HandleType {
    pub fn try_from(value: u8) -> Result<Self, EnclaveError> {
        match value {
            0 => Ok(HandleType::HANDLE_TYPE_EXECUTE),
            1 => Ok(HandleType::HANDLE_TYPE_REPLY),
            _ => {
                error!("unrecognized handle type: {}", value);
                Err(EnclaveError::FailedToDeserialize)
            }
        }
    }
}

// This is called `VerificationInfo` on the Go side
#[derive(Deserialize, Clone, Debug, PartialEq)]
pub struct SigInfo {
    pub sign_bytes: Binary,
    #[serde(with = "SignModeDef")]
    pub sign_mode: proto::tx::signing::SignMode,
    pub mode_info: Binary,
    pub public_key: Binary,
    pub signature: Binary,
    pub callback_sig: Option<Binary>,
}

// Should be in sync with https://github.com/cosmos/cosmos-sdk/blob/v0.38.3/x/auth/types/stdtx.go#L216
#[derive(Deserialize, Clone, Default, Debug, PartialEq)]
pub struct StdSignDoc {
    pub account_number: String,
    pub chain_id: String,
    pub memo: String,
    pub msgs: Vec<StdCosmWasmMsg>,
    pub sequence: String,
}

#[derive(Debug)]
pub struct SignDoc {
    pub body: TxBody,
    pub auth_info: AuthInfo,
    pub chain_id: String,
    pub account_number: u64,
}

impl SignDoc {
    pub fn from_bytes(bytes: &[u8]) -> Result<Self, EnclaveError> {
        let raw_sign_doc = proto::tx::tx::SignDoc::parse_from_bytes(bytes).map_err(|err| {
            warn!(
                "got an error while trying to deserialize sign doc bytes from protobuf: {}: {}",
                err,
                Binary(bytes.into()),
            );
            EnclaveError::FailedToDeserialize
        })?;

        let body = TxBody::from_bytes(&raw_sign_doc.body_bytes)?;
        let auth_info = AuthInfo::from_bytes(&raw_sign_doc.auth_info_bytes)?;

        Ok(Self {
            body,
            auth_info,
            chain_id: raw_sign_doc.chain_id,
            account_number: raw_sign_doc.account_number,
        })
    }
}

#[derive(Debug)]
pub struct TxBody {
    pub messages: Vec<CosmWasmMsg>,
    // Leaving this here for discoverability. We can use this, but don't verify it today.
    memo: (),
    timeout_height: (),
}

impl TxBody {
    pub fn from_bytes(bytes: &[u8]) -> Result<Self, EnclaveError> {
        let tx_body = proto::tx::tx::TxBody::parse_from_bytes(bytes).map_err(|err| {
            warn!(
                "got an error while trying to deserialize cosmos message body bytes from protobuf: {}: {}",
                err,
                Binary(bytes.into()),
            );
            EnclaveError::FailedToDeserialize
        })?;

        let messages = tx_body
            .messages
            .into_iter()
            .map(|any| CosmWasmMsg::from_bytes(&any.value))
            .collect::<Result<Vec<_>, _>>()?;

        Ok(TxBody {
            messages,
            memo: (),
            timeout_height: (),
        })
    }
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq)]
#[serde(rename_all = "snake_case", tag = "type", content = "value")]
pub enum StdCosmWasmMsg {
    #[serde(alias = "wasm/MsgExecuteContract")]
    Execute {
        sender: HumanAddr,
        contract: HumanAddr,
        /// msg is the json-encoded ExecuteMsg struct (as raw Binary)
        msg: String,
        funds: Vec<Coin>,
        callback_sig: Option<Vec<u8>>,
    },
    #[serde(alias = "wasm/MsgInstantiateContract")]
    Instantiate {
        sender: HumanAddr,
        code_id: String,
        msg: String,
        auto_msg: String,
        funds: Vec<Coin>,
        contract_id: String,
        duration: String,
        interval: String,
        start_duration_at: u64,
        callback_sig: Option<Vec<u8>>,
    },
}

impl StdCosmWasmMsg {
    pub fn into_cosmwasm_msg(self) -> Result<CosmWasmMsg, EnclaveError> {
        match self {
            Self::Execute {
                sender,
                contract,
                msg,
                funds,
                callback_sig,
            } => {
                let sender = CanonicalAddr::from_human(&sender).map_err(|err| {
                    warn!("failed to turn human addr to canonical addr when parsing CosmWasmMsg: {:?}", err);
                    EnclaveError::FailedToDeserialize
                })?;
                let msg = Binary::from_base64(&msg).map_err(|err| {
                    warn!(
                        "failed to parse base64 msg when parsing CosmWasmMsg: {:?}",
                        err
                    );
                    EnclaveError::FailedToDeserialize
                })?;
                let msg = msg.0;
                Ok(CosmWasmMsg::Execute {
                    sender,
                    contract,
                    msg,
                    funds,
                    callback_sig,
                })
            }
            Self::Instantiate {
                sender,
                msg,
                auto_msg,
                funds,
                contract_id,
                duration,
                interval,
                start_duration_at,
                callback_sig,
                code_id: _,
            } => {
                let sender = CanonicalAddr::from_human(&sender).map_err(|err| {
                    warn!("failed to turn human addr to canonical addr when parsing CosmWasmMsg: {:?}", err);
                    EnclaveError::FailedToDeserialize
                })?;
                let msg = Binary::from_base64(&msg).map_err(|err| {
                    warn!(
                        "failed to parse base64 msg when parsing CosmWasmMsg: {:?}",
                        err
                    );
                    EnclaveError::FailedToDeserialize
                })?;
                let auto_msg = Binary::from_base64(&auto_msg).map_err(|err| {
                    warn!(
                        "failed to parse base64 auto_msg when parsing last CosmWasmMsg: {:?}",
                        err
                    );
                    EnclaveError::FailedToDeserialize
                })?;
                let msg = msg.0;
                let auto_msg = auto_msg.0;
                Ok(CosmWasmMsg::Instantiate {
                    sender,
                    msg,
                    auto_msg,
                    funds,
                    contract_id,
                    duration,
                    interval,
                    start_duration_at,
                    callback_sig,
                })
            }
        }
    }
}

#[derive(Debug)]
pub enum CosmWasmMsg {
    Execute {
        sender: CanonicalAddr,
        contract: HumanAddr,
        msg: Vec<u8>,
        funds: Vec<Coin>,
        callback_sig: Option<Vec<u8>>,
    },
    Instantiate {
        sender: CanonicalAddr,
        msg: Vec<u8>,
        auto_msg: Vec<u8>,
        funds: Vec<Coin>,
        contract_id: String,
        duration: String,
        interval: String,
        start_duration_at: u64,
        callback_sig: Option<Vec<u8>>,
    },

    Other,
}

impl CosmWasmMsg {
    pub fn from_bytes(bytes: &[u8]) -> Result<Self, EnclaveError> {
        Self::try_parse_execute(bytes)
            .or_else(|_| Self::try_parse_instantiate(bytes))
            .or_else(|_| {
                warn!(
                    "got an error while trying to deserialize cosmwasm message bytes from protobuf: {}",
                    Binary(bytes.into())
                );
                Ok(CosmWasmMsg::Other)
            })
    }

    fn try_parse_instantiate(bytes: &[u8]) -> Result<Self, EnclaveError> {
        use proto::cosmwasm::msg::MsgInstantiateContract;

        let raw_msg = MsgInstantiateContract::parse_from_bytes(bytes)
            .map_err(|_| EnclaveError::FailedToDeserialize)?;

        trace!(
            "try_parse_instantiate sender: len={} val={:?}",
            raw_msg.sender.len(),
            raw_msg.sender
        );
        let sender = CanonicalAddr::from_human(&HumanAddr(raw_msg.sender)).map_err(|err| {
            warn!("failed to turn human addr to canonical addr when try_parse_instantiate CosmWasmMsg: {:?}", err);
            EnclaveError::FailedToDeserialize
        })?;

        let funds = Self::parse_funds(raw_msg.funds)?;

        let callback_sig = Some(raw_msg.callback_sig);

        Ok(CosmWasmMsg::Instantiate {
            sender, //: CanonicalAddr(Binary(raw_msg.sender)),
            msg: raw_msg.msg,
            auto_msg: raw_msg.auto_msg,
            funds,
            contract_id: raw_msg.contract_id,
            duration: raw_msg.duration,
            interval: raw_msg.interval,
            start_duration_at: raw_msg.start_duration_at,
            callback_sig,
        })
    }

    fn try_parse_execute(bytes: &[u8]) -> Result<Self, EnclaveError> {
        use proto::cosmwasm::msg::MsgExecuteContract;

        let raw_msg = MsgExecuteContract::parse_from_bytes(bytes)
            .map_err(|_| EnclaveError::FailedToDeserialize)?;

        trace!(
            "try_parse_execute sender: len={} val={:?}",
            raw_msg.sender.len(),
            raw_msg.sender
        );
        trace!(
            "try_parse_execute contract: len={} val={:?}",
            raw_msg.contract.len(),
            raw_msg.contract
        );

        let sender = CanonicalAddr::from_human(&HumanAddr(raw_msg.sender)).map_err(|err| {
            warn!("failed to turn human addr to canonical addr when try_parse_execute CosmWasmMsg: {:?}", err);
            EnclaveError::FailedToDeserialize
        })?;
     

        if raw_msg.contract.clone().len() == 0  {
              warn!(
                  "Contract address was empty: {}",
                  raw_msg.contract.len(),
              );
        };
       
        let funds = Self::parse_funds(raw_msg.funds)?;

        let callback_sig = Some(raw_msg.callback_sig);

        Ok(CosmWasmMsg::Execute {
            sender,
            contract: HumanAddr(raw_msg.contract),
            msg: raw_msg.msg,
            funds,
            callback_sig,
        })
    }

    fn parse_funds(
        raw_funds: protobuf::RepeatedField<proto::base::coin::Coin>,
    ) -> Result<Vec<Coin>, EnclaveError> {
        let mut funds = Vec::with_capacity(raw_funds.len());
        for raw_coin in raw_funds {
            let amount: u128 = raw_coin.amount.parse().map_err(|_err| {
                warn!(
                    "instantiate message funds were not a numeric string: {:?}",
                    raw_coin.amount,
                );
                EnclaveError::FailedToDeserialize
            })?;
            let coin = Coin {
                amount: Uint128::new(amount),
                denom: raw_coin.denom,
            };
            funds.push(coin);
        }

        Ok(funds)
    }

    pub fn sender(&self) -> Option<&CanonicalAddr> {
        match self {
            CosmWasmMsg::Execute { sender, .. } | CosmWasmMsg::Instantiate { sender, .. } => {
                Some(sender)
            }
            CosmWasmMsg::Other => None,
        }
    }
}

#[derive(Debug)]
pub struct AuthInfo {
    pub signer_infos: Vec<SignerInfo>,
    // Leaving this here for discoverability. We can use this, but don't verify it today.
    fee: (),
}

impl AuthInfo {
    pub fn from_bytes(bytes: &[u8]) -> Result<Self, EnclaveError> {
        let raw_auth_info = proto::tx::tx::AuthInfo::parse_from_bytes(&bytes).map_err(|err| {
            warn!("Could not parse AuthInfo from protobuf bytes: {:?}", err);
            EnclaveError::FailedToDeserialize
        })?;

        let mut signer_infos = vec![];
        for raw_signer_info in raw_auth_info.signer_infos {
            let signer_info = SignerInfo::from_proto(raw_signer_info)?;
            signer_infos.push(signer_info);
        }

        if signer_infos.is_empty() {
            warn!("No signature information provided for this TX. signer_infos empty");
            return Err(EnclaveError::FailedToDeserialize);
        }

        Ok(Self {
            signer_infos,
            fee: (),
        })
    }

    pub fn sender_public_key(&self, sender: &CanonicalAddr) -> Option<&CosmosPubKey> {
        self.signer_infos
            .iter()
            .find(|signer_info| &signer_info.public_key.get_address() == sender)
            .map(|si| &si.public_key)
    }
}

#[derive(Debug)]
pub struct SignerInfo {
    pub public_key: CosmosPubKey,
    pub sequence: u64,
}

impl SignerInfo {
    pub fn from_proto(raw_signer_info: proto::tx::tx::SignerInfo) -> Result<Self, EnclaveError> {
        if !raw_signer_info.has_public_key() {
            warn!("One of the provided signers had no public key");
            return Err(EnclaveError::FailedToDeserialize);
        }

        // unwraps valid after checks above
        let any_public_key = raw_signer_info.public_key.get_ref();

        let public_key = CosmosPubKey::from_proto(any_public_key)
            .map_err(|_| EnclaveError::FailedToDeserialize)?;

        let signer_info = Self {
            public_key,
            sequence: raw_signer_info.sequence,
        };
        Ok(signer_info)
    }
}
