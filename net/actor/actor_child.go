package cherryActor

import (
	"strings"
	"sync"

	cfacade "github.com/cherry-game/cherry/facade"
)

type actorChild struct {
	thisActor   *Actor
	childActors *sync.Map // key:childActorId, value:*actor
}

func newChild(thisActor *Actor) actorChild {
	return actorChild{
		thisActor:   thisActor,
		childActors: &sync.Map{},
	}
}

func (p *actorChild) onStop() {
	p.childActors.Range(func(key, value any) bool {
		if childActor, ok := value.(*Actor); ok {
			childActor.Exit()
		}
		return true
	})

	//p.childActors = nil
	p.thisActor = nil
}

// Create bind new child actor handler with child id
func (p *actorChild) Create(childID string, handler cfacade.IActorHandler) (cfacade.IActor, error) {
	// not allow to create child in parent
	if p.thisActor.path.IsChild() {
		return nil, ErrForbiddenCreateChildActor
	}
	// must need none empty id
	if strings.TrimSpace(childID) == "" {
		return nil, ErrActorIDIsNil
	}

	// not allowed repeated id
	if thisActor, ok := p.Get(childID); ok {
		return thisActor, nil
	}
	// create Actor struct with handler
	childActor, err := newActor(p.thisActor.ActorID(), childID, handler, p.thisActor.system)
	if err != nil {
		return nil, err
	}
	// insert to map
	p.childActors.Store(childID, &childActor)
	// call onInit and star loop
	go childActor.run()

	return &childActor, nil
}

// Get return child actor with childID
func (p *actorChild) Get(childID string) (cfacade.IActor, bool) {
	return p.GetActor(childID)
}

// GetActor 返回按childID指定的下一级 actorChild
// return nil if childID not exist
func (p *actorChild) GetActor(childID string) (*Actor, bool) {
	if actorValue, ok := p.childActors.Load(childID); ok {
		actor, found := actorValue.(*Actor)
		return actor, found
	}

	return nil, false
}

// Remove removes actorChild with childID
func (p *actorChild) Remove(childID string) {
	p.childActors.Delete(childID)
}

// Each invoke func fn for all child actors
func (p *actorChild) Each(fn func(cfacade.IActor)) {
	p.childActors.Range(func(key, value any) bool {
		if actor, found := value.(*Actor); found {
			fn(actor)
		}
		return true
	})
}
