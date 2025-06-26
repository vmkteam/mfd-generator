# MFD Generator

[![Release](https://img.shields.io/github/release/vmkteam/mfd-generator.svg)](https://github.com/vmkteam/mfd-generator/releases/latest)
[![Build Status](https://github.com/vmkteam/mfd-generator/actions/workflows/go.yml/badge.svg?branch=master)](https://github.com/vmkteam/mfd-generator/actions)
[![Linter Status](https://github.com/vmkteam/mfd-generator/actions/workflows/golangci-lint.yml/badge.svg?branch=master)](https://github.com/vmkteam/mfd-generator/actions)

**mfd generator** призван облегчить работу с базой данных путем генерирования моделей, поисков и валидаторов, а также сопутствующих сущностей вплоть до интерфейса админки.
Проект включает в себя несколько генераторов, каждый из которых генерирует xml, go, js разных уровней.

Для редактирования xml файлов, новых сущностей и кода [доступен UI](https://github.com/vmkteam/mfd-ui): `mfd-generator server`.

**Первая группа:**  
[xml](/generators/xml) - генератор основы проекта: mfd файла, неймспейсов и сущностей в них.  
[xml-vt](/generators/xml-vt) - генератор неймспейсов и сущностей в них для vt-часть проекта.   
[xml-lang](/generators/xml-lang) - генератор языковых xml файлов.  

#### Кастомизация переводов через `<dict>`

Для гибкой настройки и переопределения стандартных переводов в `.mfd` файле можно использовать секцию `<Dictionary>`. Это позволяет изменять или добавлять переводы для ключей, не внося изменений в код генератора.
Генератор считывает этот блок и обновляет им глобальную карту переводов.
Пример использования в `.mfd` файле:
```xml
<Project xmlns:xsi="" xmlns:xsd="">
    <Dictionary>
        <user>Пользователь (кастомный)</user>
        <myCustomButton>Моя новая кнопка</myCustomButton>
    </Dictionary>
</Project>
```
В сгенерированном коде вызов `Translate(RuLang, "user")` вернет "Пользователь (кастомный)".

**Вторая группа:**  
[model](/generators/model) - генератор golang модели для взаимодействия с базой данных. В качестве источника данных используется результат xml генератора.  
[repo](/generators/repo) - генератор golang репозиториев для манипуляций с данными в базе с помощью моделей.  

**Третья группа:**  
[vt](/generators/vt) - генератор golang файлов для создания vt-сервиса, серверной части интерфейса vt.  
[template](/generators/vt-template) - генератор js шаблонов, которые используются для создания интерфейса vt.  

Результат работы генераторов может зависеть друг от друга, часть генераторов работает на основе результатов других генераторов. Далее приведена справка по каждому из генераторов с разбором их работы.  

Описание форматов xml файлов можно найти в соответствующих генераторах [xml](/generators/xml), [xml-vt](/generators/xml-vt) и [xml-lang](/generators/xml-lang)

# Command line usage
```
Usage:
  mfd-generator [flags]
  mfd-generator [command]

Available Commands:
  help        Help about any command
  model       Create golang model from xml
  repo        Create repo from xml
  server      Run web server with generators
  template    Create vt template from xml
  version     Print mfd-generator version
  vt          Create vt from xml
  xml         Create or update project base with namespaces and entities
  xml-lang    Create lang xml from mfd
  xml-vt      Create vt xml from mfd

Flags:
  -h, --help   help for mfd

Use "mfd-generator [command] --help" for more information about a command.
```
