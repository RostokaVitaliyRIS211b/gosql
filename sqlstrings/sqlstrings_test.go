package sqlstrings

import (
	"testing"
)

type dbItem struct {
	Id                 int     `db:"Id"`
	ObjectId           int     `db:"Objectid"`
	TypeUnitId         int     `db:"TypeUnitId"`
	CountOfUnits       float64 `db:"CountOfUnits"`
	PricePerUnit       float64 `db:"PricePerUnit"`
	ExcpectedCost      float64 `db:"ExcpectedCost"`
	ProducerId         int     `db:"ProducerId"`
	Description        string  `db:"Description"`
	TypeOfItem         int     `db:"TypeOfItemId"`
	NameId             int     `db:"NameId"`
	CountOfUsedUnits   float64 `db:"CountOfUsedUnits"`
	CountOfUsedUnits1  float64 `db:"CountOfUsedUnits"`
	CountOfUsedUnits2  float64 `db:"CountOfUsedUnits"`
	CountOfUsedUnits3  float64 `db:"CountOfUsedUnits"`
	CountOfUsedUnits4  float64 `db:"CountOfUsedUnits"`
	CountOfUsedUnits5  float64 `db:"CountOfUsedUnits"`
	CountOfUsedUnits6  float64 `db:"CountOfUsedUnits"`
	CountOfUsedUnits7  float64 `db:"CountOfUsedUnits"`
	CountOfUsedUnits8  float64 `db:"CountOfUsedUnits"`
	CountOfUsedUnits9  float64 `db:"CountOfUsedUnits"`
	CountOfUsedUnits10 float64 `db:"CountOfUsedUnits"`
	CountOfUsedUnits11 float64 `db:"CountOfUsedUnits"`
	CountOfUsedUnitsq  float64 `db:"CountOfUsedUnits"`
	CountOfUsedUnitsw  float64 `db:"CountOfUsedUnits"`
	CountOfUsedUnitse  float64 `db:"CountOfUsedUnits"`
	CountOfUsedUnitsd  float64 `db:"CountOfUsedUnits"`
	CountOfUsedUnitsf  float64 `db:"CountOfUsedUnits"`
	CountOfUsedUnitsg  float64 `db:"CountOfUsedUnits"`
	CountOfUsedUnitsh  float64 `db:"CountOfUsedUnits"`
	CountOfUsedUnitsj  float64 `db:"CountOfUsedUnits"`
	CountOfUsedUnitsk  float64 `db:"CountOfUsedUnits"`
	CountOfUsedUnitsl  float64 `db:"CountOfUsedUnits"`
	CountOfUsedUnitsz  float64 `db:"CountOfUsedUnits"`
	CountOfUsedUnitsx  float64 `db:"CountOfUsedUnits"`
	CountOfUsedUnitsc  float64 `db:"CountOfUsedUnits"`
	CountOfUsedUnitsv  float64 `db:"CountOfUsedUnits"`
	CountOfUsedUnitsb  float64 `db:"CountOfUsedUnits"`
	CountOfUsedUnitsn  float64 `db:"CountOfUsedUnits"`
	CountOfUsedUnitsm  float64 `db:"CountOfUsedUnits"`
	CountOfUsedUnits09 float64 `db:"CountOfUsedUnits"`
}

type users struct {
	Id          int     `db:"Id"`
	Name        string  `db:"Name"`
	Password    string  `db:"Password"`
	Description string  `db:"Description"`
	Roles       []int32 `ffff:"gg"`
}

type user2 struct {
	Id          int    `dbcn:"Id"`
	Name        string `dbcn:"Name"`
	Password    string `dbcn:"Password"`
	Description string `dbcn:"Description"`
	Roles       []int32
}

