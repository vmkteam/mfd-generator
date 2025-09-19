//nolint:dupl
package vt

import (
	"time"

	"github.com/vmkteam/mfd-generator/generators/testdata/expected/db"
)

type Category struct {
	ID          int    `json:"id"`
	Title       string `json:"title" validate:"required,max=255"`
	OrderNumber int    `json:"orderNumber" validate:"required"`
	StatusID    int    `json:"statusId" validate:"required,status"`

	Status *Status `json:"status"`
}

func (c *Category) ToDB() *db.Category {
	if c == nil {
		return nil
	}

	category := &db.Category{
		ID:          c.ID,
		Title:       c.Title,
		OrderNumber: c.OrderNumber,
		StatusID:    c.StatusID,
	}

	return category
}

type CategorySearch struct {
	ID          *int    `json:"id"`
	Title       *string `json:"title"`
	OrderNumber *int    `json:"orderNumber"`
	StatusID    *int    `json:"statusId"`
	IDs         []int   `json:"ids"`
}

func (cs *CategorySearch) ToDB() *db.CategorySearch {
	if cs == nil {
		return nil
	}

	return &db.CategorySearch{
		ID:          cs.ID,
		TitleILike:  cs.Title,
		OrderNumber: cs.OrderNumber,
		StatusID:    cs.StatusID,
		IDs:         cs.IDs,
	}
}

type CategorySummary struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	OrderNumber int    `json:"orderNumber"`

	Status *Status `json:"status"`
}

type News struct {
	ID          int        `json:"id"`
	Title       string     `json:"title" validate:"required,max=255"`
	Preview     *string    `json:"preview" validate:"omitempty,max=255"`
	Content     *string    `json:"content"`
	CategoryID  int        `json:"categoryId" validate:"required"`
	CountryID   *int       `json:"countryId"`
	RegionID    *int       `json:"regionId"`
	CityID      *int       `json:"cityId"`
	TagIDs      []int      `json:"tagIds"`
	CreatedAt   time.Time  `json:"createdAt"`
	PublishedAt *time.Time `json:"publishedAt"`
	StatusID    int        `json:"statusId" validate:"required,status"`

	Category *CategorySummary `json:"category"`
	Country  *CountrySummary  `json:"country"`
	Region   *RegionSummary   `json:"region"`
	City     *CitySummary     `json:"city"`
	Status   *Status          `json:"status"`
}

func (n *News) ToDB() *db.News {
	if n == nil {
		return nil
	}

	news := &db.News{
		ID:          n.ID,
		Title:       n.Title,
		Preview:     n.Preview,
		Content:     n.Content,
		CategoryID:  n.CategoryID,
		CountryID:   n.CountryID,
		RegionID:    n.RegionID,
		CityID:      n.CityID,
		TagIDs:      n.TagIDs,
		CreatedAt:   n.CreatedAt,
		PublishedAt: n.PublishedAt,
		StatusID:    n.StatusID,
	}

	return news
}

type NewsSearch struct {
	ID          *int       `json:"id"`
	Title       *string    `json:"title"`
	Preview     *string    `json:"preview"`
	Content     *string    `json:"content"`
	CategoryID  *int       `json:"categoryId"`
	CountryID   *int       `json:"countryId"`
	RegionID    *int       `json:"regionId"`
	CityID      *int       `json:"cityId"`
	CreatedAt   *time.Time `json:"createdAt"`
	PublishedAt *time.Time `json:"publishedAt"`
	StatusID    *int       `json:"statusId"`
	IDs         []int      `json:"ids"`
}

func (ns *NewsSearch) ToDB() *db.NewsSearch {
	if ns == nil {
		return nil
	}

	return &db.NewsSearch{
		ID:           ns.ID,
		TitleILike:   ns.Title,
		PreviewILike: ns.Preview,
		ContentILike: ns.Content,
		CategoryID:   ns.CategoryID,
		CountryID:    ns.CountryID,
		RegionID:     ns.RegionID,
		CityID:       ns.CityID,
		CreatedAt:    ns.CreatedAt,
		PublishedAt:  ns.PublishedAt,
		StatusID:     ns.StatusID,
		IDs:          ns.IDs,
	}
}

type NewsSummary struct {
	ID          int        `json:"id"`
	Title       string     `json:"title"`
	Preview     *string    `json:"preview"`
	Content     *string    `json:"content"`
	CategoryID  int        `json:"categoryId"`
	CountryID   *int       `json:"countryId"`
	RegionID    *int       `json:"regionId"`
	CityID      *int       `json:"cityId"`
	CreatedAt   time.Time  `json:"createdAt"`
	PublishedAt *time.Time `json:"publishedAt"`

	Category *CategorySummary `json:"category"`
	Country  *CountrySummary  `json:"country"`
	Region   *RegionSummary   `json:"region"`
	City     *CitySummary     `json:"city"`
	Status   *Status          `json:"status"`
}

type Tag struct {
	ID       int    `json:"id"`
	Title    string `json:"title" validate:"required,max=255"`
	StatusID int    `json:"statusId" validate:"required,status"`

	Status *Status `json:"status"`
}

func (t *Tag) ToDB() *db.Tag {
	if t == nil {
		return nil
	}

	tag := &db.Tag{
		ID:       t.ID,
		Title:    t.Title,
		StatusID: t.StatusID,
	}

	return tag
}

type TagSearch struct {
	ID       *int    `json:"id"`
	Title    *string `json:"title"`
	StatusID *int    `json:"statusId"`
	IDs      []int   `json:"ids"`
}

func (ts *TagSearch) ToDB() *db.TagSearch {
	if ts == nil {
		return nil
	}

	return &db.TagSearch{
		ID:         ts.ID,
		TitleILike: ts.Title,
		StatusID:   ts.StatusID,
		IDs:        ts.IDs,
	}
}

type TagSummary struct {
	ID    int    `json:"id"`
	Title string `json:"title"`

	Status *Status `json:"status"`
}
