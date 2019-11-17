# Cachet plugin

This plugin can provide notifications for services that are failed in
[Cachet](https://cachethq.io/).

## Configuration

There are two environment variables that this plugin needs:

* CACHET_API - URL of your Cachet top-level API endpoint.
  Example API URL `https://status.company.com/api`
* CACHET_ALERT_CONFIG - Mapping of component name to channels for alerting

Format of notification configuration is as follows:

* Each notification is `<service_name>:<notification target>`
* To notify on any service outage use `any` as service name
* Multiple notifications can be separated with semicolons

Example:
export CACHET_ALERT_CONFIG="service1:#chan1;any:#outages"

Above configuration would sent alerts to `#chan1` every time `service1` is in
outage and also send all outage notifications to `#outages`.

## Duplicate alerts

Cachet is checked every minute and outages are reported as per
configuration. This means that channels will keep receiving duplicate alerts
every minute. It's recommended to combine this plugin with `dedup` and
configure the dedup plugin to remove duplicate notification for some period of
time (say 1 day).
