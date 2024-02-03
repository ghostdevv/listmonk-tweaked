# Listmonk Tweaked

This is a fork of [listmonk](https://github.com/knadh/listmonk).

## Intentions/Motivations

We needed a few features such as "drip campaigns" for a project and had major issues running other open source listmonk alternatives that had these features. I intend to keep this project alive until listmonk has the features we need or I find a suitable alternative. I don't intend to maintain this beyond our own internal needs due to time constraints/my lack of knowledge of go.

## Changes

### `/api/txc`

An endpoint for sending marketing emails without having to have an attached campaign. It works by using a [transactional template](https://listmonk.app/docs/templating/#transactional-templates) which must be written in markdown when sent with txc endpoint, combining the template contents with the default campaign template and then sending it to the user. It supports *most* of the [template expressions](https://listmonk.app/docs/templating/#template-expressions) available in a campaign template.

```
POST <host>/api/txc

{
    "list_id": 3,
    "subscriber_id": 3,
    "template_id": 8
}
```

### New Subscriber Webhooks

Webhooks that are sent when a user is added to the list for the first time (i.e. they get added to the list, this means if you remove them and then re-add them it will trigger the webhook). We use this with [n8n](https://n8n.io/) to and the txc endpoint mentioned above to send drip/automated campaigns (without an actual listmonk campaign attached) to users.

### Unsubscribe Page

I've tweaked the unsubscribe page to allow passing in a list uuid or a campaign uuid, this is so that emails you send with the txc endpoint can still have a fully functioning unsubscribe system. It'll also try it's best to find the list/campaign title and add it under `{{ .Data.ListTitle }}`.

### Misc

- I've replaced the frontend package manager with [pnpm](https://pnpm.io/) because I couldn't get yarn to run.
- I'm publishing a docker image to `ghcr.io/ghostdevv/listmonk-tweaked`
- For places that I've mad show/have a list name, if that list is private then it'll add something like "list" as the name.
