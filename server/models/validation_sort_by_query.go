package models

import (
	"10-typing/errors"
	"fmt"
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
	const op errors.Op = "models.BindSortByQuery"

	sortByValues := c.QueryArray("sort_by")

	var sortOptions = make([]SortOption, 0, len(sortByValues))

	for _, sortByValue := range sortByValues {
		options := strings.Split(sortByValue, ".")
		if len(options) != 2 {
			err := fmt.Errorf("invalid sort_by query")
			return nil, errors.E(op, err)
		}

		sortOptionWithValidationStructTag := reflect.New(reflect.TypeOf(sortOptionStruct)).Elem()

		sortOptionWithValidationStructTag.FieldByName("Column").SetString(options[0])
		sortOptionWithValidationStructTag.FieldByName("Order").SetString(options[1])

		if err := validateStruct(sortOptionWithValidationStructTag.Interface()); err != nil {
			return nil, errors.E(op, err)
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
	const op errors.Op = "models.validateStruct"

	validate := validator.New()
	err := validate.Struct(s)
	if err != nil {
		var b strings.Builder
		for _, err := range err.(validator.ValidationErrors) {
			fmt.Fprintf(&b, "\n%s field validation failed on the %s rule", err.Field(), err.Tag())
		}
		err := fmt.Errorf("%w: %s", err, b.String())

		return errors.E(op, err)
	}

	return nil
}
