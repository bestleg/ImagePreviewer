#"Превьювер изображений"

## Общее описание
Сервис предназначен для изготовления preview (создания изображения
с новыми размерами на основе имеющегося изображения).

## Архитектура
Сервис представляет собой web-сервер (прокси), загружающий изображения,
масштабирующий/обрезающий их до нужного формата и возвращающий пользователю.

## Основной обработчик
http://cut-service.com/fill/300/200/www.audubon.org/sites/default/files/a1_1902_16_barred-owl_sandra_rothenberg_kk.jpg

<---- микросервис ----><- размеры превью -><--------- URL исходного изображения --------------------------------->

В URL выше мы видим:
- http://cut-service.com/fill/300/200/ - endpoint нашего сервиса,
  в котором 300x200 - это размеры финального изображения.
- fill - обрезка картинки по центру до указанных размеров, resize - преобразование до нужного размера.
- www.audubon.org/sites/default/files/a1_1902_16_barred-owl_sandra_rothenberg_kk.jpg -
  адрес исходного изображения; сервис должен скачать его, произвести resize, закэшировать и отдать клиенту.

## Полезное ##
- make test - запуск unit-тестов
- make lint - запуск линтера
- make test-integration - запуск интеграционных тестов
- make run - запуск в docker контейнере
- make down - остановка сервиса
- make build - компиляция сервиса
