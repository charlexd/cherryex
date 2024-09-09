package simple

import (
	"encoding/binary"
	"net"
	"time"

	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	cactor "github.com/cherry-game/cherry/net/actor"
	cproto "github.com/cherry-game/cherry/net/proto"
	"github.com/nats-io/nuid"
	"go.uber.org/zap/zapcore"
)

type (
	Actor struct {
		cactor.Base
		agentActorID   string
		connectors     []cfacade.IConnector
		onNewAgentFunc OnNewAgentFunc
	}
)

type OnNewAgentFunc func(newAgent *Agent)

func NewActor(agentActorID string) *Actor {
	if agentActorID == "" {
		panic("agentActorID is empty.")
	}

	parser := &Actor{
		agentActorID: agentActorID,
		connectors:   make([]cfacade.IConnector, 0),
	}

	return parser
}

// OnInit Actor初始化前触发该函数
func (p *Actor) OnInit() {
	p.Remote().Register(ResponseFuncName, p.response)
}

func (p *Actor) Load(app cfacade.IApplication) {
	if len(p.connectors) < 1 {
		panic("Connectors is nil. Please call the AddConnector(...) method add IConnector.")
	}

	//  Create agent SimpleActor
	if _, err := app.ActorSystem().CreateActor(p.agentActorID, p); err != nil {
		clog.Panicf("Create agent SimpleActor fail. err = %+v", err)
	}

	for _, connector := range p.connectors {
		connector.OnConnect(p.defaultOnConnectFunc)
		go connector.Start() // start connector!
	}
}

func (p *Actor) AddConnector(connector cfacade.IConnector) {
	p.connectors = append(p.connectors, connector)
}

func (p *Actor) Connectors() []cfacade.IConnector {
	return p.connectors
}

func (p *Actor) AddNodeRoute(mid uint32, nodeRoute *NodeRoute) {
	AddNodeRoute(mid, nodeRoute)
}

// defaultOnConnectFunc 创建新连接时，通过当前agentActor创建child agent SimpleActor
func (p *Actor) defaultOnConnectFunc(conn net.Conn) {
	session := &cproto.Session{
		Sid:       nuid.Next(),
		AgentPath: p.Path().String(),
		Data:      map[string]string{},
	}

	agent := NewAgent(p.App(), conn, session)

	if p.onNewAgentFunc != nil {
		p.onNewAgentFunc(&agent)
	}

	BindSID(&agent)
	agent.Run()
}

func (p *Actor) SetOnNewAgent(fn OnNewAgentFunc) {
	p.onNewAgentFunc = fn
}

func (p *Actor) SetHeartbeatTime(t time.Duration) {
	SetHeartbeatTime(t)
}

func (p *Actor) SetWriteBacklog(backlog int) {
	SetWriteBacklog(backlog)
}

func (p *Actor) SetEndian(e binary.ByteOrder) {
	SetEndian(e)
}

func (*Actor) SetOnDataRoute(fn DataRouteFunc) {
	if fn != nil {
		onDataRouteFunc = fn
	}
}

func (p *Actor) response(rsp *cproto.PomeloResponse) {
	agent, found := GetAgent(rsp.Sid)
	if !found {
		if clog.PrintLevel(zapcore.DebugLevel) {
			clog.Debugf("[response] Not found agent. [rsp = %+v]", rsp)
		}
		return
	}

	agent.Response(rsp.Mid, rsp.Data)
}
