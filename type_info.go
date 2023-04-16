package scf

import (
	"reflect"
	"sort"
	"sync"
)

var (
	DefaultTypeInfoCache = NewTypeInfoCache()
)

func GetTypeInfo(t reflect.Type) *TypeInfo {
	return DefaultTypeInfoCache.GetTypeInfo(t)
}

type TypeInfo struct {
	Type   reflect.Type
	Fields Fields
}

type TypeInfoCache struct {
	sync.RWMutex
	cache map[reflect.Type]*TypeInfo
}

func NewTypeInfoCache() *TypeInfoCache {
	return &TypeInfoCache{
		cache: make(map[reflect.Type]*TypeInfo),
	}
}

func (cache *TypeInfoCache) GetTypeInfo(t reflect.Type) *TypeInfo {
	t = removeIndirect(t)

	cache.RLock()
	typeInfo, ok := cache.cache[t]
	cache.RUnlock()

	if ok {
		return typeInfo
	}

	typeInfo = &TypeInfo{Type: t}
	if t.Kind() == reflect.Struct {
		typeInfo.Fields = make(Fields, 0).Append(nil, t)
		sort.Sort(typeInfo.Fields)
	}

	cache.Lock()
	cache.cache[t] = typeInfo
	cache.Unlock()

	return typeInfo
}
