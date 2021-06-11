# MDF Generator

**mfd generator** призван облегчить работу с базой данных путем генерирования моделей, поисков и валидаторов, а также сопутствующих сущностей вплоть до интерфейса админки.
Проект включает в себя несколько генераторов, каждый из которых генерирует xml, go, js разных уровней.

Для редактирования xml файлов, новых сущностей и кода [доступен UI](https://github.com/vmkteam/mfd-ui): `mfd-generator server`.

**Первая группа:**  
[xml](/generators/xml) - генератор основы проекта: mfd файла, неймспейсов и сущностей в них.  
[xml-vt](/generators/xml-vt) - генератор неймспейсов и сущностей в них для vt-часть проекта.   
[xml-lang](/generators/xml-lang) - генератор языковых xml файлов.  

**Вторая группа:**  
[model](/generators/model) - генератор golang модели для взаимодействия с базой данных. В качестве источника данных используется результат xml генератора.  
[repo](/generators/repo) - генератор golang репозиториев для манипуляций с данными в базе с помощью моделей.  

**Третья группа:**  
[vt](/generators/vt) - генератор golang файлов для создания vt-сервиса, серверной части vt интерфейса.  
[template](/generators/vt-template) - генератор js шаблонов, которые используются для создания интерфйса vt.  

Результат работы генераторов может зависеть друг от друга, часть генераторов работает на основе результатов других генераторов. Далее приведена справка по каждому из генераторов с разбором их работы.  

Описание форматов xml файлов можно найти в соответствующих генераторах [xml](/generators/xml), [xml-vt](/generators/xml-vt) и [xml-lang](/generators/xml-lang)

### CHANGELOG

#### v0.0.1  
**xml**
- `-o --output` renamed to `-m --mfd` 
- `-p --pkgs`  renamed to `-n --namespaces`  
- added `-p --print` flag to print value for `namespaces` flag based on current project. flag was moved from `model` generator

**xml-vt**
- generator command renamed from `xml vt` to `xml-vt`
- `-p --ns`  renamed to `-n --namespaces`    

**xml-lang**
- generator command renamed from `xml lang` to `xml-lang`

**model**
- `-n --ns` flag renamed to `-p --print` and moved to **xml** generator
- `-p --package` flag now not required. if flag not set - last element of output path will be used

#### v0.0.2 

**repo**
- `-n --ns`  renamed to `-n --namespaces` 
- `-p --package` flag now not required. if flag not set - last element of output path will be used

**vt**
- `-n --ns`  renamed to `-n --namespaces`
- `-x --model-pkg`  renamed to `-x --model`

**vt-template** 
- `-n --ns`  renamed to `-n --namespaces`
- disabled custom templates flags
