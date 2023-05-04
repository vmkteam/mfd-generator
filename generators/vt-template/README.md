## VT-TEMPLATE

vt-template - генератор клиетнской части vt. В качестве источника данных используется mfd файл. На выходе - несколько js, json, файлов

### Использование

Генератор считывает информацию из mfd файла о vt-неймспейсах, загружает каждый их них.     
Файлы записываются в папку указанную в параметре `-o --output`  
Для каждой vt-сущности будет сгенерирован набор файлов, включающий в себя:  
- routes.ts - настройка роутов интерфейса  
- List.vue - шаблон списка, используются записи из `VTEntities->Entity->Template` атрибут List=true
- Form.vue - шаблон создания/редактирования, используются записи из `VTEntities->Entity->Template` атрибут Form определяет внешний вид контрола
- ListFilters.vue - шаблон фильтров, используются записи из `VTEntities->Entity->Template` атрибут Search определяет внешний вид контролаx
- .json файлы переводов. В качестве источника - lang.xml файлы. Будут сгенерированы только те переводы, что указаны в mfd файле
                                                                      
### CLI

```
Create vt template from xml

Usage:
  mfd template [flags]

Flags:
  -o, --output string        output dir path
  -m, --mfd string           mfd file path
  -n, --namespaces strings   namespaces to generate. separate by comma
  -h, --help                 help for template

```

#### MODE

Значение Mode vt-сущности в vt.xml определяет какие файлы будут сгенерировны.
- "Full" - все файлы
- "ReadOnlyWithTemplates" - все файлы в read-only режиме, Form.vue генерироваться не будет
- "ReadOnly" - только модели model.go
- "None" -  файлы генерироваться не будут

#### Особенности работы с существующими моделями

Все файлы будут перезаписаны при каждой генерации.
