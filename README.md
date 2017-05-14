## 基于区块链技术的新一代众筹平台

 
说明

- 该demo展示了如何在众筹的场景中使用区块链


结构说明

- chaincode

fabric chaincode 源码，支持fabric 1.0版本

- restapi

基于fabric的nodejs SDK 封装了一个基于json rpc的接口，通过CURL可以访问区块链

nodejs的版本必须为 6.10.x

- application

用java开发的众筹流程模拟关键，运行环境为：jdk1.8及以上 tomcat7级以上版本

模拟账号是通过fabric ca 动态注册生成


