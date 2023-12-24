package client

import (
	"github.com/cosmos/cosmos-sdk/x/distribution/client/rest"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"

	"github.com/iqlusioninc/liquidity-staking-module/x/distribution/client/cli"
)

// ProposalHandler is the community spend proposal handler.
var (
	// ProposalHandler = govclient.NewProposalHandler(cli.GetCmdSubmitProposal)
	ProposalHandler = govclient.NewProposalHandler(cli.GetCmdSubmitProposal, rest.ProposalRESTHandler)
)
