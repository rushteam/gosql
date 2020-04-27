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
		t.Errorf("SQLSegment.BuildSelect() = %v, want %v", result, want)
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
		t.Errorf("SQLSegment.BuildInsert() = %v, want %v", result, want)
	}
}
