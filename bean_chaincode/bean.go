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
	"errors"
	"strconv"
	"encoding/json"
)

//var beanLogger = logging.MustGetLogger("bean_cc")

type Resp_AccountInfo struct {
	Address	string	`json:"address"`
	Bean 	int	`json:"bean"`
}

/*
0 	RecvAddres 	String
1	Timestamp	int64(long)
2	SendAddress	String
3	TransferBean	int32
4	Certificate	Byte
*/
type TransactionInfo struct {
	SendAddress		string	`json:"sendAddr"`
	TransactionTime		int64	`json:"transactionTime"`
	TransferredBean		int32	`json:"transferredBean"`
}

type TransactionList struct {
	Transactions	[]TransactionInfo	`json:"transactions"`
}


type BeanChaincode struct {
}

var TableforBeanTransation = "BeanTransaction"

func (bc *BeanChaincode) Init(stub shim.ChaincodeStubInterface,
		function string, args []string) ([]byte, error) {
//	beanLogger.Debug("Entered Init func..")

	err := stub.CreateTable("BeanTransaction", []*shim.ColumnDefinition{
		&shim.ColumnDefinition{Name:"RecvAddress", Type: shim.ColumnDefinition_STRING, Key:false},
		&shim.ColumnDefinition{Name:"Timestamp", Type: shim.ColumnDefinition_INT64, Key:true},
		&shim.ColumnDefinition{Name:"SendAddress", Type: shim.ColumnDefinition_STRING, Key:false},
		&shim.ColumnDefinition{Name:"TransferBean", Type: shim.ColumnDefinition_INT32, Key:false},
		&shim.ColumnDefinition{Name:"Certificate", Type: shim.ColumnDefinition_BYTES, Key:false},
	})
	if err != nil {
		return nil, errors.New("Failed creating BeanTransaction Table..")
	}

	return nil, nil
}

func (bc *BeanChaincode) checkCallerCert (stub shim.ChaincodeStubInterface, certificate []byte) (bool, error) {
	// ref : asset_management.go
	sigma, err := stub.GetCallerMetadata()
	if err != nil {
		return false, errors.New("Failed Getting Metadata..")
	}
	// Transaction payload, which is a `ChaincodeSpec` defined in fabric/protos/chaincode.proto
	payload, err := stub.GetPayload()
	if err != nil {
		return false, errors.New("Failed Getting Payload..")
	}
	// Transaction Binding?!..
	binding, err := stub.GetBinding()
	if err != nil {
		return false, errors.New("Failed Getting Binding..")
	}

	// certificate, signature, messages => Verify Signautre..
	res, err := stub.VerifySignature(certificate, sigma, append(payload,binding...),)
	if err != nil {
		return false, errors.New("Failed Verifying Signature..")
	}
	return res, nil
}

func convertTransactionToJson (row shim.Row) TransactionInfo {
	var transaction TransactionInfo

	transaction.SendAddress = row.Columns[2].GetString_()
	//transaction.TransferredBean = strconv.Itoa(int(row.Columns[3].GetInt32()))
	//transaction.TransactionTime = fmt.Sprintf("%d",row.Columns[2].GetInt64())//strconv.FormatInt(row.Columns[2].GetInt64(), 10)
	transaction.TransferredBean = row.Columns[3].GetInt32()
	transaction.TransactionTime = row.Columns[2].GetInt64()

	return transaction
}

func (bc *BeanChaincode) queryTransactions (stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	recvAddr := args[0]
	str_fromPeriod := args[1]
	str_toPeriod := args[2]

	var columns []shim.Column
//	col1 := shim.Column {Value: &shim.Column_String_{String_:recvAddr}}
//	columns = append(columns, col1)

	/*
	row, err := stub.GetRow("BeanTransaction", columns)
	if err != nil {
		return nil, fmt.Errorf("getRow operation failed. %s", err)
	}
	rowString := fmt.Sprintf("%s", row)
	return []byte(rowString), nil
	*/


	rowChannel, err := stub.GetRows("BeanTransaction", columns)
	if err != nil {
		return nil, fmt.Errorf("Failed retrieving Transfer Record")
	}
	// Timestamp, RecvAddress, SendAddress, TransferBean, Certificate

	var transactions TransactionList
	var rows []shim.Row
	for {
		select {
		case row, ok := <- rowChannel :
			if !ok {
				rowChannel = nil
			} else {
				/*
0 	RecvAddres 	String
1	Timestamp	int64(long)
2	SendAddress	String
3	TransferBean	int32
4	Certificate	Byte
				 */

				if recvAddr == row.Columns[0].GetString_() {
					fromPeriod,_ := strconv.ParseInt(str_fromPeriod, 10, 64)
					toPeriod,_ := strconv.ParseInt(str_toPeriod, 10, 64)
					rowTimeStamp := row.Columns[1].GetInt64()
					if str_toPeriod == "0" {
						if str_fromPeriod == "0" {
							//rows = append(rows, row)
							transactions.Transactions = append(transactions.Transactions, convertTransactionToJson(row))
						} else if rowTimeStamp > fromPeriod {
							rows = append(rows, row)
						}
					} else {
						if str_fromPeriod == "0" && rowTimeStamp < toPeriod {
							//rows = append(rows, row)
							transactions.Transactions = append(transactions.Transactions, convertTransactionToJson(row))
						} else if rowTimeStamp > fromPeriod && rowTimeStamp < toPeriod {
							//rows = append(rows, row)
							transactions.Transactions = append(transactions.Transactions, convertTransactionToJson(row))
						}
					}
				}
			}
		}
		if rowChannel == nil {
			break
		}
	}

	jsonRows, err := json.Marshal(transactions)
	if err != nil {
		return nil, fmt.Errorf("queryTransactions operation failed. Error marshaling JSON: %s", err)
	}
	return jsonRows, nil

}

