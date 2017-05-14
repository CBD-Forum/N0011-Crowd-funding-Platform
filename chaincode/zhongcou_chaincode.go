/*
onechain zhongchou chaincode
*/

package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

const (
	ORDER_STATUS_CREATE    = iota //订单创建
	ORDER_STATUS_CANINVEST        //订单有效
	ORDER_STATUS_FULL             //订单满额
	ORDER_STATUS_LOAN             //订单放款
	ORDER_STATUS_REFUND           //订单还款
	ORDER_STATUS_FINISHED         //订单完成
)
const (
	ADMIN         = iota //管理员
	SIMPLE_PERSON        //普通用户
)

//存放用户信息
var primaryKeyToUser = map[string]*User{}

var nameToUser = map[string]*User{}

//存放用户众筹订单信息
var creatorToOrder = map[string]map[string]*Order{}

//存放用户投资订单信息
var creatorToInvestRecord = map[string][]*InvestRecord{}

//存放用户还款信息
var creatorToRefundRecord = map[string][]*RefundRecord{}

//用户信息
type User struct {
	ID     string  `json:"id"`     //用户id
	Name   string  `json:"name"`   //用户名
	Mobile string  `json:"mobile"` //用户电话
	Amount float64 `json:"amount"` //账户
	Role   int     `json:"role"`   //用户属性
}

//众筹订单信息
type Order struct {
	ID               string         `json:"id"`               //众筹订单id
	Title            string         `json:"title"`            //众筹项目名称
	Amount           float64        `json:"amount"`           //众筹金额
	Current          float64        `json:"current"`          //已筹金额
	Status           int            `json:"status"`           //众筹状态
	Rate             float64        `json:"rate"`             //收益率
	CreatorId        string         `json:"creatorId"`        //发起人id
	CreateTime       string         `json:"createTime"`       //发起时间
	EndTime          string         `json:"endTime"`          //结束时间
	InvestRecords    []InvestRecord `json:"investRecords"`    //投资记录
	RefundRecords    []RefundRecord `json:"refundRecords"`    //还款记录
	TradeCertificate string         `json:"tradeCertificate"` //交易凭证
}

//投资记录信息
type InvestRecord struct {
	ID         string  `json:"id"`         //投资记录id
	CreatorId  string  `json:"creatorId"`  //投资人id
	UserName   string  `json:"userName"`   //投资人名
	OrderId    string  `json:"orderId"`    //众筹订单id
	Amount     float64 `json:"amount"`     //投资金额
	Title      string  `json:"title"`      //项目名称
	CreateTime string  `json:"createTime"` //创建时间
}

//还款记录信息
type RefundRecord struct {
	ID         string  `json:"id"`         //还款记录id
	CreatorId  string  `json:"creatorId"`  //还款人id
	OrderId    string  `json:"orderId"`    //众筹订单id
	Amount     float64 `json:"amount"`     //还款金额
	Title      string  `json:"title"`      //项目名称
	CreateTime string  `json:"createTime"` //创建时间
}

type SimpleChaincode struct {
}

func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("########### example_cc Init ###########")

	CreateUser("admin", "18888888888", 200000, ADMIN)

	CreateUser("apple", "18673692416", 100000, SIMPLE_PERSON)

	CreateUser("jack", "18673692435", 400000, SIMPLE_PERSON)

	CreateUser("tom", "1333692435", 500000, SIMPLE_PERSON)

	return shim.Success(nil)

}

//调用智能合约
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("########### example_cc Invoke ###########")
	function, args := stub.GetFunctionAndParameters()
	if function != "invoke" {
		return shim.Error("Unknown function call")
	}
	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting at least 2")
	}
	//创建用户
	if args[0] == "createUser" {
		return t.createUser(stub, args)
	}
	//账户充值
	if args[0] == "recharge" {
		return t.recharge(stub, args)
	}
	//创建众筹
	if args[0] == "createOrder" {
		return t.createOrder(stub, args)
	}
	//发布众筹
	if args[0] == "publish" {
		return t.publish(stub, args)
	}
	//投资项目
	if args[0] == "invest" {
		return t.invest(stub, args)
	}
	//项目放款
	if args[0] == "loan" {
		return t.loan(stub, args)
	}
	//项目回款
	if args[0] == "refund" {
		return t.refund(stub, args)
	}
	//查询操作
	if args[0] == "query" {
		// queries an entity state
		return t.query(stub, args)
	}
	//凭证上传
	if args[0] == "uploadTradeCertificate" {
		return t.uploadTradeCertificate(stub, args)
	}

	return shim.Error("Unknown action, check the first argument, must be one of  ‘createUser’，‘recharge’，'createOrder', 'publish', 'invest', 'loan','uploadTradeCertificate' or 'refund'")
}

