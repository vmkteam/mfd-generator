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

Встроенные шаблоны существуют в трёх вариантах:
- **`vue2`** (по умолчанию) — class-based компоненты на `vue-property-decorator` + `mobx-vue`, Vuetify 2.
- **`composition`** — Vue 2 + Composition API: компоненты через `defineComponent`/`setup` и composables (`useEntityList`, `useEntityForm`, `useI18n`), которые должны существовать во фронтенд-проекте (`@/composables/...`); в `routes.ts` добавляется поле `meta.title`. Vuetify 2.
- **`vue3`** — Vue 3 + Vuetify 3: те же composables плюс `useErrorMessage`, компоненты обёрток `vt-row`/`vt-col`/`vt-btn`/`vt-icon`/`vt-card`/`vt-hover`/`vt-data-table`, `$vuetify.display` вместо `$vuetify.breakpoint`, `v-tabs-window`/`v-tabs-window-item` вместо `v-tabs-items`/`v-tab-item`, mdi-иконки. `routes.ts` и `MultiListFilters.vue` совпадают с вариантом `composition`.

Выбор варианта задаётся на уровне проекта в mfd-файле. В корневом `<Project>` добавьте элемент `<VTTemplate>vue3</VTTemplate>`. Допустимые значения — `vue2`, `composition`, `vue3`; если элемент отсутствует, пуст или содержит неизвестное значение, используется `vue2`.

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
