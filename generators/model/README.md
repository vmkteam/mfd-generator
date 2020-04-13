## MODEL

model - генератор модели: структур призванных облегчить работу с базой данных и с сущностями в ней. В качестве источника данных используется mfd файл. На выходе - 4 golang файла

### Использование

Генератор считывает информацию из mfd файла о неймспейсах, загружает каждый их них. Генерирует golang файлы:  
- model.go - описание всех сущностей из xml в виде структур и списка колонок
- model_search.go - описание всех поисков из xml в виде структур
- model_validate.go - функции для валидации структур. используются при записи в базу
- model_params.go - структуры для json(b) атрибутов.  

Файлы записываются в папку указанную в параметре `-o --output`

### CLI
```
Create golang model from xml

Usage:
  mfd model [flags]

Flags:
  -o, --output string    output dir path
  -m, --mfd string       mfd file path
  -p, --package string   package name that will be used in golang files. if not set - last element of output path will be used
  -h, --help             help for model
```

`-p, --package` задаёт имя пакета для генерируемого файла. Если не задан - в качестве значения будет использоваться последний элемент значения флага `-o --output`

#### model.go 

Приведено в сокращённом варианте

```go
//nolint
//lint:file-ignore U1000 ignore unused code, it's generated
package db // значение параметра -p --package

import (
	"time"
)

// Список колонок для построения sql запросов, например в функции Where(`? = 'test'`, pg.F(Columns.Post.Title))
var Columns = struct {
    // сущность, используется параметр значение Name сущности 
	Post struct {
        // список колонок, генерируется из секции Attributes, используется значение Name атрибута
		ID, Alias, Title, Text, Views, CreatedAt, UserID, TagIDs, StatusID string
        
        // список fk-колонок, генерируется из секции Attributes если задано значение FK, используется значение Name атрибута
		User string
	}
}{
	Post: struct {
		ID, Alias, Title, Text, Views, CreatedAt, UserID, TagIDs, StatusID string

		User string
	}{
        // список колонок, генерируется из секции Attributes, используется значение DBName атрибута
		ID:        "postId",
		Alias:     "alias",
		Title:     "title",
		Text:      "text",
		Views:     "views",
		CreatedAt: "createdAt",
		UserID:    "userId",
		TagIDs:    "tagIds",
		StatusID:  "statusId",

		User: "User",
	},
}

// Список сущностей с именами таблиц и алиасами, для построения sql запросов.
var Tables = struct {
    // сущность, используется параметр значение Name сущности
	Post struct {
		Name, Alias string
	}
}{
	Post: struct {
		Name, Alias string
	}{
		Name:  "posts", // имя таблицы, используется значение Table
		Alias: "t", // алиас таблицы, используется в фильтрах, чтобы ссылаться на текущую таблицу. значение всегда "t"
	},
}

// Сущность, используется параметр значение Name сущности. Используется как модель в pg.Model(&Post{})
type Post struct {
	// sql:"posts" <- используется значение Table сущности
	tableName struct{} `sql:"posts,alias:t" pg:",discard_unknown_columns"`

    // ID - Name атрибута
    // int - GoType атрибута
    // pg:"postId" - DBName атрибута
    // pg:"postId,pk" - для PK
    // pg:"alias,use_zero" - для Nullable=No
    // pg:"tagIds,array" - для IsArray=true
    // pg:"fk:userId" - для FK
    ID        int       `pg:"postId,pk"`
	Alias     string    `pg:"alias,use_zero"`
	Title     string    `pg:"title,use_zero"`
	Text      string    `pg:"text,use_zero"`
	Views     int       `pg:"views,use_zero"`
	CreatedAt time.Time `pg:"createdAt,use_zero"`
	UserID    int       `pg:"userId,use_zero"`
	TagIDs    []int     `pg:"tagIds,array"`
	StatusID  int       `pg:"statusId,use_zero"`

	User *User `pg:"fk:userId"`
}

type User struct {
	tableName struct{} `sql:"users,alias:t" pg:",discard_unknown_columns"`

	ID          int         `sql:"userId,pk"`
	Email       string      `sql:"email,notnull"`
	Password    string      `sql:"password,notnull"`
	Active      bool        `sql:"active,notnull"`
    // для json(b) типов будут сгенерированы специальные структуры в отдельный файд model_params.go
	Params      *UserParams `sql:"params"`
	StatusID    int         `sql:"statusId,notnull"`
	LastLoginAt *time.Time  `sql:"lastLoginAt"`
}
```

#### model_search.go 

