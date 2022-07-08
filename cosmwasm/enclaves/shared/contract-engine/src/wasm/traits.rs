use wasmi::{RuntimeValue, Trap};

/// These functions are imported to WASM code
pub trait WasmiApi {

    fn read_db(&mut self, state_key_ptr: i32) -> Result<Option<RuntimeValue>, Trap>;


    fn remove_db(&mut self, state_key_ptr: i32) -> Result<Option<RuntimeValue>, Trap>;


    fn write_db(
        &mut self,
        state_key_ptr: i32,
        value_ptr: i32,
    ) -> Result<Option<RuntimeValue>, Trap>;

    fn query_chain(&mut self, query_ptr: i32) -> Result<Option<RuntimeValue>, Trap>;


    fn addr_canonicalize(
        &mut self,
        canonical_ptr: i32,
        human_ptr: i32,
    ) -> Result<Option<RuntimeValue>, Trap>;


    fn addr_humanize(
        &mut self,
        canonical_ptr: i32,
        human_ptr: i32,
    ) -> Result<Option<RuntimeValue>, Trap>;


    fn addr_validate(&mut self, source_ptr: i32) -> Result<Option<RuntimeValue>, Trap>;


    fn secp256k1_verify(
        &mut self,
        message_hash_ptr: i32,
        signature_ptr: i32,
        public_key_ptr: i32,
    ) -> Result<Option<RuntimeValue>, Trap>;


    fn secp256k1_recover_pubkey(
        &mut self,
        message_hash_ptr: i32,
        signature_ptr: i32,
        recovery_param: i32,
    ) -> Result<Option<RuntimeValue>, Trap>;


    fn ed25519_verify(
        &mut self,
        message_ptr: i32,
        signature_ptr: i32,
        public_key_ptr: i32,
    ) -> Result<Option<RuntimeValue>, Trap>;


    fn ed25519_batch_verify(
        &mut self,
        messages_ptr: i32,
        signatures_ptr: i32,
        public_keys_ptr: i32,
    ) -> Result<Option<RuntimeValue>, Trap>;

    /// internal
    fn gas(&mut self, gas_amount: i32) -> Result<Option<RuntimeValue>, Trap>;

    fn debug_print_index(&self, message: i32) -> Result<Option<RuntimeValue>, Trap>;


    fn secp256k1_sign(
        &mut self,
        message_ptr: i32,
        private_key_ptr: i32,
    ) -> Result<Option<RuntimeValue>, Trap>;


    fn ed25519_sign(
        &mut self,
        message_ptr: i32,
        private_key_ptr: i32,
    ) -> Result<Option<RuntimeValue>, Trap>;
}
