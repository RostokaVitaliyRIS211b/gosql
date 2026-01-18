package gosql

import (
	"reflect"
	"slices"
	"testing"
)

var it = reflect.TypeFor[int]()

func TestConvertionOfTypeToNonRef(t *testing.T) {
	ty := ConversionTypeToNonRefType(reflect.TypeFor[*int]())

	if it != ty {
		t.Errorf("types not matching result %s", ty.Name())
	}

	ty = ConversionTypeToNonRefType(reflect.TypeFor[**int]())

	if it != ty {
		t.Errorf("types not matching 2 %s", ty.Name())
	}

	ty = ConversionTypeToNonRefType(reflect.TypeFor[***int]())

	if it != ty {
		t.Errorf("types not matching 3 %s", ty.Name())
	}

	ty = ConversionTypeToNonRefType(reflect.TypeFor[****int]())

	if it != ty {
		t.Errorf("types not matching 4 %s", ty.Name())
	}
}

func TestConvertionOfValToNonRef(t *testing.T) {

	ty := ConversionValToNonRefType(reflect.Zero(reflect.TypeFor[*int]()).Interface())
	if it != ty {
		t.Errorf("types not matching")
	}

	ty = ConversionValToNonRefType(reflect.Zero(reflect.TypeFor[**int]()).Interface())
	if it != ty {
		t.Errorf("types not matching 2")
	}

	ty = ConversionValToNonRefType(reflect.Zero(reflect.TypeFor[***int]()).Interface())
	if it != ty {
		t.Errorf("types not matching 3")
	}

	ty = ConversionValToNonRefType(reflect.Zero(reflect.TypeFor[****int]()).Interface())
	if it != ty {
		t.Errorf("types not matching 4")
	}
}

func TestConversionToOgType(t *testing.T) {
	i := 4
	og := reflect.TypeFor[*int]()

	res, err := ConversionToOgType(i, og)
	if err != nil || reflect.TypeOf(res) != og {
		t.Errorf("either type not match or error: %s", err)
	}

	og = reflect.TypeFor[**int]()

	res, err = ConversionToOgType(i, og)

	if err != nil || reflect.TypeOf(res) != og {
		t.Errorf("either type not match or error: %s", err)
	}

	og = reflect.TypeFor[***int]()

	res, err = ConversionToOgType(i, og)

	if err != nil || reflect.TypeOf(res) != og {
		t.Errorf("either type not match or error: %s", err)
	}

	og = reflect.TypeFor[****int]()

	res, err = ConversionToOgType(i, og)

	if err != nil || reflect.TypeOf(res) != og {
		t.Errorf("either type not match or error: %s", err)
	}
}

func TestMapping(t *testing.T) {
	type user struct {
		Name  string  `db:"Name"`
		Name2 int     `db:"Name2"`
		Name3 float64 `db:"Name3"`
		Name4 []byte  `db:"Name4"`
		Name5 []int
	}

	type user2 struct {
		Name  string  `dbcn:"Name"`
		Name2 int     `dbcn:"Name2"`
		Name3 float64 `dbcn:"Name3"`
		Name4 []byte  `dbcn:"Name4"`
		Name6 []int   `db:"gay"`
	}

	item := user{}

	const tagName, tagName2 = "db", "dbcn"

	compareRes := func(tm *TypeMap, tgName string, err error, it any) {
		if err != nil {
			t.Errorf("Error: %s", err)
		}

		if len(tm.Fields) != 4 {
			t.Errorf("NumOfFields dont match")
		}

		if tm.NonRefType != reflect.TypeOf(it) {
			t.Errorf("types dont match")
		}

		if tm.TagName != tgName {
			t.Errorf("tags dont match")
		}

		if !slices.ContainsFunc(tm.Fields, func(f *FieldInfo) bool { return f.FTag == "Name" }) {
			t.Errorf("dont contains Name")
		}

		if !slices.ContainsFunc(tm.Fields, func(f *FieldInfo) bool { return f.FTag == "Name2" }) {
			t.Errorf("dont contains Name2")
		}

		if !slices.ContainsFunc(tm.Fields, func(f *FieldInfo) bool { return f.FTag == "Name3" }) {
			t.Errorf("dont contains Name3")
		}

		if !slices.ContainsFunc(tm.Fields, func(f *FieldInfo) bool { return f.FTag == "Name4" }) {
			t.Errorf("dont contains Name4")
		}
	}

	typeMap, err := MapFunc(reflect.TypeOf(item), tagName)

	compareRes(typeMap, tagName, err, item)

	item2 := user2{}

	typeMap, err = MapFunc(reflect.TypeOf(item2), tagName2)

	compareRes(typeMap, tagName2, err, item2)

}

