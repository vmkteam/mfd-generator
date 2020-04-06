## XML-VT

xml-vt генератор неймспейсов и сущностей в них для vt-часть проекта. В качестве источника данных используются xml файлы неймспейсов. Результат - несколько xml файлов.

### Использование

Генератор считывает информацию из mfd файла о неймспейсах, зашружает каждый их них. Генерирует xml учитывая значение флага `-n --namesapces`    
Для кажой сущности сгенерируется vt-сущность. Для каждого атрибута и некоторых поисков сгенерируются vt-атрибуты и описание шаблона админки (vt-template).
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
                <Attribute Name="TagID" AttrName="TagID" SearchName="TagID" Summary="false" Search="false" Max="0" Min="0" Required="false" Validate=""></Attribute>
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