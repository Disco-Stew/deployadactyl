![](https://raw.githubusercontent.com/compozed/images/master/deployadactyl_logo.png)

[![Release](https://img.shields.io/github/release/compozed/deployadactyl.svg)](https://github.com/compozed/deployadactyl/releases/latest)
[![CircleCI](https://circleci.com/gh/compozed/deployadactyl.svg?style=svg&circle-token=0eab8bce42440217fb24ffd8ffdc2b44932125d5)](https://circleci.com/gh/compozed/deployadactyl)
[![Go Report Card](https://goreportcard.com/badge/github.com/compozed/deployadactyl)](https://goreportcard.com/report/github.com/compozed/deployadactyl)
[![codecov](https://codecov.io/gh/compozed/deployadactyl/branch/master/graph/badge.svg?token=r9yd1cwtbH)](https://codecov.io/gh/compozed/deployadactyl)
[![Stories in Ready](https://badge.waffle.io/compozed/deployadactyl.png?label=ready&title=Ready)](https://waffle.io/compozed/deployadactyl)
[![Gitter](https://badges.gitter.im/compozed/deployadactyl.svg)](https://gitter.im/compozed/deployadactyl?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge)
[![Slack Status](https://deployadactyl-invite.cfapps.io/badge.svg)](https://deployadactyl-invite.cfapps.io)
[![GoDoc](https://godoc.org/github.com/compozed/deployadactyl?status.svg)](https://godoc.org/github.com/compozed/deployadactyl)

Deployadactyl is a Go library for deploying applications to multiple [Cloud Foundry](https://www.cloudfoundry.org/) instances. Deployadactyl utilizes [blue green deployments](https://docs.pivotal.io/pivotalcf/devguide/deploy-apps/blue-green.html) and if it's unable to push your application it will rollback to the previous version. It also utilizes Go channels for concurrent deployments across the multiple Cloud Foundry instances.

<!-- TOC depthFrom:2 depthTo:6 withLinks:1 updateOnSave:1 orderedList:0 -->

- [How It Works](#how-it-works)
- [Why Use Deployadactyl?](#why-use-deployadactyl)
- [Usage Requirements](#usage-requirements)
	- [Dependencies](#dependencies)
	- [Configuration File](#configuration-file)
		- [Example Configuration yml](#example-configuration-yml)
		- [Environment Variables](#environment-variables)
- [How to Download Dependencies](#how-to-download-dependencies)
- [How To Run Deployadactyl](#how-to-run-deployadactyl)
- [How to Push Deployadactyl to Cloud Foundry](#how-to-push-deployadactyl-to-cloud-foundry)
	- [Available Flags](#available-flags)
	- [API](#api)
		- [Example Curl](#example-curl)
- [Event Handling](#event-handling)
	- [Available Emitted Event Types](#available-emitted-event-types)
	- [Event Handler Example](#event-handler-example)
	- [Event Handling Example](#event-handling-example)
- [Contributing](#contributing)

<!-- /TOC -->

## How It Works

Deployadactyl works by utilizing the [Cloud Foundry CLI](http://docs.cloudfoundry.org/cf-cli/) to push your application. The general flow is to get a list of Cloud Foundry instances, check that the instances are available, download your artifact, log into each instance, and concurrently call `cf push` in the deploying applications directory. If your application fails to deploy on any instance, Deployadactyl will automatically roll the application back to the previous version.

## Why Use Deployadactyl?

As an application grows, it will have multiple foundations for each environment. These scaling foundations make deploying an application time consuming and difficult to manage. If any errors occur during a deployment it can greatly increase downtime.

Deployadactyl makes the process easy and efficient with:

- Management of multiple environment configurations
- Concurrent deployments across environment foundations
- Automatic rollbacks for failures or errors
- Prechecking foundation availablity before deployments
- Event handlers for third-party services


## Usage Requirements


### Dependencies

Deployadactyl has the following dependencies within the environment:

- [ CloudFoundry CLI](https://github.com/cloudfoundry/cli)
- [Go 1.6](https://golang.org/dl/) or later


### Configuration File

Deployadactyl needs a `yml` configuration file to specify your environments. Each environment has a name, domain and a list of foundations.

The configuration file can be placed anywhere within your project directory as long as you specify the location.

|**Param**|**Necessity**|**Type**|**Description**|
|---|:---:|---|---|
|`name`|**Required**|`string`| Used in the deploy when the users are sending a request to Deployadactyl to specify which environment from the config they want to use.|
|`domain`|**Required**|`string`| Used to specify a load balanced URL that has previously been created on the Cloud Foundry instances.|
|`foundations` |**Required**|`[]string`|A list of Cloud Foundry instance URLs.|
|`authenticate` |*Optional*|`bool`| Used to specify if basic authentication is required for users. See the [authentication section](https://github.com/compozed/deployadactyl/wiki/Deployadactyl-API-v1.0.0#authentication) in the [API documentation](https://github.com/compozed/deployadactyl/wiki/Deployadactyl-API-Versions) for more details|
|`skip_ssl` |*Optional*|`bool`| Used to skip SSL verification when Deployadactyl logs into Cloud Foundry.|
|`disable_first_deploy_rollback` |*Optional*|`bool`| Used to disable automatic rollback on first deploy so that initial logs are kept.|
|`instances` |*Optional*|`int`| Used to set the number of instances an application is deployed with. If the number of instances is specified in a Cloud Foundry manifest, that will be used instead. |

#### Example Configuration yml

```yaml
---
environments:
  - name: preproduction
    domain: preproduction.example.com
    foundations:
    - https://preproduction.foundation-1.example.com
    - https://preproduction.foundation-2.example.com
    authenticate: false
    skip_ssl: true
    disable_first_deploy_rollback: true
    instances: 2

  - name: production
    domain: production.example.com
    foundations:
    - https://production.foundation-1.example.com
    - https://production.foundation-2.example.com
    - https://production.foundation-3.example.com
    - https://production.foundation-4.example.com
    authenticate: true
    skip_ssl: false
    disable_first_deploy_rollback: false
    instances: 4
```

#### Environment Variables

Authentication is optional as long as `CF_USERNAME` and `CF_PASSWORD` environment variables are exported. We recommend making a generic user account that is able to push to each Cloud Foundry instance.

```bash
$ export CF_USERNAME=some-username
$ export CF_PASSWORD=some-password
```

*Optional:* The log level can be changed by defining `DEPLOYADACTYL_LOGLEVEL`. `DEBUG` is the default log level.

## How to Download Dependencies

We have our dependencies using [git submodules](https://git-scm.com/docs/git-submodule). To download the dependencies when you clone Deployadactyl you can use the following command:

```bash
$ git clone --recursive https://github.com/compozed/deployadactyl
```

If you have already cloned Deployadactyl you can use our Makefile to get the dependencies with this command:

```bash
$ make dependencies
```

## How To Run Deployadactyl

After a configuration yml has been created and environment variables have been set, the server can be run using the following commands:

```bash
$ go run server.go
```

or

```bash
$ go build && ./deployadactyl
```

## How to Push Deployadactyl to Cloud Foundry

To push Deployadactyl to Cloud Foundry, edit the `manifest.yml` to include your `CF_USERNAME` and `CF_PASSWORD` environment variables. In addition, be sure to create a `config.yml`.

```bash
$ make push
```

or

```bash
$ git submodule update --init --recursive
$ cf login
$ cf push
```

### Available Flags

|**Flag**|**Usage**|
|---|---|
|`-config`|location of the config file (default "./config.yml")|

### API

A deployment by hitting the API using `curl` or other means. For more information on using the Deployadactyl API visit the [API documentation](https://github.com/compozed/deployadactyl/wiki/Deployadactyl-API-Versions) in the wiki.

#### Example Curl

```bash
curl -X POST \
     -u your_username:your_password \
     -H "Accept: application/json" \
     -H "Content-Type: application/json" \
     -d '{ "artifact_url": "https://example.com/lib/release/my_artifact.jar"}' \
     https://preproduction.example.com/v1/apps/environment/org/space/t-rex
```

## Event Handling

With Deployadactyl you can optionally register event handlers to perform any additional actions your deployment flow may require. For us, this meant adding handlers that would open and close change records, as well as notify anyone on pager duty of significant events.

### Available Emitted Event Types

|**Event Type**|**Returned Struct**|**Emitted**|
|---|---|---|---|---|
|`deploy.start`|[DeployEventData](structs/deploy_event_data.go)|Before deployment starts
|`deploy.success`|[DeployEventData](structs/deploy_event_data.go)|When a deployment succeeds
|`deploy.failure`|[DeployEventData](structs/deploy_event_data.go)|When a deployment fails
|`deploy.error`|[DeployEventData](structs/deploy_event_data.go)|When a deployment throws an error
|`deploy.finish`|[DeployEventData](structs/deploy_event_data.go)|When a deployment finishes, regardless of success or failure
|`validate.foundationsUnavailable`|[PrecheckerEventData](structs/prechecker_event_data.go)|When a foundation you're deploying to is down

### Event Handler Example

```go
package pagehandler

type Pager interface {
  Page(description string)
}

import (
	DS "github.com/compozed/deployadactyl/structs"
)

type PageHandler struct {
	Pager        Pager
	Environments map[string]bool
}

func (p PageHandler) OnEvent(event DS.Event) error {
	var (
		precheckerEventData = event.Data.(DS.PrecheckerEventData)
		environmentName     = precheckerEventData.Environment.Name
		allowPage           = p.Environments[environmentName]
	)

	if allowPage {
		p.Pager.Page(precheckerEventData.Description)
	}

	return nil
}
```

```go
package page

type Page struct {
  Token string
  Log   I.Logger
}

func (p *Page) Page(description string) {
  // pagerduty code
}
```

### Event Handling Example

```go
  // server.go

  p := pagehandler.PageHandler{Pager: pager, Config: config}

  em := creator.CreateEventManager()
  em.AddHandler(p, "deploy.start")
  em.AddHandler(p, "deploy.success")
  em.AddHandler(p, "deploy.failure")
  em.AddHandler(p, "deploy.error")
  em.AddHandler(p, "deploy.finish")
```

## Contributing

See our [CONTRUBUTING](CONTRIBUTING.md) section for more information.