// 查询操作
func (t *SimpleChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	var subQueryMethod string = args[1]
	var param = args[2]
	switch subQueryMethod {
	//用户列表查询
	case "userList":
		{
			var userList []*User
			for _, v := range primaryKeyToUser {

				userList = append(userList, v)

			}
			bytes, err := json.Marshal(userList)
			if err != nil {
				return shim.Error("find UserList error !")
			}
			return shim.Success(bytes)
		}
	//用户查询
	case "user":
		{
			bytes, err := json.Marshal(primaryKeyToUser[param])
			if err != nil {
				return shim.Error("find User error !")
			}
			return shim.Success(bytes)
		}
	//未审核众筹查询
	case "orderList":
		{
			//非admin账户不能看到CREATE状态下的订单
			var isAdmin bool = false
			user, ok := primaryKeyToUser[param]
			if ok {
				if user.Role == ADMIN {
					isAdmin = true
				}
			}
			var orderList []*Order
			for _, v := range creatorToOrder {
				for _, v2 := range v {
					if !isAdmin && v2.Status == ORDER_STATUS_CREATE {
						continue
					}
					orderList = append(orderList, v2)
				}
			}
			bytes, err := json.Marshal(orderList)
			if err != nil {
				return shim.Error("find OrderList error !")
			}
			return shim.Success(bytes)
		}
	//用户众筹查询
	case "userOrderList":
		{
			var orderList []*Order
			for _, v := range creatorToOrder {
				for _, v2 := range v {
					if v2.CreatorId != param {
						continue
					}
					orderList = append(orderList, v2)
				}
			}
			bytes, err := json.Marshal(orderList)
			if err != nil {
				return shim.Error("find userOrderList error !")
			}
			return shim.Success(bytes)
		}
	case "order":
		{
			return t.queryOrder(stub, args)
		}
	//投资记录查询
	case "investRecord":
		{
			bytes, err := json.Marshal(creatorToInvestRecord[param])
			if err != nil {
				return shim.Error("find investRecord error !")
			}
			return shim.Success(bytes)
		}
	//还款记录查询
	case "refundRecord":
		{
			bytes, err := json.Marshal(creatorToRefundRecord[param])
			if err != nil {
				return shim.Error("find refundRecord error !")
			}
			return shim.Success(bytes)
		}
	default:
		{
			return shim.Error("support subQuery method is 'userList','user','orderList','userOrderList','order','refundRecord','investRecord'!")
		}

	}
}

//随机字符串
func getUUID() string {
	var size = 32
	var kind = 3
	ikind, kinds, result := kind, [][]int{[]int{10, 48}, []int{26, 97}, []int{26, 65}}, make([]byte, size)
	is_all := kind > 2 || kind < 0
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < size; i++ {
		if is_all { // random ikind
			ikind = rand.Intn(3)
		}
		scope, base := kinds[ikind][0], kinds[ikind][1]
		result[i] = uint8(base + rand.Intn(scope))
	}
	return string(result)
}

//创建用户
func CreateUser(name string, mobile string, amount float64, role int) *User {
	_, ok := nameToUser[name]
	if ok {
		fmt.Println("User is exist!")
		return nil
	}
	var user User
	user.ID = getUUID()
	user.Name = name
	user.Mobile = mobile
	user.Amount = amount
	user.Role = role
	fmt.Printf("Create User successfully User = %v\n", user)

	primaryKeyToUser[user.ID] = &user
	nameToUser[user.Name] = &user
	return &user
}

func (t *SimpleChaincode) createUser(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var name = args[1]
	var mobile = args[2]
	_, ok := nameToUser[name]
	if ok {
		return shim.Error("User name has exist!")
	}
	var user User
	user.ID = getUUID()
	user.Name = name
	user.Mobile = mobile
	user.Amount = 0
	user.Role = SIMPLE_PERSON

	primaryKeyToUser[user.ID] = &user
	nameToUser[user.Name] = &user
	return shim.Success(nil)
}

//账户充值
func (t *SimpleChaincode) recharge(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var userId = args[1]
	amount, err := strconv.ParseFloat(args[2], 64)
	if err != nil {
		return shim.Error("recharge amount illegal!")
	}

	user, ok := primaryKeyToUser[userId]
	if !ok {
		shim.Error("User not exist!")
	}
	user.Amount = user.Amount + amount
	nameToUser[user.Name] = user

	return shim.Success(nil)
}

