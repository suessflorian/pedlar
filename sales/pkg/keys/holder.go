package keys

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"time"

	"github.com/google/uuid"
)

var (
	// ErrNoActiveKeySet is a sentinel that will typically trigger the path of registering
	// a new keyset.
	ErrNoActiveKeySet = errors.New("no active keyset")

	// ErrHolderRevoked is a critical error reserved to force a system crash in the event
	// of an inability to verify external id's and map them to internal ones.
	ErrHolderRevoked = errors.New("holder revoked")
)

type store interface {
	GetActiveKeySet(context.Context) (KeySet, error)
	RevokeKeySet(context.Context, uuid.UUID) error
	KeySets(context.Context, ...uuid.UUID) ([]KeySet, error)
	RegisterKeySet(context.Context, string, string, string) (KeySet, error)
}

type Holder struct {
	store store

	curr  KeySet               // WARNING: concurrent write/read
	chain map[uuid.UUID]KeySet // WARNING: concurrent write/read

	poll   time.Time
	revoke bool
}

func NewHolder(ctx context.Context, store store) (*Holder, error) {
	holder := &Holder{
		store: store,
		chain: make(map[uuid.UUID]KeySet, 0),
		poll:  time.Now().Add(5 * time.Second),
	}

	return holder, holder.setCurrent(ctx)
}

func (k *Holder) sync() {
	if time.Now().After(k.poll) {
		k.poll = k.poll.Add(5 * time.Second)
		err := k.update(context.Background())
		if err != nil {
			slog.Error(fmt.Sprintf("failed to update chain, revoking holder: %v", err.Error()))
			k.revoke = true
			return
		}
	}
}

// holding checks the curr key is active, if so returns it, if not, gets a new active key.
func (k *Holder) holding(ctx context.Context) (KeySet, error) {
	if k.curr.active() {
		return k.curr, nil
	}

	return k.store.GetActiveKeySet(ctx)
}

// update checks that the set of active keys being held are still active.
func (k *Holder) update(ctx context.Context) error {
	var check = make([]uuid.UUID, 0, len(k.chain))
	for _, key := range k.chain {
		if key.active() {
			check = append(check, key.ID)
		}
	}

	keys, err := k.store.KeySets(ctx, check...)
	if err != nil {
		return fmt.Errorf("failed to check for key updates: %w", err)
	}

	var updated = make(map[uuid.UUID]KeySet, len(keys))
	for _, key := range keys {
		updated[key.ID] = key
	}

	maps.Copy(k.chain, updated)
	return k.setCurrent(ctx)
}

// setCurrent ensures the current key is active with respect to what the current
// holding chain specifies as active. If it isn't we first try update with something
// from the chain, if there are no active keys, we retrieve a new one from the store.
func (k *Holder) setCurrent(ctx context.Context) error {
	if key, ok := k.chain[k.curr.ID]; ok && key.active() {
		return nil
	}

	var updatedWithActiveFromChain bool
	for _, key := range k.chain {
		if key.active() {
			k.curr, updatedWithActiveFromChain = key, true
			break
		}
	}

	if updatedWithActiveFromChain {
		return nil
	}

	currentActiveKey, err := k.store.GetActiveKeySet(ctx)
	if errors.Is(err, ErrNoActiveKeySet) {
		encryptionKey, err := generateAESKey()
		if err != nil {
			return fmt.Errorf("could not create new encryption keys: %w", err)
		}

		signingKey, publicKey, err := generateRSAKeyPair()
		if err != nil {
			return fmt.Errorf("could not create new signing keys: %w", err)
		}

		newKey, err := k.store.RegisterKeySet(ctx, signingKey, publicKey, encryptionKey)
		if err != nil {
			return fmt.Errorf("failed to register new keyset: %w", err)
		}
		k.curr = newKey
		return nil
	} else if err != nil {
		return fmt.Errorf("could not get new active keyset: %w", err)
	} else {
		k.curr = currentActiveKey
		return nil
	}
}
