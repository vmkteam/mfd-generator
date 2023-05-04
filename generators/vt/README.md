## VT

vt - генератор серверной части vt. В качестве источника данных используется mfd файл. На выходе - несколько golang файлов

### Использование

Генератор считывает информацию из mfd файла о vt-неймспейсах, загружает каждый их них. Генерирует golang файлы c неймспейсом в качестве префикса.  
Файлы записываются в папку указанную в параметре `-o --output`  
Результат генеририрования ссылается на модели, которые сгенерированы генератором [model](/generators/model) и [repo](/generators/repo), следовательно, код не должен располагаться в том же пакете  
Сгенерированный код предназначен для использования в качестве сервиса zenrpc. В консоль выводится код, который можно использовать в настройке zenrpc  
Зависит от пакета "embedlog" (например, "blog-api/pkg/embedlog") и пакета "github.com/vmkteam/zenrpc", который используется как сервис  

### CLI

```
Create vt from xml

Usage:
  mfd vt [flags]

Flags:
  -o, --output string        output dir path
  -m, --mfd string           mfd file path
  -x, --model string         package containing model files got with model generator
  -p, --package string       package name that will be used in golang files. if not set - last element of output path will be used
  -n, --namespaces strings   namespaces to generate. separate by comma
  -h, --help                 help for vt
```

`-p, --package` задаёт имя пакета для генерируемого файла. Если не задан - в качестве значения будет использоваться последний элемент значения флага `-o --output`    
`-x, --model` задаёт имя пакета, который будет использоваться для ссылок на результат генерирования [модели](/generators/model)

#### console output

```go
const (
        NSAuth = "auth" // стандартный неймспейс для авторизации
        
        NSPost = "post"   // на vt-сущность создаётся свой неймспейс в zenrpc
        NSTag = "tag"     // будут перечислены только те vt-сущности, которые попадают в vt-неймспейсы, 
        NSUser = "user"   // указанные в соотвествуюем параметре --namespaces
)

// services
rpc.RegisterAll(map[string]zenrpc.Invoker{
        NSAuth: NewAuthService(dbo, logger),
        
        NSPost: NewPostService(dbo, logger), // каждая сущность регистрируется в zen-rpc
        NSTag: NewTagService(dbo, logger),
        NSUser: NewUserService(dbo, logger),
})
```

#### MODE

Значение Mode vt-сущности в vt.xml определяет какие файлы будут сгенерировны.  
- "Full" - все файлы
- "ReadOnlyWithTemplates" - все файлы в read-only режиме
- "ReadOnly" - только модели model.go
- "None" -  файлы генерироваться не будут

#### namespace_model.go

