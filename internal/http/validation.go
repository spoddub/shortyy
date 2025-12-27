package httpapi

import (
	"errors"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func setupValidator() {
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return
	}

	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := fld.Tag.Get("json")
		if name == "" {
			return fld.Name
		}
		name = strings.Split(name, ",")[0]
		if name == "-" {
			return ""
		}
		return name
	})

	_ = v.RegisterValidation("shortname", func(fl validator.FieldLevel) bool {
		s := fl.Field().String()
		return shortNameRe.MatchString(s)
	})
}

func writeBindError(c *gin.Context, err error) bool {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		out := make(map[string]string, len(ve))
		for _, fe := range ve {
			field := fe.Field()
			if field == "" {
				field = strings.ToLower(fe.StructField())
			}
			out[field] = fe.Error()
		}

		c.JSON(422, gin.H{"errors": out})
		return true
	}

	c.JSON(400, gin.H{"error": "invalid request"})
	return true
}

func writeUniqueShortNameError(c *gin.Context) {
	c.JSON(422, gin.H{"errors": gin.H{"short_name": "short name already in use"}})
}
