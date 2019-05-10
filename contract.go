package main

import (
        "bytes"
        "encoding/json"
        "fmt"
        "github.com/hyperledger/fabric/core/chaincode/shim"
        "github.com/hyperledger/fabric/protos/peer"
        "strconv"
)

type ContractChaincode struct {
}

type Employee struct {
        EmployeeId     int     `json:"employeeId"`
        EmployeeName   string  `json:"employeeName"`
        Salary         int     `json:"salary"`
        WorkingHours   float64 `json:"workingHours"`
        EmployeeType   string  `json:"employeeType"`
        ParentCompany  string  `json:"parentCompany"`
        CurrentCompany string  `json:"currentCompany"`
}

func (contract *ContractChaincode) Init(stub shim.ChaincodeStubInterface) peer.Response {
        fmt.Println("Init executed")
        return shim.Success(nil)
}

func (contract *ContractChaincode) Invoke(stub shim.ChaincodeStubInterface) peer.Response {

        function, args := stub.GetFunctionAndParameters()
        fmt.Println("Invoke is running " + function)
        if function == "addEmployee" {
                return contract.addEmployee(stub)
        } else if function == "createEmployee" {
                return contract.createEmployee(stub, args)
        } else if function == "getEmployees" {
                return contract.getEmployees(stub)
        } else if function == "sendEmployee" {
                return contract.sendEmployee(stub, args)
        }

        return shim.Error("invalid function name")
        // return peer.Response{Status:401, Message: "unAuthorized", Payload: payload}
}

func (contract *ContractChaincode) addEmployee(stub shim.ChaincodeStubInterface) peer.Response {
        employees := []Employee{
                Employee{EmployeeId: 1, EmployeeName: "karthik", Salary: 5000, WorkingHours: 9.30, EmployeeType: "contract", ParentCompany: "human", CurrentCompany: "human"},
                Employee{EmployeeId: 2, EmployeeName: "mithun", Salary: 4000, WorkingHours: 9.30, EmployeeType: "contract", ParentCompany: "human", CurrentCompany: "human"},
                Employee{EmployeeId: 3, EmployeeName: "kawakami", Salary: 10000, WorkingHours: 9.30, EmployeeType: "permanent", ParentCompany: "human", CurrentCompany: "human"},
                Employee{EmployeeId: 4, EmployeeName: "ozawa", Salary: 15000, WorkingHours: 8.30, EmployeeType: "permanent", ParentCompany: "hitachi", CurrentCompany: "hitachi"},
                Employee{EmployeeId: 5, EmployeeName: "sakura", Salary: 10000, WorkingHours: 8.30, EmployeeType: "permanent", ParentCompany: "hitachi", CurrentCompany: "hitachi"},
        }

        i := 0
        for i < len(employees) {
                employeeAsBytes, _ := json.Marshal(employees[i])
                stub.PutState(strconv.Itoa(employees[i].EmployeeId), employeeAsBytes)
                //stub.PutState("empId"+strconv.Itoa(i), employeeAsBytes)
                fmt.Println("Added", employees[i])
                i = i + 1
        }

        payload := []byte("Employee details added successfully , count: " + strconv.Itoa(i))
        return shim.Success(payload)

}

func (contract *ContractChaincode) createEmployee(stub shim.ChaincodeStubInterface, args []string) peer.Response {
        if len(args) != 7 {
                return shim.Error("Incorrect number of arguments, required: 7")
        }

        employeeId, _ := strconv.Atoi(args[1])
        salary, _ := strconv.Atoi(args[3])
        workingHours, _ := strconv.ParseFloat(args[4], 64)

        var employee = Employee{EmployeeId: employeeId, EmployeeName: args[2], Salary: salary, WorkingHours: workingHours, EmployeeType: args[5], ParentCompany: args[6], CurrentCompany: args[7]}

        employeeAsBytes, _ := json.Marshal(employee)
        //stub.PutState("emp-id"+strconv.Itoa(employee.employeeId), employeeAsBytes)
        stub.PutState(strconv.Itoa(employee.EmployeeId), employeeAsBytes)
        fmt.Println("Created ", employee)
        payload := []byte("Employee created successfully")
        return shim.Success(payload)
}

func (contract *ContractChaincode) getEmployees(stub shim.ChaincodeStubInterface) peer.Response {

        startKey := "0"
        endKey := "999"

        resultsIterator, err := stub.GetStateByRange(startKey, endKey)

        if err != nil {
                return shim.Error(err.Error())
        }
        defer resultsIterator.Close()

        var buffer bytes.Buffer
        buffer.WriteString("[")
        bArrayMemberAlreadyWritten := false

        for resultsIterator.HasNext() {
                queryResponse, err := resultsIterator.Next()
                if err != nil {
                        return shim.Error(err.Error())
                }
                // Add a comma before array members, suppress it for the first array member
                if bArrayMemberAlreadyWritten == true {
                        buffer.WriteString(",")
                }
                buffer.WriteString("{\"Key\":")
                buffer.WriteString("\"")
                buffer.WriteString(queryResponse.Key)
                buffer.WriteString("\"")

                buffer.WriteString(", \"Record\":")
                // Record is a JSON object, so we write as-is
                buffer.WriteString(string(queryResponse.Value))
                buffer.WriteString("}")
                bArrayMemberAlreadyWritten = true

        }

        buffer.WriteString("]")

        fmt.Printf("All employees:\n%s\n", buffer.String())
        result := buffer.String()
        payload := []byte(result)
        return peer.Response{Status: 200, Message: result, Payload: payload}
}

func (contract *ContractChaincode) sendEmployee(stub shim.ChaincodeStubInterface, args []string) peer.Response { //function, args := stub.GetFunctionAndParameters()
        if len(args) != 3 {
                return shim.Error("Incorrect number of arguments, required 3")
        }

        if args[1] == args[2] {
                return shim.Error("Invalid arguments, Trying to transfer in the same company")
        }
        // args[0]-empID args[1]-from args[2]-To
        employeeAsBytes, err := stub.GetState(args[0])

        if err != nil {
                return shim.Error(err.Error())
        }
        employee := Employee{}
        json.Unmarshal(employeeAsBytes, &employee)

        if employee.ParentCompany != "human" {
                return shim.Error("Only Human employees can be transferred")
        }

        if employee.EmployeeType != "contract" {
                return shim.Error("Only contract employees can be transferred")
        }

        if employee.CurrentCompany != args[1] {
                return shim.Error("This employee is not currently working in " + args[1] + " company")
        }

        employee.CurrentCompany = args[2]
        employeeAsBytes, _ = json.Marshal(employee)
        stub.PutState(args[0], employeeAsBytes)
        payload := []byte("Employee transfered Successfully")
        stub.SetEvent("sendEmployee", payload)
        return peer.Response{Status: 200, Message: "Record updated", Payload: payload}

}

func main() {
        fmt.Println("Started chain code")
        err := shim.Start(new(ContractChaincode))
        if err != nil {
                fmt.Println("Error Starting chain code : %s", err)
        }
}