```go
//nolint:dupl
package vt  // значение параметра -p --package

import (
	"time"

	"apisrv/db"  // значение параметра -x --model
)

// каждая vt-сущность генерирует свою структуру. структура используется для общения с интерфейсоной частью.
type Post struct {
    // здесь перечислены только те vt-атрибуты, которые имеют не пустое значение AttrName
    // ID - Name vt-атрибута, тип берется из соотвествующего атрибута и поля GoType
	ID        int       `json:"id"`                                      
	Alias     string    `json:"alias" validate:"required,alias,max=255"` // опции валидации добавляются в аннотации
	Title     string    `json:"title" validate:"required,max=255"`       // json имя - Name vt-атрибута с маленькой буквы
	Text      string    `json:"text" validate:"required"`
	Views     int       `json:"views" validate:"required"`
	CreatedAt time.Time `json:"createdAt"`
	UserID    int       `json:"userId" validate:"required"`
	TagIDs    []int     `json:"tagIds"`                                   
	StatusID  int       `json:"statusId" validate:"required,status"`

	User   *User   `json:"user"`   // дополнительно сгенерируется список внешних vt-моделей, которые указаны в параметре FK соотвествующего атрибута 
	Status *Status `json:"status"` // и поле Status если есть vt-атрибут StatusID
}

// конвертер из vt-модели модель базы данных. используется для выполнения действий над данными в бд
func (p *Post) ToDB() *db.Post {
	if p == nil {
		return nil
	}

	post := &db.Post{
		ID:        p.ID,
		Alias:     p.Alias,
		Title:     p.Title,
		Text:      p.Text,
		Views:     p.Views,
		CreatedAt: p.CreatedAt,
		UserID:    p.UserID,
		TagIDs:    p.TagIDs,
		StatusID:  p.StatusID,
	}

    // некоторые поля будут включать проверку на nil если необходимо
    // например для внешних vt-моделей, т.к. на них надо вызывать метод ToDB
    // а так же для полей, где требуется разименование указателя
	if p.User != nil {
		post.User = p.User.ToDB()
	}

	return post
}

// каждая vt-сущность генерирует свою структуру для поиска. структура используется для общения с интерфейсоной частью.
type PostSearch struct {
    // vt-атрибуты у которых указано Search=true попадут в структуру
    // ID - Name vt-атрибута, тип берется из соотвествующего атрибута и поля GoType
	ID        *int       `json:"id"`     // к каждому типу, кроме массивов будет добавлен указатель
	Alias     *string    `json:"alias"`  // таким образом поиск будет учитываться только для полей != nil в этой структуре
	Title     *string    `json:"title"`  // json имя - Name vt-атрибута с маленькой буквы
	Text      *string    `json:"text"`
	Views     *int       `json:"views"`
	CreatedAt *time.Time `json:"createdAt"`
	UserID    *int       `json:"userId"`
	StatusID  *int       `json:"statusId"`
	IDs       []int      `json:"ids"`
	NotID     *int       `json:"notId"`
}

// конвертер в db поиск. транслирует поиск полученный из интерфейса до базы данных
func (ps *PostSearch) ToDB() *db.PostSearch {
	if ps == nil {
		return nil
	}

	return &db.PostSearch{
        // здесь перечислены только те атрибуты, у которых Search=true
		ID:         ps.ID,
		Alias:      ps.Alias,
        // если для vt-атрибута задан параметр SearchName в этой структуре будет использоваться ссылка на его значение
		TitleILike: ps.Title,
		TextILike:  ps.Text,
		Views:      ps.Views,
		CreatedAt:  ps.CreatedAt,
		UserID:     ps.UserID,
		StatusID:   ps.StatusID,
		IDs:        ps.IDs,
		NotID:      ps.NotID,
	}
}

// каждая vt-сущность генерирует свою структуру для вывода в списке. структура используется для общения с интерфейсоной частью.
type PostSummary struct {
    // vt-атрибуты у которых указано Summary=true попадут в структуру
    // ID - Name vt-атрибута, тип берется из соотвествующего атрибута и поля GoType
	ID        int       `json:"id"` // json имя - Name vt-атрибута с маленькой буквы
	Alias     string    `json:"alias"`
	Title     string    `json:"title"`
	Text      string    `json:"text"`
	Views     int       `json:"views"`
	CreatedAt time.Time `json:"createdAt"`
	UserID    int       `json:"userId"`

	User   *UserSummary `json:"user"`
	Status *Status      `json:"status"`
}

``` 

#### namespace_converter.go

