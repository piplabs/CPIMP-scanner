GET /addresses/{address_hash}

200
```
{
  "creator_address_hash": "0xEb533ee5687044E622C69c58B1B12329F56eD9ad",
  "creation_transaction_hash": "0x1f610ff9c1efad6b5a8bb6afcc0786cd7343f03f9a61e2544fcff908cedee924",
  "token": {
    "circulating_market_cap": "83606435600.3635",
    "icon_url": "https://raw.githubusercontent.com/trustwallet/assets/master/blockchains/ethereum/assets/0xdAC17F958D2ee523a2206206994597C13D831ec7/logo.png",
    "name": "Tether USD",
    "decimals": "6",
    "symbol": "USDT",
    "address": "0x394c399dbA25B99Ab7708EdB505d755B3aa29997",
    "type": "ERC-20",
    "holders": "837494234523",
    "exchange_rate": "0.99",
    "total_supply": "10000000"
  },
  "coin_balance": "10000000",
  "exchange_rate": "1.01",
  "implementation_address": "0xEb533ee5687044E622C69c58B1B12329F56eD9ad",
  "block_number_balance_updated_at": 27656552,
  "hash": "0xEb533ee5687044E622C69c58B1B12329F56eD9ad",
  "implementation_name": "implementationName",
  "name": "contractName",
  "is_contract": true,
  "private_tags": [
    {
      "address_hash": "0xEb533ee5687044E622C69c58B1B12329F56eD9ad",
      "display_name": "name to show",
      "label": "label"
    }
  ],
  "watchlist_names": [
    {
      "display_name": "name to show",
      "label": "label"
    }
  ],
  "public_tags": [
    {
      "address_hash": "0xEb533ee5687044E622C69c58B1B12329F56eD9ad",
      "display_name": "name to show",
      "label": "label"
    }
  ],
  "is_verified": true,
  "has_beacon_chain_withdrawals": true,
  "has_logs": true,
  "has_token_transfers": true,
  "has_tokens": true,
  "has_validated_blocks": true
}
```

GET /transactions/{transaction_hash}

