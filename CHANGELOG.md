# Changelog

## Unreleased

- Fix: Fix reconnect race condition on ping failure. (#33)
- Dev: Add moderation event tests. (#21)
- Dev: Add bits event tests. (#22, #23)

## v0.1.0

- Major: Changed minimum required Go version from 1.16 to 1.19. (#19)
- Minor: Add support for subscribe events. See more information here: https://dev.twitch.tv/docs/pubsub/#example-channel-subscriptions-event-message (#15)
- Dev: Update github.com/gorilla/websocket from v1.4.2 to v1.5.0 (#17)
- Dev: Add https://github.com/frankban/quicktest as a dependency for tests. (#15)
- Dev: Add CI linting & testing. (#16)
- Dev: Fix linting issues. (#17)

## v0.0.4

- Fix: Fix Whisper events. (#13)

## v0.0.3

- Minor: Add support for AutoMod Queue events. (#12)

## v0.0.2

- Minor: Add UserInput and Status fields to ChannelPoints `PointsEvent` Topic. (#10)

## v0.0.1

- Minor: Add support for ChannelPoints Topic. (#7)
- Dev: Remove dependency on github.com/pajbot/utils library. (#6)
