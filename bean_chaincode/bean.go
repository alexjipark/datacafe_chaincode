/*
Copyright IBM Corp. 2016 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"fmt"
	"github.com/op/go-logging"
	"errors"
	"strconv"
	"encoding/binary"
)

var beanLogger = logging.MustGetLogger("bean_cc")

type BeanChaincode struct {
}

func (bc *BeanChaincode) Init(stub shim.ChaincodeStubInterface,
		function string, args []string) ([]byte, error) {
	beanLogger.Debug("Entered Init func..")
	return nil, nil
}


func (bc *BeanChaincode) getBeanBalance(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	beanLogger.Debug("=============== getBeanBalance ===============")
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expected 1..")
	}
	requestAddress := args[0]

	//Get the state from the ledger
	beanBytes, err := stub.GetState(requestAddress)
	if err != nil {
		return []byte(strconv.Itoa(0)), errors.New("No BeanBalance stored in the ledger")
	}

	// binary.BigEndian.Uint64(mySlice)
	beanLogger.Info("Address[%x]'s Balance : %d", requestAddress, binary.BigEndian.Uint64(beanBytes))
	return beanBytes, nil
}

func (bc *BeanChaincode) transferBean(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	var remainBean4Sender, newBean4Receiver int

	beanLogger.Debug("=============== transferBean =================")
	if len(args) != 3 {
		return nil, errors.New("Incorrect number of arguments. Expected 3..")
	}

	sendAddr := args[0]
	recvAddr := args[1]
	beanAmount, err := strconv.ParseInt(args[2], 10, 64)
	if err != nil {
		return nil, errors.New("Error in Coverting beanAmount")
	}

	sendBeanBytes, err := stub.GetState(sendAddr)
	if err != nil {
		return nil, errors.New("Incorrect Address for Sender")
	}
	sendBean := binary.BigEndian.Uint64(sendBeanBytes)

	recvBeanBytes, err := stub.GetState(recvAddr)
	if err != nil {
		beanLogger.Debug("Address[%x] doesn't have the stored balance..", recvAddr)
		recvBeanBytes = []byte(strconv.Itoa(0))
	}
	recvBean := binary.BigEndian.Uint64(recvBeanBytes)

	if sendBean < beanAmount {
		return nil, errors.New("Not enough for Sender..")
	}
	remainBean4Sender = sendBean - beanAmount
	newBean4Receiver = recvBean + beanAmount

	// Store new Bean Amounts
	err = stub.PutState(sendAddr, []byte(strconv.Itoa(remainBean4Sender)))
	if err != nil {
		return nil, errors.New("Error in putting State with sendAddress")
	}
	err = stub.PutState(recvAddr, []byte(strconv.Itoa(newBean4Receiver)))
	if err != nil {
		// [AJ] Problem : what if PutState with sendAddr
		return nil, errors.New("Error in putting State with recvAddress")
	}

	return nil, nil
}

func (bc *BeanChaincode) Invoke(stub shim.ChaincodeStubInterface,
		function string, args []string) ([]byte, error) {
	beanLogger.Debug("Entered Invoke func..")

	if function == "transferBean" {
		return bc.transferBean(stub, args)
	}

	return nil, errors.New("Received Unknown Function Invokation")
}

func (bc *BeanChaincode) Query(stub shim.ChaincodeStubInterface,
		function string, args []string) ([]byte, error) {
	beanLogger.Debug("Entered Query func..")

	if function == "getBeanBalance" {
		return bc.getBeanBalance(stub, args)
	}

	return nil, errors.New("Received Unknown Function Query")

}

func main() {
	err := shim.Start(new(BeanChaincode))
	if err != nil {
		fmt.Printf("Error starting Bean chaincode: %s", err)
	}
}
