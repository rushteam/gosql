package gosql

import (
	"testing"
)

func TestBuildSelect(t *testing.T) {
	s := NewSQLSegment()
	s.Flag("DISTINCT")
	s.Field("*")
	s.Table("tbl1.t1")
	s.Where("t1.status", "0")
	s.Where("type", "A")
	s.Where("[in]sts", []string{"1", "2", "3", "4"})
	s.Where("[in]sts2", 1)
	s.Where(func(s *Clause) {
		s.Where("a", "200")
		s.Where("b", "100")
	})
	s.Where("aaa = 999")
	s.Where("[#]ccc = ?", 888)
	s.Join("tbl3", "a", "=", "b")
	s.Having("ss", "1")
	s.Where("[~]a", "AA")
	s.Where("[exists]", "select 1")
	s.Where("[exists]", func(s *SQLSegments) {
		s.Table("tbl2.t2")
		s.Where("xx", 10000)
	})
	s.GroupBy("id")
	s.OrderBy("id desc", "id asc")
	s.Limit(30)
	s.Offset(10)
	s.ForUpdate()
	result := s.BuildSelect()
	want := "SELECT DISTINCT * FROM `tbl1`.`t1` JOIN `tbl3` ON `a` = `b` WHERE `t1`.`status` = ? AND `type` = ? AND `sts` IN (? ,? ,? ,?) AND `sts2` IN (?) AND ( `a` = ? AND `b` = ?) AND aaa = 999 AND ccc = ? AND `a` LIKE ? AND EXISTS (select 1) AND EXISTS (SELECT * FROM `tbl2`.`t2` WHERE `xx` = ?) GROUP BY `id` HAVING `ss` = ? ORDER BY `id desc`, `id asc` LIMIT 30 OFFSET 10 FOR UPDATE"
	if result != want {
		t.Errorf("result: %v, want: %v", result, want)
	}
}

func TestBuildInsert(t *testing.T) {
	data := make(map[string]interface{})
	data["name"] = "jack"
	s := NewSQLSegment()
	s.Table("test")
	s.Insert(data)
	result := s.BuildInsert()
	want := "INSERT INTO `test` (`name`) VALUES (?)"
	if result != want {
		t.Errorf("result: %v, want: %v", result, want)
	}
}

func TestBuildUpdate(t *testing.T) {
	data := make(map[string]interface{})
	data["name"] = "jack"
	s := NewSQLSegment()
	s.Table("test")
	s.Where("[in]id", []int{1, 2, 3})
	s.Update(data)
	result := s.BuildUpdate()
	want := "UPDATE `test` SET `name` = ? WHERE `id` IN (? ,? ,?)"
	if result != want {
		t.Errorf("result: %v, want: %v", result, want)
	}
}

func TestBuildReplace(t *testing.T) {
	data := make(map[string]interface{})
	data["name"] = "jack"
	s := NewSQLSegment()
	s.Table("test")
	s.Insert(data)
	result := s.BuildReplace()
	want := "REPLACE INTO `test` (`name`) VALUES (?)"
	if result != want {
		t.Errorf("result: %v, want: %v", result, want)
	}
}

func TestBuildDelete(t *testing.T) {
	data := make(map[string]interface{})
	data["name"] = "jack"
	s := NewSQLSegment()
	s.Table("test")
	s.Where("id", 1)
	s.Delete()
	result := s.BuildDelete()
	want := "DELETE FROM `test` WHERE `id` = ?"
	if result != want {
		t.Errorf("result: %v, want: %v", result, want)
	}
}

