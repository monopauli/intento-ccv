package types

import (
	errorsmod "cosmossdk.io/errors"
)

const (
	// ModuleName defines the module name
	ModuleName = "alloc"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// Fairburn pool name
	FairburnPoolName = "fairburn_pool"

	SupplementPoolName = "supplement_pool"

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName
)

// ParamsKey stores the module params
var (
	ParamsKey = []byte{0x01}
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}

// x/alloc module sentinel errors
var (
	ErrUnauthorized = errorsmod.Register(ModuleName, 2, "sender is unauthorized to perform the operation")
)