```go
//nolint
//lint:file-ignore U1000 ignore unused code, it's generated
package db // значение параметра -p --package

import (
	"time"

	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"
)

// Стандартные фильтры, генерируется всегда

const condition = "?.? = ?"

// base filters
type applier func(query *orm.Query) (*orm.Query, error)

type search struct {
	appliers []applier
}

func (s *search) apply(query *orm.Query) {
	for _, applier := range s.appliers {
		query.Apply(applier)
	}
}

func (s *search) where(query *orm.Query, table, field string, value interface{}) {
	query.Where(condition, pg.Ident(table), pg.Ident(field), value)
}

func (s *search) WithApply(a applier) {
	if s.appliers == nil {
		s.appliers = []applier{}
	}
	s.appliers = append(s.appliers, a)
}

func (s *search) With(condition string, params ...interface{}) {
	s.WithApply(func(query *orm.Query) (*orm.Query, error) {
		return query.Where(condition, params...), nil
	})
}

// Searcher is interface for every generated filter
type Searcher interface {
	Apply(query *orm.Query) *orm.Query
	Q() applier

	With(condition string, params ...interface{})
	WithApply(a applier)
}

// Фильтры для каждой сущности
// Имя генериуется из значения Name + Search
type PostSearch struct {
	search

    // Список аттрибутов из секции <Attributes>
    // ID - Name атрибута
    // *int - GoType атрибута, всегда добавляется указатель
	ID         *int
	Alias      *string
	Title      *string
	Text       *string
	Views      *int
	CreatedAt  *time.Time
	UserID     *int
	StatusID   *int
    
    // Список поисков из секции <Search>
    // IDs - Name атрибута
    // []int - генерируется на основе GoType атрибута и SearchType поиска, см. Особенности генерирования типов для поисков
	IDs        []int
	NotID      *int
	TitleILike *string
	TextILike  *string
}

// Функция для применения поиска к orm.Query
func (ps *PostSearch) Apply(query *orm.Query) *orm.Query {
	if ps.ID != nil {
		ps.where(query, Tables.Post.Alias, Columns.Post.ID, ps.ID)
	}
	if ps.Alias != nil {
		ps.where(query, Tables.Post.Alias, Columns.Post.Alias, ps.Alias)
	}
	if ps.Title != nil {
		ps.where(query, Tables.Post.Alias, Columns.Post.Title, ps.Title)
	}
	if ps.Text != nil {
		ps.where(query, Tables.Post.Alias, Columns.Post.Text, ps.Text)
	}
	if ps.Views != nil {
		ps.where(query, Tables.Post.Alias, Columns.Post.Views, ps.Views)
	}
	if ps.CreatedAt != nil {
		ps.where(query, Tables.Post.Alias, Columns.Post.CreatedAt, ps.CreatedAt)
	}
	if ps.UserID != nil {
		ps.where(query, Tables.Post.Alias, Columns.Post.UserID, ps.UserID)
	}
	if ps.StatusID != nil {
		ps.where(query, Tables.Post.Alias, Columns.Post.StatusID, ps.StatusID)
	}
	if len(ps.IDs) > 0 {
		Filter{Columns.Post.ID, ps.IDs, SearchTypeArray, false}.Apply(query)
	}
	if ps.NotID != nil {
		Filter{Columns.Post.ID, *ps.NotID, SearchTypeEquals, true}.Apply(query)
	}
	if ps.TitleILike != nil {
		Filter{Columns.Post.Title, *ps.TitleILike, SearchTypeILike, false}.Apply(query)
	}
	if ps.TextILike != nil {
		Filter{Columns.Post.Text, *ps.TextILike, SearchTypeILike, false}.Apply(query)
	}

	ps.apply(query)

	return query
}

// Функция для использования в pg.Model().Apply()
func (ps *PostSearch) Q() applier {
	return func(query *orm.Query) (*orm.Query, error) {
		return ps.Apply(query), nil
	}
}
```

#### model_validate.go

```go
//nolint
//lint:file-ignore U1000 ignore unused code, it's generated
package db // значение параметра -p --package

import (
	"unicode/utf8"
)

// Стандартные константы для формирования массива ошибок
const (
	ErrEmptyValue = "empty"
	ErrMaxLength  = "len"
	ErrWrongValue = "value"
)

// Для каждой сущности сгенерируется функция с именем Validate, если есть атрибуты, требующие валидации
// p Post - ресивер будет сгенерирован из всех заглавных букв имени. Например pv для PostViews
func (p Post) Validate() (errors map[string]string, valid bool) {
	errors = map[string]string{}

    // Каждый атрибут, который можно валидировать сгенерирует здесь проверку на корректные значения 

	if p.ID == 0 {
		errors[Columns.Post.ID] = ErrEmptyValue
	}

    // Для проверки длинны строки. Используется значения Min и Max атрибута 
	if utf8.RuneCountInString(p.Alias) > 255 {
		errors[Columns.Post.Alias] = ErrMaxLength
	}

	if utf8.RuneCountInString(p.Title) > 255 {
		errors[Columns.Post.Title] = ErrMaxLength
	}

    // Для Nullable=No атрибутов
	if p.TagIDs == nil {
		errors[Columns.Post.TagIDs] = ErrEmptyValue
	}

	return errors, len(errors) == 0
}
```

#### model_params.go

```go
package db // значение параметра -p --package

// Для каждого атрибута с DBType json(b) сгенерируется пустая структура
// UserParams - Name атрибута + "Params"
type UserParams struct {
}
```

Для генерирования этого файла используется парсер и генератор AST, поэтому существующий код будет только дополнятся.

#### Особенности генерирования типов для поисков
- Для SEARCHTYPE_ARRAY и SEARCHTYPE_NOT_ARRAY тип поиска всегда оборачивается в массив  
- Для SEARCHTYPE_NULL и SEARCHTYPE_NOT_NULL тип поиска всегда bool
- Для остальных - GoType с обязательным указателем

[MakeSearchType](/mfd/types.go#L8)

#### Особенности работы с существующими моделями

Все файлы, кроме `model_params.go` будут перезаписаны при каждой генерации. `model_params.go` - дополняется несуществующими структурами