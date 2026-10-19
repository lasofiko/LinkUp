package usecase

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/lasofiko/LinkUp/internal/domain"
	"github.com/lasofiko/LinkUp/internal/interaction"
	"github.com/lasofiko/LinkUp/internal/repository/memory"
)

// reactFixture records the complete call sequence and only commits successful saves.
type reactFixture struct {
	calls    []string
	users    map[int]domain.User
	profiles map[int]domain.Profile
	states   map[interaction.Pair]interaction.State
	failAt   string
	err      error
	saves    int
	ctx      context.Context
}

func newReactFixture() *reactFixture {
	return &reactFixture{
		users:    map[int]domain.User{1: {ID: 1, Name: "Соня"}, 2: {ID: 2, Name: "Арина"}},
		profiles: map[int]domain.Profile{1: {UserID: 1, Available: true}, 2: {UserID: 2, Available: true}},
		states:   make(map[interaction.Pair]interaction.State),
	}
}

func (f *reactFixture) record(call string) error {
	f.calls = append(f.calls, call)
	if call == f.failAt {
		return f.err
	}
	return nil
}

func (f *reactFixture) GetByID(id int) (domain.User, error) {
	if err := f.record(fmt.Sprintf("user:%d", id)); err != nil {
		return domain.User{}, err
	}
	u, ok := f.users[id]
	if !ok {
		return domain.User{}, domain.ErrUserNotFound
	}
	return u, nil
}

func (f *reactFixture) GetProfile(ctx context.Context, id int) (domain.Profile, error) {
	f.ctx = ctx
	if err := f.record(fmt.Sprintf("profile:%d", id)); err != nil {
		return domain.Profile{}, err
	}
	p, ok := f.profiles[id]
	if !ok {
		return domain.Profile{}, domain.ErrProfileNotFound
	}
	return p, nil
}

func (f *reactFixture) Load(pair interaction.Pair) (interaction.State, bool, error) {
	if err := f.record("load"); err != nil {
		return interaction.State{}, false, err
	}
	state, found := f.states[pair]
	return state, found, nil
}

func (f *reactFixture) Save(state interaction.State) error {
	f.saves++
	if err := f.record("save"); err != nil {
		return err
	}
	f.states[state.Pair] = state
	return nil
}

func TestReactLikes(t *testing.T) {
	pair := interaction.Pair{FirstUserID: 1, SecondUserID: 2}
	for _, tt := range []struct {
		name     string
		initial  interaction.State
		from, to int
		want     interaction.State
		saves    int
	}{
		{"first", interaction.State{}, 1, 2, interaction.State{Pair: pair, FirstLikesSecond: true}, 1},
		{"first reversed", interaction.State{}, 2, 1, interaction.State{Pair: pair, SecondLikesFirst: true}, 1},
		{"reply", interaction.State{Pair: pair, FirstLikesSecond: true}, 2, 1, interaction.State{Pair: pair, FirstLikesSecond: true, SecondLikesFirst: true}, 1},
		{"repeat", interaction.State{Pair: pair, FirstLikesSecond: true}, 1, 2, interaction.State{Pair: pair, FirstLikesSecond: true}, 0},
		{"repeat mutual", interaction.State{Pair: pair, FirstLikesSecond: true, SecondLikesFirst: true}, 2, 1, interaction.State{Pair: pair, FirstLikesSecond: true, SecondLikesFirst: true}, 0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			f := newReactFixture()
			if tt.initial != (interaction.State{}) {
				f.states[pair] = tt.initial
			}
			ctx := context.Background()
			got, err := NewReactService(f, f, f).React(ctx, tt.from, tt.to)
			if err != nil || got != tt.want {
				t.Fatalf("got %#v, %v; want %#v", got, err, tt.want)
			}
			if f.saves != tt.saves || f.states[pair] != tt.want || len(f.states) != 1 {
				t.Fatalf("saves=%d states=%#v", f.saves, f.states)
			}
			if got.IsMutual() != (tt.want.FirstLikesSecond && tt.want.SecondLikesFirst) {
				t.Fatal("incorrect mutuality")
			}
			wantCalls := []string{fmt.Sprintf("user:%d", tt.from), fmt.Sprintf("user:%d", tt.to), fmt.Sprintf("profile:%d", tt.to), "load"}
			if tt.saves != 0 {
				wantCalls = append(wantCalls, "save")
			}
			if !reflect.DeepEqual(f.calls, wantCalls) || f.ctx != ctx {
				t.Fatalf("calls=%v, want %v; context=%v", f.calls, wantCalls, f.ctx)
			}
		})
	}
}

