package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"fmt"
	"io/ioutil"
	"github.com/golang/protobuf/proto"
	pb "github.com/hyperledger/fabric/protos"
	"strconv"
	"errors"
)

// 721cb8cf-3374-4f78-8c3a-4a744d85ca96

func isTransferComplete( txid string, sendAddr string, recvAddr string, beanAmount int) (int, error) {

	url := "http://52.197.104.234:7050/transactions/" + txid
	response, err := http.Get(url)
	defer response.Body.Close()

	responseBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("...")
		return 0, errors.New("Error in Reading Transaction Info from Peer")
	}

	message := pb.ChaincodeMessage{}
	err = json.Unmarshal(responseBytes, &message)
	if err != nil {
		fmt.Printf("...")
		return 0, errors.New("Error in Unmarshalling chaincode_messages")
	}

	invocationSpec := pb.ChaincodeInvocationSpec{}

	//err := pb.Unmarshal([]byte(payload), invocationSpec)
	err = proto.Unmarshal(message.Payload, &invocationSpec)
	if err != nil {
		fmt.Printf("Error: " + err.Error())
		return 0, errors.New("Error in Unmarshalling message_payload")
	}

	//fmt.Printf(invocationSpec.ChaincodeSpec.ChaincodeID.Name + "\n")
	//fmt.Printf(invocationSpec.ChaincodeSpec.CtorMsg.String() + "\n")

	function := string(invocationSpec.ChaincodeSpec.CtorMsg.Args[0])
	if function != "transferBean" {
		fmt.Printf("Not TransferBean..\n")
		return 0, errors.New("The requested function is not transferBean")
	}
	send := string(invocationSpec.ChaincodeSpec.CtorMsg.Args[1])
	recv := string(invocationSpec.ChaincodeSpec.CtorMsg.Args[2])
	amount, err := strconv.Atoi(string(invocationSpec.ChaincodeSpec.CtorMsg.Args[3]))

	if err != nil {
		fmt.Printf("Error in atoi..")
		return 0, errors.New("Error in AtoI from beanAmount..")
	}

	if send != sendAddr || recv != recvAddr || beanAmount != amount {
		fmt.Printf("Not matched..")
		return 0, errors.New("The Requested Info doesn't match..")

	} else {
		fmt.Printf("Done!!!!!!")
		return 1, nil
	}
	return 0, nil

}

type TransferInfo struct {
	SendAddr 	string	`json:"sendAddr"`
	RecvAddr	string	`json:"recvAddr"`
	BeanAmount	int	`json:"beanAmount"`
}

type CheckResult struct {
	TxID		string	`json:"txId"`
	Result		int	`json:"result"`
}

func CheckTransferComplete(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	txid := params["txid"]

	var transfer_info TransferInfo
	err := json.NewDecoder(req.Body).Decode(&transfer_info)
	if err != nil {
		fmt.Printf("Error in Decoding Json..")
		return
	}

	retVal, err := isTransferComplete (txid, transfer_info.SendAddr,
				transfer_info.RecvAddr, transfer_info.BeanAmount)

	var result CheckResult
	result.Result = retVal
	result.TxID = txid
	json.NewEncoder(w).Encode(result)
}

func HealthCheck(w http.ResponseWriter, req *http.Request) {
	var result CheckResult
	json.NewEncoder(w).Encode(result)
}

func checkTemp() {
	var trasfer_agrs []string
	trasfer_agrs = append(trasfer_agrs, "transferBean")
	trasfer_agrs = append(trasfer_agrs, "asdf")
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/", HealthCheck).Methods("GET")
	router.HandleFunc("/checkTransferComplete/{txid}", CheckTransferComplete).Methods("POST")
	log.Fatal(http.ListenAndServe(":8090", router))
}