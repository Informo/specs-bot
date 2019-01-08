# Specs bot

The specs bot is a Matrix bot that shares state updates of a specifications proposal to the configured Matrix rooms. While initially designed to shout about updates to the [Informo open specs](https://github.com/Informo/specs), we made it compatible to most specifications projects using GitHub issues or pull requests to track proposals and labels to track a proposal's state.

It works by setting up a GitHub webhook listening on pull requests and issues events. Each time it receives a matching payload, and if the event was triggered by a change in the PR/issue's list of labels, it generates an update message by selecting a configured string matching the update and processing it (along with some information specific to the PR/issue) through the configured template. It then sends the message as a [notice](https://matrix.org/docs/spec/client_server/r0.4.0.html#m-notice) to the configured Matrix rooms.

## Build

You can install the bot by building it or using one of the binaries available in the project's [releases](https://github.com/Informo/specs-bot/releases).

This project is written in Go, therefore building it requires a [Go development environment](https://golang.org/doc/install).

Building it also requires [gb](https://github.com/constabulary/gb):

```
go get github.com/constabulary/gb/...
```

Once everything is setup correctly, clone this repository, retrieve the project's dependencies and build it:

```
git clone https://github.com/Informo/specs-bot
cd specs-bot
gb vendor restore
gb build
```

The bot's binary should now be available as `bin/specs-bot`. You can run it as it is, or with the `--debug` flag to make it print debug logs, which explain the whole path of a payload's content through the different workflows.

## Configure

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

## What is Informo?

Informo is an open project using decentralisation and federation to fight online censorship of the press. Read more about it [here](https://specs.informo.network/informo/), and join the discussion online on [Matrix](https://matrix.to/#/#discuss:weu.informo.network) or [IRC](https://webchat.freenode.net/?channels=%23informo).
