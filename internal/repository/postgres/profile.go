package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lasofiko/LinkUp/internal/domain"
)

type ProfileRepository struct {
	db *sql.DB
}

func NewProfileRepository(db *sql.DB) *ProfileRepository {
	return &ProfileRepository{db: db}
}

type profileRow interface {
	Scan(dest ...any) error
}

type interestQuerier interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

func scanProfile(row profileRow, userID int) (domain.Profile, error) {
	var p domain.Profile
	err := row.Scan(&p.UserID, &p.City, &p.Available, &p.Version)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Profile{}, fmt.Errorf("profile %d: %w", userID, domain.ErrProfileNotFound)
	}
	if err != nil {
		return domain.Profile{}, fmt.Errorf("read profile %d: %w", userID, err)
	}
	return p, nil
}

func readInterests(ctx context.Context, q interestQuerier, userID int) ([]domain.Interest, error) {
	rows, err := q.QueryContext(ctx, `
		SELECT interest
		FROM profile_interests
		WHERE user_id = $1
		ORDER BY position`, userID)
	if err != nil {
		return nil, fmt.Errorf("query interests for profile %d: %w", userID, err)
	}
	defer func() { _ = rows.Close() }()

	interests := make([]domain.Interest, 0)
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			return nil, fmt.Errorf("scan interest for profile %d: %w", userID, err)
		}
		interests = append(interests, domain.Interest(value))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read interests for profile %d: %w", userID, err)
	}
	return interests, nil
}

func (r *ProfileRepository) GetProfile(ctx context.Context, userID int) (domain.Profile, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelRepeatableRead,
		ReadOnly:  true,
	})
	if err != nil {
		return domain.Profile{}, fmt.Errorf("start profile read: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	p, err := scanProfile(tx.QueryRowContext(ctx, `
		SELECT user_id, city, available, version
		FROM profiles
		WHERE user_id = $1`, userID), userID)
	if err != nil {
		return domain.Profile{}, err
	}

	p.Interests, err = readInterests(ctx, tx, userID)
	if err != nil {
		return domain.Profile{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.Profile{}, fmt.Errorf("finish profile read: %w", err)
	}
	return p, nil
}

func (r *ProfileRepository) UpdateProfile(
	ctx context.Context,
	userID int,
	expectedVersion int64,
	city string,
	interests []domain.Interest,
) (domain.Profile, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return domain.Profile{}, fmt.Errorf("start profile update: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	p, err := scanProfile(tx.QueryRowContext(ctx, `
		SELECT user_id, city, available, version
		FROM profiles
		WHERE user_id = $1
		FOR UPDATE`, userID), userID)
	if err != nil {
		return domain.Profile{}, err
	}
	if p.Version != expectedVersion {
		return domain.Profile{}, fmt.Errorf("profile %d has version %d, expected %d: %w",
			userID, p.Version, expectedVersion, domain.ErrVersionConflict)
	}

	err = tx.QueryRowContext(ctx, `
		UPDATE profiles
		SET city = $1, version = version + 1
		WHERE user_id = $2 AND version = $3
		RETURNING city, version`, city, userID, expectedVersion).Scan(&p.City, &p.Version)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Profile{}, fmt.Errorf("profile %d changed during update: %w", userID, domain.ErrVersionConflict)
	}
	if err != nil {
		return domain.Profile{}, fmt.Errorf("update profile %d: %w", userID, err)
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM profile_interests WHERE user_id = $1`, userID); err != nil {
		return domain.Profile{}, fmt.Errorf("remove interests for profile %d: %w", userID, err)
	}
	for i, interest := range interests {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO profile_interests (user_id, interest, position)
			VALUES ($1, $2, $3)`, userID, string(interest), i); err != nil {
			return domain.Profile{}, fmt.Errorf("insert interest %d for profile %d: %w", i, userID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return domain.Profile{}, fmt.Errorf("commit profile %d update: %w", userID, err)
	}
	p.Interests = append([]domain.Interest{}, interests...)
	return p, nil
}
