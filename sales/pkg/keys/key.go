package keys

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

type KeySet struct {
	ID uuid.UUID

	EncryptionKey string
	SigningKey    string
	PublicKey     string

	Expiry  time.Time
	Revoked bool
}

func (k KeySet) Active() bool {
	return k.Revoked || time.Now().After(k.Expiry)
}

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

	expiry time.Time
	revoke bool
}

func NewHolder(ctx context.Context, store store) (*Holder, error) {
	holder := &Holder{
		store:  store,
		chain:  make(map[uuid.UUID]KeySet, 0),
		expiry: time.Now().Add(5 * time.Second),
	}

	return holder, holder.setCurrent(ctx)
}

func (k *Holder) sync() {
	if time.Now().After(k.expiry) {
		k.expiry = k.expiry.Add(5 * time.Second)
		err := k.update(context.Background())
		if err != nil {
			slog.Error(fmt.Sprintf("failed to update chain, revoking holder: %v", err.Error()))
			k.revoke = true
			return
		}
	}
}

// ExternalID takes in a internal serial id of an object and returns the current external facing ID.
func (k *Holder) ExternalID(ctx context.Context, object string, internalID int) (string, error) {
	go k.sync()

	key, err := k.holding(ctx)
	if err != nil {
		return "", fmt.Errorf("could not retrieve latest keyset: %w", err)
	}

	encrypted, err := k.encrypt(object, strconv.Itoa(internalID))
	if err != nil || len(encrypted) != 2 {
		return "", fmt.Errorf("failed to encrypt object %q and internal id %q: %w", object, internalID, err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"kid":             key.ID,
		"internal_object": encrypted[0],
		"internal_id":     encrypted[1],
		"exp":             key.Expiry.Unix(),
	})

	block, _ := pem.Decode([]byte(key.SigningKey))
	if block == nil {
		return "", errors.New("could not decode signing key")
	}

	signingKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("could not parse signing key: %w", err)
	}

	id, err := token.SignedString(signingKey)
	if err != nil {
		return "", fmt.Errorf("could not sign token: %w", err)
	}

	return id, nil
}

// InternalID takes in a externalID and returns the internal facing ID.
func (k *Holder) InternalID(ctx context.Context, externalID string) (int, error) {
	if k.revoke {
		return -1, ErrHolderRevoked
	}
	go k.sync()

	return -1, nil
}

// holding checks the curr key is active, if so returns it, if not, gets a new active key.
func (k *Holder) holding(ctx context.Context) (KeySet, error) {
	if k.curr.Active() {
		return k.curr, nil
	}

	return k.store.GetActiveKeySet(ctx)
}

// update checks that the set of active keys being held are still active.
func (k *Holder) update(ctx context.Context) error {
	var check = make([]uuid.UUID, 0, len(k.chain))
	for _, key := range k.chain {
		if key.Active() {
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
	if key, ok := k.chain[k.curr.ID]; ok && key.Active() {
		return nil
	}

	var updatedWithActiveFromChain bool
	for _, key := range k.chain {
		if key.Active() {
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
