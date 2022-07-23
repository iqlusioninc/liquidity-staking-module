rm -rf ~/.simapp

set -o errexit -o nounset

liquidstakingd init --chain-id test test
liquidstakingd keys add validator --keyring-backend="test"
liquidstakingd add-genesis-account $(liquidstakingd keys show validator -a --keyring-backend="test") 100000000000000stake
liquidstakingd gentx validator 100000000stake --keyring-backend="test" --chain-id test
liquidstakingd collect-gentxs

liquidstakingd start
# liquidstakingd start --home=home --mode=validator
