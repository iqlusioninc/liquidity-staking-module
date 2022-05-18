package types

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func (r TokenizeShareRecord) GetModuleAddress() sdk.AccAddress {
	return authtypes.NewModuleAddress(r.ModuleAccount)
}

func (r TokenizeShareRecord) GetShareTokenDenom() string {
	return r.Validator + strconv.Itoa(int(r.Id))
}
