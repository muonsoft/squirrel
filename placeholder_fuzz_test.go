package squirrel

import (
	"fmt"
	"strings"
	"testing"
)

func FuzzPlaceholderComposition(f *testing.F) {
	f.Add(byte(0), "x = ? AND y = ?")
	f.Add(byte(0), "meta ?? 'format'")
	f.Add(byte(0), "meta->'tags' ??| array[?, ?]")
	f.Add(byte(0), "meta->'keys' ??& array[?]")
	f.Add(byte(0), "")
	f.Add(byte(0), "no placeholders")

	f.Add(byte(1), "tenant_id = ?")
	f.Add(byte(1), "state = ? AND kind = ?")
	f.Add(byte(1), "payload ??| array[?, ?]")
	f.Add(byte(1), "meta ?? 'format'")

	f.Add(byte(2), "body = ?")
	f.Add(byte(2), "region = ? AND active = ?")
	f.Add(byte(2), "tags ??& array[?]")

	f.Add(byte(3), "child_col = ?")
	f.Add(byte(3), "tenant_key = ? AND active = ?")
	f.Add(byte(3), "scope ??| array[?, ?]")

	f.Fuzz(func(t *testing.T, mode byte, fragment string) {
		if len(fragment) > 512 {
			fragment = fragment[:512]
		}

		switch mode % 4 {
		case 0:
			fuzzDirectPlaceholderReplacement(t, fragment)
		case 1:
			fuzzWhereComposition(t, fragment)
		case 2:
			fuzzPrefixSuffixComposition(t, fragment)
		default:
			fuzzNestedComposition(t, fragment)
		}
	})
}

func fuzzDirectPlaceholderReplacement(t *testing.T, sql string) {
	t.Helper()

	bindingCount := countBindingPlaceholders(sql)
	replaced, err := Dollar.ReplacePlaceholders(sql)
	mustNoError(t, err)

	want, err := replacePositionalPlaceholders(sql, "$")
	mustNoError(t, err)
	assertEqual(t, want, replaced)

	if strings.Contains(sql, "??") && bindingCount == strings.Count(sql, "?") {
		t.Fatalf("escaped ?? must not count as binding placeholders in %q", sql)
	}
}

func fuzzWhereComposition(t *testing.T, fragment string) {
	t.Helper()

	bindingCount := countBindingPlaceholders(fragment)
	args := makeFuzzArgs(bindingCount)

	q := Select("id").From("events").PlaceholderFormat(Dollar)
	if fragment != "" {
		if bindingCount > 0 {
			q = q.Where(fragment, args...)
		} else {
			q = q.Where(fragment)
		}
	}

	gotSQL, gotArgs, err := q.ToSql()
	if err != nil {
		return
	}

	if len(gotArgs) != bindingCount {
		t.Fatalf("where args: want %d got %d for fragment %q", bindingCount, len(gotArgs), fragment)
	}
	assertEscapedOperatorsConsumeNoArgs(t, fragment, len(gotArgs))
	_ = gotSQL
}

func fuzzPrefixSuffixComposition(t *testing.T, fragment string) {
	t.Helper()

	const prefixSQL = "WITH scope AS (SELECT ? AS tenant_key)"
	const suffixSQL = "LIMIT ?"

	bodyCount := countBindingPlaceholders(fragment)
	bodyArgs := makeFuzzArgs(bodyCount)

	q := Select("id").
		From("events").
		Prefix(prefixSQL, "pfx-0").
		PlaceholderFormat(Dollar)
	if fragment != "" {
		if bodyCount > 0 {
			q = q.Where(fragment, bodyArgs...)
		} else {
			q = q.Where(fragment)
		}
	}
	q = q.Suffix(suffixSQL, 25)

	gotSQL, gotArgs, err := q.ToSql()
	if err != nil {
		return
	}

	expectedArgs := 1 + bodyCount + 1
	if len(gotArgs) != expectedArgs {
		t.Fatalf("prefix/suffix args: want %d got %d for fragment %q", expectedArgs, len(gotArgs), fragment)
	}
	assertEscapedOperatorsConsumeNoArgs(t, fragment, bodyCount)
	_ = gotSQL
}

func fuzzNestedComposition(t *testing.T, childFragment string) {
	t.Helper()

	childCount := countBindingPlaceholders(childFragment)
	childArgs := makeFuzzArgs(childCount)

	child := Select("account_id").From("memberships").PlaceholderFormat(Dollar)
	if childFragment != "" {
		if childCount > 0 {
			child = child.Where(childFragment, childArgs...)
		} else {
			child = child.Where(childFragment)
		}
	}

	parent := Select("user_id").
		From("users").
		Where(Expr("account_id IN (?)", child)).
		Where("state = ?", "parent-state").
		PlaceholderFormat(Dollar)

	gotSQL, gotArgs, err := parent.ToSql()
	if err != nil {
		return
	}

	expectedArgs := childCount + 1
	if len(gotArgs) != expectedArgs {
		t.Fatalf("nested args: want %d got %d for fragment %q", expectedArgs, len(gotArgs), childFragment)
	}
	assertEscapedOperatorsConsumeNoArgs(t, childFragment, childCount)
	_ = gotSQL
}

func assertEscapedOperatorsConsumeNoArgs(t *testing.T, fragment string, bindingCount int) {
	t.Helper()
	if !strings.Contains(fragment, "??") {
		return
	}
	if bindingCount == strings.Count(fragment, "?") {
		t.Fatalf("escaped ?? must not consume bind arguments in %q", fragment)
	}
}

func makeFuzzArgs(count int) []any {
	args := make([]any, count)
	for i := range args {
		args[i] = fmt.Sprintf("fuzz-arg-%d", i)
	}
	return args
}

func countBindingPlaceholders(sql string) int {
	count := 0
	for i := 0; i < len(sql); i++ {
		if sql[i] != '?' {
			continue
		}
		if i+1 < len(sql) && sql[i+1] == '?' {
			i++
			continue
		}
		count++
	}
	return count
}
