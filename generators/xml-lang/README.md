## XML-LANG

xml-lang генератор xml файлов для хранения переводов vt-сущностей и vt-шаблонов. В качестве источника данных используются xml файлы vt-неймспейсов. Результат - несколько xml файлов с именем языка

### Использование

Генератор считывает информацию из mfd файла о vt-неймспейсах, загружает каждый их них. Для каждой vt-entity будет сгенерирован раздел с заголовками и хлебными крошками. Для каждого vt-шаблона (раздел `<Templates>`) будет сгенерирован перевод. Также будут добавлены некоторые служебные переводы. 

### CLI
```
Create lang xml from mfd

Usage:
  mfd xml-lang [flags]

Flags:
  -m, --mfd string           mfd file path
  -l, --langs strings        languages to generate, use two letters code, eg. ru,en,de. separate by comma
  -h, --help                 help for xml-lang
  -n, --namespaces strings   namespaces to generate, must be in mfd file. separate by comma
  -e, --entities strings     entities to generate, must be in vt.xml file. separate by comma
```

`-l, --langs` - генерировать только из перечисленных языков. Через запятую. Указанные языки будут добавлены в mfd файл

`-n, --namespaces` - генерировать только из перечесиленных vt-неймспейсов. Через запятую.

`-e, --entities` - генерировать только из перечисленных сущностей. Можно использовать с флагом (`-n`)

#### lang файл

```xml
<Translation xmlns:xsi="" xmlns:xsd="">
    <Language>en</Language> <!-- код языка -->
    <Namespaces>
        <Namespace Name="blog"> <!-- vt-неймспейс -->
            <Entities>
                <Entity Name="Post" Key="post"> <!-- vt-сущность -->
                    <Crumbs> <!-- хлебные крошки -->
                        <postList>Posts</postList>
                        <postAdd>Add</postAdd>
                        <postEdit>Edit</postEdit>
                    </Crumbs>
                    <Form> <!-- элементы формы -->
                        <statusIdLabel>Status Id</statusIdLabel>
                        <tagIdsLabel>Tags</tagIdsLabel>
                        <aliasLabel>Alias</aliasLabel>
                        <titleLabel>Title</titleLabel>
                        <textLabel>Text</textLabel>
                        <viewsLabel>Views</viewsLabel>
                        <userIdLabel>User</userIdLabel>
                    </Form>
                    <List> <!-- элементы списка -->
                        <Title>Posts</Title> <!-- заголовок -->
                        <Filter> <!-- элементы фильтров -->
                            <createdAt>Created at</createdAt>
                            <title>Title</title>
                            <userId>User</userId>
                            <notId>Not</notId>
                            <ids>Ids</ids>
                            <quickFilterPlaceholder></quickFilterPlaceholder>
                            <alias>Alias</alias>
                            <text>Text</text>
                            <views>Views</views>
                            <statusId>Status</statusId>
                        </Filter>
                        <Headers> <!-- заголовки таблицы -->
                            <actions>Actions</actions>
                            <alias>Alias</alias>
                            <title>Title</title>
                            <text>Text</text>
                            <views>Views</views>
                            <user>User</user>
                            <status>Status</status>
                        </Headers>
                    </List>
                </Entity>
            </Entities>
        </Namespace>
    </Namespaces>
</Translation>
``` 

Для каждой vt-сущности и шаблона будет сгенерирован перевод или набор переводов. В качестве источника данных используется секция `<Templates>` соответствующей vt-сущности  
**Entity** - Описание переводов vt-сущности   
Атрибут **Name** - Генерируется из имени vt-сущности. Так же - ссылка на сущности по имени.  
Атрибут **Key** - Генерируется из имени vt-сущности приведением к нижнему регистру. Используется для генерирования json-файлов переводов    

#### Crumbs

Для каждой vt-сущности генерируется 3 заголовка для хлебных крошек (в имя ключа добавляется имя vt-сущности как указано в поле **Key**)
- List - для списка
- Add - для формы добавления
- Edit - для формы редактирования

#### Form/List

Для каждой vt-сущности будет сгенерирован перевод листа и формы.
Для каждого vt-шаблона будет сгенерирован лейбл для инпута в форме, лейбл для инпута в фильтрах, заголовок колонки в таблице 

#### Особенности генерирования переводов

При запуске генератора выбранные языка будут добавлены в mfd файл, если язык ещё там не указан. При дальнейшей генерации будет использовать список в mfd файле, если lang файл присутствует, но в mfd файле он не указан - для него не будут сгенерированы json переводы при генерировании [vt-template](/generators/vt-template)   
Для русского и английского языка существуют предустановленные переводы. Перевод подбирается по имени vt-шаблона. Если перевод не удалось подобрать (неустановленное имя колонки или другой язык) - то в xml генерируется пустой элемент для vt-шаблона.  
Все существующие переводы будут сохранены. То есть, возможно использовать lang файлы как хранилище любых переводов, не только сгенерированных.  
