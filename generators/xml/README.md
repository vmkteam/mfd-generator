## XML

xml - генератор основы проекта: mfd файла, неймспейсов и сущностей в них. В качестве источника данных используется база данных. Результат - несколько xml файлов. 

### Использование

Генератор подключается к базе данных, считывает информацию о таблицах, отношениях между ними и генерирует xml на основе пользовательского ввода.
Если проект уже существует сначала генератор его загрузит и будет использовать как основу для будущих xml. Также при вводе неймспейсов для таблиц предлагается выбрать из существующих в проекте.

В результате выбранные таблицы сгенерируют сущность (entity). В сущности описаны все поля таблицы в виде атрибутов (attribute). Так же автоматически обновятся стандартные поиски (search). Сущности сгруппированы по неймспейсам, которые сохраняются в файлы с именем неймспейса.
Все неймспейсы также запишутся в mfd файл, который сохранится по указанному в команде пути.

### CLI
```
mfd-generator xml -h 

Create or update project base with namespaces and entities

Usage:
  mfd xml [flags]

Flags:
  -v, --verbose             print sql queries
  -c, --conn string         connection string to postgres database, e.g. postgres://usr:pwd@localhost:5432/db
  -m, --mfd string          mfd file path
  -t, --tables strings      table names for model generation separated by comma
                            use 'schema_name.*' to generate model for every table in model (default [public.*])
  -n, --namespaces string   use this parameter to set table & namespace in format "users=users,projects;shop=orders,prices"
  -p, --print               print namespace - tables association
  -h, --help                help for xml
```
  
`-t, --tables` - позволяет вводить исходные таблицы для генератора через запятую, если не указана схема для таблицы, то будет использоваться public.   
`*` - для генерирования всех таблиц в схеме, например: `public.*,geo.locations,geo.cities`      
`-n, --namespaces` - сайлент-режим, позволяет задать ассоциацию неймспейс - таблица. Формат; `namespace1=table1,table2;namespace2=table3,table4`  
`-p, --print` - на основе загруженного проекта выводит ассоциации неймспейс - таблица в формате, подходящем для флага `-n, --namespaces`. Не запускает генератор    
 
### MFD файл

Основной файл проекта, содержит в себе настройки и список неймспейсов. Генерируется с нуля или дополняется.  
```xml
<Project xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema">
    <Name>example.mfd</Name> <!-- имя проекта -->
    <PackageNames> <!-- список неймспейсов -->
        <string>blog</string>
        <string>common</string>
    </PackageNames>
    <Languages> <!-- список языков, см. генераторы xml-lang и template -->
        <string>en</string>
    </Languages> 
    <GoPGVer>8</GoPGVer> <!-- версия go-pg -->
</Project>
```

**PackageNames** - Указанные неймспейсы будут использоваться для дальнейшей генерации. Если неймспейс не указан в списке, даже если файл с неймспейсом присутствует, то он генерироваться не будет  
**Languages** Управление этим полем происходит в генераторе [xml-lang](/generators/xml-lang). В дальнейшем генератор [template](/generators/vt-template) будет использовать этот список, чтобы сгенерировать языковые файлы для интерфейса vt       
**GoPGVer** - Версия go-pg. Поддерживаемые значения 8, 9 и 10. От этого параметра зависят все генераторы golang кода:
  - импорты (`"github.com/go-pg/pg"` vs `"github.com/go-pg/pg/v9"` vs `"github.com/go-pg/pg/v10"`)  
  - аннотации к структурам (`sql:"title"` vs `pg:"title"`)  
  - функции (`pg.F` и `pg.Q` vs `pg.Ident` и `pg.SafeQuery`)  
   
#### Namespace файл и сущности

