Promise    = require 'lie'
tl         = require 'talio'
superagent = (require 'superagent-promise')((require 'superagent'), Promise)

State   = require './state'
baseURL = require './baseurl'

handlers =
  createRelationship: (State, data) ->
    Promise.resolve().then(->
      superagent
        .post(baseURL + '/rel/')
        .set('Content-Type', 'application/x-www-form-urlencoded')
        .send(
          source: State.get 'here.url'
          target: data.target
          rel: data.rel
        )
    ).then(->

    ).catch((e) ->
      console.log e
    )
  createEquality: (State, data) ->
    if not confirm('Is this equal to "' + data.target + '" ?')
      return

    Promise.resolve().then(->
      superagent
        .post(baseURL + '/eql/')
        .set('Content-Type', 'application/x-www-form-urlencoded')
        .send(
          source: State.get 'here.url'
          target: data.target
        )
    ).then((res) ->

    ).catch((e) ->
      console.log e
    )
  openRelationshipInput: (State, data) ->
    for url, tab of State.get 'otherTabs'
      if typeof tab.rel == 'string'
        State.silentlyUpdate {
          otherTabs:
            "#{url}":
               rel: null
        }
    State.change {
      otherTabs:
        "#{data.target}":
          rel: ''
    }
  changeRelationship: (State, data) ->
    State.change {
      otherTabs:
        "#{data.target}":
          rel: data.rel
    }

# append container
container = document.createElement 'div'
container.id = 'nodes'
document.body.appendChild container
tl.run container, (require './vrender-main'), handlers, State

# append css
link = document.createElement 'link'
link.rel = 'stylesheet'
link.href = baseURL + '/embed.css'
document.querySelector('head').appendChild(link)
