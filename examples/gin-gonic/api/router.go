package api

import (
	"github.com/gin-gonic/gin"
	"github.com/jakoblorz/scf"
)

func GetParams[T interface{}](c *gin.Context) *T {
	val, ok := c.Get("params")
	if !ok {
		return new(T)
	}
	return val.(*T)
}

func PutParams(c *gin.Context, params interface{}) {
	c.Set("params", params)
}

func GetQuery[T interface{}](c *gin.Context) *T {
	val, ok := c.Get("query")
	if !ok {
		return new(T)
	}
	return val.(*T)
}

func PutQuery(c *gin.Context, query interface{}) {
	c.Set("query", query)
}

func GetPayload[T interface{}](c *gin.Context) *T {
	val, ok := c.Get("payload")
	if !ok {
		return new(T)
	}
	return val.(*T)
}

func PutPayload(c *gin.Context, body interface{}) {
	c.Set("payload", body)
}

var (
	router = scf.Registry[gin.HandlerFunc]{}
)

func Router() *scf.Registry[gin.HandlerFunc] {
	return &router
}
