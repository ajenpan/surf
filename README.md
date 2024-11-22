# surf

services for poker games
棋牌游戏后台服务集

## RoadMap

目标是完成以下组件开发:

- uauth: 用户认证
  - 用户注册
  - 用户登录
- gate: 网关
- lobby: 大厅
- battle: 对局
- mailbox: 邮件
- IM: 及时消息
- GM: 游戏管理
- race: 比赛

- 游戏玩法:
  - ddz: 斗地主
  - guandan: 掼蛋

### version 1

### todolist

- [x] gate conn 状态通知 (使用消息队列需要引入第三方消息队列)
- [x] 拆分 gate
- [x] 服务注册
- [ ] 服务发现
- [x] 日志改进
- [ ] 服务配置
- [ ] 节点之间通信, 暂时使用 gate 转发
- [ ] guandan 玩法
- [ ] lobby 模块
- [ ] auth 改进
- [ ] 消息订阅设计
- [ ] IM 模块
- [ ] gate 多节点支持
- [ ] gate 限流
- [ ] gate 黑名单
