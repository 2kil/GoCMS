package handlers

import (
	"sync"
	"time"

	"shuaitesteel.com/cms/database"
	"shuaitesteel.com/cms/models"
)

type cacheEntry[T any] struct {
	data    T
	expires time.Time
}

type cache struct {
	mu         sync.RWMutex
	settings   cacheEntry[map[string]string]
	customTags cacheEntry[map[string]string]
	columns    cacheEntry[[]models.Column]
	columnsAll cacheEntry[[]models.Column]
}

var frontCache = &cache{}

const cacheTTL = 60 * time.Second

func (c *cache) getSettings() map[string]string {
	c.mu.RLock()
	if !c.settings.expires.IsZero() && time.Now().Before(c.settings.expires) {
		defer c.mu.RUnlock()
		return c.settings.data
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.settings.expires.IsZero() && time.Now().Before(c.settings.expires) {
		return c.settings.data
	}

	var settings []models.Setting
	database.DB.Find(&settings)
	m := make(map[string]string)
	for _, s := range settings {
		m[s.Key] = s.Value
	}
	c.settings = cacheEntry[map[string]string]{data: m, expires: time.Now().Add(cacheTTL)}
	return m
}

func (c *cache) getCustomTags() map[string]string {
	c.mu.RLock()
	if !c.customTags.expires.IsZero() && time.Now().Before(c.customTags.expires) {
		defer c.mu.RUnlock()
		return c.customTags.data
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.customTags.expires.IsZero() && time.Now().Before(c.customTags.expires) {
		return c.customTags.data
	}

	var tags []models.CustomTag
	database.DB.Find(&tags)
	m := make(map[string]string)
	for _, tag := range tags {
		m[tag.Key] = tag.Value
	}
	c.customTags = cacheEntry[map[string]string]{data: m, expires: time.Now().Add(cacheTTL)}
	return m
}

func (c *cache) getColumns() []models.Column {
	c.mu.RLock()
	if !c.columns.expires.IsZero() && time.Now().Before(c.columns.expires) {
		defer c.mu.RUnlock()
		return c.columns.data
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.columns.expires.IsZero() && time.Now().Before(c.columns.expires) {
		return c.columns.data
	}

	var columns []models.Column
	database.DB.Where("parent_id IS NULL").Order("sort_order asc").Find(&columns)
	c.columns = cacheEntry[[]models.Column]{data: columns, expires: time.Now().Add(cacheTTL)}
	return columns
}

func (c *cache) getColumnsAll() []models.Column {
	c.mu.RLock()
	if !c.columnsAll.expires.IsZero() && time.Now().Before(c.columnsAll.expires) {
		defer c.mu.RUnlock()
		return c.columnsAll.data
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.columnsAll.expires.IsZero() && time.Now().Before(c.columnsAll.expires) {
		return c.columnsAll.data
	}

	var all []models.Column
	database.DB.Order("sort_order asc").Find(&all)
	c.columnsAll = cacheEntry[[]models.Column]{data: all, expires: time.Now().Add(cacheTTL)}
	return all
}

func InvalidateCache() {
	frontCache.mu.Lock()
	frontCache.settings = cacheEntry[map[string]string]{}
	frontCache.customTags = cacheEntry[map[string]string]{}
	frontCache.columns = cacheEntry[[]models.Column]{}
	frontCache.columnsAll = cacheEntry[[]models.Column]{}
	frontCache.mu.Unlock()
}
