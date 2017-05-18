/*
Copyright IBM Corp 2016 All Rights Reserved.

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
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

type Patient struct {
	SourceId    string `json:"sourceId"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	DateOfBirth string `json:"dateOfBirth"`
	Sex         string `json:"sex"`
	PhoneNumber string `json:"phoneNumber"`
	EntityId    string `json:"entityId"`
	Meds        []Medication   'json:"meds"'
}

type Medication struct {
	MedName    string `json:"medName"`
	Dosage     string `json:"dosage"`
	FillDate   string `json:"fillDate"`
	Form       string `json:"form"`
	Quantity   string `json:"quantity"`
	Count      string `json:"count"`
	FillLoc    string `json:"fillLoc"`
	Prescriber string `json:"prescriber"`
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// Init resets all the things
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("hl3 Is Starting Up")
	_, args := stub.GetFunctionAndParameters()
	var Aval int
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	// convert numeric string to integer
	Aval, err = strconv.Atoi(args[0])
	if err != nil {
		return shim.Error("Expecting a numeric string argument to Init()")
	}

	// store compaitible hl3 application version
	err = stub.PutState("hl3_ui", []byte("1.0.0"))
	if err != nil {
		return shim.Error(err.Error())
	}

	// this is a very simple dumb test.  let's write to the ledger and error on any errors
	err = stub.PutState("selftest", []byte(strconv.Itoa(Aval))) //making a test var "selftest", its handy to read this right away to test the network
	if err != nil {
		return shim.Error(err.Error())                          //self-test fail
	}

	fmt.Println(" - ready for action")                          //self-test pass
	return shim.Success(nil)
}

// Invoke isur entry point to invoke a chaincode function
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println(" ")
	fmt.Println("starting invoke, for - " + function)

	// Handle different functions
	if function == "init" {                    //initialize the chaincode state, used as reset
		return t.Init(stub)
	} else if function == "read" {             //generic read ledger
		return read(stub, args)
	} else if function == "write" {            //generic writes to ledger
		return write(stub, args)
	} else if function == "addPatient" {
		return t.addPatient(stub, args)
	} else if function == "addMedication" {
		return t.addMedication(stub, args)
	} else if function == "removeMedication" {
		return t.removeMedication(stub, args)
	}
	fmt.Println("Received unknown invoke function name - " + function)
	return shim.Error("Received unknown invoke function name - '" + function + "'")
}

func read(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var key, jsonResp string
	var err error
	fmt.Println("starting read")

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting key of the var to query")
	}

	// input sanitation
	err = sanitize_arguments(args)
	if err != nil {
		return shim.Error(err.Error())
	}

	key = args[0]
	valAsbytes, err := stub.GetState(key)           //get the var from ledger
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + key + "\"}"
		return shim.Error(jsonResp)
	}

	fmt.Println("- end read")
	return shim.Success(valAsbytes)                  //send it onward
}

func write(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var key, value string
	var err error
	fmt.Println("starting write")

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2. key of the variable and value to set")
	}

	// input sanitation
	err = sanitize_arguments(args)
	if err != nil {
		return shim.Error(err.Error())
	}

	key = args[0]                                   //rename for funsies
	value = args[1]
	err = stub.PutState(key, []byte(value))         //write the variable into the ledger
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end write")
	return shim.Success(nil)
}

// write - invoke function to write key/value pair
func (t *SimpleChaincode) addMedication(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	bytes, _ := stub.GetState(args[0])
	fmt.Println(bytes)

	var pat Patient
	json.Unmarshal(bytes, &pat)

	var med Medication

	med.MedName = args[1]
	med.Count = args[2]
	med.Dosage = args[3]
	med.FillDate = args[4]
	med.Form = args[5]
	med.Quantity = args[6]
	med.FillLoc = args[7]
	med.Prescriber = args[8]

	pat.Meds = append(pat.Meds, med)

	bytes, _ = json.Marshal(pat)

	stub.PutState(args[0], bytes)

	return shim.Success(nil)

}

func (t *SimpleChaincode) removeMedication(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	bytes, _ := stub.GetState(args[0])
	fmt.Println(bytes)

	var pat Patient
	json.Unmarshal(bytes, &pat)

	for index, element := range pat.Meds {
		// fmt.Println(element.MedName)
		// fmt.Println(args[1])
		if element.MedName == args[1] && element.FillDate == args[2] {
			var blankMed Medication
			fmt.Println("FOUND TO REMOVE!")
			pat.Meds[index] = blankMed
			break
		}
	}
	fmt.Println(index)

	bytes, _ = json.Marshal(pat)

	stub.PutState(args[0], bytes)

	return shim.Success(nil)

}

func RemoveIndex(s []Medication, index int) []Medication {
	return append(s[:index], s[index+1:]...)
}

func (t *SimpleChaincode) addPatient(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	var err error
	fmt.Println("running addPatient()")

	// if len(args) != 2 {
	// 	return nil, errors.New("Incorrect number of arguments. Expecting 2. name of the key and value to set")
	// }

	var patient Patient
	var sourceid = args[0] //rename for funsies
	var firstname = args[1]
	var lastname = args[2]
	var dob = args[3]
	var sex = args[4]
	var phone = args[5]

	patient.SourceId = sourceid
	patient.FirstName = firstname
	patient.LastName = lastname
	patient.DateOfBirth = dob
	patient.Sex = sex
	patient.PhoneNumber = phone
	patient.Meds.Meds = make([]Medication, 200)

	fmt.Println("Before Patient Marshall:")
	fmt.Println(patient)

	bytes, err := json.Marshal(patient)

	if err != nil {
		return nil, errors.New("Error creating Patient record")
	}

	err = stub.PutState(sourceid, bytes)
	if err != nil {
		return nil, err
	}

	return shim.Success(nil)
}

