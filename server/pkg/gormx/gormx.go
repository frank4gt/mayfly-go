package gormx

import (
	"fmt"
	"mayfly-go/pkg/biz"
	"mayfly-go/pkg/global"
	"mayfly-go/pkg/model"
	"strings"
	"time"

	"gorm.io/gorm"
)

// 伪删除之未删除过滤条件
func UndeleteScope(db *gorm.DB) *gorm.DB {
	return db.Where(model.DeletedColumn, model.ModelUndeleted)
}

// 根据id获取实体对象。model需为指针类型（需要将查询出来的值赋值给model）
//
// 若error不为nil则为不存在该记录
// @param model  数据库映射实体模型
func GetById(model any, id uint64, cols ...string) error {
	return NewQuery(model).Eq("id", id).Undeleted().GenGdb().First(model).Error
}

// 根据id列表查询实体信息
// @param model  数据库映射实体模型
func GetByIdIn(model any, list any, ids []uint64, orderBy ...string) {
	NewQuery(model).In("id", ids).Undeleted().GenGdb().Find(list)
}

// 获取满足model中不为空的字段值条件的单个对象。model需为指针类型（需要将查询出来的值赋值给model）
//
// 若 error不为nil，则为不存在该记录
// @param model  数据库映射实体模型
func GetBy(model any, cols ...string) error {
	return global.Db.Select(cols).Where(model).Scopes(UndeleteScope).First(model).Error
}

// 根据model指定条件统计数量
func CountBy(model any) int64 {
	var count int64
	NewQuery(model).WithCondModel(model).Undeleted().GenGdb().Count(&count)
	return count
}

// 根据条件model指定条件统计数量
func CountByCond(model any, condModel any) int64 {
	var count int64
	NewQuery(model).WithCondModel(condModel).Undeleted().GenGdb().Count(&count)
	return count
}

// 根据查询条件分页查询数据
// 若未指定查询列，则查询列以toModels字段为准
func PageQuery[T any](q *QueryCond, pageParam *model.PageParam, toModels T) *model.PageResult[T] {
	q.Undeleted()
	gdb := q.GenGdb()
	var count int64
	err := gdb.Count(&count).Error
	biz.ErrIsNilAppendErr(err, "查询错误: %s")
	if count == 0 {
		return model.EmptyPageResult[T]()
	}

	page := pageParam.PageNum
	pageSize := pageParam.PageSize
	err = gdb.Limit(pageSize).Offset((page - 1) * pageSize).Find(toModels).Error
	biz.ErrIsNilAppendErr(err, "查询失败: %s")
	return &model.PageResult[T]{Total: count, List: toModels}
}

// 获取满足model中不为空的字段值条件的所有数据.
//
// @param list为数组类型 如 var users *[]User，可指定为非model结构体，即只包含需要返回的字段结构体
func ListBy(model any, list any, cols ...string) {
	global.Db.Model(model).Select(cols).Where(model).Scopes(UndeleteScope).Order("id desc").Find(list)
}

// 获取满足model中不为空的字段值条件的所有数据.
//
// @param list为数组类型 如 var users *[]User，可指定为非model结构体
// @param model  数据库映射实体模型
func ListByOrder(model any, list any, order ...string) {
	var orderByStr string
	if order == nil {
		orderByStr = "id desc"
	} else {
		orderByStr = strings.Join(order, ",")
	}
	global.Db.Model(model).Where(model).Scopes(UndeleteScope).Order(orderByStr).Find(list)
}

func GetListBySql2Model(sql string, toEntity any, params ...any) error {
	return global.Db.Raw(sql, params...).Find(toEntity).Error
}

func ExecSql(sql string, params ...any) {
	global.Db.Exec(sql, params...)
}

// 插入model
// @param model  数据库映射实体模型
func Insert(model any) error {
	return InsertWithDb(*global.Db, model)
}

// 使用指定gormDb插入model
func InsertWithDb(db gorm.DB, model any) error {
	return db.Create(model).Error
}

// 批量插入
func BatchInsert[T any](models []T) {
	global.Db.CreateInBatches(models, len(models))
}

// 根据id更新model，更新字段为model中不为空的值，即int类型不为0，ptr类型不为nil这类字段值
// @param model  数据库映射实体模型
func UpdateById(model any) error {
	return UpdateByIdWithDb(global.Db, model)
}

func UpdateByIdWithDb(db *gorm.DB, model any) error {
	return db.Model(model).Updates(model).Error
}

// 根据实体条件，更新参数udpateFields指定字段
func Updates(condition any, udpateFields map[string]any) error {
	return global.Db.Model(condition).Updates(udpateFields).Error
}

// 根据id删除model
// @param model  数据库映射实体模型
func DeleteById(model_ any, id uint64) error {
	return DeleteByIdWithDb(global.Db, model_, id)
}

// 根据id使用指定gromDb删除
func DeleteByIdWithDb(db *gorm.DB, model_ any, id uint64) error {
	return db.Model(model_).Where("id = ?", id).Updates(getDeleteColumnValue()).Error
}

// 根据条件删除
// @param model  数据库映射实体模型
func DeleteByCondition(model_ any) error {
	return DeleteByConditionWithDb(global.Db, model_)
}

// 根据条件使用指定gromDb删除
// @param model  数据库映射实体模型
func DeleteByConditionWithDb(db *gorm.DB, model_ any) error {
	return db.Model(model_).Where(model_).Updates(getDeleteColumnValue()).Error
}

func getDeleteColumnValue() map[string]any {
	columnValue := make(map[string]any)
	columnValue[model.DeletedColumn] = model.ModelDeleted
	columnValue[model.DeleteTimeColumn] = time.Now()
	return columnValue
}

func Tx(funcs ...func(db *gorm.DB) error) (err error) {
	tx := global.Db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			err = fmt.Errorf("%v", err)
		}
	}()
	for _, f := range funcs {
		err = f(tx)
		if err != nil {
			tx.Rollback()
			return
		}
	}
	err = tx.Commit().Error
	return
}
