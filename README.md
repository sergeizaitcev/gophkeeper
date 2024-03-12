# gophkeeper

## Описание

gophkeeper - это выпускной проект на курсе "Продвинутый Go-разработчик" от Yandex Practucum.

## Компоненты
- Сервер gophkeeper
- Консольный клиент gk

## Зависимости
- MongoDB

## Описание сервера gophkeeper

Сервер регистрирует пользователей и хранит их данные в зашифрованном виде. Также сервер синхронизирует данные между устройствами одного пользователя. В качестве транспорта используется https протокол.

## Описание консольного клиента gk

Клиент реализован в git-like стиле.

Возможности клиента:
- регистрация/авторизация пользователя на удалённом сервере
- добавление, удаление и отображение секретных данных
- синхронизация данных с данными на удалённом сервере

Поддерживаемые типы данных:
- банковские карты
- учетные данные пользователя
- любые файлы

Для шифрования данных используется алгоритм AES-256.

### Работа с gk

-  После установки при помощи `make install` будет доступна команда `gk`
```sh
$ gk
Description: gophkeeper client

Usage: gk [version] | <command> ...

List of commands:
	remote	remote server settings
	login	authorization on a remote server
	add	adding new data with encryption to the vault
	rm	deleting data from the vault
	sync	synchronizing files with a remote server
	show	show data in the vault
	ls	show a list of all data in the vault
```

- Просмотр версии
```sh
$ gk version
Version: v1.0.0
```

- Добавление данных банковской карты
```sh
$ gk add card -d 'some card' 4720-4755-3562-9559
Password: ******
the data has been successfully added
```

- Добавление данных учетной записи
```sh
$ gk add logpass -d 'some logpass' user password
Password: ******
the data has been successfully added
```

- Добавление файлов
```sh
$ gk add file -d 'some file' file.txt
Password: ******
the data has been successfully added
```

- Список добавленных файлов
```sh
$ gk ls
ID            TYPE     DESCRIPTION
d9706bb621a4  CARD     some card
aa623b6b3c27  LOGPASS  some logpass
b5c90eed1d19  FILE     some file
```

- Удаление данных
```sh
$ gk rm aa623b6b3c27
the data was successfully deleted
```

- Просмотр данных
```sh
$ gk show d9706bb621a4
Password: ******
4720-4755-3562-9559
```

- Добавление удалённого репозитория
```sh
$ gk remote set $(REMOTE_ADDRESS)
```

- Регистрация/авторизация на удалённом сервере
```sh
$ gk login -u $(USERNAME) -p $(PASSWORD)
successful login
```

- Синхронизация данных с данными на удалённом севере
```sh
$ gk sync
sync up-to-date
```

## Дальнейшее развитие проекта

- Добавление автодополнения и подсказок в gk
- Добавление возможности редактирования секретной информации
- Добавление S3 хранилища
- Добавление e2e тестов
