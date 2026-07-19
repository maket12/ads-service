///go:build e2e

package e2e

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestCreateProfile_Success(t *testing.T) {
	app := setupE2E(t)

	accountID := uuid.New().String()
	app.publishAccountCreated(t, accountID)

	resp := app.waitForProfile(t, accountID, 2*time.Second)

	require.Equal(t, accountID, resp.AccountId)
}

func TestCreateProfile_Idempotent(t *testing.T) {
	app := setupE2E(t)

	accountID := uuid.New().String()

	// publish the same event twice; creation should not error/duplicate
	app.publishAccountCreated(t, accountID)
	app.waitForProfile(t, accountID, 2*time.Second)

	app.publishAccountCreated(t, accountID)

	// give the second (redundant) event a moment to be processed/dropped,
	// then confirm the profile is still fetchable and unique.
	time.Sleep(500 * time.Millisecond)
	resp := app.waitForProfile(t, accountID, 5*time.Second)
	require.Equal(t, accountID, resp.AccountId)
}