func TestReactErrorsStopExecution(t *testing.T) {
	root := errors.New("technical failure")
	pair := interaction.Pair{FirstUserID: 1, SecondUserID: 2}
	for _, tt := range []struct {
		name    string
		prepare func(*reactFixture)
		to      int
		cause   error
		calls   []string
		saves   int
	}{
		{"sender missing", func(f *reactFixture) { delete(f.users, 1) }, 2, domain.ErrUserNotFound, []string{"user:1"}, 0},
		{"candidate missing", func(f *reactFixture) { delete(f.users, 2) }, 2, domain.ErrUserNotFound, []string{"user:1", "user:2"}, 0},
		{"profile missing", func(f *reactFixture) { delete(f.profiles, 2) }, 2, domain.ErrProfileNotFound, []string{"user:1", "user:2", "profile:2"}, 0},
		{"sender read", func(f *reactFixture) { f.failAt = "user:1" }, 2, root, []string{"user:1"}, 0},
		{"candidate read", func(f *reactFixture) { f.failAt = "user:2" }, 2, root, []string{"user:1", "user:2"}, 0},
		{"profile read", func(f *reactFixture) { f.failAt = "profile:2" }, 2, root, []string{"user:1", "user:2", "profile:2"}, 0},
		{"load", func(f *reactFixture) { f.failAt = "load" }, 2, root, []string{"user:1", "user:2", "profile:2", "load"}, 0},
		{"save", func(f *reactFixture) { f.failAt = "save" }, 2, root, []string{"user:1", "user:2", "profile:2", "load", "save"}, 1},
		{"unavailable", func(f *reactFixture) { f.profiles[2] = domain.Profile{UserID: 2} }, 2, ErrReactionUnavailable, []string{"user:1", "user:2", "profile:2", "load"}, 0},
		{"self", func(f *reactFixture) {}, 1, interaction.ErrSameUser, []string{"user:1", "user:1", "profile:1"}, 0},
		{"wrong stored pair", func(f *reactFixture) {
			f.states[pair] = interaction.State{Pair: interaction.Pair{FirstUserID: 1, SecondUserID: 3}}
		}, 2, interaction.ErrStatePair, []string{"user:1", "user:2", "profile:2", "load"}, 0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			f := newReactFixture()
			f.err = fmt.Errorf("storage: %w", root)
			tt.prepare(f)
			before := make(map[interaction.Pair]interaction.State)
			for key, state := range f.states {
				before[key] = state
			}
			got, err := NewReactService(f, f, f).React(context.Background(), 1, tt.to)
			if got != (interaction.State{}) || !errors.Is(err, tt.cause) {
				t.Fatalf("got %#v, %v; want zero and %v", got, err, tt.cause)
			}
			if f.saves != tt.saves || !reflect.DeepEqual(f.calls, tt.calls) {
				t.Fatalf("saves=%d calls=%v; want %d %v", f.saves, f.calls, tt.saves, tt.calls)
			}
			if !reflect.DeepEqual(f.states, before) {
				t.Fatal("failed reaction changed state")
			}
		})
	}
}

func TestReactRetryAfterSaveFailure(t *testing.T) {
	f := newReactFixture()
	f.failAt, f.err = "save", errors.New("save failed")
	service := NewReactService(f, f, f)
	got, err := service.React(context.Background(), 1, 2)
	if got != (interaction.State{}) || !errors.Is(err, f.err) || len(f.states) != 0 || f.saves != 1 {
		t.Fatalf("failed save: %#v, %v, saves=%d", got, err, f.saves)
	}
	f.failAt = ""
	got, err = service.React(context.Background(), 1, 2)
	if err != nil || !got.HasLike(1, 2) || got.IsMutual() || f.saves != 2 {
		t.Fatalf("retry: %#v, %v, saves=%d", got, err, f.saves)
	}
	repeated, err := service.React(context.Background(), 1, 2)
	if err != nil || repeated != got || f.saves != 2 {
		t.Fatalf("repeat: %#v, %v, saves=%d", repeated, err, f.saves)
	}
}

func TestReactCanceledContext(t *testing.T) {
	f := newReactFixture()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	got, err := NewReactService(f, f, f).React(ctx, 1, 2)
	if got != (interaction.State{}) || !errors.Is(err, context.Canceled) || len(f.calls) != 0 {
		t.Fatalf("got %#v, %v, calls=%v", got, err, f.calls)
	}
}

func TestReactWithExistingUserAndProfileRepositories(t *testing.T) {
	users := memory.NewUserRepository()
	for _, user := range []domain.User{{ID: 1, Name: "Соня"}, {ID: 2, Name: "Арина"}} {
		if err := users.Create(user); err != nil {
			t.Fatal(err)
		}
	}
	profiles := memory.NewProfiles([]domain.Profile{{UserID: 1, Available: true}, {UserID: 2, Available: true}})
	store := newReactFixture()
	service := NewReactService(users, profiles, store)
	for _, command := range [][2]int{{1, 2}, {2, 1}, {1, 2}, {2, 1}} {
		if _, err := service.React(context.Background(), command[0], command[1]); err != nil {
			t.Fatal(err)
		}
	}
	if store.saves != 2 || len(store.states) != 1 {
		t.Fatalf("saves=%d states=%#v", store.saves, store.states)
	}
	for _, state := range store.states {
		if !state.IsMutual() {
			t.Fatal("expected mutual like")
		}
	}
}
