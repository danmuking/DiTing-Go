package cursor

import (
	"fmt"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"reflect"
	"regexp"
)

type PageReq struct {
	Cursor   *string `json:"cursor" form:"cursor"`
	PageSize int     `json:"page_size" form:"page_size"`
}
type PageResp struct {
	Cursor *string `json:"cursor" form:"cursor"`
	IsLast bool    `json:"is_last" form:"is_last"`
	Data   any     `json:"data" form:"data"`
}

// Paginate 是通用的游标分页函数
// TODO: select部分字段
func Paginate(db *gorm.DB, params PageReq, result interface{}, cursorFieldName string, isAsc bool, conditions ...interface{}) (*PageResp, error) {
	var resp PageResp

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
		cursorStr := fmt.Sprint(cursorValue.Interface())
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
