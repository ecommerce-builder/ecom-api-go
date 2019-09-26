package postgres

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// ErrPPAssocNotFound error
var ErrPPAssocNotFound = errors.New("postgres: product to product association not found")

// PPAssocJoinRow contains a single row from the pp_assoc table
// joined with the pp_assoc_group and product table.
type PPAssocJoinRow struct {
	id               int
	UUID             string
	ppAssocGroupID   int
	PPAssocGroupUUID string
	productFrom      int
	ProductFromUUID  string
	productTo        int
	ProductToUUID    string
	Created          time.Time
	Modified         time.Time
}

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

	// 2. Check the product with product_from_id exists.
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
			// TODO: differentiate between missing product_from_id
			// and a missing product to.
			return ErrProductNotFound
		}
	}

	// 4. Delete any old associations for this group and product_from_id combo.
	q4 := "DELETE FROM pp_assoc WHERE pp_assoc_group_id = $1 AND product_from_id = $2"
	_, err = tx.ExecContext(ctx, q4, ppAssocsGroupID, productFromID)
	if err != nil {
		tx.Rollback()
		return errors.Wrapf(err, "postgres: exec context q4=%q", q4)
	}

	// 5. Insert the new associations for this group and product_from_id combo.
	q5 := `
		INSERT INTO pp_assoc (pp_assoc_group_id, product_from_id, product_to_id)
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

// GetPPAssocs returns a list of pp_assoc rows joined with the
// pp_assoc_group and product tables
func (m *PgModel) GetPPAssocs(ctx context.Context, ppAssocGroupUUID, productFromUUID string) ([]*PPAssocJoinRow, error) {
	contextLogger := log.WithContext(ctx)
	contextLogger.Debugf("postgres: GetPPAssocs(ctx, ppAssocGroupUUID=%q, productFromUUID=%q)", ppAssocGroupUUID, productFromUUID)

	// 1. Check the product to product associations group exists.
	q1 := "SELECT id FROM pp_assoc_group WHERE uuid = $1"
	var ppAssocGroupID int
	err := m.db.QueryRowContext(ctx, q1, ppAssocGroupUUID).Scan(&ppAssocGroupID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrPPAssocGroupNotFound
		}
		return nil, errors.Wrapf(err, "postgres: query row context failed for q1=%q", q1)
	}

	// 2. If the product from uuid is passed, then check the product_from_id
	var productFromID int
	if productFromUUID != "" {
		q2 := "SELECT id FROM product WHERE uuid = $1"
		err = m.db.QueryRowContext(ctx, q2, productFromUUID).Scan(&productFromID)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, ErrProductNotFound
			}
			return nil, errors.Wrapf(err, "postgres: m.db.QueryRowContext(ctx, q2=%q, productFromUUID=%q) failed", q2, productFromUUID)
		}
	}

	// 3. Get a list of all product to product associations in this group
	// joined with the product table.
	var where string
	where = "WHERE a.pp_assoc_group_id = $1"
	if productFromUUID != "" {
		where = where + " AND a.product_from_id = $2"
	}

	q3 := `
		SELECT
		  a.id, a.uuid, a.pp_assoc_group_id, a.product_from_id, p1.uuid as product_from_uuid,
		  a.product_to_id, p2.uuid as product_to_uuid, a.created, a.modified
		FROM
		  pp_assoc AS a
		INNER JOIN product AS p1
		  ON a.product_from_id = p1.id
		INNER JOIN product AS p2
		  ON a.product_to_id = p2.id
		%WHERECLAUSE%
	 `
	q3 = strings.Replace(q3, "%WHERECLAUSE%", where, 1)
	var rows *sql.Rows
	if productFromUUID != "" {
		rows, err = m.db.QueryContext(ctx, q3, ppAssocGroupID, productFromID)
	} else {
		rows, err = m.db.QueryContext(ctx, q3, ppAssocGroupID)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "postgres: m.db.QueryContext(ctx) failed")
	}
	defer rows.Close()

	list := make([]*PPAssocJoinRow, 0, 128)
	for rows.Next() {
		var a PPAssocJoinRow
		if err = rows.Scan(&a.id, &a.UUID, &a.ppAssocGroupID, &a.productFrom, &a.ProductFromUUID, &a.productTo, &a.ProductToUUID, &a.Created, &a.Modified); err != nil {
			return nil, errors.Wrap(err, "postgres: scan failed")
		}
		a.PPAssocGroupUUID = ppAssocGroupUUID
		list = append(list, &a)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "postgres: rows.Err()")
	}

	return list, nil
}

// DeletePPAssoc attempts to delete a product to product association row
// from the pp_assoc table. Returns `ErrPPAssocNotFound` if the row is
// not matched.
func (m *PgModel) DeletePPAssoc(ctx context.Context, ppAssocUUID string) error {
	// 1. Check the product to product assocations exists
	q1 := "SELECT id FROM pp_assoc WHERE uuid = $1"
	var ppAssocID int
	err := m.db.QueryRowContext(ctx, q1, ppAssocUUID).Scan(&ppAssocID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrPPAssocNotFound
		}
		return errors.Wrapf(err, "postgres: query row context failed for q1=%q", q1)
	}

	// 2. Delete the association
	q2 := "DELETE FROM pp_assoc WHERE id = $1"
	_, err = m.db.ExecContext(ctx, q2, ppAssocID)
	if err != nil {
		return errors.Wrapf(err, "postgres: exec context q2=%q", q2)
	}
	return nil
}
