package postgres

import (
	"context"
	"database/sql"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// BatchUpdatePPAssocs attempts to batch update a set of product to product associations.
func (m *PgModel) BatchUpdatePPAssocs(ctx context.Context, ppAssocsGroupUUID, productFromUUID string, productUUIDs []string) error {
	contextLogger := log.WithContext(ctx)

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "postgres: db.BeginTx")
	}
	contextLogger.Debugf("postgres: db begin transaction")

	// 1. Check the product to product association group exists.
	q1 := "SELECT id FROM pp_assoc_group WHERE uuid = $1"
	var ppAssocsGroupID int
	err = tx.QueryRowContext(ctx, q1, ppAssocsGroupUUID).Scan(&ppAssocsGroupID)
	if err != nil {
		if err == sql.ErrNoRows {
			tx.Rollback()
			return ErrPPAssocGroupNotFound
		}
		tx.Rollback()
		return errors.Wrapf(err, "postgres: query row context failed for q1=%q", q1)
	}

	// 2. Check the product with product_from exists.
	q2 := "SELECT id FROM product WHERE uuid = $1"
	var productFromID int
	err = tx.QueryRowContext(ctx, q2, productFromUUID).Scan(&productFromID)
	if err != nil {
		if err == sql.ErrNoRows {
			tx.Rollback()
			return ErrProductNotFound
		}
		tx.Rollback()
		return errors.Wrapf(err, "postgres: query row context failed for q2=%q", q2)
	}

	// 3. Check for missing products in the to set.
	q3 := `
		SELECT
		  id, uuid
		FROM product WHERE uuid = ANY($1::UUID[])
	`
	// TODO: sanitise productUUIDs
	rows, err := m.db.QueryContext(ctx, q3, "{"+strings.Join(productUUIDs, ",")+"}")
	if err != nil {
		return errors.Wrapf(err, "postgres: m.db.QueryContext(ctx) failed q3=%q", q3)
	}
	defer rows.Close()

	pmap := make(map[string]*ProductRow)
	for rows.Next() {
		var p ProductRow
		if err = rows.Scan(&p.id, &p.UUID); err != nil {
			return errors.Wrap(err, "postgres: scan failed")
		}
		pmap[p.UUID] = &p
	}
	if err = rows.Err(); err != nil {
		return errors.Wrap(err, "postgres: rows.Err()")
	}

	// check for missing products
	for _, puuid := range productUUIDs {
		if _, ok := pmap[puuid]; !ok {
			// TODO: differentiate between missing product_from
			// and a missing product to.
			return ErrProductNotFound
		}
	}

	// 4. Delete any old associations for this group and product_from combo.
	q4 := "DELETE FROM pp_assoc WHERE pp_assoc_group_id = $1 AND product_from = $2"
	_, err = tx.ExecContext(ctx, q4, ppAssocsGroupID, productFromID)
	if err != nil {
		tx.Rollback()
		return errors.Wrapf(err, "postgres: exec context q4=%q", q4)
	}

	// 5. Insert the new associations for this group and product_from combo.
	q5 := `
		INSERT INTO pp_assoc (pp_assoc_group_id, product_from, product_to)
		VALUES ($1, $2, $3)
		RETURNING id
	`
	stmt5, err := tx.PrepareContext(ctx, q5)
	if err != nil {
		tx.Rollback()
		return errors.Wrapf(err, "postgres: tx prepare for q5=%q", q5)
	}
	defer stmt5.Close()

	for _, puuid := range productUUIDs {
		var ppAssocID int
		if err := stmt5.QueryRowContext(ctx, ppAssocsGroupID, productFromID, pmap[puuid].id).Scan(&ppAssocID); err != nil {
			tx.Rollback()
			return errors.Wrapf(err, "postgres: stmt4.QueryRowContext(ctx, ppAssocsGroupID=%d, productFromID=%d, pmap[puuid].id=%d) failed", ppAssocsGroupID, productFromID, pmap[puuid].id)
		}
	}

	if err = tx.Commit(); err != nil {
		return errors.Wrap(err, "postgres: tx.Commit")
	}

	contextLogger.Debugf("postgres: db commit succeeded")
	return nil
}
