package squirrel

import "testing"

func TestNestedPlaceholderRegressionSuite(t *testing.T) {
	t.Run("basic dollar single placeholder", func(t *testing.T) {
		q := Select("event_id").
			From("events").
			Where("tenant_id = ?", "basic-single-1001").
			PlaceholderFormat(Dollar)

		assertNestedPlaceholderSQL(
			t,
			q,
			"SELECT event_id FROM events WHERE tenant_id = $1",
			[]any{"basic-single-1001"},
		)
	})

	t.Run("basic dollar multiple placeholders", func(t *testing.T) {
		q := Select("event_id").
			From("events").
			Where("tenant_id = ?", 1101).
			Where("state = ?", "basic-state-1102").
			PlaceholderFormat(Dollar)

		assertNestedPlaceholderSQL(
			t,
			q,
			"SELECT event_id FROM events WHERE tenant_id = $1 AND state = $2",
			[]any{1101, "basic-state-1102"},
		)
	})

	t.Run("select in where", func(t *testing.T) {
		accountIDs := Select("account_id").
			From("memberships").
			Where("tenant_key = ?", "where-child-2101")

		q := Select("user_id").
			From("users").
			Where(Expr("account_id IN (?)", accountIDs)).
			Where("state = ?", "where-parent-2102").
			PlaceholderFormat(Dollar)

		assertNestedPlaceholderSQL(
			t,
			q,
			"SELECT user_id FROM users WHERE account_id IN (SELECT account_id FROM memberships WHERE tenant_key = $1) AND state = $2",
			[]any{"where-child-2101", "where-parent-2102"},
		)
	})

	t.Run("select in eq", func(t *testing.T) {
		accountIDs := Select("account_id").
			From("memberships").
			Where("tenant_key = ?", "eq-child-3101")

		q := Select("user_id").
			From("users").
			Where(Eq{"account_id": accountIDs}).
			Where("state = ?", "eq-parent-3102").
			PlaceholderFormat(Dollar)

		assertNestedPlaceholderSQL(
			t,
			q,
			"SELECT user_id FROM users WHERE account_id IN (SELECT account_id FROM memberships WHERE tenant_key = $1) AND state = $2",
			[]any{"eq-child-3101", "eq-parent-3102"},
		)
	})

	t.Run("select in join clause", func(t *testing.T) {
		teamOwners := Select("owner_id").
			From("teams").
			Where("region = ?", "join-child-4101")

		q := Select("users.id").
			From("users").
			JoinClause(teamOwners.
				Prefix("JOIN (").
				Suffix(") team_filter ON team_filter.owner_id = users.id")).
			Where("users.state = ?", "join-parent-4102").
			PlaceholderFormat(Dollar)

		assertNestedPlaceholderSQL(
			t,
			q,
			"SELECT users.id FROM users JOIN ( SELECT owner_id FROM teams WHERE region = $1 ) team_filter ON team_filter.owner_id = users.id WHERE users.state = $2",
			[]any{"join-child-4101", "join-parent-4102"},
		)
	})

	t.Run("from select", func(t *testing.T) {
		memberships := Select("account_id").
			From("memberships").
			Where("tenant_key = ?", "from-child-5102")

		q := Select("derived.account_id").
			Prefix("WITH request_scope AS (SELECT ? AS tenant_key)", "from-prefix-5101").
			FromSelect(memberships, "derived").
			Where("derived.state = ?", "from-parent-5103").
			PlaceholderFormat(Dollar)

		assertNestedPlaceholderSQL(
			t,
			q,
			"WITH request_scope AS (SELECT $1 AS tenant_key) SELECT derived.account_id FROM (SELECT account_id FROM memberships WHERE tenant_key = $2) AS derived WHERE derived.state = $3",
			[]any{"from-prefix-5101", "from-child-5102", "from-parent-5103"},
		)
	})

	t.Run("column select", func(t *testing.T) {
		latestScore := Select("score").
			From("scores").
			Where("scores.user_key = ?", "column-child-6101")

		q := Select("users.id").
			Column(Alias(latestScore, "latest_score")).
			From("users").
			Where("users.state = ?", "column-parent-6102").
			PlaceholderFormat(Dollar)

		assertNestedPlaceholderSQL(
			t,
			q,
			"SELECT users.id, (SELECT score FROM scores WHERE scores.user_key = $1) AS latest_score FROM users WHERE users.state = $2",
			[]any{"column-child-6101", "column-parent-6102"},
		)
	})

	t.Run("update set select", func(t *testing.T) {
		settingValue := Select("value").
			From("settings").
			Where("settings.key = ?", "update-child-7101")

		q := Update("profiles").
			Set("setting_value", settingValue).
			Set("updated_by", "update-set-7102").
			Where("profile_key = ?", "update-parent-7103").
			PlaceholderFormat(Dollar)

		assertNestedPlaceholderSQL(
			t,
			q,
			"UPDATE profiles SET setting_value = (SELECT value FROM settings WHERE settings.key = $1), updated_by = $2 WHERE profile_key = $3",
			[]any{"update-child-7101", "update-set-7102", "update-parent-7103"},
		)
	})

	t.Run("three level and or", func(t *testing.T) {
		auditMatch := Select("1").
			From("audit_log").
			Where("audit_log.marker = ?", "logic-child-8104").
			Prefix("EXISTS (").
			Suffix(")")

		q := Select("record_id").
			From("records").
			Where(And{
				Expr("root_flag = ?", "logic-root-8101"),
				Or{
					Expr("branch_code = ?", "logic-or-8102"),
					And{
						Expr("leaf_state = ?", "logic-and-8103"),
						auditMatch,
					},
				},
			}).
			Where("tail_key = ?", "logic-tail-8105").
			PlaceholderFormat(Dollar)

		assertNestedPlaceholderSQL(
			t,
			q,
			"SELECT record_id FROM records WHERE (root_flag = $1 AND (branch_code = $2 OR (leaf_state = $3 AND EXISTS ( SELECT 1 FROM audit_log WHERE audit_log.marker = $4 )))) AND tail_key = $5",
			[]any{
				"logic-root-8101",
				"logic-or-8102",
				"logic-and-8103",
				"logic-child-8104",
				"logic-tail-8105",
			},
		)
	})
}

func assertNestedPlaceholderSQL(t *testing.T, q Sqlizer, wantSQL string, wantArgs []any) {
	t.Helper()

	gotSQL, gotArgs, err := q.ToSql()
	mustNoError(t, err)
	assertEqual(t, wantSQL, gotSQL)
	assertEqual(t, wantArgs, gotArgs)
}
