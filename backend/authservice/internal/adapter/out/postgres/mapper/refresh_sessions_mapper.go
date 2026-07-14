package mapper

import (
	"net/netip"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/maket12/ads-service/authservice/internal/adapter/out/postgres/sqlc"
	"github.com/maket12/ads-service/authservice/internal/domain/model"
	"github.com/maket12/ads-service/authservice/pkg/utils"

	"github.com/google/uuid"
)

func MapRefreshSessionToSQLCCreate(session *model.RefreshSession) sqlc.CreateSessionParams {
	var (
		revokedAt    pgtype.Timestamptz
		revokeReason pgtype.Text
		rotatedFrom  pgtype.UUID
		ip           *netip.Addr
		userAgent    pgtype.Text
	)

	if session.RevokedAt() != nil {
		revokedAt = pgtype.Timestamptz{
			Time:  *session.RevokedAt(),
			Valid: true,
		}
	}

	if session.RevokeReason() != nil {
		revokeReason = pgtype.Text{
			String: session.RevokeReason().String(),
			Valid:  true,
		}
	}
	if session.RotatedFrom() != nil {
		rotatedFrom = pgtype.UUID{
			Bytes: *session.RotatedFrom(),
			Valid: true,
		}
	}

	if session.IP() != nil && *session.IP() != "" {
		parsedIP, _ := netip.ParseAddr(*session.IP())
		ip = &parsedIP
	}

	if session.UserAgent() != nil {
		userAgent = pgtype.Text{String: *session.UserAgent(), Valid: true}
	}

	return sqlc.CreateSessionParams{
		ID:               pgtype.UUID{Bytes: session.ID(), Valid: true},
		AccountID:        pgtype.UUID{Bytes: session.AccountID(), Valid: true},
		RefreshTokenHash: session.RefreshTokenHash(),
		CreatedAt:        pgtype.Timestamptz{Time: session.CreatedAt(), Valid: true},
		ExpiresAt:        pgtype.Timestamptz{Time: session.ExpiresAt(), Valid: true},
		RevokedAt:        revokedAt,
		RevokeReason:     revokeReason,
		RotatedFrom:      rotatedFrom,
		Ip:               ip,
		UserAgent:        userAgent,
	}
}

func MapRefreshSessionToSQLCUpdate(session *model.RefreshSession) sqlc.UpdateSessionParams {
	var (
		revokedAt    pgtype.Timestamptz
		revokeReason pgtype.Text
		rotatedFrom  pgtype.UUID
		ip           *netip.Addr
		userAgent    pgtype.Text
	)

	if session.RevokedAt() != nil {
		revokedAt = pgtype.Timestamptz{
			Time:  *session.RevokedAt(),
			Valid: true,
		}
	}

	if session.RevokeReason() != nil {
		revokeReason = pgtype.Text{
			String: session.RevokeReason().String(),
			Valid:  true,
		}
	}

	if session.RotatedFrom() != nil {
		rotatedFrom = pgtype.UUID{
			Bytes: *session.RotatedFrom(),
			Valid: true,
		}
	}

	if session.IP() != nil {
		ipStr, _ := netip.ParseAddr(*session.IP())
		ip = &ipStr
	}

	if session.UserAgent() != nil {
		userAgent = pgtype.Text{
			String: *session.UserAgent(),
			Valid:  true,
		}
	}

	return sqlc.UpdateSessionParams{
		ID: pgtype.UUID{
			Bytes: session.ID(),
			Valid: true,
		},
		AccountID: pgtype.UUID{
			Bytes: session.AccountID(),
			Valid: true,
		},
		RefreshTokenHash: session.RefreshTokenHash(),
		CreatedAt: pgtype.Timestamptz{
			Time:  session.CreatedAt(),
			Valid: true,
		},
		ExpiresAt: pgtype.Timestamptz{
			Time:  session.ExpiresAt(),
			Valid: true,
		},
		RevokedAt:    revokedAt,
		RevokeReason: revokeReason,
		RotatedFrom:  rotatedFrom,
		Ip:           ip,
		UserAgent:    userAgent,
	}
}

