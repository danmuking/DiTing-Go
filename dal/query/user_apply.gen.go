// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package query

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"

	"gorm.io/gen"
	"gorm.io/gen/field"

	"gorm.io/plugin/dbresolver"

	"DiTing-Go/dal/model"
)

func newUserApply(db *gorm.DB, opts ...gen.DOOption) userApply {
	_userApply := userApply{}

	_userApply.userApplyDo.UseDB(db, opts...)
	_userApply.userApplyDo.UseModel(&model.UserApply{})

	tableName := _userApply.userApplyDo.TableName()
	_userApply.ALL = field.NewAsterisk(tableName)
	_userApply.ID = field.NewInt64(tableName, "id")
	_userApply.UID = field.NewInt64(tableName, "uid")
	_userApply.Type = field.NewInt32(tableName, "type")
	_userApply.TargetID = field.NewInt64(tableName, "target_id")
	_userApply.Msg = field.NewString(tableName, "msg")
	_userApply.Status = field.NewInt32(tableName, "status")
	_userApply.ReadStatus = field.NewInt32(tableName, "read_status")
	_userApply.CreateTime = field.NewTime(tableName, "create_time")
	_userApply.UpdateTime = field.NewTime(tableName, "update_time")

	_userApply.fillFieldMap()

	return _userApply
}

// userApply 用户申请表
type userApply struct {
	userApplyDo userApplyDo

	ALL        field.Asterisk
	ID         field.Int64  // id
	UID        field.Int64  // 申请人uid
	Type       field.Int32  // 申请类型 1加好友
	TargetID   field.Int64  // 接收人uid
	Msg        field.String // 申请信息
	Status     field.Int32  // 申请状态 1待审批 2同意
	ReadStatus field.Int32  // 阅读状态 1未读 2已读
	CreateTime field.Time   // 创建时间
	UpdateTime field.Time   // 修改时间

	fieldMap map[string]field.Expr
}

func (u userApply) Table(newTableName string) *userApply {
	u.userApplyDo.UseTable(newTableName)
	return u.updateTableName(newTableName)
}

func (u userApply) As(alias string) *userApply {
	u.userApplyDo.DO = *(u.userApplyDo.As(alias).(*gen.DO))
	return u.updateTableName(alias)
}

func (u *userApply) updateTableName(table string) *userApply {
	u.ALL = field.NewAsterisk(table)
	u.ID = field.NewInt64(table, "id")
	u.UID = field.NewInt64(table, "uid")
	u.Type = field.NewInt32(table, "type")
	u.TargetID = field.NewInt64(table, "target_id")
	u.Msg = field.NewString(table, "msg")
	u.Status = field.NewInt32(table, "status")
	u.ReadStatus = field.NewInt32(table, "read_status")
	u.CreateTime = field.NewTime(table, "create_time")
	u.UpdateTime = field.NewTime(table, "update_time")

	u.fillFieldMap()

	return u
}

func (u *userApply) WithContext(ctx context.Context) IUserApplyDo {
	return u.userApplyDo.WithContext(ctx)
}

func (u userApply) TableName() string { return u.userApplyDo.TableName() }

func (u userApply) Alias() string { return u.userApplyDo.Alias() }

func (u userApply) Columns(cols ...field.Expr) gen.Columns { return u.userApplyDo.Columns(cols...) }

func (u *userApply) GetFieldByName(fieldName string) (field.OrderExpr, bool) {
	_f, ok := u.fieldMap[fieldName]
	if !ok || _f == nil {
		return nil, false
	}
	_oe, ok := _f.(field.OrderExpr)
	return _oe, ok
}

func (u *userApply) fillFieldMap() {
	u.fieldMap = make(map[string]field.Expr, 9)
	u.fieldMap["id"] = u.ID
	u.fieldMap["uid"] = u.UID
	u.fieldMap["type"] = u.Type
	u.fieldMap["target_id"] = u.TargetID
	u.fieldMap["msg"] = u.Msg
	u.fieldMap["status"] = u.Status
	u.fieldMap["read_status"] = u.ReadStatus
	u.fieldMap["create_time"] = u.CreateTime
	u.fieldMap["update_time"] = u.UpdateTime
}

