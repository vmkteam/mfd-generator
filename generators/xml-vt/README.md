## XML-VT

xml-vt генератор неймспейсов и сущностей в них для vt-часть проекта. В качестве источника данных используются xml файлы неймспейсов. Результат - несколько xml файлов.

### Использование

Генератор считывает информацию из mfd файла о неймспейсах, загружает каждый их них. Генерирует xml учитывая значение флага `-n --namesapces`    
Для каждой сущности сгенерируется vt-сущность. Для каждого атрибута и некоторых поисков сгенерируются vt-атрибуты и описание шаблона админки (vt-template).
VT-сущности сгруппированы в vt-неймспейсы так же как и исходные. VT-неймспейсы сохраняются в файлы с именем vt-неймспейса + '.vt.xml' 

### CLI
```
Create vt xml from mfd

Usage:
  mfd xml-vt [flags]

Flags:
  -m, --mfd string           mfd file
  -n, --namespaces strings   namespaces
  -h, --help                 help for xml-vt
```

`-n, --namespaces` - генерировать только из перечисленных неймспейсов. Через запятую

#### VT-namespace файл и vt-сущности

```xml
<VTNamespace xmlns:xsi="" xmlns:xsd="">
    <Name>blog</Name> <!-- имя vt-неймспейса -->
    <VTEntities> <!-- массив vt-сущностей -->
        <Entity Name="Post" Mode="Full">
            <TerminalPath>posts</TerminalPath>
            <Attributes> <!-- атрибуты vt-сущности -->
                <Attribute Name="ID" AttrName="ID" SearchName="ID" Summary="true" Search="true" Max="0" Min="0" Required="false" Validate=""></Attribute>
                <Attribute Name="Alias" AttrName="Alias" SearchName="Alias" Summary="true" Search="true" Max="255" Min="0" Required="true" Validate="alias"></Attribute>
                <Attribute Name="Title" AttrName="Title" SearchName="TitleILike" Summary="true" Search="true" Max="255" Min="0" Required="true" Validate=""></Attribute>
                <Attribute Name="Text" AttrName="Text" SearchName="TextILike" Summary="true" Search="true" Max="0" Min="0" Required="true" Validate=""></Attribute>
                <Attribute Name="Views" AttrName="Views" SearchName="Views" Summary="true" Search="true" Max="0" Min="0" Required="true" Validate=""></Attribute>
                <Attribute Name="CreatedAt" AttrName="CreatedAt" SearchName="CreatedAt" Summary="true" Search="true" Max="0" Min="0" Required="false" Validate=""></Attribute>
                <Attribute Name="UserID" AttrName="UserID" SearchName="UserID" Summary="true" Search="true" Max="0" Min="0" Required="true" Validate=""></Attribute>
                <Attribute Name="TagIDs" AttrName="TagIDs" SearchName="TagIDs" Summary="false" Search="false" Max="0" Min="0" Required="false" Validate=""></Attribute>
                <Attribute Name="StatusID" AttrName="StatusID" SearchName="StatusID" Summary="true" Search="true" Max="0" Min="0" Required="true" Validate="status"></Attribute>
                <Attribute Name="IDs" SearchName="IDs" Summary="false" Search="true" Max="0" Min="0" Required="false" Validate=""></Attribute>
                <Attribute Name="NotID" SearchName="NotID" Summary="false" Search="true" Max="0" Min="0" Required="false" Validate=""></Attribute>
            </Attributes>
            <Template> <!-- параметры шаблона -->
                <Attribute Name="Alias" VTAttrName="Alias" List="true" Form="HTML_INPUT" Search="HTML_INPUT"></Attribute>
                <Attribute Name="Title" VTAttrName="Title" List="true" Form="HTML_INPUT" Search="HTML_INPUT"></Attribute>
                <Attribute Name="Text" VTAttrName="Text" List="true" Form="HTML_TEXT" Search="HTML_TEXT"></Attribute>
                <Attribute Name="Views" VTAttrName="Views" List="true" Form="HTML_INPUT" Search="HTML_INPUT"></Attribute>
                <Attribute Name="CreatedAt" VTAttrName="CreatedAt" List="false" Form="HTML_NONE" Search="HTML_DATETIME"></Attribute>
                <Attribute Name="UserID" VTAttrName="UserID" List="false" FKOpts="id" Form="HTML_INPUT" Search="HTML_INPUT"></Attribute>
                <Attribute Name="User" VTAttrName="UserID" List="true" FKOpts="id" Form="" Search="HTML_NONE"></Attribute>
                <Attribute Name="TagIDs" VTAttrName="TagIDs" List="false" FKOpts="alias" Form="HTML_SELECT" Search="HTML_NONE"></Attribute>
                <Attribute Name="StatusID" VTAttrName="StatusID" List="true" Form="HTML_INPUT" Search="HTML_INPUT"></Attribute>
                <Attribute Name="IDs" VTAttrName="IDs" List="false" Form="HTML_NONE" Search="HTML_SELECT"></Attribute>
                <Attribute Name="NotID" VTAttrName="NotID" List="false" Form="HTML_NONE" Search="HTML_INPUT"></Attribute>
            </Template>
        </Entity>
        <Entity Name="Tag" Mode="Full">
            <TerminalPath>tags</TerminalPath>
            <Attributes>
                <Attribute Name="ID" AttrName="ID" SearchName="ID" Summary="true" Search="true" Max="0" Min="0" Required="false" Validate=""></Attribute>
                <Attribute Name="Alias" AttrName="Alias" SearchName="Alias" Summary="true" Search="true" Max="255" Min="0" Required="true" Validate="alias"></Attribute>
                <Attribute Name="Title" AttrName="Title" SearchName="TitleILike" Summary="true" Search="true" Max="255" Min="0" Required="true" Validate=""></Attribute>
                <Attribute Name="Weight" AttrName="Weight" SearchName="Weight" Summary="true" Search="true" Max="0" Min="0" Required="false" Validate=""></Attribute>
                <Attribute Name="StatusID" AttrName="StatusID" SearchName="StatusID" Summary="true" Search="true" Max="0" Min="0" Required="true" Validate="status"></Attribute>
                <Attribute Name="IDs" SearchName="IDs" Summary="false" Search="true" Max="0" Min="0" Required="false" Validate=""></Attribute>
                <Attribute Name="NotID" SearchName="NotID" Summary="false" Search="true" Max="0" Min="0" Required="false" Validate=""></Attribute>
            </Attributes>
            <Template>
                <Attribute Name="Alias" VTAttrName="Alias" List="true" Form="HTML_INPUT" Search="HTML_INPUT"></Attribute>
                <Attribute Name="Title" VTAttrName="Title" List="true" Form="HTML_INPUT" Search="HTML_INPUT"></Attribute>
                <Attribute Name="Weight" VTAttrName="Weight" List="true" Form="HTML_INPUT" Search="HTML_INPUT"></Attribute>
                <Attribute Name="StatusID" VTAttrName="StatusID" List="true" Form="HTML_INPUT" Search="HTML_INPUT"></Attribute>
                <Attribute Name="IDs" VTAttrName="IDs" List="false" Form="HTML_NONE" Search="HTML_SELECT"></Attribute>
                <Attribute Name="NotID" VTAttrName="NotID" List="false" Form="HTML_NONE" Search="HTML_INPUT"></Attribute>
            </Template>
        </Entity>
    </VTEntities>
</VTNamespace>
``` 

