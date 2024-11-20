# surf

surf is an utils lib for my golang serivce

## RoadMap

### version 1

1. 客户端和节点收发消息
1. 只需要支持单节点部署. 不做路由, 不染色
1. 不做服务发现.写死

## todolist

- [x] gate conn 状态通知 (使用消息队列感觉有延迟)
- [x] 拆分 gate
- [x] 服务注册
- [ ] 服务发现
- [x] 日志改进
- [ ] guandan 玩法
- [ ] lobby 模块
- [ ] auth 改进
- [ ] 消息订阅设计
- [ ] IM 模块
- [ ] gate 多节点支持
- [ ] gate 限流
- [ ] gate 黑名单
