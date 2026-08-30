///go:build e2e

package e2e

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCreateAdIndex_Success(t *testing.T) {
	app := setupE2E(t)
	accountID := uuid.New().String()
	app.publishAdPublished(t, accountID)
	app.waitForAdIdxCreated(t, accountID, 5*time.Second)
}
