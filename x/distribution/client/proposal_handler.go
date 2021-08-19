package client

import (
	"github.com/iqlusioninc/liquidity-staking-module/x/distribution/client/cli"
	govclient "github.com/iqlusioninc/liquidity-staking-module/x/gov/client"
)

// ProposalHandler is the community spend proposal handler.
var (
	ProposalHandler = govclient.NewProposalHandler(cli.GetCmdSubmitProposal)
)
