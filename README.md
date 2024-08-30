# Instaman

**⚠️ Important:** this heavily relies on [subzeroid/aiograpi](https://github.com/subzeroid/aiograpi), which is NOT an official library, nor it is supported by Instagram.
On the contrary, using such kind of clients goes against [Instagram's TOS](https://help.instagram.com/581066165581870/) and might result in your account being restricted or permanently banned.
Use at your own risk!

## Overview

Instaman is a simple Instagram account manager, with some automated tasks and a nice micro-service architecture. It was authored mostly for fun, and for showcasing how a multi-service, multi-platform repository should be structured and especially how it should be maintained.
See, for instance:

* The list of [pull requests](./pulls?q=is%3Apr)
* The [commit history](./commits/main/)

## Set up

### Dependencies

Make sure you have installed

* [Docker](https://www.docker.com/) with [Docker Compose](https://docs.docker.com/compose/)
* _[optional]_ Python 3 - only required by some tools like the linter and PyTest
* _[optional]_ GNU Make (or Make on OSX) - only required if you want to leverage the existing Makefiles

### Bootstrapping the environment

1. Create a new Telegram private group, following [these instructions](https://telegram.org/faq#q-how-do-i-create-a-group).
2. Send a message to [BotFather](https://t.me/botfather) to create a new Telegram bot (it's super easy).
3. Send another message to [IDBot](https://t.me/username_to_id_bot) to retrieve your new bot's Token and the private group's ID, you will need these in the fifth step!
4. Make a copy of [docker-compose.override.example.yml](./docker-compose.override.example.yml) and name it `docker-compose.override.yml`
5. Edit the new file and fill the placeholder environment variables.
6. Run `docker compose up` and - voilà - your Instaman app is now running!

## Repo Directories

### `/instaproxy`

A small webserver that acts as a proxy (with cache) for the Instagram GraphQL APIs. Written in Python and powered by [FastAPI](https://github.com/fastapi/fastapi), it is completely asynchronous and does not require many dependencies (not even a database).

## Other Directories

### `/data`

Data persisted from the application, that should never be committed. This directory is automatically created by `docker compose` and contains at least:

* `/instagram/`: directory with data fetched by the instaproxy application, including the main user's sessions.