func (u userApply) clone(db *gorm.DB) userApply {
	u.userApplyDo.ReplaceConnPool(db.Statement.ConnPool)
	return u
}

func (u userApply) replaceDB(db *gorm.DB) userApply {
	u.userApplyDo.ReplaceDB(db)
	return u
}

type userApplyDo struct{ gen.DO }

type IUserApplyDo interface {
	gen.SubQuery
	Debug() IUserApplyDo
	WithContext(ctx context.Context) IUserApplyDo
	WithResult(fc func(tx gen.Dao)) gen.ResultInfo
	ReplaceDB(db *gorm.DB)
	ReadDB() IUserApplyDo
	WriteDB() IUserApplyDo
	As(alias string) gen.Dao
	Session(config *gorm.Session) IUserApplyDo
	Columns(cols ...field.Expr) gen.Columns
	Clauses(conds ...clause.Expression) IUserApplyDo
	Not(conds ...gen.Condition) IUserApplyDo
	Or(conds ...gen.Condition) IUserApplyDo
	Select(conds ...field.Expr) IUserApplyDo
	Where(conds ...gen.Condition) IUserApplyDo
	Order(conds ...field.Expr) IUserApplyDo
	Distinct(cols ...field.Expr) IUserApplyDo
	Omit(cols ...field.Expr) IUserApplyDo
	Join(table schema.Tabler, on ...field.Expr) IUserApplyDo
	LeftJoin(table schema.Tabler, on ...field.Expr) IUserApplyDo
	RightJoin(table schema.Tabler, on ...field.Expr) IUserApplyDo
	Group(cols ...field.Expr) IUserApplyDo
	Having(conds ...gen.Condition) IUserApplyDo
	Limit(limit int) IUserApplyDo
	Offset(offset int) IUserApplyDo
	Count() (count int64, err error)
	Scopes(funcs ...func(gen.Dao) gen.Dao) IUserApplyDo
	Unscoped() IUserApplyDo
	Create(values ...*model.UserApply) error
	CreateInBatches(values []*model.UserApply, batchSize int) error
	Save(values ...*model.UserApply) error
	First() (*model.UserApply, error)
	Take() (*model.UserApply, error)
	Last() (*model.UserApply, error)
	Find() ([]*model.UserApply, error)
	FindInBatch(batchSize int, fc func(tx gen.Dao, batch int) error) (results []*model.UserApply, err error)
	FindInBatches(result *[]*model.UserApply, batchSize int, fc func(tx gen.Dao, batch int) error) error
	Pluck(column field.Expr, dest interface{}) error
	Delete(...*model.UserApply) (info gen.ResultInfo, err error)
	Update(column field.Expr, value interface{}) (info gen.ResultInfo, err error)
	UpdateSimple(columns ...field.AssignExpr) (info gen.ResultInfo, err error)
	Updates(value interface{}) (info gen.ResultInfo, err error)
	UpdateColumn(column field.Expr, value interface{}) (info gen.ResultInfo, err error)
	UpdateColumnSimple(columns ...field.AssignExpr) (info gen.ResultInfo, err error)
	UpdateColumns(value interface{}) (info gen.ResultInfo, err error)
	UpdateFrom(q gen.SubQuery) gen.Dao
	Attrs(attrs ...field.AssignExpr) IUserApplyDo
	Assign(attrs ...field.AssignExpr) IUserApplyDo
	Joins(fields ...field.RelationField) IUserApplyDo
	Preload(fields ...field.RelationField) IUserApplyDo
	FirstOrInit() (*model.UserApply, error)
	FirstOrCreate() (*model.UserApply, error)
	FindByPage(offset int, limit int) (result []*model.UserApply, count int64, err error)
	ScanByPage(result interface{}, offset int, limit int) (count int64, err error)
	Scan(result interface{}) (err error)
	Returning(value interface{}, columns ...string) IUserApplyDo
	UnderlyingDB() *gorm.DB
	schema.Tabler
}

