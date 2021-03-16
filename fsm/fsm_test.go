package fsm_test

import (
	"context"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/hokonco/kitgo"
	"github.com/hokonco/kitgo/fsm"
	. "github.com/onsi/gomega"
)

func TestMain(m *testing.M) { os.Exit(kitgo.ShouldCover(m.Run(), 1.00)) }

func Test_pkg_fsm(t *testing.T) {
	t.Parallel()
	Expect := NewWithT(t).Expect

	m := fsm.New(StateOff, fsm.States{
		StateOff: fsm.State{&OffAction{}, fsm.Events{
			EventBroken: StateBroken,
			EventRandom: StateRandom,
			EventOn:     StateOn,
		}},
		StateOn: fsm.State{&OnAction{}, fsm.Events{
			EventBroken: StateBroken,
			EventRandom: StateRandom,
			EventOff:    StateOff,
		}},
		StateRandom: fsm.State{&RandomAction{}, fsm.Events{
			EventBroken: StateBroken,
			EventOn:     StateOn,
			EventOff:    StateOff,
		}},
		StateBroken: fsm.State{},
	})
	m.OnTransition = func(current, next fsm.StateType) { t.Log(current, next) }

	ctx := context.Background()
	Expect(m.SendEvent(ctx, EventOn)).NotTo(HaveOccurred())
	Expect(m.SendEvent(ctx, EventOn)).To(HaveOccurred())
	Expect(m.SendEvent(ctx, EventOn)).To(HaveOccurred())
	prev, curr := m.GetStates()
	Expect(prev).NotTo(Equal(curr))
	Expect(prev).NotTo(Equal(""))
	Expect(curr).NotTo(Equal(""))
	Expect(m.SendEvent(ctx, EventOff)).NotTo(HaveOccurred())
	Expect(m.SendEvent(ctx, EventRandom)).NotTo(HaveOccurred())
	Expect(m.SendEvent(ctx, EventBroken)).To(HaveOccurred())
}

const (
	StateOff    = fsm.StateType("Off")
	StateOn     = fsm.StateType("On")
	StateBroken = fsm.StateType("Broken")
	StateRandom = fsm.StateType("Random")
	EventOff    = fsm.EventType("SwitchToOff")
	EventOn     = fsm.EventType("SwitchToOn")
	EventBroken = fsm.EventType("SwitchBroken")
	EventRandom = fsm.EventType("SwitchRandom")
)

// OffAction represents the action executed on entering the Off state.
type OffAction struct{}

func (a *OffAction) Execute(ctx context.Context) fsm.EventType {
	// fmt.Println("The light has been switched off")
	return ""
}

// OnAction represents the action executed on entering the On state.
type OnAction struct{}

func (a *OnAction) Execute(ctx context.Context) fsm.EventType {
	// fmt.Println("The light has been switched on")
	return ""
}

// RandomAction represents the action executed on entering the On state.
type RandomAction struct{}

func (a *RandomAction) Execute(ctx context.Context) fsm.EventType {
	if rand.New(rand.NewSource(time.Now().UnixNano())).Int63()%2 > 0 {
		// fmt.Println("The light has been switched off")
		return EventOff
	}
	// fmt.Println("The light has been switched on")
	return EventOn
}
