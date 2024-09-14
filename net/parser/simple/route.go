package simple

import (
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	cproto "github.com/cherry-game/cherry/net/proto"
)

var (
	nodeRouteMap    = map[uint32]*NodeRoute{}
	onDataRouteFunc = ForwardDataRoute
	authMessageId   = uint32(2700)
)

type (
	NodeRoute struct {
		NodeType string
		ActorID  string
		FuncName string
	}

	DataRouteFunc func(agent *Agent, msg *Message, route *NodeRoute)
)

func AddNodeRoute(mid uint32, nodeRoute *NodeRoute) {
	if nodeRoute == nil {
		return
	}

	nodeRouteMap[mid] = nodeRoute
}

func GetNodeRoute(mid uint32) (*NodeRoute, bool) {
	routeActor, found := nodeRouteMap[mid]
	return routeActor, found
}

// 1.(建立连接)客户端建立连接，服务端对应创建一个agent用于处理玩家消息,actorID == sid
// 2.(用户登录)客户端进行帐号登录验证，通过uid绑定当前sid
// 3.(角色登录)客户端通过'beforeLoginRoutes'中的协议完成角色登录
func onClientDataRoute(agent *Agent, msg *Message, route *NodeRoute) {
	session := agent.session

	// agent没有"用户登录",且请求不是第一条协议，则踢掉agent，断开连接
	if !session.IsBind() && msg.MID != authMessageId {
		agent.Kick(-1, true)
		return
	}

	if agent.NodeType() == route.NodeType {
		targetPath := cfacade.NewChildPath(agent.NodeId(), route.ActorID, session.Sid)
		LocalDataRoute(agent, session, msg, route, targetPath)
	} else {
		//gameNodeRoute(agent, session, msg, route)
	}
}

// gameNodeRoute 实现agent路由消息到游戏节点
func gameNodeRoute(agent *Agent, session *cproto.Session, msg *Message, route *NodeRoute) {
	if !session.IsBind() {
		return
	}

	// 如果agent没有完成"角色登录",则禁止转发到game节点
	if !session.Contains("Load") {
		// 如果不是角色登录协议则踢掉agent
		if msg.MID < 1000 {
			agent.Kick(-1, true)
			return
		}
	}

	//serverId := session.GetString(sessionKey.ServerID)
	//if serverId == "" {
	//	return
	//}
	//
	//childId := cstring.ToString(session.Uid)
	//targetPath := cfacade.NewChildPath(serverId, route.HandleName(), childId)
	//pomelo.ClusterLocalDataRoute(agent, session, route, msg, serverId, targetPath)
}

// 消息路由处理函数
func ForwardDataRoute(agent *Agent, msg *Message, route *NodeRoute) {
	session := agent.session
	session.Mid = msg.MID

	// 在当前节点上进行本地路由
	if agent.NodeType() == route.NodeType {
		// 本地转发直接使用session定位子agent actor
		targetPath := cfacade.NewChildPath(agent.NodeId(), route.ActorID, session.Sid)
		LocalDataRoute(agent, session, msg, route, targetPath)
		return
	}

	if !session.IsBind() {
		clog.Warnf("[sid = %s,uid = %d] Session is not bind with UID. failed to forward message.[route = %+v]",
			agent.SID(),
			agent.UID(),
			route,
		)
		return
	}

	// TODO 根据玩家UID定位节点
	member, found := agent.Discovery().Random(route.NodeType)
	if !found {
		return
	}

	var nodeId = member.GetNodeId()
	// 前向路由使用UID拓展
	targetPath := cfacade.NewChildPath(nodeId, route.ActorID, agent.UIDString())
	ClusterLocalDataRoute(agent, session, msg, route, member.GetNodeId(), targetPath)
}

func LocalDataRoute(agent *Agent, session *cproto.Session, msg *Message, nodeRoute *NodeRoute, targetPath string) {
	message := cfacade.GetMessage()
	message.Source = session.AgentPath
	message.Target = targetPath
	message.FuncName = nodeRoute.FuncName
	message.Session = session
	message.Args = msg.Data

	agent.ActorSystem().PostLocal(&message)
}

func ClusterLocalDataRoute(agent *Agent, session *cproto.Session, msg *Message, nodeRoute *NodeRoute, nodeID, targetPath string) error {
	clusterPacket := cproto.GetClusterPacket()
	clusterPacket.SourcePath = session.AgentPath
	clusterPacket.TargetPath = targetPath
	clusterPacket.FuncName = nodeRoute.FuncName
	clusterPacket.Session = session   // agent session
	clusterPacket.ArgBytes = msg.Data // packet -> message -> data

	return agent.Cluster().PublishLocal(nodeID, clusterPacket)
}

func ClusterRemoteDataRoute(agent *Agent, session *cproto.Session, msg *Message, nodeRoute *NodeRoute, nodeID, targetPath string) error {
	clusterPacket := cproto.GetClusterPacket()
	clusterPacket.SourcePath = session.AgentPath
	clusterPacket.TargetPath = targetPath
	clusterPacket.FuncName = nodeRoute.FuncName
	clusterPacket.Session = session   // agent session
	clusterPacket.ArgBytes = msg.Data // packet -> message -> data

	return agent.Cluster().PublishRemote(nodeID, clusterPacket)
}
