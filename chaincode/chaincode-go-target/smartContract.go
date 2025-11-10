/*
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"log"

	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
	order "github.com/hyperledger/fabric-samples/biologics-admin/chaincode-go/smart-contract"
)

func main() {
	orderSmartContract, err := contractapi.NewChaincode(&order.SmartContract{})
	if err != nil {
		log.Panicf("Error creating order chaincode: %v", err)
	}

	if err := orderSmartContract.Start(); err != nil {
		log.Panicf("Error starting order chaincode: %v", err)
	}
}
