package client

import (
	govclient "github.com/iqlusioninc/liquidity-staking-module/x/gov/client"
	"github.com/iqlusioninc/liquidity-staking-module/x/params/client/cli"
)

// ProposalHandler is the param change proposal handler.
var ProposalHandler = govclient.NewProposalHandler(cli.NewSubmitParamChangeProposalTxCmd)
