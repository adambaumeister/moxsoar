# Getting started Guide
## Basic usage
Start by running the docker container, [here](../README.md). You should be able to open a browser.

You'll be greeted by the content pack page.

<img src="docs/img/start_screen.PNG" width="600">

## Content Packs

Moxsoar uses a concept of "content packs" to manage mock configurations.

Content packs are Git repositories that are accessible to the Moxsoar engine.

Content packs comprise the following:
* A "runner" configuration - which mocks (known internally as **integrations** to run) and which ports to use
* Any number of **integration** configurations

An **integration** is a collection of URL routes, handling logic and response files.

Put it together, and the layout of a complete content pack is as such:
```bash
/runner.yml
    integration1/
        routes.json
        response.json
    integration2/
        routes.json
        response.json
```