//创建众筹
func (t *SimpleChaincode) createOrder(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var title = args[1]
	var creatorId = args[4]
	var createTime = args[5]
	var endTime = args[6]
	amount, err := strconv.ParseFloat(args[2], 32)
	if err != nil {
		return shim.Error("param amount must be integer!")
	}
	rate, err := strconv.ParseFloat(args[3], 32)
	if err != nil {
		return shim.Error("param rate parse error,please have a good check!")
	}
	order := Order{}
	order.ID = getUUID()
	order.Title = title
	order.Amount = amount
	order.Current = 0
	order.Status = ORDER_STATUS_CREATE
	order.Rate = rate
	order.CreatorId = creatorId
	order.CreateTime = createTime
	order.EndTime = endTime
	order.InvestRecords = []InvestRecord{}
	order.RefundRecords = []RefundRecord{}

	subMap := creatorToOrder[creatorId]
	if subMap == nil {
		subMap = map[string]*Order{}
	}
	subMap[order.ID] = &order
	creatorToOrder[creatorId] = subMap

	bytes, err := json.Marshal(&order)
	if err != nil {
		shim.Error("Json marshal error!")
	}
	stub.PutState(order.ID, bytes)
	byt, err := stub.GetState(order.ID)
	if err != nil {
		return shim.Error("Order is not exist!")
	}

	return shim.Success(byt)
}

//发布众筹
func (t *SimpleChaincode) publish(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var orderId = args[1]

	var order1 Order

	bytes, err := stub.GetState(orderId)
	if err != nil {
		return shim.Error("Get Order State error!")
	}
	json.Unmarshal(bytes, &order1)
	order := &order1
	if order == nil || order.ID == "" {
		return shim.Error("Order is not exist!")
	}
	if order.Status != ORDER_STATUS_CREATE {
		return shim.Error("Order can't publish! status not create!")
	}
	order.Status = ORDER_STATUS_CANINVEST
	b, err := json.Marshal(order)
	if err != nil {
		shim.Error("Json marshal error!")
	}
	stub.PutState(order.ID, b)

	var tempOrder *Order
	for _, m := range creatorToOrder {
		for _, o := range m {
			if orderId == o.ID {
				tempOrder = o
				break
			}
		}
	}
	tempOrder.Status = order.Status
	return shim.Success(nil)
}

//认筹
func (t *SimpleChaincode) invest(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var orderId = args[1]
	var creatorId = args[2]
	amount, err := strconv.ParseFloat(args[3], 32)
	if err != nil {
		return shim.Error("param amount must be float!")
	}

	var order1 Order

	bytes, err := stub.GetState(orderId)
	if err != nil {
		return shim.Error("Get Order State error!")
	}
	json.Unmarshal(bytes, &order1)
	order := &order1
	if order == nil || order.ID == "" {
		return shim.Error("Order is not exist!")
	}
	if order.Status != ORDER_STATUS_CANINVEST {
		return shim.Error("Order can't invest! status can not invest")
	}
	if order.Amount < amount {
		return shim.Error("Invest amount can't greater than order left amount!")
	}
	//用户扣款
	var user *User
	user2, ok := primaryKeyToUser[creatorId]
	if ok {
		user = user2
	}
	if user == nil {
		return shim.Error("User is not exist!")
	}
	if user.Amount < amount {
		return shim.Error("User has not enough money!")
	}
	//remove "amount" from user account
	if user.Amount < amount {
		return shim.Error("balance is not enough !")
	}
	user.Amount = user.Amount - amount

	var investRecord InvestRecord
	investRecord.ID = getUUID()
	investRecord.CreatorId = creatorId
	investRecord.Amount = amount
	investRecord.OrderId = order.ID
	investRecord.UserName = user.Name
	investRecord.Title = order.Title
	investRecord.CreateTime = time.Now().Format("2006-01-02 15:04:05")

	order.InvestRecords = append(order.InvestRecords, investRecord)
	order.Current += amount

	creatorToInvestRecord[creatorId] = append(creatorToInvestRecord[creatorId], &investRecord)

	//投资金额达到目标,修改订单状态为投资满额
	if order.Amount == order.Current {
		order.Status = ORDER_STATUS_FULL
	}

	b, err := json.Marshal(order)
	if err != nil {
		shim.Error("Json marshal error!")
	}
	stub.PutState(order.ID, b)

	var tempOrder *Order
	for _, m := range creatorToOrder {
		for _, o := range m {
			if orderId == o.ID {
				tempOrder = o
				break
			}
		}
	}
	tempOrder.Current = order.Current
	tempOrder.Status = order.Status
	tempOrder.InvestRecords = order.InvestRecords

	return shim.Success(nil)
}