func (u userApplyDo) Debug() IUserApplyDo {
	return u.withDO(u.DO.Debug())
}

func (u userApplyDo) WithContext(ctx context.Context) IUserApplyDo {
	return u.withDO(u.DO.WithContext(ctx))
}

func (u userApplyDo) ReadDB() IUserApplyDo {
	return u.Clauses(dbresolver.Read)
}

func (u userApplyDo) WriteDB() IUserApplyDo {
	return u.Clauses(dbresolver.Write)
}

func (u userApplyDo) Session(config *gorm.Session) IUserApplyDo {
	return u.withDO(u.DO.Session(config))
}

func (u userApplyDo) Clauses(conds ...clause.Expression) IUserApplyDo {
	return u.withDO(u.DO.Clauses(conds...))
}

func (u userApplyDo) Returning(value interface{}, columns ...string) IUserApplyDo {
	return u.withDO(u.DO.Returning(value, columns...))
}

func (u userApplyDo) Not(conds ...gen.Condition) IUserApplyDo {
	return u.withDO(u.DO.Not(conds...))
}

func (u userApplyDo) Or(conds ...gen.Condition) IUserApplyDo {
	return u.withDO(u.DO.Or(conds...))
}

func (u userApplyDo) Select(conds ...field.Expr) IUserApplyDo {
	return u.withDO(u.DO.Select(conds...))
}

func (u userApplyDo) Where(conds ...gen.Condition) IUserApplyDo {
	return u.withDO(u.DO.Where(conds...))
}

func (u userApplyDo) Order(conds ...field.Expr) IUserApplyDo {
	return u.withDO(u.DO.Order(conds...))
}

func (u userApplyDo) Distinct(cols ...field.Expr) IUserApplyDo {
	return u.withDO(u.DO.Distinct(cols...))
}

func (u userApplyDo) Omit(cols ...field.Expr) IUserApplyDo {
	return u.withDO(u.DO.Omit(cols...))
}

func (u userApplyDo) Join(table schema.Tabler, on ...field.Expr) IUserApplyDo {
	return u.withDO(u.DO.Join(table, on...))
}

func (u userApplyDo) LeftJoin(table schema.Tabler, on ...field.Expr) IUserApplyDo {
	return u.withDO(u.DO.LeftJoin(table, on...))
}

func (u userApplyDo) RightJoin(table schema.Tabler, on ...field.Expr) IUserApplyDo {
	return u.withDO(u.DO.RightJoin(table, on...))
}

func (u userApplyDo) Group(cols ...field.Expr) IUserApplyDo {
	return u.withDO(u.DO.Group(cols...))
}

func (u userApplyDo) Having(conds ...gen.Condition) IUserApplyDo {
	return u.withDO(u.DO.Having(conds...))
}

func (u userApplyDo) Limit(limit int) IUserApplyDo {
	return u.withDO(u.DO.Limit(limit))
}

func (u userApplyDo) Offset(offset int) IUserApplyDo {
	return u.withDO(u.DO.Offset(offset))
}

func (u userApplyDo) Scopes(funcs ...func(gen.Dao) gen.Dao) IUserApplyDo {
	return u.withDO(u.DO.Scopes(funcs...))
}

func (u userApplyDo) Unscoped() IUserApplyDo {
	return u.withDO(u.DO.Unscoped())
}

func (u userApplyDo) Create(values ...*model.UserApply) error {
	if len(values) == 0 {
		return nil
	}
	return u.DO.Create(values)
}

func (u userApplyDo) CreateInBatches(values []*model.UserApply, batchSize int) error {
	return u.DO.CreateInBatches(values, batchSize)
}

