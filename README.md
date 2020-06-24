# Cadwallader
An ELK-adjacent public status page

## Welcome

We use the Elasticsearch-Logstash-Kibana stack almost exclusively for our internal monitoring. It's also nice to provide users with a status page to help them self-diagnose issues and build confidence in our systems.

Rather than rely on a second set of monitors, we thought it would be great to tie in with a (for now) read-only status page. By accessing Elasticsearch on the server, we avoid opening our monitoring data to the world and can more tightly control what people can see into our system.

Given the ubiquity of both status pages and the ELK stack, we thought we'd open source this code and share it with the world as we develop it. Please be gentle! This is our first open source project and we're definitely still learning.

This is very much alpha-state code. While we are using it in production, if a status page is a mission-critical tool for your organization, we'd recommend holding off deploying this in lieu of what you're using now. It's also subject to massive breaking changes as we build it out, so please be careful. That said, if you do use it or want to contribute, please get in touch!

## Basic Usage

The current version expects the config.yml to be in the same working directory the application runs from. There are a few basic configuration options as outlined in `config.yml.example`. The `elasticsearch` block specifies where to read monitoring data. The `server` block specifies how to run the server, and the `services` block specifies which monitors to report.

`ELASTIC_PASSWORD=password ELASTIC_USERNAME=elastic go run main.go config.go` 

## What's Next?

We plan to make a lot of the very obvious changes in due course, such as allowing the configuration to be overridden with environment variables, making that more consistent, and allowing more customization of the monitors.
