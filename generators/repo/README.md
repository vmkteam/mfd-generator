## REPO

repo - генератор репозиториев: структур призванных облегчить работу с моделями, с функциями добавления, редактирования, получения и др. В качестве источника данных используется mfd файл. На выходе - несколько golang файлов

### Использование

Генератор считывает информацию из mfd файла о неймспейсах, загружает каждый их них. Генерирует golang файлы c именем неймспейса.  
Файлы записываются в папку указанную в параметре `-o --output`  
Результат генеририрования зависит от кода, который сгенерирован генератором [model](/generators/model), следовательно, код должен располагаться в одном пакете  
Так же генератор использует общие компоненты: Filter, SortField и другие  

### CLI

```
Create repo from xml

Usage:
  mfd repo [flags]

Flags:
  -o, --output string        output dir path
  -m, --mfd string           mfd file path
  -p, --package string       package name that will be used in golang files. if not set - last element of output path will be used
  -n, --namespaces strings   namespaces to generate. separate by comma
  -h, --help                 help for repo
```

`-p, --package` задаёт имя пакета для генерируемого файла. Если не задан - в качестве значения будет использоваться последний элемент значения флага `-o --output`

#### namespace.go

```go
package db // значение параметра -p --package

import (
	"context"
 
    // если в mfd файле указана 9 версия импорты будут иметь постфикс /v9
	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"
)

// Для имени репозитория используется поле Name сущности
type BlogRepo struct {
	db      orm.DB
	filters map[string][]Filter
	sort    map[string][]SortField
	join    map[string][]string
}

// NewBlogRepo returns new repository
func NewBlogRepo(db orm.DB) BlogRepo {
	return BlogRepo{
		db: db,
		filters: map[string][]Filter{
            // StatusFilter сгенерируется только если есть атрибут StatusID                 
			Tables.Post.Name: {StatusFilter},
			Tables.Tag.Name:  {StatusFilter},
		},
		sort: map[string][]SortField{
            // при наличии атрибутов CreatedAt, Title они будут добавлены в сортировки по-умолчанию
			Tables.Post.Name: {{Column: Columns.Post.CreatedAt, Direction: SortDesc}},
			Tables.Tag.Name:  {{Column: Columns.Tag.Title, Direction: SortAsc}},
		},
		join: map[string][]string{
            // все атрибуты с FK будут указаны в join для соответствующей операции
			Tables.Post.Name: {TableColumns, Columns.Post.User},
			Tables.Tag.Name:  {TableColumns},
		},
	}
}

// WithTransaction is a function that wraps BlogRepo with pg.Tx transaction.
// для имени ресивера будут использоваться заглавные буквы сущности + r
func (br BlogRepo) WithTransaction(tx *pg.Tx) BlogRepo {
	br.db = tx
	return br
}

// WithEnabledOnly is a function that adds "statusId"=1 as base filter.
func (br BlogRepo) WithEnabledOnly() BlogRepo {
	f := make(map[string][]Filter, len(br.filters))
	for i := range br.filters {
		f[i] = make([]Filter, len(br.filters[i]))
		copy(f[i], br.filters[i])
		f[i] = append(f[i], StatusEnabledFilter)
	}
	br.filters = f

	return br
}

// Далее код для каждой из сущностей
/*** Post ***/

// FullPost returns full joins with all columns
func (br BlogRepo) FullPost() OpFunc {
	return WithColumns(br.join[Tables.Post.Name]...)
}

// DefaultPostSort returns default sort.
func (br BlogRepo) DefaultPostSort() OpFunc {
	return WithSort(br.sort[Tables.Post.Name]...)
}

// PostByID is a function that returns Post by ID(s) or nil.
// Если сущность имеет несколько PK атрибутов - они будут перечислены как аргументы в функциях, использующих id
func (br BlogRepo) PostByID(ctx context.Context, id int, ops ...OpFunc) (*Post, error) {
	return br.OnePost(ctx, &PostSearch{ID: &id}, ops...)
}

// OnePost is a function that returns one Post by filters. It could return pg.ErrMultiRows.
// PostSearch - структура поиска, которая генерируется в генераторе model
// Post - структура модели, которая генерируется в генераторе model
func (br BlogRepo) OnePost(ctx context.Context, search *PostSearch, ops ...OpFunc) (*Post, error) {
	obj := &Post{}
	err := buildQuery(ctx, br.db, obj, search, br.filters[Tables.Post.Name], PagerTwo, ops...).Select()

	switch err {
	case pg.ErrMultiRows:
		return nil, err
	case pg.ErrNoRows:
		return nil, nil
	}

	return obj, err
}

// PostsByFilters returns Post list.
func (br BlogRepo) PostsByFilters(ctx context.Context, search *PostSearch, pager Pager, ops ...OpFunc) (posts []Post, err error) {
	err = buildQuery(ctx, br.db, &posts, search, br.filters[Tables.Post.Name], pager, ops...).Select()
	return
}

// CountPosts returns count
func (br BlogRepo) CountPosts(ctx context.Context, search *PostSearch, ops ...OpFunc) (int, error) {
	return buildQuery(ctx, br.db, &Post{}, search, br.filters[Tables.Post.Name], PagerOne, ops...).Count()
}

// AddPost adds Post to DB.
func (br BlogRepo) AddPost(ctx context.Context, post *Post, ops ...OpFunc) (*Post, error) {
	q := br.db.ModelContext(ctx, post)
	applyOps(q, ops...)
	_, err := q.Insert()

	return post, err
}

// UpdatePost updates Post in DB.
func (br BlogRepo) UpdatePost(ctx context.Context, post *Post, ops ...OpFunc) (bool, error) {
	q := br.db.ModelContext(ctx, post).WherePK()
	applyOps(q, ops...)
	res, err := q.Update()
	if err != nil {
		return false, err
	}

	return res.RowsAffected() > 0, err
}

// DeletePost set statusId to deleted in DB.
// Код функции зависит от того, есть ли атрибут StatusID, если нет - будет сгенерирована функция физического удаления из базы
func (br BlogRepo) DeletePost(ctx context.Context, id int) (deleted bool, err error) {
	post := &Post{ID: id, StatusID: StatusDeleted}

	return br.UpdatePost(ctx, post, WithColumns(Columns.Post.StatusID))
}
``` 

#### Особенности работы с существующими моделями

Все файлы будут перезаписаны при каждой генерации.
