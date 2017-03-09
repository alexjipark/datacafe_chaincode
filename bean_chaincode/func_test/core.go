package main

import (
	"fmt"
	"strconv"
	"encoding/json"
)

var state map[string] []byte

func getBeanBalance(args []string) ([]byte, error) {
	requestAddress := args[0]
	beanBytes := state[requestAddress]
	//beanAmount := binary.BigEndian.Uint64(beanBytes)
	beanAmount, _ := strconv.Atoi(string(beanBytes))
	fmt.Printf("Addr[%x]'s state : %d\n", requestAddress, beanAmount)

	return beanBytes, nil
}

func transferBean(args []string) {
	if len(args) != 3 {
		return
	}
	sendAddr := args[0]
	recvAddr := args[1]
	//beanAmount, _ := strconv.ParseInt(args[2], 10, 32)
	beanAmount,_ := strconv.Atoi(args[2])

	sendBean,_ := strconv.Atoi(string(state[sendAddr]))
	recvBean,_ := strconv.Atoi(string(state[recvAddr]))

	if sendBean < beanAmount {
		fmt.Printf("Not enough bean for sender\n")
		return
	}

	remainBean4Sender := sendBean - beanAmount
	newBean4Receiver := recvBean + beanAmount

	//state[sendAddr] = []byte(strconv.FormatInt(remainBean4Sender, 10))
	//state[recvAddr] = []byte(strconv.FormatInt(newBean4Receiver,10))
	state[sendAddr] = []byte(strconv.Itoa(remainBean4Sender))
	state[recvAddr] = []byte(strconv.Itoa(newBean4Receiver))
}

func checkBalance(args []string) {
	for _,addr := range args {
		curBean,_ := strconv.Atoi(string(state[addr]))
		fmt.Printf("Addr[%x]'s Balance : %d\n", addr,  curBean)
	}
}

type TransactionInfo struct {
	SendAddress		string	`json:"sendAddr"`
	TransactionTime		string	`json:"transactionTime"`
	TransferredBean		int32	`json:"transferredBean"`
}

type TransactionList struct {
	Transactions	[]TransactionInfo	`json:"transactions"`
}

func convertTransactionToJson () TransactionInfo {
	var transaction TransactionInfo

	transaction.SendAddress = "sendAddress"
	transaction.TransferredBean = 12
	transaction.TransactionTime = strconv.FormatInt(1488856727,10)

	return transaction
}

func testConverting () {
	var transactions TransactionList

	temp := fmt.Sprintf("%d",12345)
	fmt.Printf(temp)

	transactions.Transactions = append(transactions.Transactions, convertTransactionToJson())
	transactions.Transactions = append(transactions.Transactions, convertTransactionToJson())
	jsonbytes, err := json.Marshal(transactions)
	if err == nil {
		fmt.Printf(string(jsonbytes))
	}
}

func main() {

	testConverting()

	state = make(map[string][]byte)
	state["abc"] = []byte(strconv.FormatInt(100,10))
	state["def"] = []byte(strconv.FormatInt(100,10))

	// test : getBeanBalance
	test_args := []string{"abc"}
	getBeanBalance(test_args)

	// test : transferBean
	transferBean([]string{"abc","def","23"})
	checkBalance([]string{"abc","def"})
	transferBean([]string{"abc","def","33"})
	checkBalance([]string{"abc","def"})
	checkBalance([]string{"abc","def"})
	transferBean([]string{"def","abc","56"})
	checkBalance([]string{"abc","def"})

}