Для каждой сущности в выбранных неймспейсах будет сгенерирован код vt-сущности.
**Entity** - Описание каждой vt-сущности   
Атрибут **Name** - Генерируется из имени vt сущности. Так же - ссылка на сущности по имени.   
**Mode** - Режим генерирования vt-сущности. [Возможные значения](#modes), значение по-умолчанию "Full". Конвертируется в "ReadOnly" из устаревшего параметра "WithoutTemplates"    
**TerminalPath** - Путь, по которому vt-сущность будет доступна из vt-интерфейса. Генерируется из имени vt-сущности с заменой символов на "-". 

#### Атрибуты 

Для каждого атрибута сущности будет сгенерирован свой атрибут vt-сущности **Attribute** 

**Name** - Имя vt-атрибута, при генерировании значения используется **Name** атрибута.  
На это имя будут ссылаться поля шаблона в разделе `<Template>`   
**AttrName** - Ссылка на атрибут сущности из которого был сгенерирован этот vt-атрибут. Такой атрибут должен существовать в сущности.   
**SearchName** - Ссылка на атрибут или поиск сущности который будет использоваться при поиске по ним.  
 * Для строк будет сгенерировано значение `ILike` поиск из секции `Searches`, если такое есть.  
 * Для всех остальных имя как в поле AttrName  
 
**Summary** - Определяет, будет ли поле присутствовать в структуре `EntitySummary`. Генерируется исходя из [имени и типа](#summarysearch) исходной сущности.    
**Search** -  Определяет, будет ли поле присутствовать в структуре `EntitySearch`. Генерируется исходя из [имени и типа](#summarysearch) исходной сущности.  
**Min** - Минимально возможное значение этого поля для чисел (например Age). Для строк - минимальное количество символов (например Description). Генерируется из значения **Min** в соответствующем атрибуте (указанном в `AttrNme`)   
**Max** - Максимально возможное значение этого поля (например Age). Для строк - максимальное количество символов (например Title). Генерируется из значения **Max** в соответствующем атрибуте (указанном в `AttrNme`)   
**Required** - Флаг обязательного значения. Генерируется значение `true` если Nullable=No в соответствующем атрибуте сущности. Возможные значения `true` и `false`  
**Validate** - Специальные опции валидации. [Возможные значения](#validate)   

#### Шаблон

Для каждого vt-атрибута генерируется шаблона. Шаблон описывает ui полей.

**Name** - Имя поля в шаблоне, используется для генерирования имён полей в шаблонах [vt](/generators/model). Генерируется из `Name` соответствующего vt-атрибута  
**VTAttrName** - ссылка на соответствующий vt-атрибут, используется значение поля `Name`  
**List** - Определяет, выводить ли поле в таблице в списке vt-сущностей. Генерируется исходя из [типа и имени](#listformsearch). Возможные значения `true` и `false`  
**Form** - Определяет тип инпута для формы создания/редактирования vt-сущности. Генерируется исходя из [типа и имени](#listformsearch). [Возможные значения](#html)  
**Search** - Определяет тип инпута для формы поиска vt-сущности. Генерируется исходя из [типа и имени](#listformsearch). [Возможные значения](#html)  

#### MODES

- None - пропустить генерирование сущности
- ReadOnly - генерируются только модели и репозитории
- ReadOnlyWithTemplates - генерируются, модели и репозитории и сервисы с методами на чтение. 
- Full - генерируется всё

#### Summary/Search

false устанавливается для полей
* Name=Password
* Для массивов, json(b) полей и hstore

#### Validate

- alias - автоматически генерируется для поля с именем `Alias`. Добавляет в генерируемый код проверку на уникальность поля
- status - автоматически генерируется для поля с именем `StatusID`. Добавляет в генерируемый код проверку на возможные значение статусов
- ip - автоматически генерируется для поля с именем `IP`.
- email - автоматически генерируется для поля с именем `Email`, `Mail`.

#### List/Form/Search
- List устанавливается в `false` для vt-атрибутов с именем Password, Primary ключей.
- Form/Search генерируются
  - HTML_IMAGE для полей c FK ключём `VfsFile` и именем содержащим `Image`
  - HTML_FILE для полей и FK ключём `VfsFile`
  - HTML_SELECT для vt-атрибута с FK ключём
  - HTML_EDITOR для полей с типом в BD `text` и именем `Description` или `Content` 
  - HTML_TEXT для полей с типом в BD `text`
  - HTML_TEXT - с типом `varchar` более 256 символов
  - HTML_INPUT - с типом `varchar`
  - HTML_PASSWORD - c именем `Password` 
  - HTML_CHECKBOX для полей с типом `boolean` и vt-атрибута с именем `StatusID`
  - HTML_DATE для полей с типом `date`
  - HTML_TIME для полей с типом `time`
  - HTML_DATETIME для полей с типом `timestamp`
  - HTML_NONE для `массивов`, `hstore` и `json(b)` полей

#### HTML
```
HTML_NONE     - Не отображается
HTML_INPUT    - `v-text-field`
HTML_TEXT     - `v-textarea`
HTML_PASSWORD - `v-text-field`
HTML_EDITOR   - `vt-tinymce-editor`
HTML_CHECKBOX - `v-checkbox`
HTML_DATETIME - `vt-datetime-picker`
HTML_DATE     - `vt-time-picker`
HTML_TIME     - `vt-date-picker`
HTML_FILE     - `vt-vfs-file-input`
HTML_IMAGE    - `vt-vfs-image-input`
HTML_SELECT   - генерирует select box.
```

### Особенности работы с существующими сущностями

При повторной генерации генератор пытается сохранить пользовательские изменения
- Если vt-сущность для генерируемой таблицы уже существует, то она будет дополнена новыми vt-атрибутами и шаблоном
- Добавляются только новые vt-атрибуты. Новые vt-атрибуты определяются по паре `AttrName` и `SearchName`
  - Если поменять Name у соответствующего атрибута, то будет добавлен новый vt-атрибут для него.
- Если vt-атрибут уже существует в xml, для него не будут сгенерированы новые шаблоны, даже если должны.
  - Если удалить шаблон из секции `<Templates>` от при повторной генерации он не будет добавлен.    

