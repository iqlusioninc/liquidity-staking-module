package client

import (
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	"github.com/iqlusioninc/liquidity-staking-module/x/distribution/client/cli"
	"github.com/iqlusioninc/liquidity-staking-module/x/distribution/client/rest"
)

// ProposalHandler is the community spend proposal handler.
var (
	ProposalHandler = govclient.NewProposalHandler(cli.GetCmdSubmitProposal, rest.ProposalRESTHandler)
)
