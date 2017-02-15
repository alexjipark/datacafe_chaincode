Reference URL : https://docs.google.com/document/d/10zz00XGxRTuXTcqgEHBE7WzN8shaIFnGV8hjtIhvh-U/edit

//================실행 코드 .. using CLI============//
ubuntu@ip-172-31-23-140:~/cuppa/fabric-exercise-security-on$ CORE_PEER_ADDRESS=localhost:7051 $GOPATH/src/github.com/hyperledger/fabric/build/bin/peer chaincode deploy -l java -p /project/rainfall-insurance -c '{"Function":"init", "Args":[]}'
Deploy chaincode: 2e6d19ef97e3910a74c9514e1914baafb60440626abdf8c6c03c381187ee9f6b478a17ba3ef80d09d0aef74c794133968fce08bf0d48f94b5e9ffab9068caa6e
ubuntu@ip-172-31-23-140:~/cuppa/fabric-exercise-security-on$ CORE_PEER_ADDRESS=localhost:7051 $GOPATH/src/github.com/hyperledger/fabric/build/bin/peer chaincode invoke -l java \
> -n 2e6d19ef97e3910a74c9514e1914baafb60440626abdf8c6c03c381187ee9f6b478a17ba3ef80d09d0aef74c794133968fce08bf0d48f94b5e9ffab9068caa6e \
> -c '{"Function":"new", "Args":["Ashok", "TN","300","1000000"]}'
ubuntu@ip-172-31-23-140:~/cuppa/fabric-exercise-security-on$ CORE_PEER_ADDRESS=localhost:7051 $GOPATH/src/github.com/hyperledger/fabric/build/bin/peer chaincode query -l java \
> -n 2e6d19ef97e3910a74c9514e1914baafb60440626abdf8c6c03c381187ee9f6b478a17ba3ef80d09d0aef74c794133968fce08bf0d48f94b5e9ffab9068caa6e \
> -c '{"Function":"get", "Args":["Ashok"]}'
Query Result: RainfallInsurance{insuranceID='Ashok', state='TN', rainThreshold=300, insuredAmount=1000000.0, premiumAmount=5000.0, paidOut=false, lapsed=true}


