# Olaris Server

## `This is all prelease code, if you somehow found this please move along, this won't work yet :) `

## What is Olaris?

Olaris is an open-source community driven media manager and transcoding server. The main interface is the [olaris-react](https://gitlab.com/olaris/olaris-react) project although in due time we hope to support multiple clients.

Our core values are:

### Community driven development
We want Olaris to be a community project which means we will heavily prioritise features based on our user’s feedback.

### Focus on one feature at a time
We will work on features until they are perfect (or as close to it as possible). We would rather have a product where three features work really well than a product with 30 unfinished features.

This does not mean we won't work in parallel, it simply means we will not start anything new until we are happy the new feature works to a high standard.

### Our users are not our product
We don't want to collect metadata, we don't want to sell metadata your data is yours and yours alone.

### Singular Focus: Video.
Our sole focus is on video and video alone, anything that does not meet this requirement will not be considered. This means for example we will never add music support due to different approach that would be required throughout. If we ever consider doing something that doesn’t include video it will be a separate entity to Olaris.

### Open-source
Everything we build should be open-source. We feel strongly that more can be achieved with free open-source software. That's why were are aiming to be and to remain open-source instead of open-core.

## Running manually

### Build dependencies
  * Install Go
	* go get github.com/jteeuwen/go-bindata/...
	* go get github.com/elazarl/go-bindata-assetfs/...

### Running

	`make run`

## Building

  `make build` to build for your local platform.

  `build-with-react` to build and pull in the latest web-interface.
