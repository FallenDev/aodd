package game

import (
	"bufio"
	"encoding/gob"
	"io"
	"log"

	"github.com/ghthor/aodd/game/datastore"
	"github.com/ghthor/engine/rpg2d"
	"github.com/ghthor/engine/rpg2d/entity"
	"golang.org/x/net/websocket"
)

// Used to determine the next type that's in the
// buffer so we can decode it into a real value.
// We'll decode an encoded type and switch on its
// value so we'll have the correct value to decode
// into.
type EncodedType int

//go:generate stringer -type=EncodedType
const (
	ET_ERROR EncodedType = iota
	ET_DISCONNECT

	ET_REQ_LOGIN
	ET_REQ_CREATE

	ET_RESP_AUTH_FAILED
	ET_RESP_ACTOR_EXISTS
	ET_RESP_ACTOR_DOESNT_EXIST

	ET_RESP_LOGIN_SUCCESS
	ET_RESP_CREATE_SUCCESS

	ET_WORLD_STATE
	ET_WORLD_STATE_DIFF

	ET_REQ_MOVE
	ET_REQ_USE
	ER_REQ_CHAT
)

type ReqLogin struct{ Name, Password string }
type ReqCreate struct{ Name, Password string }

type RespAuthFailed struct{ Name string }
type RespActorExists struct{ Name string }
type RespActorDoesntExist struct{ Name, Password string }

const RespDisconnect = "disconnected"

func init() {
	// Pre login Request/Response types
	gob.Register(ReqLogin{})
	gob.Register(ReqCreate{})

	gob.Register(RespAuthFailed{})
	gob.Register(RespActorExists{})
	gob.Register(RespActorDoesntExist{})

	// ActorEntityState used for login/create success
	gob.Register(ActorEntityState{})

	// Engine types
	gob.Register(rpg2d.WorldState{})
	gob.Register(rpg2d.WorldStateDiff{})
	gob.Register(rpg2d.TerrainMapState{})

	// Other entity states
	gob.Register(SayEntityState{})
	gob.Register(AssailEntityState{})

	// Cmd Requests. They have no responses.
	gob.Register(MoveRequest{})
	gob.Register(UseRequest{})
	gob.Register(ChatRequest{})
}

type GobConn interface {
	EncodeAndSend(EncodedType, interface{}) error
	ReadNextType() (EncodedType, error)
	Decode(interface{}) error
}

type gobConn struct {
	enc  *gob.Encoder
	wbuf *bufio.Writer

	*gob.Decoder
}

func (c gobConn) EncodeAndSend(t EncodedType, ev interface{}) error {
	err := c.enc.Encode(t)
	if err != nil {
		return err
	}

	err = c.enc.Encode(ev)
	if err != nil {
		return err
	}

	return c.wbuf.Flush()
}

func (c gobConn) ReadNextType() (t EncodedType, err error) {
	err = c.Decoder.Decode(&t)
	return
}

func NewGobConn(rw io.ReadWriter) GobConn {
	wbuf := bufio.NewWriter(rw)
	enc := gob.NewEncoder(wbuf)
	dec := gob.NewDecoder(rw)

	return gobConn{
		enc:  enc,
		wbuf: wbuf,

		Decoder: dec,
	}
}

type serverConn struct {
	GobConn

	sim       rpg2d.RunningSimulation
	datastore datastore.Datastore
	nextId    func() entity.Id

	actor *actor
}

type ActorConn interface {
	Run() error
}

type stateFn func() (stateFn, error)

func (c *serverConn) handleLogin() (stateFn, error) {
	eType, err := c.ReadNextType()
	if err != nil {
		return nil, err
	}

	switch eType {
	case ET_REQ_LOGIN:
		return c.handleLoginReq, nil
	case ET_REQ_CREATE:
		return c.handleCreateReq, nil

	default:
		log.Println("unexpected encoded type: ", eType)
	}

	return c.handleLogin, nil
}

func (c *serverConn) handleLoginReq() (stateFn, error) {
	var r ReqLogin
	err := c.Decode(&r)
	if err != nil {
		return nil, err
	}

	actor, exists := c.datastore.ActorExists(r.Name)
	if !exists {
		err := c.EncodeAndSend(ET_RESP_ACTOR_DOESNT_EXIST, RespActorDoesntExist{
			r.Name, r.Password,
		})
		if err != nil {
			return nil, err
		}

		return c.handleLogin, nil
	}

	if !actor.Authenticate(r.Name, r.Password) {
		err := c.EncodeAndSend(ET_RESP_AUTH_FAILED, RespAuthFailed{r.Name})
		if err != nil {
			return nil, err
		}

		return c.handleLogin, nil
	}

	c.login(actor)

	err = c.EncodeAndSend(ET_RESP_LOGIN_SUCCESS, c.actor.ToState())
	if err != nil {
		return nil, err
	}

	return c.handleInputReq, nil
}

func (c *serverConn) handleCreateReq() (stateFn, error) {
	var r ReqCreate
	err := c.Decode(&r)
	if err != nil {
		return nil, err
	}

	_, exists := c.datastore.ActorExists(r.Name)
	if exists {
		err := c.EncodeAndSend(ET_RESP_ACTOR_EXISTS, RespActorExists{r.Name})
		if err != nil {
			return nil, err
		}

		return c.handleLogin, nil
	}

	actor, err := c.datastore.AddActor(r.Name, r.Password)
	if err != nil {
		// TODO Instead of terminating the connection here
		//      we should retry contacting the database a
		//      few times
		return nil, err
	}

	c.login(actor)

	err = c.EncodeAndSend(ET_RESP_CREATE_SUCCESS, c.actor.ToState())
	if err != nil {
		return nil, err
	}

	return c.handleInputReq, nil
}

func (c *serverConn) handleInputReq() (stateFn, error) {
	_, err := c.ReadNextType()
	if err != nil {
		return nil, err
	}

	return c.handleInputReq, nil
}

func (c *serverConn) login(dsactor datastore.Actor) {
	c.actor = NewActor(c.nextId(), dsactor, c)
	c.sim.ConnectActor(c.actor)
}

func (c serverConn) Run() (err error) {
	f := c.handleLogin
	for f != nil && err == nil {
		f, err = f()
	}

	return
}

func (c serverConn) WriteWorldState(s rpg2d.WorldState) error {
	return c.EncodeAndSend(ET_WORLD_STATE, s)
}

func (c serverConn) WriteWorldStateDiff(s rpg2d.WorldStateDiff) error {
	return c.EncodeAndSend(ET_WORLD_STATE_DIFF, s)
}

func NewActorGobConn(rw io.ReadWriter, sim rpg2d.RunningSimulation, ds datastore.Datastore, eIdGen func() entity.Id) ActorConn {
	return serverConn{
		GobConn:   NewGobConn(rw),
		sim:       sim,
		datastore: ds,
		nextId:    eIdGen,
	}
}

func newGobWebsocketHandler(sim rpg2d.RunningSimulation, ds datastore.Datastore, eIdGen func() entity.Id) websocket.Handler {
	return func(ws *websocket.Conn) {
		ws.PayloadType = websocket.BinaryFrame

		c := NewActorGobConn(ws, sim, ds, eIdGen)

		// Blocks until the connection is disconnected
		err := c.Run()

		if err != nil {
			log.Printf("packet handler terminated: %v", err)
		}
	}
}