func TestTbName(t *testing.T) {
	s := NewSQLSegment()
	s.Field("*")
	s.Table(TbName{"table_1", "t1"})
	result := s.BuildSelect()
	want := "SELECT * FROM `table_1` AS `t1`"
	if result != want {
		t.Errorf("result: %v, want: %v", result, want)
	}
}
func TestTbNames(t *testing.T) {
	var tables []TbName
	tables = append(tables, TbName{"table_1", "t1"})
	tables = append(tables, TbName{"table_2", "t2"})
	s := NewSQLSegment()
	s.Field("*")
	s.Table(tables)
	result := s.BuildSelect()
	want := "SELECT * FROM `table_1` AS `t1`, `table_2` AS `t2`"
	if result != want {
		t.Errorf("result: %v, want: %v", result, want)
	}
}

func TestSelectSQL(t *testing.T) {
	result, _ := SelectSQL(
		Columns("id", "name"),
		Table("table_1"),
		Where("[!=]id", 1),
		Where(func(s *Clause) {
			s.OrWhere("[<]age", "20")
			s.OrWhere("[>]age", "30")
		}),
		OrWhere(func(s *Clause) {
			s.Where("[>=]score", "90")
			s.Where("[<=]age", "100")
		}),
		Where("[=]status", 1),
		Where("[!~]desc", "%test%"),
		Where("[!is]age", nil),
		GroupBy("type"),
		OrderBy("id DESC"),
		Offset(0),
		Limit(10),
	)
	want := "SELECT `id`, `name` FROM `table_1` WHERE `id` != ? AND ( `age` < ? OR `age` > ?) OR ( `score`` >= ? AND `age` <= ?) AND `status` = ? AND `desc`` NOT LIKE ? AND `age` IS NOT NULL GROUP BY `type` ORDER BY `id DESC` LIMIT 10"
	if result != want {
		t.Errorf("result: %v, want: %v", result, want)
	}
}

func TestInsertSQL(t *testing.T) {
	result, _ := InsertSQL(
		Table("table_1"),
		Set("a", "1"),
		Set("b", "1"),
	)
	want := "INSERT INTO `table_1` (`a`,`b`) VALUES (?,?)"
	want2 := "INSERT INTO `table_1` (`b`,`a`) VALUES (?,?)"
	if result != want && result != want2 {
		t.Errorf("result: %v, want: %v", result, want)
	}
}

func TestBatchInsertSQL(t *testing.T) {
	v1 := make(map[string]interface{}, 0)
	v1["a"] = 1
	v1["b"] = "jack"
	v2 := make(map[string]interface{}, 0)
	v2["a"] = 2
	v2["b"] = "tom"
	result, _ := InsertSQL(
		Table("table_1"),
		Params(v1),
		Params(v2),
	)
	want := "INSERT INTO `table_1` (`a`,`b`) VALUES (?,?),(?,?)"
	want2 := "INSERT INTO `table_1` (`b`,`a`) VALUES (?,?)"
	if result != want && result != want2 {
		t.Errorf("result: %v, want: %v", result, want)
	}
}
func TestReplaceSQL(t *testing.T) {
	result, _ := ReplaceSQL(
		Table("table_1"),
		Set("a", "1"),
		Set("b", "2"),
	)
	want := "REPLACE INTO `table_1` (`a`,`b`) VALUES (?,?)"
	want2 := "REPLACE INTO `table_1` (`b`,`a`) VALUES (?,?)"
	if result != want && result != want2 {
		t.Errorf("result: %v, want: %v", result, want)
	}
}
func TestBatchReplaceSQL(t *testing.T) {
	v1 := make(map[string]interface{}, 0)
	v1["a"] = 1
	v1["b"] = "jack"
	v2 := make(map[string]interface{}, 0)
	v2["a"] = 2
	v2["b"] = "tom"
	result, _ := ReplaceSQL(
		Table("table_1"),
		Params(v1),
		Params(v2),
	)
	want := "REPLACE INTO `table_1` (`a`,`b`) VALUES (?,?),(?,?)"
	want2 := "REPLACE INTO `table_1` (`b`,`a`) VALUES (?,?),(?,?)"
	if result != want && result != want2 {
		t.Errorf("result: %v, want: %v", result, want)
	}
}
