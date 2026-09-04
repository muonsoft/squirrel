package squirrel

import "testing"

// statusIDs is a named slice type used to prove In/NotIn accept typed slices.
type statusIDs []int

func TestExprHelpersRegressionSuite(t *testing.T) {
	t.Run("exists and not exists nested dollar numbering", func(t *testing.T) {
		sb := StatementBuilder.PlaceholderFormat(Dollar)
		inner := sb.Select("1").From("orders").Where("customer_id = ?", 301)
		outer := sb.Select("id").From("customers").Where(Exists(inner)).Where("region = ?", 302)

		assertNestedPlaceholderSQL(
			t,
			outer,
			"SELECT id FROM customers WHERE EXISTS (SELECT 1 FROM orders WHERE customer_id = $1) AND region = $2",
			[]any{301, 302},
		)

		outer = sb.Select("id").From("customers").Where(NotExists(inner)).Where("region = ?", 302)
		assertNestedPlaceholderSQL(
			t,
			outer,
			"SELECT id FROM customers WHERE NOT EXISTS (SELECT 1 FROM orders WHERE customer_id = $1) AND region = $2",
			[]any{301, 302},
		)
	})

	t.Run("comparison family nested dollar numbering", func(t *testing.T) {
		sb := StatementBuilder.PlaceholderFormat(Dollar)
		inner := sb.Select("total").From("invoices").Where("account_id = ?", 401)

		cases := []struct {
			name string
			cond Sqlizer
			op   string
			val  int
		}{
			{"equal", Equal(inner, 500), "=", 500},
			{"not equal", NotEqual(inner, 501), "<>", 501},
			{"greater", Greater(inner, 502), ">", 502},
			{"greater or equal", GreaterOrEqual(inner, 503), ">=", 503},
			{"less", Less(inner, 504), "<", 504},
			{"less or equal", LessOrEqual(inner, 505), "<=", 505},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				q := sb.Select("id").From("accounts").Where(tc.cond).Where("active = ?", 406)
				assertNestedPlaceholderSQL(
					t,
					q,
					"SELECT id FROM accounts WHERE (SELECT total FROM invoices WHERE account_id = $1) "+tc.op+" $2 AND active = $3",
					[]any{401, tc.val, 406},
				)
			})
		}
	})

	t.Run("not and double not nested dollar numbering", func(t *testing.T) {
		sb := StatementBuilder.PlaceholderFormat(Dollar)

		q := sb.Select("id").From("items").Where(Not(Eq{"status": "archived"})).Where("tenant_id = ?", 601)
		assertNestedPlaceholderSQL(
			t,
			q,
			"SELECT id FROM items WHERE NOT (status = $1) AND tenant_id = $2",
			[]any{"archived", 601},
		)

		q = sb.Select("id").From("items").Where(Not(Not(Eq{"status": "active"}))).Where("tenant_id = ?", 602)
		assertNestedPlaceholderSQL(
			t,
			q,
			"SELECT id FROM items WHERE status = $1 AND tenant_id = $2",
			[]any{"active", 602},
		)
	})

	t.Run("coalesce nested dollar numbering", func(t *testing.T) {
		sb := StatementBuilder.PlaceholderFormat(Dollar)
		first := sb.Select("nickname").From("profiles").Where("user_id = ?", 701)
		second := sb.Select("display_name").From("profiles").Where("user_id = ?", 702)
		co := Coalesce("anonymous", first, second)

		q := sb.Select("id").Column(co).From("users").Where("id = ?", 703)
		assertNestedPlaceholderSQL(
			t,
			q,
			"SELECT id, COALESCE((SELECT nickname FROM profiles WHERE user_id = $1), (SELECT display_name FROM profiles WHERE user_id = $2), $3) FROM users WHERE id = $4",
			[]any{701, 702, "anonymous", 703},
		)
	})

	t.Run("aggregate helpers nested dollar numbering", func(t *testing.T) {
		sb := StatementBuilder.PlaceholderFormat(Dollar)
		inner := sb.Select("amount").From("payments").Where("order_id = ?", 801)

		aggregates := []struct {
			name string
			col  Sqlizer
			fn   string
		}{
			{"sum", Sum(inner), "SUM"},
			{"count", Count(inner), "COUNT"},
			{"avg", Avg(inner), "AVG"},
			{"min", Min(inner), "MIN"},
			{"max", Max(inner), "MAX"},
		}

		for _, ag := range aggregates {
			t.Run(ag.name, func(t *testing.T) {
				q := sb.Select("id").Column(ag.col).From("orders").Where("shop_id = ?", 802)
				assertNestedPlaceholderSQL(
					t,
					q,
					"SELECT id, "+ag.fn+"(SELECT amount FROM payments WHERE order_id = $1) FROM orders WHERE shop_id = $2",
					[]any{801, 802},
				)
			})
		}
	})

	t.Run("in empty slice produces no condition", func(t *testing.T) {
		sql, args, err := In("id", []int{}).ToSql()
		mustNoError(t, err)
		assertEmpty(t, sql)
		assertEmpty(t, args)

		sql, args, err = NotIn("id", []int{}).ToSql()
		mustNoError(t, err)
		assertEmpty(t, sql)
		assertEmpty(t, args)
	})

	t.Run("in scalar one item and multi item dollar numbering", func(t *testing.T) {
		sb := StatementBuilder.PlaceholderFormat(Dollar)

		q := sb.Select("id").From("users").Where(In("role_id", 901)).Where("active = ?", 902)
		assertNestedPlaceholderSQL(
			t,
			q,
			"SELECT id FROM users WHERE role_id=$1 AND active = $2",
			[]any{901, 902},
		)

		q = sb.Select("id").From("users").Where(In("role_id", []int{903})).Where("active = ?", 904)
		assertNestedPlaceholderSQL(
			t,
			q,
			"SELECT id FROM users WHERE role_id=$1 AND active = $2",
			[]any{903, 904},
		)

		q = sb.Select("id").From("users").Where(In("role_id", []int{905, 906, 907})).Where("active = ?", 908)
		assertNestedPlaceholderSQL(
			t,
			q,
			"SELECT id FROM users WHERE role_id=ANY($1) AND active = $2",
			[]any{[]int{905, 906, 907}, 908},
		)
	})

	t.Run("not in scalar one item and multi item dollar numbering", func(t *testing.T) {
		sb := StatementBuilder.PlaceholderFormat(Dollar)

		q := sb.Select("id").From("users").Where(NotIn("role_id", 911)).Where("active = ?", 912)
		assertNestedPlaceholderSQL(
			t,
			q,
			"SELECT id FROM users WHERE role_id<>$1 AND active = $2",
			[]any{911, 912},
		)

		q = sb.Select("id").From("users").Where(NotIn("role_id", []int{913})).Where("active = ?", 914)
		assertNestedPlaceholderSQL(
			t,
			q,
			"SELECT id FROM users WHERE role_id<>$1 AND active = $2",
			[]any{913, 914},
		)

		q = sb.Select("id").From("users").Where(NotIn("role_id", []int{915, 916})).Where("active = ?", 917)
		assertNestedPlaceholderSQL(
			t,
			q,
			"SELECT id FROM users WHERE role_id<>ALL($1) AND active = $2",
			[]any{[]int{915, 916}, 917},
		)
	})

	t.Run("in and not in subquery dollar numbering", func(t *testing.T) {
		sb := StatementBuilder.PlaceholderFormat(Dollar)
		inner := sb.Select("user_id").From("bans").Where("reason = ?", 1001)

		q := sb.Select("id").From("users").Where(In("id", inner)).Where("state = ?", 1002)
		assertNestedPlaceholderSQL(
			t,
			q,
			"SELECT id FROM users WHERE id IN (SELECT user_id FROM bans WHERE reason = $1) AND state = $2",
			[]any{1001, 1002},
		)

		q = sb.Select("id").From("users").Where(NotIn("id", inner)).Where("state = ?", 1003)
		assertNestedPlaceholderSQL(
			t,
			q,
			"SELECT id FROM users WHERE id NOT IN (SELECT user_id FROM bans WHERE reason = $1) AND state = $2",
			[]any{1001, 1003},
		)
	})

	t.Run("in and not in named typed slice dollar numbering", func(t *testing.T) {
		sb := StatementBuilder.PlaceholderFormat(Dollar)
		ids := statusIDs{1101, 1102}

		q := sb.Select("id").From("tasks").Where(In("status_id", ids)).Where("owner_id = ?", 1103)
		assertNestedPlaceholderSQL(
			t,
			q,
			"SELECT id FROM tasks WHERE status_id=ANY($1) AND owner_id = $2",
			[]any{statusIDs{1101, 1102}, 1103},
		)

		q = sb.Select("id").From("tasks").Where(NotIn("status_id", ids)).Where("owner_id = ?", 1104)
		assertNestedPlaceholderSQL(
			t,
			q,
			"SELECT id FROM tasks WHERE status_id<>ALL($1) AND owner_id = $2",
			[]any{statusIDs{1101, 1102}, 1104},
		)
	})
}
