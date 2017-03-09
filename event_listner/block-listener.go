/*
Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements.  See the NOTICE file
distributed with this work for additional information
regarding copyright ownership.  The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License.  You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied.  See the License for the
specific language governing permissions and limitations
under the License.
*/

package main

import (
	"fmt"
	"os"

	"github.com/hyperledger/fabric/events/consumer"
	pb "github.com/hyperledger/fabric/protos"
	"bytes"
	"encoding/json"
	"net/http"
	"io"
	"flag"
	"encoding/binary"
	"github.com/golang/protobuf/proto"
	"io/ioutil"
	"strconv"
)

type adapter struct {
	notfy              chan *pb.Event_Block
	rejected           chan *pb.Event_Rejection
	cEvent             chan *pb.Event_ChaincodeEvent
	listenToRejections bool
	chaincodeID        string
}

type loginMsg struct {
	Name 		string	`json:"name"`
	Secret 		string	`json:"secret"`
	TargetAddr	string	`json:"targetAddr, omitempty"`
}


type Details  struct {
	BuyerAddr 	string	`json:"buyerAddr"`
	RewardAddr	string	`json:"rewardAddr"`
	RewardBean	uint32	`json:"rewardBean"`
	BuyerMsg  	string	`json:"buyerMsg,omitempty"`
}

type rewardMsg struct {
	User_hash 	string		`json:"user_hash"`
	Data 		Details	`json:"data"`
}

type T struct {
	A string
	B string
}
func testExample() {
	// Create a struct and write it.
	t := T{A: "0xEEFFEEFF", B: "3.14"}

	buf := &bytes.Buffer{}
	err := binary.Write(buf, binary.BigEndian, t)
	if err != nil {
		panic(err)
	}
	//fmt.Println([]byte(buf.Bytes()))//.Bytes())
	fmt.Println(buf.Bytes())

	// Read into an empty struct.
	t = T{}
	err = binary.Read(buf, binary.BigEndian, &t)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%x %f", t.A, t.B)
}

func testStructBytes(){

	//testExample()

	testVal := Details{
		BuyerAddr:"testBuyerAddr",
		RewardBean: 10,
		BuyerMsg:"From The Hospital",
	}

	b, err := json.Marshal(testVal)
	if err == nil {
		var recvVal Details
		err = json.Unmarshal(b,&recvVal)
		if err == nil {
			fmt.Print(recvVal)
		}

	}

}

func requestNotiforReward(rewardAddr string, buyerAddr string, bean uint32, msg string) {
	dataDetails := Details {BuyerAddr: buyerAddr, RewardBean:bean, BuyerMsg: msg}
	reqRewardMsg := rewardMsg{
		User_hash:rewardAddr,
		Data: dataDetails,
	}

	buffer := new(bytes.Buffer)
	json.NewEncoder(buffer).Encode(reqRewardMsg)
	//res, err := http.Post("http://www.mydata-market.com/v1/push/message",
	//	"application/json; charset=utf-8", buffer)

	fmt.Printf(buffer.String())
	url := "http://www.mydata-market.com/v1/push/message"
	request, err := http.NewRequest("POST", url, buffer)
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		fmt.Print("Error in client Do Request..")
		return
	}
	defer response.Body.Close()

	io.Copy(os.Stdout, response.Body)
}

func testHttpJsonPost() {
	login := &loginMsg{Name:"supernova27", Secret:"YmhtZslzEWgh", TargetAddr:"supernova27"}
	buffer := new(bytes.Buffer)
	json.NewEncoder(buffer).Encode(login)
	fmt.Printf(buffer.String())
	//http.Header.Set("Content-Type", "application/json")
	//res, err := http.Post("http://52.197.104.234:8080/getBeanBalance", "application/json;", buffer)
	url := "http://52.197.104.234:8080/getBeanBalance"
	request, err := http.NewRequest("POST", url, buffer)
	request.Header.Set("Content-Type","application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		fmt.Print("Error in Client Do Request..")
	}
	defer response.Body.Close()

	io.Copy(os.Stdout, response.Body)

}

// 721cb8cf-3374-4f78-8c3a-4a744d85ca96

func checkTransferComplete( txid string, sendAddr string, recvAddr string, beanAmount int) {

	url := "http://52.197.104.234:7050/transactions/" + txid
	response, err := http.Get(url)
	defer response.Body.Close()

	responseBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("...")
		return
	}

	message := pb.ChaincodeMessage{}
	err = json.Unmarshal(responseBytes, &message)
	if err != nil {
		fmt.Printf("...")
		return
	}

	invocationSpec := pb.ChaincodeInvocationSpec{}

	//err := pb.Unmarshal([]byte(payload), invocationSpec)
	err = proto.Unmarshal(message.Payload, &invocationSpec)
	if err != nil {
		fmt.Printf("Error: " + err.Error())
		return
	}

	fmt.Printf(invocationSpec.ChaincodeSpec.ChaincodeID.Name + "\n")
	fmt.Printf(invocationSpec.ChaincodeSpec.CtorMsg.String() + "\n")

	function := string(invocationSpec.ChaincodeSpec.CtorMsg.Args[0])
	if function != "transferBean" {
		fmt.Printf("Not TransferBean..\n")
		return
	}
	send := string(invocationSpec.ChaincodeSpec.CtorMsg.Args[1])
	recv := string(invocationSpec.ChaincodeSpec.CtorMsg.Args[2])
	amount, err := strconv.Atoi(string(invocationSpec.ChaincodeSpec.CtorMsg.Args[3]))

	if err != nil {
		fmt.Printf("Error in atoi..")
		return
	}

	if send != sendAddr || recv != recvAddr || beanAmount != amount {
		fmt.Printf("Not matched..")
		return
	} else {
		fmt.Printf("Done!!!!!!")
	}

}

