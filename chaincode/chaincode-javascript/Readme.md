#!/bin/bash
# imports
. envVar.sh

CHANNEL_NAME=treatmentorderchannel
ORG_NAME=hospital
CC_RUNTIME_LANGUAGE="node"
VERSION="9"
SEQUENCE=9
CC_SRC_PATH="../artifacts/chaincode/javascript"
CC_NAME=ordertracking

presetup() {
    echo Installing npm packages ...
    pushd ../artifacts/chaincode/javascript
    npm install
    popd
    echo Finished installing npm dependencies
}

packageChaincode() {
    rm -rf ${CC_NAME}.tar.gz
    setGlobals ${ORG_NAME} 1
    peer lifecycle chaincode package ${CC_NAME}.tar.gz \
    --path ${CC_SRC_PATH} --lang ${CC_RUNTIME_LANGUAGE} \
    --label ${CC_NAME}_${VERSION}
    echo "===================== Chaincode is packaged version: ${CC_NAME}_${VERSION} ===================== "
}

installChaincode() {
    setGlobals ${ORG_NAME} 1
    peer lifecycle chaincode install ${CC_NAME}.tar.gz

    setGlobals ${ORG_NAME} 2
    peer lifecycle chaincode install ${CC_NAME}.tar.gz
    
    # setGlobals manufacturer 1
    # peer lifecycle chaincode install ${CC_NAME}.tar.gz
    
    # setGlobals manufacturer 2
    # peer lifecycle chaincode install ${CC_NAME}.tar.gz
}

queryInstalled() {
    setGlobals ${ORG_NAME} 1
    peer lifecycle chaincode queryinstalled >&log.txt
    cat log.txt
    
    PACKAGE_ID=$(sed -n "s/^Package ID: \(.*\), Label: ${CC_NAME}_${VERSION}$/\1/p" log.txt | head -n 1)
    
    echo "PackageID is ${PACKAGE_ID}"
    echo "===================== Query installed successful on peer0.org1 on channel ===================== "
    setGlobals ${ORG_NAME} 2
    peer lifecycle chaincode queryinstalled >&log.txt
    cat log.txt
    
    PACKAGE_ID=$(sed -n "s/^Package ID: \(.*\), Label: ${CC_NAME}_${VERSION}$/\1/p" log.txt | head -n 1)
    
    echo "PackageID is ${PACKAGE_ID}"
    echo "===================== Query installed successful on peer0.org1 on channel ===================== "
}


approveForhospital() {
    setGlobals ${ORG_NAME} 1
    set -x
    peer lifecycle chaincode approveformyorg -o localhost:7050 \
    --ordererTLSHostnameOverride orderer1.com --tls \
    --cafile $ORDERER_CA --channelID ${CHANNEL_NAME} \
    --name ${CC_NAME} --version ${VERSION} \
    --package-id ${PACKAGE_ID} \
    --sequence ${SEQUENCE}
    set +x
    
    echo "===================== chaincode approved from Hospital ===================== "
    
}


approveFormanufacturer() {
    setGlobals manufacturer 1
    set -x
    peer lifecycle chaincode approveformyorg -o 10.20.10.188:7050 \
    --ordererTLSHostnameOverride orderer1.com --tls \
    --cafile $ORDERER_CA --channelID ${CHANNEL_NAME} \
    --name ${CC_NAME} --version ${VERSION} \
    --package-id ${PACKAGE_ID} \
    --sequence ${SEQUENCE}
    set +x
    
    echo "===================== chaincode approved from Manufacturer ===================== "
    
}


checkCommitReadyness() {
    setGlobals ${ORG_NAME} 1
    peer lifecycle chaincode checkcommitreadiness \
    --channelID ${CHANNEL_NAME} --name ${CC_NAME} --version ${VERSION} \
    --sequence ${SEQUENCE} --output json
    echo "===================== checking commit readyness from org 1 ===================== "
}