200
```
{
  "timestamp": "2022-08-02T07:18:05.000000Z",
  "fee": {
    "type": "maximum | actual",
    "value": "9853224000000000"
  },
  "gas_limit": 0,
  "block_number": 23484035,
  "status": "ok | error",
  "method": "transferFrom",
  "confirmations": 1035,
  "type": 2,
  "exchange_rate": "1866.51",
  "to": {
    "hash": "0xEb533ee5687044E622C69c58B1B12329F56eD9ad",
    "implementation_name": "implementationName",
    "name": "contractName",
    "ens_domain_name": "domain.eth",
    "metadata": {
      "slug": "tag_slug",
      "name": "Tag name",
      "tagType": "name",
      "ordinal": 0,
      "meta": {}
    },
    "is_contract": true,
    "private_tags": [
      {
        "address_hash": "0xEb533ee5687044E622C69c58B1B12329F56eD9ad",
        "display_name": "name to show",
        "label": "label"
      }
    ],
    "watchlist_names": [
      {
        "display_name": "name to show",
        "label": "label"
      }
    ],
    "public_tags": [
      {
        "address_hash": "0xEb533ee5687044E622C69c58B1B12329F56eD9ad",
        "display_name": "name to show",
        "label": "label"
      }
    ],
    "is_verified": true
  },
  "transaction_burnt_fee": "1099596081903840",
  "max_fee_per_gas": "55357460102",
  "result": "Error: (Awaiting internal transactions for reason)",
  "hash": "0x5d90a9da2b8da402b11bc92c8011ec8a62a2d59da5c7ac4ae0f73ec51bb73368",
  "gas_price": "26668595172",
  "priority_fee": "2056916056308",
  "base_fee_per_gas": "26618801760",
  "from": {
    "hash": "0xEb533ee5687044E622C69c58B1B12329F56eD9ad",
    "implementation_name": "implementationName",
    "name": "contractName",
    "ens_domain_name": "domain.eth",
    "metadata": {
      "slug": "tag_slug",
      "name": "Tag name",
      "tagType": "name",
      "ordinal": 0,
      "meta": {}
    },
    "is_contract": true,
    "private_tags": [
      {
        "address_hash": "0xEb533ee5687044E622C69c58B1B12329F56eD9ad",
        "display_name": "name to show",
        "label": "label"
      }
    ],
    "watchlist_names": [
      {
        "display_name": "name to show",
        "label": "label"
      }
    ],
    "public_tags": [
      {
        "address_hash": "0xEb533ee5687044E622C69c58B1B12329F56eD9ad",
        "display_name": "name to show",
        "label": "label"
      }
    ],
    "is_verified": true
  },
  "token_transfers": [
    {
      "block_hash": "0xf569ec751152b2f814001fc730f7797aa155e4bc3ba9cb6ba24bc2c8c9468c1a",
      "from": {
        "hash": "0xEb533ee5687044E622C69c58B1B12329F56eD9ad",
        "implementation_name": "implementationName",
        "name": "contractName",
        "ens_domain_name": "domain.eth",
        "metadata": {
          "slug": "tag_slug",
          "name": "Tag name",
          "tagType": "name",
          "ordinal": 0,
          "meta": {}
        },
        "is_contract": true,
        "private_tags": [
          {
            "address_hash": "0xEb533ee5687044E622C69c58B1B12329F56eD9ad",
            "display_name": "name to show",
            "label": "label"
          }
        ],
        "watchlist_names": [
          {
            "display_name": "name to show",
            "label": "label"
          }
        ],
        "public_tags": [
          {
            "address_hash": "0xEb533ee5687044E622C69c58B1B12329F56eD9ad",
            "display_name": "name to show",
            "label": "label"
          }
        ],
        "is_verified": true
      },
      "log_index": 16,
      "method": "transfer",
      "timestamp": "2023-07-03T20:09:59.000000Z",
      "to": {
        "hash": "0xEb533ee5687044E622C69c58B1B12329F56eD9ad",
        "implementation_name": "implementationName",
        "name": "contractName",
        "ens_domain_name": "domain.eth",
        "metadata": {
          "slug": "tag_slug",
          "name": "Tag name",
          "tagType": "name",
          "ordinal": 0,
          "meta": {}
        },
        "is_contract": true,
        "private_tags": [
          {
            "address_hash": "0xEb533ee5687044E622C69c58B1B12329F56eD9ad",
            "display_name": "name to show",
            "label": "label"
          }
        ],
        "watchlist_names": [
          {
            "display_name": "name to show",
            "label": "label"
          }
        ],
        "public_tags": [
          {
            "address_hash": "0xEb533ee5687044E622C69c58B1B12329F56eD9ad",
            "display_name": "name to show",
            "label": "label"
          }
        ],
        "is_verified": true
      },
      "token": {
        "circulating_market_cap": "83606435600.3635",
        "icon_url": "https://raw.githubusercontent.com/trustwallet/assets/master/blockchains/ethereum/assets/0xdAC17F958D2ee523a2206206994597C13D831ec7/logo.png",
        "name": "Tether USD",
        "decimals": "6",
        "symbol": "USDT",
        "address": "0x394c399dbA25B99Ab7708EdB505d755B3aa29997",
        "type": "ERC-20",
        "holders": "837494234523",
        "exchange_rate": "0.99",
        "total_supply": "10000000"
      },
      "total": {
        "decimals": "18",
        "value": "1000"
      },
      "transaction_hash": "0x6662ad1ad2ea899e9e27832dc202fd2ef915a5d2816c1142e6933cff93f7c592",
      "type": "token_transfer"
    }
  ],
  "transaction_types": [
    "token_transfer",
    "contract_creation",
    "contract_call",
    "token_creation",
    "coin_transfer"
  ],
  "gas_used": "41309",
  "created_contract": {
    "hash": "0xEb533ee5687044E622C69c58B1B12329F56eD9ad",
    "implementation_name": "implementationName",
    "name": "contractName",
    "ens_domain_name": "domain.eth",
    "metadata": {
      "slug": "tag_slug",
      "name": "Tag name",
      "tagType": "name",
      "ordinal": 0,
      "meta": {}
    },
    "is_contract": true,
    "private_tags": [
      {
        "address_hash": "0xEb533ee5687044E622C69c58B1B12329F56eD9ad",
        "display_name": "name to show",
        "label": "label"
      }
    ],
    "watchlist_names": [
      {
        "display_name": "name to show",
        "label": "label"
      }
    ],
    "public_tags": [
      {
        "address_hash": "0xEb533ee5687044E622C69c58B1B12329F56eD9ad",
        "display_name": "name to show",
        "label": "label"
      }
    ],
    "is_verified": true
  },
  "position": 117,
  "nonce": 115,
  "has_error_in_internal_transactions": false,
  "actions": [
    {
      "data": {
        "debt_amount": "1.289548595490270429",
        "debt_symbol": "AAVE",
        "debt_address": "0x7Fc66500c84A76Ad7e9c93437bFc5Ac33E2DDaE9",
        "collateral_amount": "110.824768",
        "collateral_symbol": "USDC",
        "collateral_address": "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48",
        "block_number": 1
      },
      "protocol": "aave_v3",
      "type": "liquidation_call"
    },
    {
      "data": {
        "amount": "1.289548595490270429",
        "symbol": "AAVE",
        "address": "0x7Fc66500c84A76Ad7e9c93437bFc5Ac33E2DDaE9",
        "block_number": 1
      },
      "protocol": "aave_v3",
      "type": "borrow | supply | withdraw | repay | flash_loan"
    },
    {
      "data": {
        "symbol": "AAVE",
        "address": "0x7Fc66500c84A76Ad7e9c93437bFc5Ac33E2DDaE9",
        "block_number": 1
      },
      "protocol": "aave_v3",
      "type": "enable_collateral | disable_collateral"
    },
    {
      "data": {
        "name": "Uniswap V3: Positions NFT",
        "symbol": "UNI-V3-POS",
        "address": "0x1F98431c8aD98523631AE4a59f267346ea31F984",
        "to": "0x7Fc66500c84A76Ad7e9c93437bFc5Ac33E2DDaE9",
        "ids": [
          "1",
          "2"
        ],
        "block_number": 1
      },
      "protocol": "uniswap_v3",
      "type": "mint_nft"
    },
    {
      "data": {
        "address0": "0x7Fc66500c84A76Ad7e9c93437bFc5Ac33E2DDaE9",
        "address1": "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48",
        "amount0": "1.289548595490270429",
        "amount1": "110.824768",
        "symbol0": "AAVE",
        "symbol1": "USDC"
      },
      "protocol": "uniswap_v3",
      "type": "burn | collect | swap"
    }
  ],
  "decoded_input": {
    "method_call": "transferFrom(address _from, address _to, uint256 _value)",
    "method_id": "23b872dd",
    "parameters": [
      {
        "name": "signature",
        "type": "bytes",
        "value": "0x0"
      }
    ]
  },
  "token_transfers_overflow": false,
  "raw_input": "0xa9059cbb000000000000000000000000ef8801eaf234ff82801821ffe2d78d60a0237f97000000000000000000000000000000000000000000000000000000003178cb80",
  "value": "0",
  "max_priority_fee_per_gas": "49793412",
  "revert_reason": "Error: (Awaiting internal transactions for reason)",
  "confirmation_duration": [
    0,
    17479
  ],
  "transaction_tag": "private_transaction_tag"
}
```