//GetInterestedEvents implements consumer.EventAdapter interface for registering interested events
func (a *adapter) GetInterestedEvents() ([]*pb.Interest, error) {
	if a.chaincodeID != "" {
		return []*pb.Interest{
			{EventType: pb.EventType_BLOCK},
			{EventType: pb.EventType_REJECTION},
			{EventType: pb.EventType_CHAINCODE,
				RegInfo: &pb.Interest_ChaincodeRegInfo{
					ChaincodeRegInfo: &pb.ChaincodeReg{
						ChaincodeID: a.chaincodeID,
						EventName:   ""}}}}, nil
	}
	return []*pb.Interest{{EventType: pb.EventType_BLOCK}, {EventType: pb.EventType_REJECTION}}, nil
}

//Recv implements consumer.EventAdapter interface for receiving events
func (a *adapter) Recv(msg *pb.Event) (bool, error) {
	if o, e := msg.Event.(*pb.Event_Block); e {
		a.notfy <- o
		return true, nil
	}
	if o, e := msg.Event.(*pb.Event_Rejection); e {
		if a.listenToRejections {
			a.rejected <- o
		}
		return true, nil
	}
	if o, e := msg.Event.(*pb.Event_ChaincodeEvent); e {
		a.cEvent <- o
		return true, nil
	}
	return false, fmt.Errorf("Receive unkown type event: %v", msg)
}

//Disconnected implements consumer.EventAdapter interface for disconnecting
func (a *adapter) Disconnected(err error) {
	fmt.Printf("Disconnected...exiting\n")
	os.Exit(1)
}

func createEventClient(eventAddress string, listenToRejections bool, cid string) *adapter {
	var obcEHClient *consumer.EventsClient

	done := make(chan *pb.Event_Block)
	reject := make(chan *pb.Event_Rejection)
	adapter := &adapter{notfy: done, rejected: reject, listenToRejections: listenToRejections, chaincodeID: cid, cEvent: make(chan *pb.Event_ChaincodeEvent)}
	obcEHClient, _ = consumer.NewEventsClient(eventAddress, 5, adapter)
	if err := obcEHClient.Start(); err != nil {
		fmt.Printf("could not start chat %s\n", err)
		obcEHClient.Stop()
		return nil
	}

	return adapter
}


//requestNotiforReward(addr string, bean uint32, msg string)

func main() {

	// Test..
	checkTransferComplete(	"721cb8cf-3374-4f78-8c3a-4a744d85ca96",
				"2F6865DFE2FC2D4B52768A01C4D93316829D052DFF6413C582A171B53FAE7C05",
				"EA1B384B7D8B33F1A63013CF4EE698ED0F32B3AFEB71D7144A73D651E5A20E51",
				1000)

	// requestNotiforReward("nuclecker27", "supernova27", 200, "Bought from the one..")
	testStructBytes()

	var eventAddress string
	var listenToRejections bool
	var chaincodeID string
	flag.StringVar(&eventAddress, "events-address", "0.0.0.0:7053", "address of events server")
	flag.BoolVar(&listenToRejections, "listen-to-rejections", false, "whether to listen to rejection events")
	flag.StringVar(&chaincodeID, "events-from-chaincode", "", "listen to events from given chaincode")
	flag.Parse()

	fmt.Printf("Event Address: %s\n", eventAddress)
	if chaincodeID != "" {
		fmt.Printf("chaincodeID: %s\n", chaincodeID)
	}

	a := createEventClient(eventAddress, listenToRejections, chaincodeID)
	if a == nil {
		fmt.Printf("Error creating event client\n")
		return
	}

	for {
		select {
		case b := <-a.notfy:
			fmt.Printf("\n")
			fmt.Printf("\n")
			fmt.Printf("Received block\n")
			fmt.Printf("--------------\n")
			for _, r := range b.Block.Transactions {
				fmt.Printf("Transaction:\n\t[%v]\n", r)
			}
		case r := <-a.rejected:
			fmt.Printf("\n")
			fmt.Printf("\n")
			fmt.Printf("Received rejected transaction\n")
			fmt.Printf("--------------\n")
			fmt.Printf("Transaction error:\n%s\t%s\n", r.Rejection.Tx.Txid, r.Rejection.ErrorMsg)
		case ce := <-a.cEvent:
			fmt.Printf("\n")
			fmt.Printf("\n")
			fmt.Printf("Received chaincode event\n")
			fmt.Printf("------------------------\n")
			fmt.Printf("Chaincode Event:%v\n", ce)

			// ce.ChaincodeEvent.EventName
			processChaincodeEvent(ce.ChaincodeEvent)
		}
	}
	/*
	Received chaincode event
	------------------------
	Chaincode Event:&{
		chaincodeID:"b5e36a896c6178d13314fd91a831b61d3b4cf389632478fa1474ffe432697b76ca1234cb80f57491a9c5cc70e3b51bf625118fc2422ece91ce5ca6664f1fa104"
		txID:"a0168b57-35b7-47d9-b03e-eb8397f28b31"
		eventName:"BeanChanged"
		payload:"def" }
	*/
}
//func requestNotiforReward(rewardAddr string, buyerAddr string, bean uint32, msg string) {
func processChaincodeEvent(ce *pb.ChaincodeEvent){
	if ce.EventName == "BeanChanged" {
		var eventDetails Details
		err := json.Unmarshal(ce.Payload, &eventDetails)
		if err != nil {
			fmt.Errorf("Error in unmarshalling payload..%s", err)
			return
		}
		requestNotiforReward(eventDetails.RewardAddr, eventDetails.BuyerAddr, eventDetails.RewardBean, eventDetails.BuyerMsg)
	}
}
