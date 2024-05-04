package utils

import (
	pkgReq "DiTing-Go/pkg/domain/vo/req"
	pkgResp "DiTing-Go/pkg/domain/vo/resp"
	"fmt"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"reflect"
	"regexp"
	"time"
)

// Paginate 是通用的游标分页函数
// TODO: select部分字段
func Paginate(db *gorm.DB, params pkgReq.PageReq, result interface{}, cursorFieldName string, isAsc bool, conditions ...interface{}) (*pkgResp.PageResp, error) {
	var resp pkgResp.PageResp

	query := db
	if len(conditions) > 0 {
		query = query.Where(conditions[0], conditions[1:]...)
	}

	if params.Cursor != nil && *params.Cursor != "" {
		query = query.Where(fmt.Sprintf("%s < ?", cursorFieldName), *params.Cursor)
	}

	if isAsc {
		query = query.Order(fmt.Sprintf("%s ASC", cursorFieldName))
	} else {
		query = query.Order(fmt.Sprintf("%s DESC", cursorFieldName))
	}
	query = query.Limit(params.PageSize).Find(result)
	if query.Error != nil {
		return &resp, query.Error
	}

	// 获取查询结果的切片值
	slice := reflect.ValueOf(result).Elem()
	// 根据记录条数是否等于页大小判断是否是最后一页
	lastItemIndex := slice.Len()
	if lastItemIndex < params.PageSize {
		resp.IsLast = true
	} else {
		resp.IsLast = false
	}

	// 通过反射获取cursorFieldName对应的值
	if lastItemIndex > 0 {
		lastItem := slice.Index(lastItemIndex - 1)
		fieldsMap, err := GetTagName(result)
		if err != nil {
			return nil, err
		}
		cursorValue := lastItem.FieldByName(fieldsMap[cursorFieldName])
		//获取cursorValue的类型
		cursorType := cursorValue.Type()
		// 如果是Time.time类型，转换为时间戳
		cursorStr := ""
		if cursorType.String() == "time.Time" {
			cursorStr = fmt.Sprint(cursorValue.Interface().(time.Time).UnixNano())
		} else {
			cursorStr = fmt.Sprint(cursorValue.Interface())
		}
		resp.Cursor = &cursorStr
	}

	resp.Data = result
	return &resp, nil
}

// GetTagName 获取结构体中Tag的值，如果没有tag则返回字段值
func GetTagName(structName interface{}) (map[string]string, error) {
	t := reflect.TypeOf(structName).Elem().Elem()
	fieldNum := t.NumField()
	result := make(map[string]string, fieldNum)
	for i := 0; i < fieldNum; i++ {
		fieldName := t.Field(i).Name
		tag := t.Field(i).Tag.Get("gorm")
		if tag == "" {
			result[fieldName] = fieldName
		} else {
			// 定义正则表达式
			re := regexp.MustCompile(`column:([^;]+)`)
			// 使用正则表达式查找匹配项
			match := re.FindStringSubmatch(tag)
			column := fieldName
			if len(match) > 1 {
				column = match[1]
			} else {
				return nil, errors.New("without column error")
			}
			result[column] = fieldName
		}
	}
	return result, nil
}