func (bc *BeanChaincode) assignNewTransaction (stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	sendAddr := args[0]
	recvAddr := args[1]
	transferBean,err := strconv.Atoi(args[2])
	if err != nil {
		return nil, errors.New("Invalid value for transferBean")
	}
	timeStamp, err := stub.GetTxTimestamp()
	if err != nil {
		return nil, errors.New("Failed to get Tx Timestamp..")
	}
	cert, err := stub.GetCallerMetadata()
	if err != nil {
		return nil, errors.New("Error Getting Caller Metadata")
	}

	ok, err := stub.InsertRow("BeanTransaction", shim.Row {
		Columns: []*shim.Column {
			&shim.Column{Value:&shim.Column_String_{String_:recvAddr}},
			&shim.Column{Value: &shim.Column_Int64{Int64:timeStamp.Seconds}},
			&shim.Column{Value:&shim.Column_String_{String_:sendAddr}},
			&shim.Column{Value:&shim.Column_Int32{Int32:int32(transferBean)}},
			&shim.Column{Value: &shim.Column_Bytes{Bytes: cert}}},
	})
	if !ok && err == nil {
		return nil, errors.New("BeanTransaction was already assigned..")
	}

	return nil, nil
}



func (bc *BeanChaincode) queryAndDeleteTransaction (stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	recvAddr := args[0]

	var columns []shim.Column
	col1 := shim.Column{Value:&shim.Column_String_{String_:recvAddr}}
	columns = append(columns, col1)

	row, err := stub.GetRow("BeanTransaction", columns)
	if err != nil {
		return nil, errors.New("Error Getting Row from the Table..")
	}
	prvCert := row.Columns[4].GetBytes()
	if len(prvCert) == 0 {
		return nil, errors.New("Invalid Previous Certificate stored in the Table..")
	}

	ok, err := bc.checkCallerCert(stub, prvCert)
	if err != nil {
		return nil, errors.New("Error occured when checking Caller's Certificate")
	}
	if ok == false {
		return nil, errors.New("Invalid Caller's Cert")
	}

	//Delete the selected column..
	err = stub.DeleteRow("BeanTransaction",
				[]shim.Column{shim.Column{Value: &shim.Column_String_{String_: recvAddr}}})
	if err != nil {
		return nil, errors.New("Error Deleting the Selected Rows..")
	}
	return nil,nil

}

func (bc *BeanChaincode) getBeanBalance(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
//	beanLogger.Debug("=============== getBeanBalance ===============")
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expected 1..")
	}
	requestAddress := args[0]

	//Get the state from the ledger
	beanBytes, err := stub.GetState(requestAddress)
	if err != nil {
		return []byte(strconv.Itoa(0)), errors.New("No BeanBalance stored in the ledger")
	}
	bean,_ := strconv.Atoi(string(beanBytes))
	resp_info := &Resp_AccountInfo{
		Address: requestAddress,
		Bean: bean,
	}
	resp_bytes, _ := json.Marshal(resp_info)
	return resp_bytes, nil

	// binary.BigEndian.Uint64(mySlice)
//	beanLogger.Info("Address[%x]'s Balance : %d", requestAddress, binary.BigEndian.Uint64(beanBytes))
	//return beanBytes, nil
}