// Save : !!! underlying implementation is different with GORM
// The method is equivalent to executing the statement: db.Clauses(clause.OnConflict{UpdateAll: true}).Create(values)
func (u userApplyDo) Save(values ...*model.UserApply) error {
	if len(values) == 0 {
		return nil
	}
	return u.DO.Save(values)
}

func (u userApplyDo) First() (*model.UserApply, error) {
	if result, err := u.DO.First(); err != nil {
		return nil, err
	} else {
		return result.(*model.UserApply), nil
	}
}

func (u userApplyDo) Take() (*model.UserApply, error) {
	if result, err := u.DO.Take(); err != nil {
		return nil, err
	} else {
		return result.(*model.UserApply), nil
	}
}

func (u userApplyDo) Last() (*model.UserApply, error) {
	if result, err := u.DO.Last(); err != nil {
		return nil, err
	} else {
		return result.(*model.UserApply), nil
	}
}

func (u userApplyDo) Find() ([]*model.UserApply, error) {
	result, err := u.DO.Find()
	return result.([]*model.UserApply), err
}

func (u userApplyDo) FindInBatch(batchSize int, fc func(tx gen.Dao, batch int) error) (results []*model.UserApply, err error) {
	buf := make([]*model.UserApply, 0, batchSize)
	err = u.DO.FindInBatches(&buf, batchSize, func(tx gen.Dao, batch int) error {
		defer func() { results = append(results, buf...) }()
		return fc(tx, batch)
	})
	return results, err
}

func (u userApplyDo) FindInBatches(result *[]*model.UserApply, batchSize int, fc func(tx gen.Dao, batch int) error) error {
	return u.DO.FindInBatches(result, batchSize, fc)
}

func (u userApplyDo) Attrs(attrs ...field.AssignExpr) IUserApplyDo {
	return u.withDO(u.DO.Attrs(attrs...))
}

func (u userApplyDo) Assign(attrs ...field.AssignExpr) IUserApplyDo {
	return u.withDO(u.DO.Assign(attrs...))
}

func (u userApplyDo) Joins(fields ...field.RelationField) IUserApplyDo {
	for _, _f := range fields {
		u = *u.withDO(u.DO.Joins(_f))
	}
	return &u
}

func (u userApplyDo) Preload(fields ...field.RelationField) IUserApplyDo {
	for _, _f := range fields {
		u = *u.withDO(u.DO.Preload(_f))
	}
	return &u
}

func (u userApplyDo) FirstOrInit() (*model.UserApply, error) {
	if result, err := u.DO.FirstOrInit(); err != nil {
		return nil, err
	} else {
		return result.(*model.UserApply), nil
	}
}

func (u userApplyDo) FirstOrCreate() (*model.UserApply, error) {
	if result, err := u.DO.FirstOrCreate(); err != nil {
		return nil, err
	} else {
		return result.(*model.UserApply), nil
	}
}

func (u userApplyDo) FindByPage(offset int, limit int) (result []*model.UserApply, count int64, err error) {
	result, err = u.Offset(offset).Limit(limit).Find()
	if err != nil {
		return
	}

	if size := len(result); 0 < limit && 0 < size && size < limit {
		count = int64(size + offset)
		return
	}

	count, err = u.Offset(-1).Limit(-1).Count()
	return
}

func (u userApplyDo) ScanByPage(result interface{}, offset int, limit int) (count int64, err error) {
	count, err = u.Count()
	if err != nil {
		return
	}

	err = u.Offset(offset).Limit(limit).Scan(result)
	return
}

func (u userApplyDo) Scan(result interface{}) (err error) {
	return u.DO.Scan(result)
}

func (u userApplyDo) Delete(models ...*model.UserApply) (result gen.ResultInfo, err error) {
	return u.DO.Delete(models)
}

func (u *userApplyDo) withDO(do gen.Dao) *userApplyDo {
	u.DO = *do.(*gen.DO)
	return u
}
