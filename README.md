# Specs bot

[![GoDoc](https://godoc.org/github.com/babolivier/go-doh-client?status.svg)](https://godoc.org/github.com/babolivier/go-doh-client) [![#discuss:weu.informo.network on Matrix](https://img.shields.io/matrix/discuss:weu.informo.network.svg?logo=matrix)](https://matrix.to/#/#discuss:weu.informo.network)

The specs bot is a Matrix bot that shares state updates of a specifications proposal to the configured Matrix rooms. While initially designed to shout about updates to the [Informo open specs](https://github.com/Informo/specs), we made it compatible to most specifications projects using GitHub issues or pull requests to track proposals and labels to track a proposal's state.

It works by setting up a GitHub webhook listening on pull requests and issues events. Each time it receives a matching payload, and if the event was triggered by a change in the PR/issue's list of labels, it generates an update message by selecting a configured string matching the update and processing it (along with some information specific to the PR/issue) through the configured template. It then sends the message as a [notice](https://matrix.org/docs/spec/client_server/r0.4.0.html#m-notice) to the configured Matrix rooms.

## Build

You can install the bot by building it or using one of the binaries available in the project's [releases](https://github.com/Informo/specs-bot/releases).

This project is written in Go, therefore building it requires a [Go development environment](https://golang.org/doc/install). It requires Go 1.11+.

In order to build the bot, you should first clone the repository:

```
git clone https://github.com/Informo/specs-bot
cd specs-bot
go build
```

The bot's binary should now be available as `./specs-bot`. You can run it as it is, or with the `--debug` flag to make it print debug logs, which explain the whole path of a payload's content through the different workflows.

## Configure

### Webhook configuration

The bot currently only support GitHub webhooks and expects them to be send with the `application/json` content type.

### Configuration files

The bot needs two configuration files to do its job.

### `config.yaml`

The first configuration file is the general one, which will allow defining the settings to connect to Matrix, set the webhook up, and generate Matrix notices. All of its configuration keys are documented in [the sample file](config.sample.yaml). The path to the configuration file can be provided as follows:

```
/path/to/specs-bot --config /path/to/config.yaml
```

If no value for the `--config` flag is provided, it will default to `./config.yaml`.

### `strings.json`

The second configuration file contains the strings to use when generating the notice message, in the JSON format. An example file is available [here](strings.json). This file is split in three sections: `typo` and `behaviour` contains strings that match states from the [Informo SCSP](https://specs.informo.network/introduction/scsp/), respectively for [typo](https://specs.informo.network/introduction/scsp/#typo-wording-and-phrasing) and [behavioural](https://specs.informo.network/introduction/scsp/#behaviour-change) changes. The `global` section contains strings that are either common between the two types, or not related to the Informo SCSP. Therefore, an instance of the bot set up to follow proposals to a specifications project that doesn't follow Informo's SCSP must have all of its strings defined in the `global` section.

The file path can be configured in the general configuration file (`config.yaml`)

#### Example: Informo

The strings to use in Informo's use case are defined in the [strings.json](strings.json) file located in this repository.

In this example, and using the pattern in the sample configuration file, doing this in GitHub's interface:

![github](https://user-images.githubusercontent.com/34184120/47513717-c228e280-d876-11e8-96d0-6b74abd34114.png)

Triggers the sending of this notice to Matrix:

![matrix](https://user-images.githubusercontent.com/34184120/47513870-0a480500-d877-11e8-9c48-f9bb58cffa26.png)

#### Example: Matrix

Let's consider using the bot to follow updates to proposals to the [Matrix specifications](https://github.com/matrix-org/matrix-doc). The JSON strings file should then look like:

```json
{
	"global": {
		"proposal-in-review": "is currently in review",
		"proposal-passed-review": "passed review as worth implementing and then being added to the spec",
		"proposal-ready-for-review": "is now ready and waiting for review by the core team and community",
		[...]
	}
}
```

In this example, and using the pattern in the sample configuration file, doing this in GitHub's interface:

![github](https://user-images.githubusercontent.com/34184120/47514337-0799df80-d878-11e8-8fcd-0a93f9ad8af3.png)

Triggers the sending of this notice to Matrix:

![matrix](https://user-images.githubusercontent.com/34184120/47514484-68291c80-d878-11e8-9b21-11e1da5c7ebb.png)

## Scripts

### DB Label Seeder (SQLite only)

This release includes a python3 script called `fill-db.py` in the `scripts/fill-db` directory. Its purpose is to initially seed a database with proposals and their labels. The reason for it is that specs-bot only tracks changes between labels, and thus if a proposal is already halfway through completion when specs-bot is activated, it will end up outputting multiple events as it finds out about all the labels that were already on the issue prior to specs-bot coming online.

The script fixes this issue by initially downloading all information about all issues/PRs (or other those with certain labels) and their label information, so that specs-bot can be up-to-date about the repo's proposals as soon as it comes online.

To use, make sure you have python3 and pip installed, then install the script's python requirements:

```
pip3 install -r scripts/fill-db/requirements.txt
```

Then, open `scripts/fill-db/fill-db.py` and enter in your repository information (ex: `"Informo/specs"`), your Github [personal access token](https://github.com/settings/tokens), your sqlite3 DB location (ex: `"./specs-bot.db"`) and the labels you'd like to filter issues/PRs by as a list of strings (or leave as an empty list to download all issues/PRs). Once done, simply run the script from this repo's root directory:

```
python3 scripts/fill-db/fill-db.py
```

Your sqlite3 DB file should now be seeded with all proposals and their label information. You can now start specs-bot and be confident it'll show changes as intended.

## Docker

The bot can be run inside a Docker container. The image can be found on the [Docker Hub](https://hub.docker.com/r/informo/specs-bot). It exposes the port 8080, on which the bot is expected to listen.

Building the image from scratch can be done with:

```bash
docker build -t informo/specs-bot -f docker/Dockerfile .
```

A configuration file and a strings file are needed. Examples can be found in the [Docker configuration directory](/docker/data).

Once both files are ready, run the image with the configuration file and mount the configuration directory:

```bash
docker run --name specs-bot -p 8080:8080 -v /path/to/config/directory:/etc/specs-bot specs-bot
``` 

### Running the DB Label Seeder script (SQLite only)

By running the image with the configuration directory mounted, the bot will initialise a SQLite database in the mounted directory. Therefore, run the image once then kill the container. The SQLite database will have been created in the directory where the configuration lives.

Run the script on that database (by setting `DB_PATH` to the right value), then you can start the bot again, with the same directory mounted as it was during the initial run, and it should start using your updated database.

## What is Informo?

Informo is an open project using decentralisation and federation to fight online censorship of the press. Read more about it [here](https://specs.informo.network/informo/), and join the discussion online on [Matrix](https://matrix.to/#/#discuss:weu.informo.network) or [IRC](https://webchat.freenode.net/?channels=%23informo).