const (
	idColumnName = "Id"
	wrapper      = "\""
	tableName    = "users"
	columnName   = "Id"
	tagName      = "db"
	insertQuery1 = "INSERT INTO " + tableName + " (Name, Password, Description) VALUES ($1,$2,$3)"
	insertQuery2 = "INSERT INTO " + wrapper + tableName + wrapper + " (" + wrapper + "Name" + wrapper + ", " + wrapper + "Password" + wrapper + ", " + wrapper + "Description" + wrapper + ") VALUES ($1,$2,$3)"
	insertQuery3 = "INSERT INTO " + tableName + " (Name, Password, Description) VALUES ($1,$2,$3) RETURNING " + idColumnName
	updateQuery1 = "UPDATE " + tableName + " SET Name = $1, Password = $2, Description = $3"
	updateQuery2 = "UPDATE " + wrapper + tableName + wrapper + " SET " + wrapper + "Name" + wrapper + " = $1, " + wrapper + "Password" + wrapper + " = $2, " + wrapper + "Description" + wrapper + " = $3"
	updateQuery3 = "UPDATE " + tableName + " SET Name = $2, Password = $3, Description = $4 WHERE " + columnName + " = $1"
	selectQuery1 = "SELECT Name, Password, Description FROM " + tableName
	selectQuery2 = "SELECT " + wrapper + "Name" + wrapper + ", " + wrapper + "Password" + wrapper + ", " + wrapper + "Description" + wrapper + " FROM " + wrapper + tableName + wrapper
	selectQuery3 = "SELECT Name, Password, Description FROM " + tableName + " WHERE " + columnName + " = $1"
	deleteQuery1 = "DELETE FROM " + tableName
	deleteQuery2 = "DELETE FROM " + wrapper + tableName + wrapper
	deleteQuery3 = "DELETE FROM " + tableName + " WHERE " + columnName + " = $1"
)

var config = QueryConfig{
	TableName:    tableName,
	NameWrapper:  wrapper,
	ColumnName:   columnName,
	ItemToAdd:    dbItem{},
	TagName:      "db",
	ExcludedTags: []string{"Id"},
}

func TestInsertQuery(t *testing.T) {
	query := QueryConfig{
		TableName:    tableName,
		NameWrapper:  "",
		ColumnName:   "",
		TagName:      "",
		ItemToAdd:    user2{},
		ExcludedTags: []string{"Id"},
	}
	res := GetInsertQuery(query)

	if res != insertQuery1 {
		t.Errorf("%s", "QUERIES NOT MATCH\n"+insertQuery1+"\n"+res)
	}

	res = GetInsertQuery(query.ChangeTagName(tagName).ChangeItem(users{}))

	if res != insertQuery1 {
		t.Errorf("%s", "QUERIES NOT MATCH\n"+insertQuery1+"\n"+res)
	}

	res = GetInsertQuery(query.ChangeNameWrapper(wrapper))

	if res != insertQuery2 {
		t.Errorf("%s", "QUERIES NOT MATCH\n"+insertQuery2+"\n"+res)
	}

	res = GetInsertQuery(query.ChangeColumnName(columnName))

	if res != insertQuery3 {
		t.Errorf("%s", "QUERIES NOT MATCH\n"+insertQuery3+"\n"+res)
	}

	res = GetInsertQuery(query.ChangeTable("", users{}).ChangeTagName(tagName).ChangeNameWrapper(wrapper))

	if res != insertQuery2 {
		t.Errorf("%s", "QUERIES NOT MATCH\n"+insertQuery2+"\n"+res)
	}
}

