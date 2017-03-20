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
	"flag"
	"github.com/hyperledger/fabric/events/consumer"
	"os"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

// 721cb8cf-3374-4f78-8c3a-4a744d85ca96

func isTransferComplete( txid string, sendAddr string, recvAddr string, beanAmount string) (int, error) {

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
	amount := string(invocationSpec.ChaincodeSpec.CtorMsg.Args[3])

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
	BeanAmount	string	`json:"beanAmount"`
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

type eventAdater struct {
	notify 			chan *pb.Event_Block
	rejected		chan *pb.Event_Rejection
	cEvent 			chan *pb.Event_ChaincodeEvent
	listenToRejctions	bool
	chaincodeID		string
}

func createEventClient (eventAddress string, rejection bool, cid string) *eventAdater {
	var obcEHClient *consumer.EventsClient

	done := make(chan *pb.Event_Block)
	reject := make(chan *pb.Event_Rejection)
	adapter := &eventAdater{
		notify: done,
		rejected: reject,
		chaincodeID: cid,
		cEvent: make(chan *pb.Event_ChaincodeEvent),
	}

	obcEHClient, _ = consumer.NewEventsClient(eventAddress, 5, adapter)
	if err := obcEHClient.Start(); err != nil {
		fmt.Printf("could not start eventClient.. %s\n", err)
		obcEHClient.Stop()
		return nil
	}

	return adapter

}

func decodeBeanTransferEvent (event []byte) {
	var transferInfo TransferInfo
	err := json.Unmarshal(event, &transferInfo)
	if err != nil {
		fmt.Errorf("Error in Unmarshalling the received TransferBean Event.. : %s\n", err)
		return
	}
	// Storing TransferBean Event into "Mysql"
}

func execRoutineForEvents (eventAddress string, rejection bool, chaincodeID string) {

	eventClient := createEventClient( eventAddress, rejection, chaincodeID)
	if eventClient == nil {
		fmt.Printf("Error in Creating Event Client..\n")
		return
	}

	go func() {
		for {
			select {
			case block := <- eventClient.notify:
				fmt.Printf("=== Receved Blocks === \n")
				for _, r := range block.Block.Transactions {
					fmt.Printf("Transaction:\n\t[%v]\n", r)
				}
			case reject := <- eventClient.rejected:
				fmt.Printf("Rejected Txid[%s] : %s\n", reject.Rejection.Tx.Txid, reject.Rejection.ErrorMsg)

			case ce := <- eventClient.cEvent:
				fmt.Printf("Chaincode Event:%v\n", ce)
				if ce.ChaincodeEvent.EventName == "BeanTransfer" {
					decodeBeanTransferEvent(ce.ChaincodeEvent.Payload)
				}
			}
		}
	}()
}

func (adapter *eventAdater) GetInterestedEvents() ([]*pb.Interest, error) {
	if adapter.chaincodeID != "" {
		return []*pb.Interest {
			{EventType: pb.EventType_BLOCK},
			{EventType: pb.EventType_REJECTION},
			{EventType: pb.EventType_CHAINCODE,
				RegInfo: &pb.Interest_ChaincodeRegInfo{
					ChaincodeRegInfo:&pb.ChaincodeReg{
						ChaincodeID: adapter.chaincodeID,
						EventName: ""}}}},nil
	}
	return []*pb.Interest{{EventType:pb.EventType_BLOCK}, {EventType:pb.EventType_REJECTION}}, nil
}

func (adapter *eventAdater) Recv(msg *pb.Event) (bool, error) {
	if info, err := msg.Event.(*pb.Event_Block); err {
		adapter.notify <- info
		return true, nil
	}
	if info, err := msg.Event.(*pb.Event_Rejection); err {
		if adapter.listenToRejctions {
			adapter.rejected <- info
		}
		return true, nil
	}
	if info, err := msg.Event.(*pb.Event_ChaincodeEvent); err {
		adapter.cEvent <- info
		return true, nil
	}

	return false, fmt.Errorf("Receive unknown type event: %v", msg)
}

func (adapter *eventAdater) Disconnected(err error) {
	fmt.Printf("Disconnected.. ")
	os.Exit(1)
}

/*
type TransferInfo struct {
	SendAddr 	string	`json:"sendAddr"`
	RecvAddr	string	`json:"recvAddr"`
	BeanAmount	int	`json:"beanAmount"`
}

CREATE TABLE beanrecords (  	sendAddress VARCHAR(128) NOT NULL,
				recvAddress VARCHAR(128) NOT NULL,
				beanAmount  INT(11) unsigned NOT NULL,
				transferTime timestamp not null);
 */
//
func setupBeanStorage(info TransferInfo) {
	db, err := sql.Open("mysql", "root:supernova27@tcp(127.0.0.1:3306)/cuppadata")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
/*
	var sendAddr string
	err = db.QueryRow("SELECT sendAddress FROM beanRecords").Scan(&sendAddr)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(sendAddr)
*/
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()

	_,err = db.Exec("Insert into beanrecords values (?,?,?,NOW())", info.SendAddr, info.RecvAddr, info.BeanAmount)
	if err != nil {
		log.Fatal(err)
	}

	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}
}

func main() {

	setupBeanStorage(TransferInfo{SendAddr:"qazwsx", RecvAddr:"edcrfv", BeanAmount:100})

	// Event Register
	var eventAddress string
	var listenToRejections bool
	var chaincodeID string

	flag.StringVar (&eventAddress, "events-address", "0.0.0.0:9053", "address of events Server..")
	flag.BoolVar( &listenToRejections, "listen-to-rejections", false, "whether to listen to rejection events")
	flag.StringVar (&chaincodeID, "events-from-chaincode", "", "listen to events from the given chaincode")

	fmt.Printf("Events Address: %s\n", eventAddress)
	fmt.Printf("Events From ChaincodeID : %s\n", chaincodeID)

	// Trigger Execution for Events
	execRoutineForEvents(eventAddress, listenToRejections, chaincodeID)

	// Support for TXID check
	router := mux.NewRouter()
	router.HandleFunc("/", HealthCheck).Methods("GET")
	router.HandleFunc("/checkTransferComplete/{txid}", CheckTransferComplete).Methods("POST")
	log.Fatal(http.ListenAndServe(":8090", router))
}