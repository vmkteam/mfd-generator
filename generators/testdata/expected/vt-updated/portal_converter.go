package vt

import (
	"github.com/vmkteam/mfd-generator/generators/testdata/expected/db"
)

func NewCategory(in *db.Category) *Category {
	if in == nil {
		return nil
	}

	category := &Category{
		ID:          in.ID,
		Title:       in.Title,
		OrderNumber: in.OrderNumber,
		StatusID:    in.StatusID,

		Status: NewStatus(in.StatusID),
	}

	return category
}

func NewCategorySummary(in *db.Category) *CategorySummary {
	if in == nil {
		return nil
	}

	return &CategorySummary{
		ID:          in.ID,
		Title:       in.Title,
		OrderNumber: in.OrderNumber,

		Status: NewStatus(in.StatusID),
	}
}

func NewNewsSummary(in *db.News) *NewsSummary {
	if in == nil {
		return nil
	}

	return &NewsSummary{
		ID:          in.ID,
		Title:       in.Title,
		Preview:     in.Preview,
		Content:     in.Content,
		CategoryID:  in.CategoryID,
		CreatedAt:   in.CreatedAt,
		PublishedAt: in.PublishedAt,

		Category: NewCategorySummary(in.Category),
		Status:   NewStatus(in.StatusID),
	}
}

func NewNews(in *db.News) *News {
	if in == nil {
		return nil
	}

	news := &News{
		ID:          in.ID,
		Title:       in.Title,
		Preview:     in.Preview,
		Content:     in.Content,
		CategoryID:  in.CategoryID,
		TagIDs:      in.TagIDs,
		CreatedAt:   in.CreatedAt,
		PublishedAt: in.PublishedAt,
		StatusID:    in.StatusID,

		Category: NewCategorySummary(in.Category),
		Status:   NewStatus(in.StatusID),
	}

	return news
}

func NewTag(in *db.Tag) *Tag {
	if in == nil {
		return nil
	}

	tag := &Tag{
		ID:       in.ID,
		Title:    in.Title,
		StatusID: in.StatusID,

		Status: NewStatus(in.StatusID),
	}

	return tag
}

func NewTagSummary(in *db.Tag) *TagSummary {
	if in == nil {
		return nil
	}

	return &TagSummary{
		ID:    in.ID,
		Title: in.Title,

		Status: NewStatus(in.StatusID),
	}
}
