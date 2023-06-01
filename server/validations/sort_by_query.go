package validations

import (
	"errors"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type SortOption struct {
	Column string
	Order  string
}

func BindSortByQuery(c *gin.Context, sortOptionStruct interface{}) ([]SortOption, error) {
	sortByValues := c.QueryArray("sort_by")

	var sortOptions = make([]SortOption, 0, len(sortByValues))

	for _, sortByValue := range sortByValues {
		options := strings.Split(sortByValue, ".")
		if len(options) != 2 {
			return nil, errors.New("invalid sort_by query")
		}

		sortOptionWithValidationStructTag := reflect.New(reflect.TypeOf(sortOptionStruct)).Elem()

		sortOptionWithValidationStructTag.FieldByName("Column").SetString(options[0])
		sortOptionWithValidationStructTag.FieldByName("Order").SetString(options[1])

		if err := validateStruct(sortOptionWithValidationStructTag.Interface()); err != nil {
			return nil, err
		}

		sortOption := SortOption{
			Column: options[0],
			Order:  options[1],
		}

		sortOptions = append(sortOptions, sortOption)
	}

	return sortOptions, nil
}

func validateStruct(s interface{}) error {
	validate := validator.New()
	err := validate.Struct(s)
	if err != nil {
		var errMsg string
		for _, err := range err.(validator.ValidationErrors) {
			errMsg = err.Field() + " field validation failed on the " + err.Tag() + " rule"
		}
		return errors.New(errMsg)
	}
	return nil
}