commitChaincodeDefination() {
    setGlobals ${ORG_NAME} 1
    peer lifecycle chaincode commit -o localhost:7050 --ordererTLSHostnameOverride orderer1.com \
    --tls $CORE_PEER_TLS_ENABLED --cafile $ORDERER_CA \
    --channelID ${CHANNEL_NAME} --name ${CC_NAME} \
    --version ${VERSION} --sequence ${SEQUENCE} \
    --peerAddresses localhost:7051 --tlsRootCertFiles $hospital_CA \
    --peerAddresses localhost:8051 --tlsRootCertFiles $hospital_CA \
    --peerAddresses peer1.manufacturer.com:9051 --tlsRootCertFiles $manufacturer_CA \
    --peerAddresses peer2.manufacturer.com:10051 --tlsRootCertFiles $manufacturer_CA \
    
}


chaincodeInvoke() {
    setGlobals ${ORG_NAME} 1

    # Create ORDER
    peer chaincode invoke -o localhost:7050 \
    --ordererTLSHostnameOverride orderer1.com \
    --tls $CORE_PEER_TLS_ENABLED \
    --cafile $ORDERER_CA \
    -C ${CHANNEL_NAME} -n ${CC_NAME}  \
    -c '{"function":"createOrder","Args":["ORDER-123","CAR-T","manufacturer-1","hospital-123","logistics-1","slot-789","user-1","2025-08-06T12:00:00Z","therapy_confirmed","2025-08-06T12:00:00Z"]}' \
    --peerAddresses localhost:7051 --tlsRootCertFiles $hospital_CA \
    --peerAddresses localhost:8051 --tlsRootCertFiles $hospital_CA \
    --peerAddresses peer1.manufacturer.com:9051 --tlsRootCertFiles $manufacturer_CA \
    --peerAddresses peer2.manufacturer.com:10051 --tlsRootCertFiles $manufacturer_CA \
}

chaincodeInvokeManufacturer() {
    setGlobals manufacturer 1
    # Create ORDER
    peer chaincode invoke -o 172.32.3.132:7050 \
    --ordererTLSHostnameOverride orderer1.com \
    --tls $CORE_PEER_TLS_ENABLED \
    --cafile $ORDERER_CA \
    -C ${CHANNEL_NAME} \
    -n ${CC_NAME} \
    -c '{
      "function": "createOrder",
      "Args": [
        "ORDER-1234",
        "CAR-T",
        "manufacturer-1",
        "hospital-123",
        "logistics-1",
        "slot-789",
        "user-1",
        "2025-08-06T12:00:00Z",
        "therapy_confirmed",
        "2025-08-06T12:00:00Z"
      ]
    }' \
    --peerAddresses peer1.hospital.com:7051 --tlsRootCertFiles $hospital_CA \
    --peerAddresses peer2.hospital.com:8051 --tlsRootCertFiles $hospital_CA \
    --peerAddresses localhost:9051 --tlsRootCertFiles $manufacturer_CA \
    --peerAddresses localhost:10051 --tlsRootCertFiles $manufacturer_CA
}


chaincodeQuery() {
    setGlobals ${ORG_NAME} 1
    peer chaincode query -C ${CHANNEL_NAME} -n ${CC_NAME} -c '{"function": "getAllOrdersWithPagination","Args":["10", ""]}'
}

#presetup
#packageChaincode
#installChaincode
#queryInstalled
#approveForhospital
#approveFormanufacturer
#checkCommitReadyness
#commitChaincodeDefination
#chaincodeInvoke
#chaincodeInvokeManufacturer
#sleep 3
#chaincodeQuery


To Update chaincode 

---- Run the following commands in hospital ----
1. Update ORG_NAME = hospital, VERSION = +1, SEQUENCE = +1
2. Enable functions 
presetup
packageChaincode
installChaincode
queryInstalled
approveForhospital

3. Repeat the same for manufacturer with ORG_NAME = manufacturer and step 2 as well
But in step 2 uncomment approveFormanufacturer
presetup
packageChaincode
installChaincode
queryInstalled
approveFormanufacturer

4. Un-comment checkCommitReadyness and comment other function and run script in hospital
5. Un-comment checkCommitReadyness and comment other function and run script in manufacturer
6. Un-comment commitChaincodeDefination and run script in hospital
7. Un-comment chaincodeInvoke & chaincodeQuery and run script in hospital
8. IF wanted to check same in manufacturer then un-comment chaincodeInvokeManufacturer and run script in manufacturer