func (Bc *BeanChaincode) depositBean(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expected 2..")
	}
	targetAddr := args[0]
	beanAmount,err := strconv.Atoi(args[1])
	if err != nil {
		return nil, err
	}
	currentBeanBytes, err := stub.GetState(targetAddr)
	var currentBean int
	if err != nil {
		//return nil, errors.New("Error in Getting current BeanBytes")
		currentBean = 0
	} else {
		currentBean,_ = strconv.Atoi(string(currentBeanBytes))
	}
	// assign new bean
	err = stub.PutState(targetAddr, []byte(strconv.Itoa(currentBean+beanAmount)))
	if err != nil {
		return nil, errors.New("Error in Putting new BeanBytes")
	}

	return nil, nil
}


func (bc *BeanChaincode) transferBean(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	var remainBean4Sender, newBean4Receiver int

//	beanLogger.Debug("=============== transferBean =================")
	if len(args) != 3 {
		return nil, errors.New("Incorrect number of arguments. Expected 3..")
	}

	sendAddr := args[0]
	recvAddr := args[1]
	//beanAmount, err := strconv.ParseInt(args[2], 10, 64)
	beanAmount, err := strconv.Atoi(args[2])
	if err != nil {
		return nil, errors.New("Error in Coverting beanAmount")
	}

	sendBeanBytes, err := stub.GetState(sendAddr)
	if err != nil {
		return nil, errors.New("Incorrect Address for Sender")
	}
	//sendBean := binary.BigEndian.Uint64(sendBeanBytes)
	sendBean,_ := strconv.Atoi(string(sendBeanBytes))

	recvBeanBytes, err := stub.GetState(recvAddr)
	if err != nil {
//		beanLogger.Debug("Address[%x] doesn't have the stored balance..", recvAddr)
		recvBeanBytes = []byte(strconv.Itoa(0))
	}
	//recvBean := binary.BigEndian.Uint64(recvBeanBytes)
	recvBean,_ := strconv.Atoi(string(recvBeanBytes))

	//uBeanAmount := uint64(beanAmount)
	if sendBean < beanAmount {
		return nil, errors.New("Not enough for Sender..")
	}
	remainBean4Sender = sendBean - beanAmount
	newBean4Receiver = recvBean + beanAmount

	// Store new Bean Amounts
	//err = stub.PutState(sendAddr, []byte(strconv.FormatUint(remainBean4Sender,10)))
	err = stub.PutState(sendAddr, []byte(strconv.Itoa(remainBean4Sender)))
	if err != nil {
		return nil, errors.New("Error in putting State with sendAddress")
	}
	err = stub.SetEvent("BeanChanged", []byte(sendAddr))
	if err != nil {
		fmt.Printf("Error in Setting event for Addr[%x]\n", sendAddr)
	}

	//err = stub.PutState(recvAddr, []byte(strconv.FormatUint(newBean4Receiver,10)))
	err = stub.PutState(recvAddr, []byte(strconv.Itoa(newBean4Receiver)))
	if err != nil {
		// [AJ] Problem : what if PutState with sendAddr
		return nil, errors.New("Error in putting State with recvAddress")
	}
	err = stub.SetEvent("BeanChanged", []byte(recvAddr))
	if err != nil {
		fmt.Printf("Error in Setting event for Addr[%x]", recvAddr)
	}

	//====================== Update Table ====================//
	_, err = bc.assignNewTransaction(stub, args)
	if err != nil {
		//fmt.Printf("Error in Assign New Transaction in the Table..")
		return nil, errors.New("Error in Assign New Transaction in the Table..")
	}

	return nil, nil
}

func (bc *BeanChaincode) Invoke(stub shim.ChaincodeStubInterface,
		function string, args []string) ([]byte, error) {
//	beanLogger.Debug("Entered Invoke func..")

	if function == "transferBean" {
		return bc.transferBean(stub, args)
	} else if function == "depositBean" {
		return bc.depositBean(stub, args)
	} else if function =="transferMultipleBean" {
		// 1 to N..
	}

	return nil, errors.New("Received Unknown Function Invokation")
}

func (bc *BeanChaincode) Query(stub shim.ChaincodeStubInterface,
		function string, args []string) ([]byte, error) {
//	beanLogger.Debug("Entered Query func..")
	var jsonRespByte []byte
	if function == "getBeanBalance" {
		jsonRespByte, _ = bc.getBeanBalance(stub, args)
	} else if function == "getTransactionList" {
		jsonRespByte, _ = bc.queryTransactions(stub, args)
	}

	if len(jsonRespByte) == 0 {
		return nil, errors.New("Received Unknown Function Query")
	} else {
		fmt.Printf("Query Response:%s\n", string(jsonRespByte))
		return jsonRespByte, nil
	}
}

func main() {
	err := shim.Start(new(BeanChaincode))
	if err != nil {
		fmt.Printf("Error starting Bean chaincode: %s", err)
	}
}