Файл с неймспейсом, содержит все входящие в него сущности. Сущности будут сгруппированы в файлы по неймспейсам и в дальнейшей генерации
```xml
<Package xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema">
    <Name>blog</Name> <!-- имя неймспейса -->
    <Entities> <!-- список сущностей -->
        <Entity Name="Post" Namespace="blog" Table="posts">
            <Attributes> <!-- список атрибутов -->
                <Attribute Name="ID" DBName="postId" DBType="int4" GoType="int" PK="true" Nullable="Yes" Addable="true" Updatable="true" Min="0" Max="0"></Attribute>
                <Attribute Name="Alias" DBName="alias" DBType="varchar" GoType="string" PK="false" Nullable="No" Addable="true" Updatable="true" Min="0" Max="255"></Attribute>
                <Attribute Name="Title" DBName="title" DBType="varchar" GoType="string" PK="false" Nullable="No" Addable="true" Updatable="true" Min="0" Max="255"></Attribute>
                <Attribute Name="Text" DBName="text" DBType="text" GoType="string" PK="false" Nullable="No" Addable="true" Updatable="true" Min="0" Max="0"></Attribute>
                <Attribute Name="Views" DBName="views" DBType="int4" GoType="int" PK="false" Nullable="No" Addable="true" Updatable="true" Min="0" Max="0"></Attribute>
                <Attribute Name="CreatedAt" DBName="createdAt" DBType="timestamp" GoType="time.Time" PK="false" Nullable="No" Addable="false" Updatable="false" Min="0" Max="0"></Attribute>
                <Attribute Name="UserID" DBName="userId" DBType="int4" GoType="int" PK="false" FK="User" Nullable="No" Addable="true" Updatable="true" Min="0" Max="0"></Attribute>
                <Attribute Name="TagIDs" DBName="tagIds" IsArray="true" DBType="int4" GoType="[]int" PK="false" FK="Tag" Nullable="Yes" Addable="true" Updatable="true" Min="0" Max="0"></Attribute>
                <Attribute Name="StatusID" DBName="statusId" DBType="int4" GoType="int" PK="false" Nullable="No" Addable="true" Updatable="true" Min="0" Max="0"></Attribute>
            </Attributes>
            <Searches> <!-- список поисков -->
                <Search Name="IDs" AttrName="ID" SearchType="SEARCHTYPE_ARRAY"></Search>
                <Search Name="NotID" AttrName="ID" SearchType="SEARCHTYPE_NOT_EQUALS"></Search>
                <Search Name="TitleILike" AttrName="Title" SearchType="SEARCHTYPE_ILIKE"></Search>
                <Search Name="TextILike" AttrName="Text" SearchType="SEARCHTYPE_ILIKE"></Search>
            </Searches>
        </Entity>
        <Entity Name="Tag" Namespace="blog" Table="tags">
            <Attributes> 
                <Attribute Name="ID" DBName="tagId" DBType="int4" GoType="int" PK="true" Nullable="Yes" Addable="true" Updatable="true" Min="0" Max="0"></Attribute>
                <Attribute Name="Alias" DBName="alias" DBType="varchar" GoType="string" PK="false" Nullable="No" Addable="true" Updatable="true" Min="0" Max="255"></Attribute>
                <Attribute Name="Title" DBName="title" DBType="varchar" GoType="string" PK="false" Nullable="No" Addable="true" Updatable="true" Min="0" Max="255"></Attribute>
                <Attribute Name="Weight" DBName="weight" DBType="float8" GoType="*float64" PK="false" Nullable="Yes" Addable="true" Updatable="true" Min="0" Max="0"></Attribute>
                <Attribute Name="StatusID" DBName="statusId" DBType="int4" GoType="int" PK="false" Nullable="No" Addable="true" Updatable="true" Min="0" Max="0"></Attribute>
            </Attributes>
            <Searches>
                <Search Name="IDs" AttrName="ID" SearchType="SEARCHTYPE_ARRAY"></Search>
                <Search Name="NotID" AttrName="ID" SearchType="SEARCHTYPE_NOT_EQUALS"></Search>
                <Search Name="TitleILike" AttrName="Title" SearchType="SEARCHTYPE_ILIKE"></Search>
            </Searches>
        </Entity>
    </Entities>
</Package>
``` 

**Entity** - Описание каждой сущности, содержит в себе имя (Name), неймспейс (Namespace) и соответствующую таблицу в бд (Table). 
В поле Table если не указана схема будет использоваться public  
Атрибут **Name** - содержит имя сущности, соответствует имени таблицы, капитализированное и приведённое к единственному числу.
  
#### Атрибуты 

**Attribute** - Содержит описание поля таблицы в бд  

