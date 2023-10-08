package slogx_test

// TODO: implement testing and benchmarks
/*
func BenchmarkSimple(b *testing.B) {
	b.ReportAllocs()
	formatter := attrformatter.New(attrformatter.Options{IncludeSource: true})
	record := slog.NewRecord(time.Now().UTC(), slog.LevelInfo, "this is a test message", 0)
	formattedRecord, err := formatter.HandleRecord(context.TODO(), record)
	if err != nil {
		b.Error(err)
		return
	}
	formattedRecord.Attrs(func(attr slog.Attr) bool {
		b.Logf("%s = %+v", attr.Key, attr.Value.Any())
		return true
	})
}

func BenchmarkWithPC(b *testing.B) {
	b.ReportAllocs()
	formatter := attrformatter.New(attrformatter.Options{IncludeSource: true})
	record := slog.NewRecord(time.Now().UTC(), slog.LevelInfo, "this is a test message", getCallerPC())
	formattedRecord, err := formatter.HandleRecord(context.TODO(), record)
	if err != nil {
		b.Error(err)
		return
	}
	formattedRecord.Attrs(func(attr slog.Attr) bool {
		b.Logf("%s = %+v", attr.Key, attr.Value.Any())
		return true
	})
}

func BenchmarkWithAttrs(b *testing.B) {
	b.ReportAllocs()
	formatter := attrformatter.New(attrformatter.Options{IncludeSource: true})
	record := slog.NewRecord(time.Now().UTC(), slog.LevelInfo, "this is a test message", getCallerPC())
	record.AddAttrs(
		slog.String("key1", "value1"),
		slog.Time("local_time", time.Now().Local()),
	)
	formattedRecord, err := formatter.HandleRecord(context.TODO(), record)
	if err != nil {
		b.Error(err)
		return
	}
	formattedRecord.Attrs(func(attr slog.Attr) bool {
		b.Logf("%s = %+v", attr.Key, attr.Value.Any())
		return true
	})
}

func BenchmarkWithGroups(b *testing.B) {
	b.ReportAllocs()
	formatter := attrformatter.New(attrformatter.Options{IncludeSource: true})
	record := slog.NewRecord(time.Now().UTC(), slog.LevelInfo, "this is a test message", getCallerPC())
	record.AddAttrs(
		slog.Group("nested",
			slog.String("key1", "value1"),
			slog.Time("local_time", time.Now().Local()),
		),
	)
	formattedRecord, err := formatter.HandleRecord(context.TODO(), record)
	if err != nil {
		b.Error(err)
		return
	}
	formattedRecord.Attrs(func(attr slog.Attr) bool {
		b.Logf("%s = %+v", attr.Key, attr.Value.Any())
		return true
	})
}

func BenchmarkWithCustomFormatter(b *testing.B) {
	b.ReportAllocs()
	formatter := attrformatter.New(attrformatter.Options{FormatAttrFn: formatAttr, IncludeSource: true})
	record := slog.NewRecord(time.Now().UTC(), slog.LevelInfo, "this is a test message", getCallerPC())
	record.AddAttrs(
		slog.String("key1", "value1"),
		slog.Time("local_time", time.Now().Local()),
	)
	formattedRecord, err := formatter.HandleRecord(context.TODO(), record)
	if err != nil {
		b.Error(err)
		return
	}
	formattedRecord.Attrs(func(attr slog.Attr) bool {
		b.Logf("%s = %+v", attr.Key, attr.Value.Any())
		return true
	})
}

func BenchmarkWithCustomFormatter2(b *testing.B) {
	b.ReportAllocs()
	u := User{
		Username: "johndoe",
		Password: "mypassword",
	}
	formatter := attrformatter.New(attrformatter.Options{
		FormatAttrFn: formatAttr,
		FormatSpecificAttrFn: map[string]attrformatter.FormatAttrFn{
			"nested.one":             formatNestedOne,
			"nested.two.three.super": formatNestedOne,
		},
		IncludeSource: true,
	})
	record := slog.NewRecord(time.Now().UTC(), slog.LevelInfo, "this is a test message", getCallerPC())
	record.AddAttrs(
		slog.String("key1", "value1"),
		slog.Time("local_time", time.Now().Local()),
		slog.Any("user", u),
		slog.Group("nested",
			slog.String("one", "two"),
			slog.Group("two",
				slog.Group("three",
					slog.String("super", "duper"),
				),
			),
		),
	)
	formattedRecord, err := formatter.HandleRecord(context.TODO(), record)
	if err != nil {
		b.Error(err)
		return
	}
	formattedRecord.Attrs(func(attr slog.Attr) bool {
		b.Logf("%s = %+v", attr.Key, attr.Value.Any())
		return true
	})
}

type InternalUser struct {
	Username string
	Password string
}

func (u InternalUser) LogValue() slog.Value {
	return slog.GroupValue(slog.String("user", u.Username), slog.String("pass", "****"))
}

type User struct {
	Username string
	Password string
}

func (u User) LogValue() slog.Value {
	user := InternalUser(u)
	return slog.AnyValue(user)
}

func getCallerPC() uintptr {
	pc, _, _, _ := runtime.Caller(1)
	return pc
}

func formatAttr(ctx context.Context, attrKey string, attrValue slog.Value) (string, slog.Value, error) {
	return fmt.Sprintf("__%s__", attrKey), slog.StringValue(fmt.Sprintf("| %s |", attrValue.String())), nil
}

func formatNestedOne(ctx context.Context, attrKey string, attrValue slog.Value) (string, slog.Value, error) {
	return fmt.Sprintf("___%s___", attrKey), slog.StringValue(fmt.Sprintf("!! %s !!", attrValue.String())), nil
}
*/
