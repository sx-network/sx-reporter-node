package abis

const OutcomeReporterJSONABI = `[
{
	"anonymous": false,
	"inputs": [{
			"indexed": false,
			"internalType": "address",
			"name": "previousAdmin",
			"type": "address"
		},
		{
			"indexed": false,
			"internalType": "address",
			"name": "newAdmin",
			"type": "address"
		}
	],
	"name": "AdminChanged",
	"type": "event"
},
{
	"anonymous": false,
	"inputs": [{
		"indexed": true,
		"internalType": "address",
		"name": "beacon",
		"type": "address"
	}],
	"name": "BeaconUpgraded",
	"type": "event"
},
{
	"anonymous": false,
	"inputs": [{
		"indexed": false,
		"internalType": "uint8",
		"name": "version",
		"type": "uint8"
	}],
	"name": "Initialized",
	"type": "event"
},
{
	"anonymous": false,
	"inputs": [{
			"indexed": false,
			"internalType": "bytes32",
			"name": "marketHash",
			"type": "bytes32"
		},
		{
			"indexed": false,
			"internalType": "enum LibOutcome.Outcome",
			"name": "outcome",
			"type": "uint8"
		}
	],
	"name": "OutcomeReported",
	"type": "event"
},
{
	"anonymous": false,
	"inputs": [{
			"indexed": false,
			"internalType": "bytes32",
			"name": "marketHash",
			"type": "bytes32"
		},
		{
			"indexed": false,
			"internalType": "address[]",
			"name": "",
			"type": "address[]"
		}
	],
	"name": "OutcomeVotingFinalized",
	"type": "event"
},
{
	"anonymous": false,
	"inputs": [{
			"indexed": true,
			"internalType": "address",
			"name": "previousOwner",
			"type": "address"
		},
		{
			"indexed": true,
			"internalType": "address",
			"name": "newOwner",
			"type": "address"
		}
	],
	"name": "OwnershipTransferred",
	"type": "event"
},
{
	"anonymous": false,
	"inputs": [{
			"indexed": false,
			"internalType": "bytes32",
			"name": "marketHash",
			"type": "bytes32"
		},
		{
			"indexed": false,
			"internalType": "enum LibOutcome.Outcome",
			"name": "outcome",
			"type": "uint8"
		},
		{
			"indexed": false,
			"internalType": "uint256",
			"name": "blockTime",
			"type": "uint256"
		}
	],
	"name": "ProposeOutcome",
	"type": "event"
},
{
	"anonymous": false,
	"inputs": [{
			"indexed": true,
			"internalType": "bytes32",
			"name": "role",
			"type": "bytes32"
		},
		{
			"indexed": true,
			"internalType": "bytes32",
			"name": "previousAdminRole",
			"type": "bytes32"
		},
		{
			"indexed": true,
			"internalType": "bytes32",
			"name": "newAdminRole",
			"type": "bytes32"
		}
	],
	"name": "RoleAdminChanged",
	"type": "event"
},
{
	"anonymous": false,
	"inputs": [{
			"indexed": true,
			"internalType": "bytes32",
			"name": "role",
			"type": "bytes32"
		},
		{
			"indexed": true,
			"internalType": "address",
			"name": "account",
			"type": "address"
		},
		{
			"indexed": true,
			"internalType": "address",
			"name": "sender",
			"type": "address"
		}
	],
	"name": "RoleGranted",
	"type": "event"
},
{
	"anonymous": false,
	"inputs": [{
			"indexed": true,
			"internalType": "bytes32",
			"name": "role",
			"type": "bytes32"
		},
		{
			"indexed": true,
			"internalType": "address",
			"name": "account",
			"type": "address"
		},
		{
			"indexed": true,
			"internalType": "address",
			"name": "sender",
			"type": "address"
		}
	],
	"name": "RoleRevoked",
	"type": "event"
},
{
	"anonymous": false,
	"inputs": [{
		"indexed": true,
		"internalType": "address",
		"name": "implementation",
		"type": "address"
	}],
	"name": "Upgraded",
	"type": "event"
},
{
	"anonymous": false,
	"inputs": [{
			"indexed": false,
			"internalType": "bytes32",
			"name": "marketHash",
			"type": "bytes32"
		},
		{
			"indexed": false,
			"internalType": "enum LibOutcome.Outcome",
			"name": "outcome",
			"type": "uint8"
		}
	],
	"name": "VoteOutcome",
	"type": "event"
},
{
	"inputs": [],
	"name": "DEFAULT_ADMIN_ROLE",
	"outputs": [{
		"internalType": "bytes32",
		"name": "",
		"type": "bytes32"
	}],
	"stateMutability": "view",
	"type": "function"
},
{
	"inputs": [],
	"name": "OUTCOME_EMERGENCY_REPORTER_ROLE",
	"outputs": [{
		"internalType": "bytes32",
		"name": "",
		"type": "bytes32"
	}],
	"stateMutability": "view",
	"type": "function"
},
{
	"inputs": [],
	"name": "_juicedReportingRewardAmount",
	"outputs": [{
		"internalType": "uint256",
		"name": "",
		"type": "uint256"
	}],
	"stateMutability": "view",
	"type": "function"
},
{
	"inputs": [],
	"name": "_totalReportedOutcomeCount",
	"outputs": [{
		"internalType": "uint256",
		"name": "",
		"type": "uint256"
	}],
	"stateMutability": "view",
	"type": "function"
},
{
	"inputs": [],
	"name": "_votingPeriod",
	"outputs": [{
		"internalType": "uint256",
		"name": "",
		"type": "uint256"
	}],
	"stateMutability": "view",
	"type": "function"
},
{
	"inputs": [{
			"internalType": "bytes32",
			"name": "marketHash",
			"type": "bytes32"
		},
		{
			"internalType": "address",
			"name": "validator",
			"type": "address"
		}
	],
	"name": "didValidatorVoteValid",
	"outputs": [{
		"internalType": "bool",
		"name": "",
		"type": "bool"
	}],
	"stateMutability": "view",
	"type": "function"
},
{
	"inputs": [{
			"internalType": "bytes32",
			"name": "marketHash",
			"type": "bytes32"
		},
		{
			"internalType": "enum LibOutcome.Outcome",
			"name": "outcome",
			"type": "uint8"
		}
	],
	"name": "emergencyReportOutcome",
	"outputs": [],
	"stateMutability": "nonpayable",
	"type": "function"
},
{
	"inputs": [{
		"internalType": "bytes32",
		"name": "marketHash",
		"type": "bytes32"
	}],
	"name": "getReportTime",
	"outputs": [{
		"internalType": "uint256",
		"name": "",
		"type": "uint256"
	}],
	"stateMutability": "view",
	"type": "function"
},
{
	"inputs": [{
		"internalType": "bytes32",
		"name": "marketHash",
		"type": "bytes32"
	}],
	"name": "getReportedOutcome",
	"outputs": [{
		"internalType": "enum LibOutcome.Outcome",
		"name": "",
		"type": "uint8"
	}],
	"stateMutability": "view",
	"type": "function"
},
{
	"inputs": [{
		"internalType": "address",
		"name": "validatorAddress",
		"type": "address"
	}],
	"name": "getReportedOutcomeStats",
	"outputs": [{
			"internalType": "uint256",
			"name": "",
			"type": "uint256"
		},
		{
			"internalType": "uint256",
			"name": "",
			"type": "uint256"
		},
		{
			"internalType": "uint256",
			"name": "",
			"type": "uint256"
		}
	],
	"stateMutability": "view",
	"type": "function"
},
{
	"inputs": [{
		"internalType": "bytes32",
		"name": "role",
		"type": "bytes32"
	}],
	"name": "getRoleAdmin",
	"outputs": [{
		"internalType": "bytes32",
		"name": "",
		"type": "bytes32"
	}],
	"stateMutability": "view",
	"type": "function"
},
{
	"inputs": [{
		"internalType": "bytes32",
		"name": "marketHash",
		"type": "bytes32"
	}],
	"name": "getValidVoteValidators",
	"outputs": [{
		"internalType": "address[]",
		"name": "",
		"type": "address[]"
	}],
	"stateMutability": "view",
	"type": "function"
},
{
	"inputs": [{
			"internalType": "bytes32",
			"name": "role",
			"type": "bytes32"
		},
		{
			"internalType": "address",
			"name": "account",
			"type": "address"
		}
	],
	"name": "grantRole",
	"outputs": [],
	"stateMutability": "nonpayable",
	"type": "function"
},
{
	"inputs": [{
			"internalType": "bytes32",
			"name": "role",
			"type": "bytes32"
		},
		{
			"internalType": "address",
			"name": "account",
			"type": "address"
		}
	],
	"name": "hasRole",
	"outputs": [{
		"internalType": "bool",
		"name": "",
		"type": "bool"
	}],
	"stateMutability": "view",
	"type": "function"
},
{
	"inputs": [{
			"internalType": "address",
			"name": "sxNode",
			"type": "address"
		},
		{
			"internalType": "address",
			"name": "wsx",
			"type": "address"
		}
	],
	"name": "initialize",
	"outputs": [],
	"stateMutability": "nonpayable",
	"type": "function"
},
{
	"inputs": [],
	"name": "owner",
	"outputs": [{
		"internalType": "address",
		"name": "",
		"type": "address"
	}],
	"stateMutability": "view",
	"type": "function"
},
{
	"inputs": [{
			"internalType": "address",
			"name": "sender",
			"type": "address"
		},
		{
			"internalType": "bytes32",
			"name": "marketHash",
			"type": "bytes32"
		},
		{
			"internalType": "enum LibOutcome.Outcome",
			"name": "outcome",
			"type": "uint8"
		}
	],
	"name": "proposeOutcome",
	"outputs": [],
	"stateMutability": "nonpayable",
	"type": "function"
},
{
	"inputs": [],
	"name": "proxiableUUID",
	"outputs": [{
		"internalType": "bytes32",
		"name": "",
		"type": "bytes32"
	}],
	"stateMutability": "view",
	"type": "function"
},
{
	"inputs": [],
	"name": "renounceOwnership",
	"outputs": [],
	"stateMutability": "nonpayable",
	"type": "function"
},
{
	"inputs": [{
			"internalType": "bytes32",
			"name": "role",
			"type": "bytes32"
		},
		{
			"internalType": "address",
			"name": "account",
			"type": "address"
		}
	],
	"name": "renounceRole",
	"outputs": [],
	"stateMutability": "nonpayable",
	"type": "function"
},
{
	"inputs": [{
		"internalType": "bytes32",
		"name": "marketHash",
		"type": "bytes32"
	}],
	"name": "reportOutcome",
	"outputs": [],
	"stateMutability": "nonpayable",
	"type": "function"
},
{
	"inputs": [{
			"internalType": "bytes32",
			"name": "role",
			"type": "bytes32"
		},
		{
			"internalType": "address",
			"name": "account",
			"type": "address"
		}
	],
	"name": "revokeRole",
	"outputs": [],
	"stateMutability": "nonpayable",
	"type": "function"
},
{
	"inputs": [{
		"internalType": "uint256",
		"name": "amount",
		"type": "uint256"
	}],
	"name": "setJuicedRewardAmount",
	"outputs": [],
	"stateMutability": "nonpayable",
	"type": "function"
},
{
	"inputs": [{
		"internalType": "address",
		"name": "sxNode",
		"type": "address"
	}],
	"name": "setSXNodeAddress",
	"outputs": [],
	"stateMutability": "nonpayable",
	"type": "function"
},
{
	"inputs": [{
		"internalType": "address",
		"name": "staking",
		"type": "address"
	}],
	"name": "setStakingAddress",
	"outputs": [],
	"stateMutability": "nonpayable",
	"type": "function"
},
{
	"inputs": [{
		"internalType": "uint256",
		"name": "duration",
		"type": "uint256"
	}],
	"name": "setVotingPeriod",
	"outputs": [],
	"stateMutability": "nonpayable",
	"type": "function"
},
{
	"inputs": [{
		"internalType": "bytes4",
		"name": "interfaceId",
		"type": "bytes4"
	}],
	"name": "supportsInterface",
	"outputs": [{
		"internalType": "bool",
		"name": "",
		"type": "bool"
	}],
	"stateMutability": "view",
	"type": "function"
},
{
	"inputs": [{
		"internalType": "address",
		"name": "newOwner",
		"type": "address"
	}],
	"name": "transferOwnership",
	"outputs": [],
	"stateMutability": "nonpayable",
	"type": "function"
},
{
	"inputs": [{
		"internalType": "address",
		"name": "newImplementation",
		"type": "address"
	}],
	"name": "upgradeTo",
	"outputs": [],
	"stateMutability": "nonpayable",
	"type": "function"
},
{
	"inputs": [{
			"internalType": "address",
			"name": "newImplementation",
			"type": "address"
		},
		{
			"internalType": "bytes",
			"name": "data",
			"type": "bytes"
		}
	],
	"name": "upgradeToAndCall",
	"outputs": [],
	"stateMutability": "payable",
	"type": "function"
},
{
	"inputs": [{
			"internalType": "address",
			"name": "sender",
			"type": "address"
		},
		{
			"internalType": "bytes32",
			"name": "marketHash",
			"type": "bytes32"
		},
		{
			"internalType": "enum LibOutcome.Outcome",
			"name": "outcome",
			"type": "uint8"
		}
	],
	"name": "voteOutcome",
	"outputs": [],
	"stateMutability": "nonpayable",
	"type": "function"
},
{
	"inputs": [{
		"internalType": "uint256",
		"name": "amount",
		"type": "uint256"
	}],
	"name": "withdrawJuicedRewards",
	"outputs": [],
	"stateMutability": "nonpayable",
	"type": "function"
}
]`
