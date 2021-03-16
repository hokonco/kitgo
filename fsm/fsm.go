package fsm // https://venilnoronha.io/a-simple-state-machine-framework-in-go

import (
	"context"
	"errors"
	"sync"
)

// StateType represents an extensible state type in the state machine.
type StateType string

// EventType represents an extensible event type in the state machine.
type EventType string

// Action represents the action to be executed in a given state.
type Action interface {
	Execute(ctx context.Context) EventType
}

// Events represents a mapping of events and states.
type Events map[EventType]StateType

// State binds a state with an action and a set of events it can handle.
type State struct {
	Action
	Events
}

// States represents a mapping of states and their implementations.
type States map[StateType]State

// Machine represents the state machine.
type Machine struct {
	mu     sync.Mutex // mutex ensures that only 1 event is processed by the state machine at any given time.
	prev   StateType  // Previous represents the previous state.
	curr   StateType  // Current represents the current state.
	states States     // States holds the configuration of states and events handled by the state machine.

	// OnTransition called when transitioning from current StateType to the nextStateType
	OnTransition func(curr, next StateType)
}

// New create new finite-state Machine with initial StateType and States mapping.
func New(curr StateType, states States) *Machine { return &Machine{curr: curr, states: states} }

// nextState get next StateType and State, return error on invalid State.
func (s *Machine) nextState(event EventType) (StateType, *State, error) {
	if state, ok := s.states[s.curr]; ok && len(state.Events) > 0 {
		if next, ok := state.Events[event]; ok {
			if state, ok = s.states[next]; ok && state.Action != nil {
				return next, &state, nil
			}
			return next, nil, errors.New("fsm: invalid state")
		}
	}
	return "", nil, errors.New("fsm: rejected event")
}

// SendEvent sends an event to the state machine.
func (s *Machine) SendEvent(ctx context.Context, event EventType) (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for {
		// Determine the next state for the event given the machine's current state.
		next, state, err := s.nextState(event)
		if err != nil {
			return err
		}

		// Transition over to the next state when event is valid.
		if s.OnTransition != nil {
			s.OnTransition(s.curr, next)
		}
		s.prev, s.curr = s.curr, next

		// Execute the next state's action and loop over again if the event returned is not a no-op.
		if nextEvent := state.Action.Execute(ctx); nextEvent != "" {
			event = nextEvent
			continue
		}
		return nil
	}
}

// GetStates tuple of previous and current state
func (s *Machine) GetStates() (prev, curr StateType) { return s.prev, s.curr }