func MapToSQLCRevokeAllForAccount(accountID uuid.UUID, reason *string) sqlc.RevokeAllAccountSessionsParams {
	var revokeReason pgtype.Text
	if reason != nil {
		revokeReason = pgtype.Text{
			String: *reason,
			Valid:  true,
		}
	}

	return sqlc.RevokeAllAccountSessionsParams{
		AccountID:    pgtype.UUID{Bytes: accountID, Valid: true},
		RevokedAt:    pgtype.Timestamptz{Time: time.Now(), Valid: true},
		RevokeReason: revokeReason,
	}
}

func MapToSQLCRevokeDescendants(sessionID uuid.UUID, reason *string) sqlc.RevokeSessionDescendantsParams {
	var revokeReason pgtype.Text
	if reason != nil {
		revokeReason = pgtype.Text{
			String: *reason,
			Valid:  true,
		}
	}
	return sqlc.RevokeSessionDescendantsParams{
		ID:           pgtype.UUID{Bytes: sessionID, Valid: true},
		RevokedAt:    pgtype.Timestamptz{Time: time.Now(), Valid: true},
		RevokeReason: revokeReason,
	}
}

func MapToSQLCRevokeAllForAccountByIPUA(
	accID uuid.UUID,
	ip, userAgent, reason *string,
) sqlc.RevokeAllSessionsForAccountByIPUAParams {
	var (
		revokeReason pgtype.Text
		parsedIP     *netip.Addr
		parsedUA     pgtype.Text
	)

	if reason != nil {
		revokeReason = pgtype.Text{
			String: *reason,
			Valid:  true,
		}
	}

	if ip != nil && *ip != "" {
		parsed, _ := netip.ParseAddr(*ip)
		parsedIP = &parsed
	}

	if userAgent != nil {
		parsedUA = pgtype.Text{String: *userAgent, Valid: true}
	}

	return sqlc.RevokeAllSessionsForAccountByIPUAParams{
		AccountID: pgtype.UUID{
			Bytes: accID,
			Valid: true,
		},
		Ip:        parsedIP,
		UserAgent: parsedUA,
		RevokedAt: pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		},
		RevokeReason: revokeReason,
	}
}

func MapSQLCToRefreshSession(rawSession sqlc.RefreshSession) *model.RefreshSession {
	var (
		revokedAt    *time.Time
		revokeReason *model.RevokeReason
		rotatedFrom  *uuid.UUID
		ip           *string
		userAgent    *string
	)

	if rawSession.RevokedAt.Valid {
		revokedAt = &rawSession.RevokedAt.Time
	}

	if rawSession.RevokeReason.Valid {
		revokeReason = utils.VPtr(model.RevokeReason(rawSession.RevokeReason.String))
	}

	if rawSession.RotatedFrom.Valid {
		rotatedFrom = (*uuid.UUID)(&rawSession.RotatedFrom.Bytes)
	}

	if rawSession.Ip != nil {
		ip = utils.VPtr(rawSession.Ip.String())
	}

	if rawSession.UserAgent.Valid {
		userAgent = &rawSession.UserAgent.String
	}

	return model.RestoreRefreshSession(
		rawSession.ID.Bytes,
		rawSession.AccountID.Bytes,
		rawSession.RefreshTokenHash,
		rawSession.CreatedAt.Time,
		rawSession.ExpiresAt.Time,
		revokedAt,
		revokeReason,
		rotatedFrom,
		ip,
		userAgent,
	)
}

func MapSQLCToListRefreshSession(rawList []sqlc.RefreshSession) []*model.RefreshSession {
	result := make([]*model.RefreshSession, 0, len(rawList))
	for i := range rawList {
		mappedSession := MapSQLCToRefreshSession(rawList[i])
		result = append(result, mappedSession)
	}
	return result
}
