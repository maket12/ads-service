//go:build e2e

package e2e

import (
	"testing"
	"time"
)

func TestDeleteAdIndex_Success(t *testing.T) {
	app := setupE2E(t)
	adID := app.createAdIndex(t, nil, nil, nil, nil, nil, nil)
	app.publishAdDeleted(t, adID)
	app.waitForAdIdxDeleted(t, adID, time.Second*5)
}