```go
package vt  // значение параметра -p --package

import (
	"apisrv/db"  // значение параметра -x --model
)

// каждая vt-сущность генерирует свою функцию - конструктор из соотвествующей сущности 
func NewPost(in *db.Post) *Post {
	if in == nil {
		return nil
	}

	post := &Post{
        // каждый vt-атрибут генерирует конвертер из исходного атрибута, указанного в AttrName
		ID:        in.ID,
		Alias:     in.Alias,
		Title:     in.Title,
		Text:      in.Text,
		Views:     in.Views,
		CreatedAt: in.CreatedAt,
		UserID:    in.UserID,
		TagIDs:    in.TagIDs,
		StatusID:  in.StatusID,

        // для внешних vt-сущностей генерируются соответвующие конструкторы
		User:   NewUser(in.User),
		Status: NewStatus(in.StatusID),
	}

	return post
}

// каждая vt-сущность генерирует свою функцию - конструктор summary (для показа в листах)
func NewPostSummary(in *db.Post) *PostSummary {
	if in == nil {
		return nil
	}

	return &PostSummary{
        // здесь перечислены только те vt-атрибуты у корорых Summary=true
		ID:        in.ID,
		Alias:     in.Alias,
		Title:     in.Title,
		Text:      in.Text,
		Views:     in.Views,
		CreatedAt: in.CreatedAt,
		UserID:    in.UserID,

        // для внешних vt-сущностей генерируются соответвующие конструкторы
		User:   NewUserSummary(in.User),
		Status: NewStatus(in.StatusID),
	}
}

```

#### namespace.go