func TestSelectQuery(t *testing.T) {
	query := QueryConfig{
		TableName:    tableName,
		NameWrapper:  "",
		ColumnName:   "",
		TagName:      "",
		ItemToAdd:    user2{},
		ExcludedTags: []string{"Id"},
	}
	res := GetSelectQuery(query)

	if res != selectQuery1 {
		t.Errorf("%s", "QUERIES NOT MATCH\n"+selectQuery1+"\n"+res)
	}

	res = GetSelectQuery(query.ChangeTagName(tagName).ChangeItem(users{}))

	if res != selectQuery1 {
		t.Errorf("%s", "QUERIES NOT MATCH\n"+selectQuery1+"\n"+res)
	}

	res = GetSelectQuery(query.ChangeNameWrapper(wrapper))

	if res != selectQuery2 {
		t.Errorf("%s", "QUERIES NOT MATCH\n"+selectQuery2+"\n"+res)
	}

	res = GetSelectQuery(query.ChangeColumnName(columnName))

	if res != selectQuery3 {
		t.Errorf("%s", "QUERIES NOT MATCH\n"+selectQuery3+"\n"+res)
	}

	res = GetSelectQuery(query.ChangeTable("", users{}).ChangeTagName(tagName).ChangeNameWrapper(wrapper))

	if res != selectQuery2 {
		t.Errorf("%s", "QUERIES NOT MATCH\n"+selectQuery2+"\n"+res)
	}
}

func TestUpdateQuery(t *testing.T) {
	query := QueryConfig{
		TableName:    tableName,
		NameWrapper:  "",
		ColumnName:   "",
		TagName:      "",
		ItemToAdd:    user2{},
		ExcludedTags: []string{"Id"},
	}
	res := GetUpdateQuery(query)

	if res != updateQuery1 {
		t.Errorf("%s", "QUERIES NOT MATCH\n"+updateQuery1+"\n"+res)
	}

	res = GetUpdateQuery(query.ChangeTagName(tagName).ChangeItem(users{}))

	if res != updateQuery1 {
		t.Errorf("%s", "QUERIES NOT MATCH\n"+updateQuery1+"\n"+res)
	}

	res = GetUpdateQuery(query.ChangeNameWrapper(wrapper))

	if res != updateQuery2 {
		t.Errorf("%s", "QUERIES NOT MATCH\n"+updateQuery2+"\n"+res)
	}

	res = GetUpdateQuery(query.ChangeColumnName(columnName))

	if res != updateQuery3 {
		t.Errorf("%s", "QUERIES NOT MATCH\n"+updateQuery3+"\n"+res)
	}

	res = GetUpdateQuery(query.ChangeTable("", users{}).ChangeTagName(tagName).ChangeNameWrapper(wrapper))

	if res != updateQuery2 {
		t.Errorf("%s", "QUERIES NOT MATCH\n"+updateQuery2+"\n"+res)
	}
}

func TestDeleteQuery(t *testing.T) {
	query := QueryConfig{
		TableName:    tableName,
		NameWrapper:  "",
		ColumnName:   "",
		TagName:      "",
		ItemToAdd:    user2{},
		ExcludedTags: []string{"Id"},
	}
	res := GetDeleteQuery(query)

	if res != deleteQuery1 {
		t.Errorf("%s", "QUERIES NOT MATCH\n"+deleteQuery1+"\n"+res)
	}

	res = GetDeleteQuery(query.ChangeTagName(tagName).ChangeItem(users{}))

	if res != deleteQuery1 {
		t.Errorf("%s", "QUERIES NOT MATCH\n"+deleteQuery1+"\n"+res)
	}

	res = GetDeleteQuery(query.ChangeNameWrapper(wrapper))

	if res != deleteQuery2 {
		t.Errorf("%s", "QUERIES NOT MATCH\n"+deleteQuery2+"\n"+res)
	}

	res = GetDeleteQuery(query.ChangeColumnName(columnName))

	if res != deleteQuery3 {
		t.Errorf("%s", "QUERIES NOT MATCH\n"+deleteQuery3+"\n"+res)
	}

	res = GetDeleteQuery(query.ChangeTable("", users{}).ChangeTagName(tagName).ChangeNameWrapper(wrapper))

	if res != deleteQuery2 {
		t.Errorf("%s", "QUERIES NOT MATCH\n"+deleteQuery2+"\n"+res)
	}

}