**Name** - Имя атрибута, сгенерировано из имени поля таблицы как поле структуры в Go. Уникально для сущности  
На это имя будут ссылаться поиски в разделе `<Searches>`, атрибуты `VTEntity` которые генерируются [xml-vt](/generators/xml-vt/README.md)   
Если поле primary key - имя будет изменено на `ID` 
**DBName** - Имя соответствующей колонки в таблице в бд.  
**DBType** - Тип соответствующей колонки в таблице в бд.  
**GoType** - Тип в Go соответствующий типу колонки в таблице. [Соответствие типов](#gotype). По-умолчанию для nullable атрибутов в типе будет присутствовать указатель, который можно убрать, если необходимо  
**DisablePointer** - Отключить указатель в поиске по этому атрибуту.  
**PK** - Флаг Primary ключа, генерируется у primary ключей
**FK** - Ссылка на сущность в проекте, для foreign ключей. Указывается как `EntityName`. Может быть сгенерирован для массивов внешних Id: полей c именем `EntityIDs`, если сущность `Entity` существует. 
**Nullable** - Может ли значение быть nil. `false` ставится для полей `NOT NULL` . Возможные значения `Yes` и `No`  
**Addable** - Можно ли указать значение этого поля, при добавлении сущности в базу (например, ID). [Addable/Updatable](#addable-updatable). Возможные значения `true` и `false`    
**Updatable** - Можно ли указать значение этого поля, при обновлении сущности в базе (например, CreatedAt). [Addable/Updatable](#addable-updatable). Возможные значения `true` и `false`   
**Min** - Минимально возможное значение этого поля для чисел (например Age). Для строк - минимальное количество символов (например Description)  
**Max** - Максимально возможное значение этого поля (например Age). Для строк - максимальное количество символов (например Title) 

#### Поиски
 
**Searches** - содержит в себе список полей для поиска по сущностям.  

**Name** - Имя поиска в структуре поиска Search. Уникально для сущности, включая атрибуты.  
**AttrName** - Ссылка на атрибут сущности. Может быть ссылкой на другую сущность, в формате Entity.Attribute, например `User.ID` или `Category.ShowOnMain`.   
**SearchType** - Тип поиска, влияет на соответствующий тип поиска при построении запросов в БД. Влияет на структуру Search и тип поля [возможные значения](#SEARCH_TYPE)  
**GoType** - Необязательный атрибут, применяется только для поиска по JSON/JSONB полям. Тип значения у JSON-ключа. Возможные значения: `int`, `int64`, `uint`, `uint64`, `float32`, `float64`, `string`, `bool` и слайсы этих типов.  

Если атрибут добавляемый в модель новый (новая колонка в базе, новая таблица, новый проект) - то для этого атрибута будут сгенерированы поиски.   
Для строковых атрибутов кроме поля `Alias` появится `SEARCHTYPE_ILIKE` поиск, добавляя к имени атрибута `ILike`, например `TitleILike`. Для `ID` - поиск по массиву `IDs`. Если присутствует поле Alias, то добавляется поиск NotID для генерирования поиска в vt- модели при проверке уникальности.  

#### SEARCH_TYPE  

Ниже приведены значения для поля SearchType и соответствующие им SQL условия.  
```
SEARCHTYPE_EQUALS             -  f = v
SEARCHTYPE_NOT_EQUALS         -  f != v
SEARCHTYPE_NULL               -  f is null
SEARCHTYPE_NOT_NULL           -  f is not null
SEARCHTYPE_GE                 -  f >= v
SEARCHTYPE_LE                 -  f <= v
SEARCHTYPE_G                  -  f > v
SEARCHTYPE_L                  -  f < v
SEARCHTYPE_LEFT_LIKE          -  f like '%v'
SEARCHTYPE_LEFT_ILIKE         -  f ilike '%v'
SEARCHTYPE_RIGHT_LIKE         -  f like 'v%'
SEARCHTYPE_RIGHT_ILIKE        -  f ilike 'v%'
SEARCHTYPE_LIKE               -  f like '%v%'
SEARCHTYPE_ILIKE              -  f ilike '%v%'
SEARCHTYPE_ARRAY              -  f in (v, v1, v2)
SEARCHTYPE_NOT_INARRAY        -  f not in (v1, v2)
SEARCHTYPE_ARRAY_CONTAINS     -  v = any (f)
SEARCHTYPE_ARRAY_NOT_CONTAINS -  v != all (f)
SEARCHTYPE_ARRAY_CONTAINED    -  ARRAY[v] <@ f
SEARCHTYPE_ARRAY_INTERSECT    -  ARRAY[v] && f
SEARCHTYPE_JSONB_PATH         -  f @> v
``` 
f - имя поля, v - значение
 
#### GoType

```
integer, serial            -> int
bigint                     -> int64
real                       -> floaf32
double, numeric            -> float64
text, varchar, uuid, point -> string
boolean                    -> bool
timestamp, date, time      -> time.Time
interval                   -> time.Duration
hstore                     -> map[string]string
inet                       -> net.IP
cidr                       -> net.IPNet
```

`json` и `jsonb` сгенерируют тип с названием имя сущности + имя поля. Например `UserParams`  
Если поле массив - к типу будет добавлены `[]`. hstore и json(b) не могут быть массивами.  
Если поле `Nullable=yes` будет добавлен `*` перед типом

Неизвестные типы генерируют interface{}

#### Addable/Updatable

Поля с именами `createdAt` и `modifiedAt` генерируют флаги `Addable` и `Updatable` со значением `false`

### Особенности проверки консистентности

При загрузке существующего проекта будет проведена проверка консистентности (это справедливо для всех генераторов).
Проверка включает в себя:
- каждый поиск в секции `<Searches>` ссылается на существующие в xml сущность и атрибут.  
- каждый FK атрибут ссылается на существующие в xml сущность и атрибут. 

В случае если проверки не пройдены - проект не загрузится с ошибкой.   
 
### Особенности работы с существующими сущностями

При повторной генерации генератор пытается сохранить пользовательские изменения:
- Если сущность для генерируемой таблицы уже существует, то она будет дополнена новыми атрибутами
- Новые атрибуты определяются по паре `DBName` и `DBType`. Это значит что поля, у которых имя и тип уже присутствуют в xml, добавляться не будут.
  - Если поменять тип колонки в таблице, то она будет добавлена в сущность как новая
- Если атрибут уже существует в xml, для него не будут сгенерированы новые поиски, даже если должны.
  - Если удалить поиск из секции `<Searches>` от при повторной генерации он не будет добавлен.

### Особенности работы с JSON полями

Поиск по json полям не поддерживает ссылки на json-поля у других сущностей.   
При работе с JSON/JSONB полями поддерживаются следующие типы поиска:
```
SEARCHTYPE_EQUALS             -  f->>'k' in (v)
SEARCHTYPE_NOT_EQUALS         -  f->>'k' not in (v)
SEARCHTYPE_NULL               -  f->>'k' is null
SEARCHTYPE_NOT_NULL           -  f->>'k' is not null
SEARCHTYPE_ARRAY              -  f->>'k' in (v, v1, v2)
SEARCHTYPE_NOT_INARRAY        -  f->>'k' not in (v1, v2)
SEARCHTYPE_ARRAY_CONTAINS     -  f @> '{"k": [v]}'
SEARCHTYPE_ARRAY_NOT_CONTAINS -  not f @> '{"k": [v]}'
SEARCHTYPE_JSONB_PATH         -  f @> v
```
f - имя json-поля, k - имя ключа в json-поле, v - значение

Для указания типа значения json поля необходимо заполнить атрибут `GoType` в XML-определении поиска. Если его не указывать, то будет использован тип `interface{}`

Поиск типа `SEARCHTYPE_ARRAY_CONTAINS` и `SEARCHTYPE_ARRAY_NOT_CONTAINS` имеет следующие ограничения:
- Поддерживается только JSONB-полями. С JSON-полями такой поиск не работает.
- Ищет только по одному значению в массиве.
- Для быстрой работы рекомендуется повесить GIN индекс с опцией `jsonb_path_ops`.

### Примеры поиска в связанной сущности и в json полях 
```xml
<!-- поиск по значению IsMain из связанной сущности Rubric -->
<Search Name="IsMain" AttrName="Rubric.IsMain" SearchType="SEARCHTYPE_EQUALS"></Search>

<!-- поиск значения в ключе smsCount из json поля Params, поле - int -->
<Search Name="SmsCount" AttrName="Params->smsCount" SearchType="SEARCHTYPE_EQUALS" GoType="int"></Search>

<!-- поиск по ключу addressHome из json поля Params, поле - string -->
<Search Name="NotAddressHome" AttrName="Params->addressHome" SearchType="SEARCHTYPE_NOT_EQUALS" GoType="string"></Search>

<!-- поиск по ключу isPasswordSent из json поля Params, поле - bool -->
<Search Name="IsPasswordSent" AttrName="Params->isPasswordSent" SearchType="SEARCHTYPE_EQUALS" GoType="bool"></Search>

<!-- поиск по наличию ключа token / отсутствию в ключе значения null в json поле Params, поле - string -->
<Search Name="TokenNotExists" AttrName="Params->token" SearchType="SEARCHTYPE_NULL" GoType="string"></Search>

<!-- поиск по вложенному ключу parent->subValue в json поле Params, поле - int -->
<Search Name="YandexSubValue" AttrName="Params->parent->subValue" SearchType="SEARCHTYPE_EQUALS" GoType="int"></Search>

<!-- поиск по наличию значения в массиве у ключа favoriteProducts в json поле Params, поле - int -->
<Search Name="FavoriteProduct" AttrName="Params->favoriteProducts" SearchType="SEARCHTYPE_ARRAY_CONTAINS" GoType="int"></Search>

<!-- поиск по отсутствию значения в массиве у ключа favoriteProducts в json поле Params, поле - int -->
<Search Name="NotFavoriteProduct" AttrName="Params->favoriteProducts" SearchType="SEARCHTYPE_ARRAY_NOT_CONTAINS" GoType="int"></Search>

<!-- поиск по нескольким значениям ключа smsCount в json поле Params, поле - []int -->
<Search Name="SmsCounts" AttrName="Params->smsCount" SearchType="SEARCHTYPE_ARRAY" GoType="[]int"></Search>

<!-- поиск по нескольким значениям ключа addressHome в json поле Params, поле - []string -->
<Search Name="AddressHomes" AttrName="Params->addressHome" SearchType="SEARCHTYPE_ARRAY" GoType="[]string"></Search>

<!-- поиск по пути в json поле Params -->
<Search Name="ParamsPath" AttrName="Params" SearchType="SEARCHTYPE_JSONB_PATH"></Search>
```
