package squirrel

import "testing"

func TestCTEUpdateFromRegressionSuite(t *testing.T) {
	t.Run("multiple ctes and final select continuous dollar numbering", func(t *testing.T) {
		sb := StatementBuilder.PlaceholderFormat(Dollar)

		q := sb.With("first").
			As(sb.Select("a").From("t1").Where("a = ?", 1201)).
			Cte("second").
			As(sb.Select("b").From("t2").Where("b = ?", 1202)).
			Select(sb.Select("a", "b").From("first").Join("second ON first.a = second.b").Where("a > ?", 1203))

		assertNestedPlaceholderSQL(
			t,
			q,
			"WITH first AS (SELECT a FROM t1 WHERE a = $1), second AS (SELECT b FROM t2 WHERE b = $2) SELECT a, b FROM first JOIN second ON first.a = second.b WHERE a > $3",
			[]any{1201, 1202, 1203},
		)
	})

	t.Run("recursive cte anchor recursive and final dollar numbering", func(t *testing.T) {
		sb := StatementBuilder.PlaceholderFormat(Dollar)

		anchor := sb.Select("id", "parent_id").
			From("nodes").
			Where("parent_id IS NULL").
			Where("tenant_id = ?", 1301)
		recursive := sb.Select("n.id", "n.parent_id").
			From("nodes n").
			Join("tree t ON n.parent_id = t.id").
			Where("n.active = ?", true)
		body := ConcatExpr(anchor, " UNION ALL ", recursive)

		q := sb.With("tree").
			Recursive(true).
			As(body).
			Select(sb.Select("id").From("tree").Where("parent_id = ?", 1302))

		assertNestedPlaceholderSQL(
			t,
			q,
			"WITH RECURSIVE tree AS (SELECT id, parent_id FROM nodes WHERE parent_id IS NULL AND tenant_id = $1 UNION ALL SELECT n.id, n.parent_id FROM nodes n JOIN tree t ON n.parent_id = t.id WHERE n.active = $2) SELECT id FROM tree WHERE parent_id = $3",
			[]any{1301, true, 1302},
		)
	})

	t.Run("cte bodies select insert update delete dollar numbering", func(t *testing.T) {
		sb := StatementBuilder.PlaceholderFormat(Dollar)

		selectBody := sb.Select("id").From("source").Where("state = ?", 1401)
		insertBody := sb.Insert("archive").
			Columns("id").
			Select(sb.Select("id").From("source").Where("state = ?", 1402)).
			Suffix("RETURNING id")
		updateBody := sb.Update("source").
			Set("state", "archived").
			Where("id IN (SELECT id FROM archive)").
			Suffix("RETURNING id")
		deleteBody := sb.Delete("archive").Where("id = ?", 1403)

		selectCTE := sb.With("scoped").As(selectBody).Select(sb.Select("id").From("scoped").Where("id > ?", 1404))
		assertNestedPlaceholderSQL(
			t,
			selectCTE,
			"WITH scoped AS (SELECT id FROM source WHERE state = $1) SELECT id FROM scoped WHERE id > $2",
			[]any{1401, 1404},
		)

		insertCTE := sb.With("inserted").As(insertBody).Select(sb.Select("id").From("inserted").Where("id = ?", 1405))
		assertNestedPlaceholderSQL(
			t,
			insertCTE,
			"WITH inserted AS (INSERT INTO archive (id) SELECT id FROM source WHERE state = $1 RETURNING id) SELECT id FROM inserted WHERE id = $2",
			[]any{1402, 1405},
		)

		updateCTE := sb.With("archived").As(insertBody).
			Cte("updated").
			As(updateBody).
			Select(sb.Select("id").From("updated").Where("id = ?", 1406))
		assertNestedPlaceholderSQL(
			t,
			updateCTE,
			"WITH archived AS (INSERT INTO archive (id) SELECT id FROM source WHERE state = $1 RETURNING id), updated AS (UPDATE source SET state = $2 WHERE id IN (SELECT id FROM archive) RETURNING id) SELECT id FROM updated WHERE id = $3",
			[]any{1402, "archived", 1406},
		)

		deleteCTE := sb.With("removed").As(deleteBody).Select(sb.Select("1").From("removed"))
		assertNestedPlaceholderSQL(
			t,
			deleteCTE,
			"WITH removed AS (DELETE FROM archive WHERE id = $1) SELECT 1 FROM removed",
			[]any{1403},
		)
	})

	t.Run("dml cte production pattern continuous dollar numbering", func(t *testing.T) {
		sb := StatementBuilder.PlaceholderFormat(Dollar)

		candidate := sb.Select("id").
			From("jobs").
			Where("status = ?", "pending").
			OrderBy("id").
			Limit(1).
			Suffix("FOR UPDATE SKIP LOCKED")
		picked := sb.Select("id").From("candidate")
		updated := sb.Update("jobs j").
			Set("status", "running").
			From("picked p").
			Where("j.id = p.id").
			Suffix("RETURNING j.id, j.status")

		q := sb.With("candidate").
			As(candidate).
			Cte("picked").
			As(picked).
			Cte("updated").
			As(updated).
			Select(sb.Select("id", "status").From("updated").Where("status = ?", "running"))

		assertNestedPlaceholderSQL(
			t,
			q,
			"WITH candidate AS (SELECT id FROM jobs WHERE status = $1 ORDER BY id LIMIT 1 FOR UPDATE SKIP LOCKED), picked AS (SELECT id FROM candidate), updated AS (UPDATE jobs j SET status = $2 FROM picked p WHERE j.id = p.id RETURNING j.id, j.status) SELECT id, status FROM updated WHERE status = $3",
			[]any{"pending", "running", "running"},
		)
	})

	t.Run("standalone update from subquery cte and returning", func(t *testing.T) {
		sb := StatementBuilder.PlaceholderFormat(Dollar)

		fromSub := sb.Select("account_id", "total").
			From("ledger").
			Where("posted_at >= ?", 1501)

		standalone := sb.Update("accounts a").
			Set("balance", Expr("a.balance + sub.total")).
			FromSelect(fromSub, "sub").
			Where("a.id = sub.account_id").
			Suffix("RETURNING a.id, a.balance")

		assertNestedPlaceholderSQL(
			t,
			standalone,
			"UPDATE accounts a SET balance = a.balance + sub.total FROM (SELECT account_id, total FROM ledger WHERE posted_at >= $1) AS sub WHERE a.id = sub.account_id RETURNING a.id, a.balance",
			[]any{1501},
		)

		wrapped := sb.With("recent").
			As(fromSub).
			Update(
				sb.Update("accounts a").
					Set("balance", Expr("a.balance + recent.total")).
					From("recent").
					Where("a.id = recent.account_id").
					Suffix("RETURNING a.id"),
			)

		assertNestedPlaceholderSQL(
			t,
			wrapped,
			"WITH recent AS (SELECT account_id, total FROM ledger WHERE posted_at >= $1) UPDATE accounts a SET balance = a.balance + recent.total FROM recent WHERE a.id = recent.account_id RETURNING a.id",
			[]any{1501},
		)
	})

	t.Run("cte validation errors", func(t *testing.T) {
		t.Run("without body", func(t *testing.T) {
			_, _, err := With("scope").Select(Select("id").From("users")).ToSql()
			assertValidationError(t, err, "common table expressions statements must have at least one label and subquery")
		})

		t.Run("without final statement", func(t *testing.T) {
			_, _, err := With("scope").As(Select("id").From("users")).ToSql()
			assertValidationError(t, err, "common table expressions must one of the following final statement")
		})
	})
}
