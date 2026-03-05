package mapper

import (
	"database/sql"
	"net"
	"time"

	"github.com/maket12/ads-service/authservice/internal/adapter/out/postgres/sqlc"
	"github.com/maket12/ads-service/authservice/internal/domain/model"

	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
)

func MapRefreshSessionToSQLCCreate(session *model.RefreshSession) sqlc.CreateRefreshSessionParams {
	var (
		revokedAt    sql.NullTime
		revokeReason sql.NullString
		rotatedFrom  uuid.NullUUID
		ip           pqtype.Inet
		userAgent    sql.NullString
	)
	if session.RevokedAt() != nil {
		revokedAt = sql.NullTime{
			Time:  *session.RevokedAt(),
			Valid: true,
		}
	}
	if session.RevokeReason() != nil {
		revokeReason = sql.NullString{
			String: *session.RevokeReason(),
			Valid:  true,
		}
	}
	if session.RotatedFrom() != nil {
		rotatedFrom = uuid.NullUUID{
			UUID:  *session.RotatedFrom(),
			Valid: true,
		}
	}
	if session.IP() != nil && *session.IP() != "" {
		parsedIP := net.ParseIP(*session.IP())
		if parsedIP != nil {
			mask := net.CIDRMask(32, 32)
			if parsedIP.To4() == nil {
				mask = net.CIDRMask(128, 128)
			}

			ip = pqtype.Inet{
				IPNet: net.IPNet{
					IP:   parsedIP,
					Mask: mask,
				},
				Valid: true,
			}
		} else {
			ip = pqtype.Inet{Valid: false}
		}
	} else {
		// IP = nil
		ip = pqtype.Inet{Valid: false}
	}
	if session.UserAgent() != nil {
		userAgent = sql.NullString{
			String: *session.UserAgent(),
			Valid:  true,
		}
	}

	return sqlc.CreateRefreshSessionParams{
		ID:               session.ID(),
		AccountID:        session.AccountID(),
		RefreshTokenHash: session.RefreshTokenHash(),
		CreatedAt:        session.CreatedAt(),
		ExpiresAt:        session.ExpiresAt(),
		RevokedAt:        revokedAt,
		RevokeReason:     revokeReason,
		RotatedFrom:      rotatedFrom,
		Ip:               ip,
		UserAgent:        userAgent,
	}
}

func MapSQLCToRefreshSession(rawSession sqlc.RefreshSession) *model.RefreshSession {
	var (
		revokedAt    *time.Time
		revokeReason *string
		rotatedFrom  *uuid.UUID
		ip           *string
		userAgent    *string
	)
	if rawSession.RevokedAt.Valid {
		revokedAt = &rawSession.RevokedAt.Time
	}
	if rawSession.RevokeReason.Valid {
		revokeReason = &rawSession.RevokeReason.String
	}
	if rawSession.RotatedFrom.Valid {
		rotatedFrom = &rawSession.RotatedFrom.UUID
	}
	if rawSession.Ip.Valid {
		s := rawSession.Ip.IPNet.IP.String()
		ip = &s
	}
	if rawSession.UserAgent.Valid {
		userAgent = &rawSession.UserAgent.String
	}

	return model.RestoreRefreshSession(
		rawSession.ID,
		rawSession.AccountID,
		rawSession.RefreshTokenHash,
		rawSession.CreatedAt,
		rawSession.ExpiresAt,
		revokedAt,
		revokeReason,
		rotatedFrom,
		ip,
		userAgent,
	)
}