```go
package vt  // значение параметра -p --package

import (
	"context"

	"apisrv/db"  // значение параметра -x --model
	"apisrv/pkg/embedlog" //TODO undo hardcode

	"github.com/vmkteam/zenrpc/v2"
)

// каждая vt-сущность генерирует свой сервис c именем Name+"Service"
type PostService struct {
	zenrpc.Service
	embedlog.Logger
    // ссылка на соотвествующий репозиторий
	blogRepo db.BlogRepo
}

// конструктор, 'New"+Name+"Service"
func NewPostService(dbo db.DB, logger embedlog.Logger) *PostService {
	return &PostService{
		Logger:   logger,
		blogRepo: db.NewBlogRepo(dbo),
	}
}

func (s PostService) dbSort(ops *ViewOps) db.OpFunc {
	v := s.blogRepo.DefaultPostSort()
	if ops == nil {
		return v
	}

	switch ops.SortColumn {
    // Здесь перечислены vt-атрибуты, кроме массиов, у которых Summary=trueж
	case db.Columns.Post.ID, db.Columns.Post.Alias, db.Columns.Post.Title, db.Columns.Post.Text, db.Columns.Post.Views, db.Columns.Post.CreatedAt, db.Columns.Post.UserID, db.Columns.Post.StatusID:
		v = db.WithSort(db.NewSortField(ops.SortColumn, ops.SortDesc))
	}

	return v
}

// Count Posts according to conditions in search params
//zenrpc:search PostSearch
//zenrpc:return int
//zenrpc:500 Internal Error
func (s PostService) Count(ctx context.Context, search *PostSearch) (int, error) {
	count, err := s.blogRepo.CountPosts(ctx, search.ToDB())
	if err != nil {
		return 0, InternalError(err)
	}
	return count, nil
}

// Get а list of Posts according to conditions in search params
//zenrpc:search PostSearch
//zenrpc:viewOps ViewOps
//zenrpc:return []PostSummary
//zenrpc:500 Internal Error
func (s PostService) Get(ctx context.Context, search *PostSearch, viewOps *ViewOps) ([]PostSummary, error) {
	list, err := s.blogRepo.PostsByFilters(ctx, search.ToDB(), viewOps.Pager(), s.dbSort(viewOps), s.blogRepo.FullPost())
	if err != nil {
		return nil, InternalError(err)
	}
	posts := make([]PostSummary, 0, len(list))
	for i := 0; i < len(list); i++ {
		if post := NewPostSummary(&list[i]); post != nil {
			posts = append(posts, *NewPostSummary(&list[i]))
		}
	}
	return posts, nil
}

// Returns a Post by its ID
//zenrpc:id int
//zenrpc:return Post
//zenrpc:500 Internal Error
//zenrpc:404 Not Found
func (s PostService) GetByID(ctx context.Context, id int) (*Post, error) {
	db, err := s.byID(ctx, id)
	if err != nil {
		return nil, err
	}
	return NewPost(db), nil
}

func (s PostService) byID(ctx context.Context, id int) (*db.Post, error) {
	db, err := s.blogRepo.PostByID(ctx, id, s.blogRepo.FullPost())
	if err != nil {
		return nil, InternalError(err)
	} else if db == nil {
		return nil, ErrNotFound
	}
	return db, nil
}

// Функции ниже не генерируеются для Mode=ReadOnlyWithTemplates

// Add a Post from from the query
//zenrpc:post Post
//zenrpc:return Post
//zenrpc:500 Internal Error
//zenrpc:400 Validation Error
func (s PostService) Add(ctx context.Context, post *Post) (*Post, error) {
	if ve := s.isValid(ctx, post, false); ve.HasErrors() {
		return nil, ve.Error()
	}

	db, err := s.blogRepo.AddPost(ctx, post.ToDB())
	if err != nil {
		return nil, InternalError(err)
	}
	return NewPost(db), nil
}

// Updates the Post data identified by id from the query
//zenrpc:posts Post
//zenrpc:return Post
//zenrpc:500 Internal Error
//zenrpc:400 Validation Error
//zenrpc:404 Not Found
func (s PostService) Update(ctx context.Context, post *Post) (bool, error) {
	if _, err := s.byID(ctx, post.ID); err != nil {
		return false, err
	}

	if ve := s.isValid(ctx, post, true); ve.HasErrors() {
		return false, ve.Error()
	}

	ok, err := s.blogRepo.UpdatePost(ctx, post.ToDB())
	if err != nil {
		return false, InternalError(err)
	}
	return ok, nil
}

// Delete the Post by its ID
//zenrpc:id int
//zenrpc:return isDeleted
//zenrpc:500 Internal Error
//zenrpc:400 Validation Error
//zenrpc:404 Not Found
func (s PostService) Delete(ctx context.Context, id int) (bool, error) {
	if _, err := s.byID(ctx, id); err != nil {
		return false, err
	}

	ok, err := s.blogRepo.DeletePost(ctx, id)
	if err != nil {
		return false, InternalError(err)
	}
	return ok, err
}

// Verifies that Post data is valid
//zenrpc:post Post
//zenrpc:return []FieldError
//zenrpc:500 Internal Error
func (s PostService) Validate(ctx context.Context, post *Post) ([]FieldError, error) {
	isUpdate := post.ID != 0

	if isUpdate {
		_, err := s.byID(ctx, post.ID)
		if err != nil {
			return nil, err
		}
	}

	ve := s.isValid(ctx, post, isUpdate)
	if ve.HasInternalError() {
		return nil, ve.Error()
	}

	return ve.Fields(), nil
}

func (s PostService) isValid(ctx context.Context, post *Post, isUpdate bool) Validator {
	var v Validator

	if v.CheckBasic(ctx, *post); v.HasInternalError() {
		return v
	}

    //для vt-атрибутов с именем alias будет добавлена следующая проверка.
	//check alias unique
	search := &db.PostSearch{
		Alias: &post.Alias,
		NotID: &post.ID,
	}
	item, err := s.blogRepo.OnePost(ctx, search)
	if err != nil {
		v.SetInternalError(err)
	} else if item != nil {
		v.Append("alias", FieldErrorUnique)
	}

    // для vt-атрибутов с внешними ключами
	// check fks
	if post.UserID != 0 {
		item, err := s.commonRepo.UserByID(ctx, post.UserID)
		if err != nil {
			v.SetInternalError(err)
		} else if item == nil {
			v.Append("userId", FieldErrorIncorrect)
		}
	}
    // для vt-атрибутов с внешними ключами в виде массов
	if len(post.TagIDs) != 0 {
		items, err := s.blogRepo.TagsByFilters(ctx, &db.TagSearch{IDs: post.TagIDs}, db.PagerNoLimit)
		if err != nil {
			v.SetInternalError(err)
		} else if len(items) != len(post.TagIDs) {
			v.Append("tagIds", FieldErrorIncorrect)
		}
	}
	//custom validation starts here
	return v
}

```

#### Особенности работы с существующими моделями

Все файлы будут перезаписаны при каждой генерации.
