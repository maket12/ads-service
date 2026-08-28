//go:build e2e

package e2e

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestDeleteProfile_Success(t *testing.T) {
	app := setupE2E(t)
	accountID := uuid.New().String()

	app.publishAccountCreated(t, accountID)
	profile := app.waitForProfileCreated(t, accountID, 5*time.Second)
	require.NotNil(t, profile)

	app.publishAccountDeleted(t, accountID)
	app.waitForProfileDeleted(t, accountID, 5*time.Second)
}
