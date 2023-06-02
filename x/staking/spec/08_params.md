<!--
order: 8
-->

# Parameters

The staking module contains the following parameters:

| Key                       | Type             | Example                |
| ------------------------- | ---------------- | ---------------------- |
| UnbondingTime             | string (time ns) | "259200000000000"      |
| MaxValidators             | uint16           | 100                    |
| KeyMaxEntries             | uint16           | 7                      |
| HistoricalEntries         | uint16           | 3                      |
| BondDenom                 | string           | "stake"                |
| MinCommissionRate         | string           | "0.000000000000000000" |
| ValidatorBondFactor       | string           | "250.0000000000000000" |
| GlobalLiquidStakingCap    | string           | "0.250000000000000000" |
| ValidatorLiquidStakingCap | string           | "0.500000000000000000" |
| LiquidStakingCapsEnabled  | bool             | true                   |