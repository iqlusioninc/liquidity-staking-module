module github.com/iqlusioninc/liquidity-staking-module

go 1.16

require (
	cosmossdk.io/api v0.1.0-alpha8 // indirect
	cosmossdk.io/math v1.0.0-beta.3 // indirect
	github.com/armon/go-metrics v0.4.1
	github.com/cosmos/cosmos-proto v1.0.0-alpha7 // indirect
	github.com/cosmos/cosmos-sdk v0.46.3
	github.com/cosmos/go-bip39 v1.0.0
	github.com/cosmos/gogoproto v1.4.2 // indirect
	github.com/creachadair/taskgroup v0.3.2 // indirect
	github.com/creachadair/tomledit v0.0.22 // indirect
	github.com/gogo/protobuf v1.3.3
	github.com/golang/mock v1.6.0
	github.com/golang/protobuf v1.5.2
	github.com/gorilla/handlers v1.5.1 // indirect
	github.com/gorilla/mux v1.8.0
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/hashicorp/go-getter v1.6.1 // indirect
	github.com/improbable-eng/grpc-web v0.15.0 // indirect
	github.com/lazyledger/smt v0.2.1-0.20210709230900-03ea40719554 // indirect
	github.com/mroth/weightedrand v0.4.1 // indirect
	github.com/oasisprotocol/curve25519-voi v0.0.0-20210609091139-0a56a4bca00b // indirect
	github.com/pkg/errors v0.9.1
	github.com/rakyll/statik v0.1.7
	github.com/regen-network/cosmos-proto v0.3.1
	github.com/rs/zerolog v1.27.0
	github.com/spf13/cast v1.5.0
	github.com/spf13/cobra v1.5.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.13.0
	github.com/stretchr/testify v1.8.0
	github.com/tendermint/tendermint v0.34.22
	github.com/tendermint/tm-db v0.6.7 // indirect
	github.com/zondax/hid v0.9.1-0.20220302062450-5552068d2266 // indirect
	golang.org/x/crypto v0.0.0-20220722155217-630584e8d5aa // indirect
	google.golang.org/genproto v0.0.0-20220815135757-37a418bb8959
	google.golang.org/grpc v1.50.0
	google.golang.org/protobuf v1.28.1
	nhooyr.io/websocket v1.8.6 // indirect
	sigs.k8s.io/yaml v1.3.0
)

replace (
	github.com/99designs/keyring => github.com/cosmos/keyring v1.1.7-0.20210622111912-ef00f8ac3d76
	github.com/gin-gonic/gin => github.com/gin-gonic/gin v1.7.0
	github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
)