func TestGetFieldsPointers(t *testing.T) {
	type user struct {
		Name  string  `db:"Name"`
		Name2 int     `db:"Name2"`
		Name3 float64 `db:"Name3"`
		Name4 []byte  `db:"Name4"`
		Name5 []int
	}

	type user2 struct {
		Name  *string  `dbcn:"Name"`
		Name2 *int     `dbcn:"Name2"`
		Name3 *float64 `dbcn:"Name3"`
		Name4 *[]byte  `dbcn:"Name4"`
		Name5 []int
	}

	item := &user{}

	const tagName, tagName2 = "db", "dbcn"

	typeMap, err := MapFunc(reflect.TypeFor[*user](), tagName)

	if err != nil {
		t.Errorf("error: %s", err)
	}

	pointers := GetFieldsPointersOfItem(reflect.ValueOf(item), typeMap, []string{})

	b1 := []byte("123")
	f2 := 42.0
	i3 := 4
	s4 := "123"

	p1 := pointers[3].(*[]byte)
	*p1 = b1

	p2 := pointers[2].(*float64)
	*p2 = f2

	p3 := pointers[1].(*int)
	*p3 = i3

	p4 := pointers[0].(*string)
	*p4 = s4

	if len(pointers) != 4 {
		t.Errorf("Tags failed")
	}

	if item.Name != s4 || p4 != &item.Name {
		t.Errorf("1 field pointer fail")
	}

	if item.Name2 != i3 || p3 != &item.Name2 {
		t.Errorf("2 field pointer fail")
	}

	if item.Name3 != f2 || p2 != &item.Name3 {
		t.Errorf("3 field pointer fail")
	}

	if string(item.Name4) != string(b1) || p1 != &item.Name4 {
		t.Errorf("4 field pointer fail")
	}

	item2 := &user2{}

	typeMap, err = MapFunc(reflect.TypeFor[*user2](), tagName2)

	if err != nil {
		t.Errorf("error: %s", err)
	}

	pointers = GetFieldsPointersOfItem(reflect.ValueOf(item2), typeMap, []string{"Name"})

	b11 := []byte("1231234")
	f22 := 567.1
	i33 := 4312

	p1 = pointers[2].(*[]byte)
	*p1 = b11

	p2 = pointers[1].(*float64)
	*p2 = f22

	p3 = pointers[0].(*int)
	*p3 = i33

	if len(pointers) != 3 {
		t.Errorf("Excluded tags failed")
	}

	if *item2.Name2 != i33 || p3 != item2.Name2 {
		t.Errorf("2 field pointer fail")
	}

	if *item2.Name3 != f22 || p2 != item2.Name3 {
		t.Errorf("3 field pointer fail")
	}

	if string(*item2.Name4) != string(b11) || p1 != item2.Name4 {
		t.Errorf("4 field pointer fail")
	}

}

type sc struct {
	Values  [][]any
	counter int
}

func (s *sc) Next() bool {
	defer func() { s.counter++ }()
	return s.counter == 0
}

func (s *sc) Scan(pointers ...any) error {
	for idx, pointer := range pointers {
		val := s.Values[s.counter/len(pointers)][idx]
		reflect.ValueOf(pointer).Elem().Set(reflect.ValueOf(val))
	}
	return nil
}

func TestScan(t *testing.T) {
	type user struct {
		Name  string  `db:"Name"`
		Name2 int     `db:"Name2"`
		Name3 float64 `db:"Name3"`
		Name4 []byte  `db:"Name4"`
		Name5 []int
	}

	type user2 struct {
		Name  *string  `dbcn:"Name"`
		Name2 *int     `dbcn:"Name2"`
		Name3 *float64 `dbcn:"Name3"`
		Name4 *[]byte  `dbcn:"Name4"`
		Name5 []int
	}

	const tagName, tagName2 = "db", "dbcn"

	queryConfig := QueryConfig{
		TableName:    "users",
		NameWrapper:  "",
		ColumnName:   "",
		TagName:      tagName,
		ItemToAdd:    nil,
		ExcludedTags: []string{},
	}
	nstr := "123"
	ni := 1
	f64n := 42.1
	bsn := []byte{0x8}
	var rowScanner RowScanner = &sc{Values: [][]any{{nstr, ni, f64n, bsn}}}

	var users []user

	stdScanner := GetScanner(tagName)

	stdScanner.Scan(&users, rowScanner, queryConfig)

	if len(users) != 1 {
		t.Errorf("scan failed")
	}

	if users[0].Name != nstr {
		t.Errorf("user Name not match")
	}

	if users[0].Name2 != ni {
		t.Errorf("user Name2 not match")
	}

	if users[0].Name3 != f64n {
		t.Errorf("user Name3 not match")
	}

	if string(users[0].Name4) != string(bsn) {
		t.Errorf("user Name4 not match")
	}

	nstr = "hello goida"
	ni = 67
	f64n = 67.67
	bsn = []byte{0xF}

	var users2 []user2

	rowScanner = &sc{Values: [][]any{{nstr, ni, f64n, bsn}}}

	stdScanner.Scan(&users2, rowScanner, queryConfig.ChangeTagName(tagName2).ChangeExcludedTags("Name4"))

	if len(users2) != 1 {
		t.Errorf("scan failed")
	}

	if *users2[0].Name != nstr {
		t.Errorf("user Name not match")
	}

	if *users2[0].Name2 != ni {
		t.Errorf("user Name2 not match")
	}

	if *users2[0].Name3 != f64n {
		t.Errorf("user Name3 not match")
	}

	if len(*users2[0].Name4) != 0 {
		t.Errorf("user Name4 not match")
	}

}
