## VT-TEMPLATE

vt-template - генератор клиентской части vt. В качестве источника данных используется mfd файл. На выходе - несколько js, json, vue файлов

### Использование

Генератор считывает информацию из mfd файла о vt-неймспейсах, загружает каждый их них.     
Файлы записываются в папку указанную в параметре `-o --output`  
Для каждой vt-сущности будет сгенерирован набор файлов, включающий в себя:  
- routes.ts - настройка роутов интерфейса  
- List.vue - шаблон списка, используются записи из `VTEntities->Entity->Template` атрибут List=true
- Form.vue - шаблон создания/редактирования, используются записи из `VTEntities->Entity->Template` атрибут Form определяет внешний вид контрола
- ListFilters.vue - шаблон фильтров, используются записи из `VTEntities->Entity->Template` атрибут Search определяет внешний вид контрола
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
  -e, --entities strings     entities to generate, must be in vt.xml file. separate by comma
  -h, --help                 help for template

```

#### Шаблоны

Встроенные шаблоны существуют в двух вариантах:
- **default** (по умолчанию) — class-based компоненты на `vue-property-decorator` + `mobx-vue`.
- **Composition API** — генерирует компоненты через `defineComponent`/`setup` и подключает composables (`useEntityList`, `useEntityForm`, `useI18n`), которые должны существовать во фронтенд-проекте (`@/composables/...`); в `routes.ts` добавляется поле `meta.title`.

Выбор варианта задаётся на уровне проекта в mfd-файле . В корневом `<Project>` добавьте элемент `<VTComposition>true</VTComposition>` — при его наличии используются Composition API шаблоны. Если элемент отсутствует или `false`, используются default class-based шаблоны.

#### MODE

Значение Mode vt-сущности в vt.xml определяет какие файлы будут сгенерированы.
- "Full" - все файлы
- "ReadOnlyWithTemplates" - все файлы в read-only режиме, Form.vue генерироваться не будет
- "None" - файлы генерироваться не будут

#### Частичная генерация

Флаги `-n --namespaces` и `-e --entities` ограничивают набор сущностей, для которых генерируются файлы. Если ни один из них не указан, обрабатываются все неймспейсы и сущности.

При частичной генерации `routes.ts` обновляется точечно: перезаписываются только блоки указанных сущностей (по маркеру `/* EntityName */`), остальная часть файла сохраняется. Если целевой сущности в файле ещё нет — её блок дописывается в конец. При полной генерации (без `-n`/`-e`) `routes.ts` перезаписывается целиком.

#### Особенности работы с существующими моделями

Файлы сущностей (`List.vue`, `Form.vue`, `MultiListFilters.vue`, переводы) перезаписываются при каждой генерации. Исключение — `routes.ts` при частичной генерации (см. выше).