//放款
func (t *SimpleChaincode) loan(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var orderId = args[1]
	var order1 Order

	bytes, err := stub.GetState(orderId)
	if err != nil {
		return shim.Error("Get Order State error!")
	}
	json.Unmarshal(bytes, &order1)
	order := &order1
	if order == nil || order.ID == "" {
		return shim.Error("Order is not exist!")
	}
	if order.Status != ORDER_STATUS_FULL || order.Current != order.Amount {
		return shim.Error("Order amount is not enough !")
	}
	//放款
	user := primaryKeyToUser[order.CreatorId]
	user.Amount = user.Amount + order.Current

	//订单存放的投资金额清零
	order.Current = 0

	//订单状态改为已放款
	order.Status = ORDER_STATUS_LOAN

	b, err := json.Marshal(order)
	if err != nil {
		shim.Error("Json marshal error!")
	}
	stub.PutState(order.ID, b)

	var tempOrder *Order
	for _, m := range creatorToOrder {
		for _, o := range m {
			if orderId == o.ID {
				tempOrder = o
				break
			}
		}
	}
	tempOrder.Current = order.Current
	tempOrder.Status = order.Status

	return shim.Success(nil)
}

//众筹回款
func (t *SimpleChaincode) refund(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var orderId = args[1]

	var order1 Order

	bytes, err := stub.GetState(orderId)
	if err != nil {
		return shim.Error("Get Order State error!")
	}
	json.Unmarshal(bytes, &order1)
	order := &order1
	if order == nil || order.ID == "" {
		return shim.Error("Order is not exist!")
	}
	if order.Status != ORDER_STATUS_LOAN {
		return shim.Error("Order can't refund current status is not loan!")
	}
	user := primaryKeyToUser[order.CreatorId]
	if user == nil {
		return shim.Error("User is not exist !")
	}
	var receiveAmount float64
	for _, v := range order.InvestRecords {
		receiveAmount += v.Amount * (1 + order.Rate)
	}
	if user.Amount < receiveAmount {
		return shim.Error("User balance is not enough !")
	}

	for _, v := range order.InvestRecords {
		u := primaryKeyToUser[v.CreatorId]

		refundRecord := RefundRecord{}
		refundRecord.ID = getUUID()
		refundRecord.CreatorId = v.CreatorId
		var realAmount = v.Amount * (1 + order.Rate)
		refundRecord.Amount = realAmount
		refundRecord.OrderId = v.OrderId
		refundRecord.Title = order.Title
		refundRecord.CreateTime = time.Now().Format("2006-01-02 15:04:05")

		if order.RefundRecords == nil {
			order.RefundRecords = []RefundRecord{}
		}
		order.RefundRecords = append(order.RefundRecords, refundRecord)
		u.Amount = u.Amount + realAmount
		user.Amount = user.Amount - realAmount

		creatorToRefundRecord[refundRecord.CreatorId] = append(creatorToRefundRecord[refundRecord.CreatorId], &refundRecord)
	}
	order.Status = ORDER_STATUS_REFUND

	var sum float64
	for _, v := range order.RefundRecords {
		sum += v.Amount
	}
	if order.Amount <= sum {
		order.Status = ORDER_STATUS_FINISHED
	}

	b, err := json.Marshal(order)
	if err != nil {
		shim.Error("Json marshal error!")
	}
	stub.PutState(order.ID, b)

	var tempOrder *Order
	for _, m := range creatorToOrder {
		for _, o := range m {
			if orderId == o.ID {
				tempOrder = o
				break
			}
		}
	}
	tempOrder.Status = order.Status
	tempOrder.RefundRecords = order.RefundRecords

	return shim.Success(nil)
}

//凭证上传
func (t *SimpleChaincode) uploadTradeCertificate(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	fmt.Println("×××××××××××××上传交易凭证××××××××××××××")

	var orderId = args[1]
	var tradeCertificate = args[2]

	var order1 Order

	bytes, err := stub.GetState(orderId)
	if err != nil {
		return shim.Error("Get Order State error!")
	}
	json.Unmarshal(bytes, &order1)
	order := &order1
	if order == nil || order.ID == "" {
		return shim.Error("Order is not exist!")
	}

	if tradeCertificate == "" {
		return shim.Error("Certificate can not be empty!")
	}
	order.TradeCertificate = tradeCertificate

	b, err := json.Marshal(order)
	if err != nil {
		shim.Error("Json marshal error!")
	}
	stub.PutState(order.ID, b)

	return shim.Success(nil)
}

//认筹记录查询
func (t *SimpleChaincode) queryOrder(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	fmt.Print("执行queryEntity方法: ")
	fmt.Println(args)

	var orderId = args[2]
	var order1 Order

	bytes, err := stub.GetState(orderId)
	if err != nil {
		return shim.Error("Get Order State error!")
	}
	json.Unmarshal(bytes, &order1)
	order := &order1

	if order.ID == "" {
		return shim.Success(nil)
	}

	b, err := json.Marshal(order)
	if err != nil {
		shim.Error("Json marshal error!")
	}
	stub.PutState(order.ID, b)

	return shim.Success(b)
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
